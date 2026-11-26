import assert from "node:assert/strict";
import test from "node:test";
import {buildDoguOverview} from "../src/dogus.js";
import {buildBlueprintStatus, buildComponentOverview, operatingState} from "../src/status.js";
import {loadOverviewSnapshot, resourceKey, resourceUrl, sameDestination} from "../src/argo-api.js";

const application = {metadata: {name: "ces-blueprint", namespace: "argocd"}, spec: {destination: {namespace: "ecosystem", server: "https://ces.example"}}};
const node = (kind, name, extra = {}) => ({group: "k8s.cloudogu.com", version: kind === "Blueprint" ? "v3" : kind === "Dogu" ? "v2" : "v1", kind, name, namespace: "ecosystem", uid: name, ...extra});
const bpNode = node("Blueprint", "test");
const doguNode = node("Dogu", "swaggerui", {health: {status: "Degraded", message: "Stopped"}});
const deploymentNode = node("Deployment", "swaggerui", {group: "apps", parentRefs: [doguNode]});
const componentNode = node("Component", "dogu-operator");
const resource = (ref, spec, status) => ({apiVersion: `${ref.group}/${ref.version}`, kind: ref.kind, metadata: {name: ref.name, namespace: ref.namespace, generation: 2}, spec, status});
const blueprint = resource(bpNode, {blueprint: {dogus: [{name: "official/swaggerui", version: "2.0.0-1"}]}}, {
  conditions: [{type: "Completed", status: "True", lastTransitionTime: "2026-10-13T12:00:00Z"}, {type: "LastApplySucceeded", status: "True", lastTransitionTime: "2026-10-13T12:00:00Z"}],
});
const dogu = resource(doguNode, {version: "2.0.0-1", stopped: true}, {installedVersion: "1.0.0-1", status: "installed", conditions: [{type: "ready", status: "False", observedGeneration: 2}]});
const deployment = resource(deploymentNode, {replicas: 1}, {availableReplicas: 0, observedGeneration: 2});
const component = resource(componentNode, {version: "3.0.0"}, {installedVersion: "3.0.0", health: "available", status: "installed"});
const detail = obj => ({resource: obj});
const tree = {nodes: [bpNode], orphanedNodes: [doguNode, deploymentNode]};

test("compares actual installed versions and distinguishes stopped readiness from execution success", () => {
  const overview = buildDoguOverview(blueprint, tree, {[resourceKey(doguNode)]: detail(dogu), [resourceKey(deploymentNode)]: detail(deployment)});
  const row = overview.present[0];
  assert.equal(row.installedVersion, "1.0.0-1");
  assert.equal(row.comparison.text, "Abweichung");
  assert.equal(overview.deviations, 1);
  assert.deepEqual(row.operations, ["Gestoppt", "Nicht bereit"]);
  assert.equal(row.replicas[0].text, "0/1");
  assert.equal(row.state, "Degraded");
  assert.equal(buildBlueprintStatus(blueprint).phase, "Abgeschlossen");
});

test("does not use the CR target version or labels as an installed-version fallback", () => {
  const data = structuredClone(dogu);
  delete data.status.installedVersion;
  const row = buildDoguOverview(blueprint, tree, {[resourceKey(doguNode)]: detail(data)}).present[0];
  assert.equal(row.installedVersion, "—");
  assert.equal(row.comparison.text, "Nicht vergleichbar");
});

test("also flags a changed CR target when the installed version still matches the Blueprint", () => {
  const data = structuredClone(dogu);
  data.status.installedVersion = "2.0.0-1";
  data.spec.version = "3.0.0-1";
  const row = buildDoguOverview(blueprint, tree, {[resourceKey(doguNode)]: detail(data)}).present[0];
  assert.equal(row.comparison.text, "CR-Ziel weicht ab");
});

test("does not attach a same-name Deployment with a different Dogu owner UID", () => {
  const different = {...deploymentNode, parentRefs: [{...doguNode, uid: "different-uid"}]};
  const row = buildDoguOverview(blueprint, {nodes: [doguNode, different]}, {[resourceKey(deploymentNode)]: detail(deployment)}).present[0];
  assert.equal(row.replicas.length, 0);
});

test("shows support mode and paused reconciliation and ignores outdated readiness conditions", () => {
  const data = resource(doguNode, {supportMode: true, pauseReconciliation: true}, {conditions: [{type: "ready", status: "True", observedGeneration: 1}]});
  assert.deepEqual(operatingState(data), ["Supportmodus", "Reconciliation pausiert", "Bereitschaft unbekannt"]);
});

test("retained detail data is marked stale and is not counted as a current version deviation", () => {
  const row = buildDoguOverview(blueprint, tree, {[resourceKey(doguNode)]: {...detail(dogu), error: "HTTP 403"}}).present[0];
  assert.equal(row.installedVersion, "1.0.0-1");
  assert.equal(row.comparison.state, "Unknown");
  assert.equal(row.detailError, "HTTP 403");
});

test("Component health fallback names its source and never hides a detail failure", () => {
  const sources = [{node: componentNode, application}];
  const details = {[resourceKey(componentNode)]: detail(component)};
  const row = buildComponentOverview(sources, details)[0];
  assert.equal(row.state, "Healthy");
  assert.equal(row.source, "Component-Status");
  details[resourceKey(componentNode)].error = "Forbidden";
  assert.equal(buildComponentOverview(sources, details)[0].state, "Unknown");
});

test("Blueprint failures, stopped execution and stale conditions are distinct", () => {
  const failed = structuredClone(blueprint);
  failed.status.conditions.find(entry => entry.type === "LastApplySucceeded").status = "False";
  assert.equal(buildBlueprintStatus(failed).phase, "Letzte Anwendung fehlgeschlagen");
  failed.spec.stopped = true;
  assert.equal(buildBlueprintStatus(failed).phase, "Gestoppt");
  failed.spec.stopped = false;
  failed.status.conditions.forEach(entry => entry.observedGeneration = 1);
  assert.equal(buildBlueprintStatus(failed).phase, "Status unbekannt");
  assert.ok(buildBlueprintStatus(failed).conditions.every(entry => entry.stale));
});

test("matches CES namespace and cluster without guessing aliases or empty destinations", () => {
  assert.equal(sameDestination(application, application), true);
  assert.equal(sameDestination(application, {...application, spec: {destination: {...application.spec.destination, server: "other"}}}), false);
  assert.equal(sameDestination(application, {...application, spec: {destination: {...application.spec.destination, namespace: "other"}}}), false);
  assert.equal(sameDestination({}, {}), false);
  assert.equal(sameDestination(application, {...application, spec: {destination: {namespace: "ecosystem", name: "alias"}}}), false);
});

test("resource URL preserves the ArgoCD base path and exact resource identity", () => {
  const url = new URL(resourceUrl(application, doguNode, "https://example.com/argo/"));
  assert.equal(url.pathname, "/argo/api/v1/applications/ces-blueprint/resource");
  assert.equal(url.searchParams.get("resourceName"), "swaggerui");
  assert.equal(url.searchParams.get("version"), "v2");
  assert.equal(url.searchParams.get("appNamespace"), "argocd");
});

function fixtureFetcher({failDetail, failList, apps: overrideApps} = {}) {
  const coreApp = {...application, metadata: {...application.metadata, name: "ces-core"}};
  const wrongApp = {...coreApp, metadata: {...coreApp.metadata, name: "other-ces"}, spec: {destination: {...coreApp.spec.destination, server: "other-cluster"}}};
  const calls = [];
  const fetcher = async (input, options) => {
    const url = new URL(input);
    calls.push(url);
    assert.equal(options.credentials, "same-origin");
    let body;
    if (url.pathname.endsWith("/resource-tree")) body = url.pathname.includes("ces-core") ? {nodes: [componentNode]} : tree;
    else if (url.pathname.endsWith("/applications")) {
      if (failList) return {ok: false, status: 403};
      body = {items: overrideApps || [application, coreApp, wrongApp]};
    } else {
      const kind = url.searchParams.get("kind");
      if (failDetail === kind) return {ok: false, status: 403};
      const data = {Blueprint: blueprint, Dogu: dogu, Component: component, Deployment: deployment}[kind];
      assert.ok(data, `Unexpected kind ${kind}`);
      body = {manifest: JSON.stringify(data)};
    }
    return {ok: true, json: async () => body};
  };
  return {fetcher, calls};
}

test("loads full details and Components through their own Application, excluding another CES", async () => {
  const {fetcher, calls} = fixtureFetcher();
  const result = await loadOverviewSnapshot({blueprint, application, baseURI: "https://example.com/argo/", fetcher});
  assert.equal(result.components.length, 1);
  assert.equal(result.components[0].application.metadata.name, "ces-core");
  assert.equal(result.details[resourceKey(doguNode)].resource.status.installedVersion, "1.0.0-1");
  assert.ok(calls.some(url => url.pathname.endsWith("ces-core/resource") && url.searchParams.get("kind") === "Component"));
  assert.ok(calls.every(url => !url.pathname.includes("other-ces")));
});

test("partial read failures retain the last values with an explicit per-resource error", async () => {
  const previous = {details: {[resourceKey(doguNode)]: detail(dogu)}};
  const {fetcher} = fixtureFetcher({failDetail: "Dogu"});
  const result = await loadOverviewSnapshot({blueprint, application, previous, baseURI: "https://example.com/", fetcher});
  assert.equal(result.details[resourceKey(doguNode)].resource.status.installedVersion, "1.0.0-1");
  assert.match(result.details[resourceKey(doguNode)].error, /Berechtigung/);
  assert.equal(result.details[resourceKey(componentNode)].error, null);
});

test("Application-list failure retains Components as stale while updating Dogus", async () => {
  const good = fixtureFetcher();
  const previous = await loadOverviewSnapshot({blueprint, application, baseURI: "https://example.com/", fetcher: good.fetcher});
  const bad = fixtureFetcher({failList: true});
  const result = await loadOverviewSnapshot({blueprint, application, previous, baseURI: "https://example.com/", fetcher: bad.fetcher});
  assert.equal(result.components.length, 1);
  assert.equal(result.components[0].stale, true);
  assert.equal(buildComponentOverview(result.components, result.details)[0].state, "Unknown");
  assert.equal(result.details[resourceKey(doguNode)].error, null);
  assert.match(result.warnings[0], /unvollständig/);
});

test("a failed tree refresh rejects instead of replacing the last snapshot with empty data", async () => {
  await assert.rejects(loadOverviewSnapshot({blueprint, application, baseURI: "https://example.com/", fetcher: async () => ({ok: false, status: 403})}), /Berechtigung/);
});
