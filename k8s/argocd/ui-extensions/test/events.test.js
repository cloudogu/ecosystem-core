import assert from "node:assert/strict";
import test from "node:test";
import {resourceEventsUrl} from "../src/argo-api.js";
import {doguEventResources, normalizeResourceEvents, recentEvents, loadDoguEvents, eventSourceKey} from "../src/events.js";
import {buildDoguOverview} from "../src/dogus.js";

const app = {metadata: {name: "ces test", namespace: "argo-custom"}};
const dogu = {group: "k8s.cloudogu.com", kind: "Dogu", namespace: "ecosystem", name: "mysql", uid: "dogu-1"};
const child = (kind, name, uid, parent, extra = {}) => ({group: ["Deployment", "ReplicaSet"].includes(kind) ? "apps" : "", kind, namespace: "ecosystem", name, uid, parentRefs: [parent], ...extra});
const deployment = child("Deployment", "mysql", "deployment-1", dogu);
const replicaSet = child("ReplicaSet", "mysql-123", "replicaset-1", deployment);
const pod = child("Pod", "mysql-123-abc", "pod-1", replicaSet);
const pvc = child("PersistentVolumeClaim", "mysql", "pvc-1", dogu);
const now = Date.parse("2026-10-21T10:00:00Z");
const event = (node = pod, overrides = {}) => ({metadata: {uid: "event-1"}, involvedObject: {...node, apiVersion: node.group ? `${node.group}/v1` : "v1"},
  type: "Warning", reason: "FailedMount", message: "Unable to mount volume", lastTimestamp: new Date(now).toISOString(), count: 3, ...overrides});

test("attributes Pod and PVC events through the complete ownership chain without name guesses", () => {
  const unrelated = child("Pod", "mysql-other", "unrelated", {...replicaSet, uid: "old-rs"});
  const otherNamespace = {...pvc, namespace: "another-ces", uid: "another-pvc"};
  const sources = doguEventResources(dogu, [pod, unrelated, pvc, replicaSet, otherNamespace, deployment]);
  assert.deepEqual(new Set(sources.map(node => node.uid)), new Set([dogu.uid, deployment.uid, replicaSet.uid, pod.uid, pvc.uid]));
  assert.deepEqual(doguEventResources({...dogu, uid: undefined}, [pod]), []);
});

test("includes source nodes in the Dogu model without requesting events eagerly", () => {
  const blueprint = {metadata: {namespace: "ecosystem"}, spec: {blueprint: {dogus: [{name: "official/mysql"}]}}};
  const row = buildDoguOverview(blueprint, {nodes: [dogu, deployment, replicaSet, pod, pvc]}).present[0];
  assert.equal(row.eventResources.length, 5);
});

test("builds a resource-scoped URL with UID, namespace, app namespace and base path", () => {
  const url = new URL(resourceEventsUrl(app, pod, "https://example.com/argo/"));
  assert.equal(url.pathname, "/argo/api/v1/applications/ces%20test/events");
  assert.deepEqual(Object.fromEntries(url.searchParams), {appNamespace: "argo-custom", resourceUID: "pod-1", resourceName: pod.name, resourceNamespace: "ecosystem"});
  assert.throws(() => resourceEventsUrl(app, {...pod, uid: undefined}, "https://example.com/"), /UID/);
});

test("normalizes series timestamps and repeat counts, deduplicating the same Event UID", () => {
  const repeated = event(pod, {lastTimestamp: "2026-10-20T10:00:00Z", series: {count: 9, lastObservedTime: "2026-10-21T09:59:00Z"}});
  const rows = normalizeResourceEvents([repeated, event(pod)], pod);
  assert.equal(rows.length, 1);
  assert.equal(rows[0].timestamp, now);
  assert.equal(rows[0].count, 3);
  assert.equal(normalizeResourceEvents([repeated], pod)[0].count, 9);
});

test("rejects event objects belonging to another UID, namespace, kind or API group", () => {
  const events = [event({...pod, uid: "previous-pod"}), event({...pod, namespace: "other"}), event({...pod, kind: "Secret"}), event({...pod, group: "other.example"})];
  assert.deepEqual(normalizeResourceEvents(events, pod), []);
});

test("shows newest warnings first, limits the list, and optionally includes normal events", () => {
  const rows = Array.from({length: 25}, (_, index) => ({id: String(index), type: "Warning", timestamp: now - index * 1000}));
  rows.push({id: "normal", type: "Normal", timestamp: now}, {id: "expired", type: "Warning", timestamp: now - 25 * 3600000});
  assert.equal(recentEvents(rows, {now}).length, 20);
  assert.equal(recentEvents(rows, {now})[0].id, "0");
  assert.equal(recentEvents(rows, {now, warningsOnly: false, limit: Infinity}).length, 26);
  assert.equal(recentEvents([{id: "undated", type: "Warning", timestamp: null}], {now}).length, 1);
});

test("uses eventTime and metadata fallback when legacy timestamps are absent", () => {
  const data = event(pod, {lastTimestamp: null, eventTime: "2026-10-21T09:00:00Z"});
  assert.equal(normalizeResourceEvents([data], pod)[0].timestamp, Date.parse(data.eventTime));
  delete data.eventTime;
  data.metadata.creationTimestamp = "2026-10-21T08:00:00Z";
  assert.equal(normalizeResourceEvents([data], pod)[0].timestamp, Date.parse(data.metadata.creationTimestamp));
});

test("loads only scoped resources, deduplicates requests and bounds concurrency to four", async () => {
  const resources = Array.from({length: 7}, (_, index) => ({...pod, uid: `pod-${index}`}));
  let active = 0, maximum = 0, calls = 0;
  const fetcher = async (input, options) => {
    const url = new URL(input);
    assert.equal(url.searchParams.get("resourceName"), pod.name);
    assert.equal(options.credentials, "same-origin");
    assert.equal(options.method || "GET", "GET");
    calls++; active++; maximum = Math.max(maximum, active);
    await new Promise(resolve => setTimeout(resolve, 2));
    active--;
    return {ok: true, json: async () => ({items: []})};
  };
  const result = await loadDoguEvents({application: app, resources: [...resources, resources[0]], baseURI: "https://example.com/", fetcher});
  assert.equal(calls, 7);
  assert.equal(maximum, 4);
  assert.equal(Object.keys(result.byResource).length, 7);
});

test("permission errors retain old events with a visible source error", async () => {
  const key = eventSourceKey(pod);
  const previous = {byResource: {[key]: {items: normalizeResourceEvents([event()], pod)}}};
  const result = await loadDoguEvents({application: app, resources: [pod, pvc], previous, baseURI: "https://example.com/",
    fetcher: async input => new URL(input).searchParams.get("resourceUID") === pod.uid ? {ok: false, status: 403} : {ok: true, json: async () => ({items: []})}});
  assert.equal(result.byResource[key].items.length, 1);
  assert.match(result.byResource[key].error, /Berechtigung/);
  assert.equal(result.byResource[eventSourceKey(pvc)].error, null);
});

test("recreated resources cannot inherit previous events if their UID changed", async () => {
  const previous = {byResource: {[eventSourceKey(pod)]: {items: normalizeResourceEvents([event()], pod)}}};
  const recreated = {...pod, uid: "replacement-pod"};
  const result = await loadDoguEvents({application: app, resources: [recreated], previous, baseURI: "https://example.com/", fetcher: async () => ({ok: false, status: 403})});
  assert.deepEqual(result.byResource[eventSourceKey(recreated)].items, []);
});

test("accepts an omitted empty items list but marks invalid responses as failures", async () => {
  const empty = await loadDoguEvents({application: app, resources: [pod], baseURI: "https://example.com/", fetcher: async () => ({ok: true, json: async () => ({})})});
  assert.deepEqual(empty.byResource[eventSourceKey(pod)].items, []);
  assert.equal(empty.byResource[eventSourceKey(pod)].error, null);
  const invalid = await loadDoguEvents({application: app, resources: [pod], baseURI: "https://example.com/", fetcher: async () => ({ok: true, json: async () => ({items: "invalid"})})});
  assert.match(invalid.byResource[eventSourceKey(pod)].error, /gültige Ereignisliste/);
});

test("aborting an event refresh rejects instead of replacing the previous snapshot", async () => {
  const controller = new AbortController();
  controller.abort();
  await assert.rejects(loadDoguEvents({application: app, resources: [pod], baseURI: "https://example.com/", signal: controller.signal,
    fetcher: () => {throw new Error("must not fetch");}}), {name: "AbortError"});
});
