import assert from "node:assert/strict";
import test from "node:test";
import {buildDoguOverview, resourceTreeUrl} from "../src/dogus.js";

const blueprint = {metadata: {name: "ces-test", namespace: "ecosystem"}, spec: {blueprint: {dogus: [
  {name: "official/swaggerui", version: "5.32.4-1"},
  {name: "official/mysql", version: "8.4.11-2"},
]}}};
const node = (name, status = "Healthy", overrides = {}) => ({group: "k8s.cloudogu.com", kind: "Dogu", namespace: "ecosystem", name, health: {status, message: `${name} status`}, ...overrides});

test("includes orphaned Dogus with their health, ordered by name", () => {
  const result = buildDoguOverview(blueprint, {orphanedNodes: [node("swaggerui", "Degraded"), node("mysql")]});
  assert.deepEqual(result.present.map(row => [row.name, row.state]), [["mysql", "Healthy"], ["swaggerui", "Degraded"]]);
  assert.deepEqual(result.counts, {Healthy: 1, Degraded: 1});
});

test("keeps unrelated kinds, namespaces, groups and Dogus out of the list", () => {
  const result = buildDoguOverview(blueprint, {nodes: [
    node("mysql", "Degraded", {namespace: "another-ces"}),
    node("swaggerui", "Healthy", {kind: "Deployment"}),
    node("mysql", "Healthy", {group: "other.example"}),
    node("other-dogu"),
  ]});
  assert.equal(result.present.length, 2);
  assert.ok(result.present.every(row => !row.found && row.state === "Unknown"));
});

test("prefers regular nodes over duplicate orphaned nodes", () => {
  const result = buildDoguOverview(blueprint, {nodes: [node("swaggerui")], orphanedNodes: [node("swaggerui", "Degraded")]});
  assert.equal(result.present.find(row => row.name === "swaggerui").state, "Healthy");
});

test("separates absent entries and does not interpret missing tree data as deleted", () => {
  const resource = structuredClone(blueprint);
  resource.spec.blueprint.dogus[0].absent = true;
  const result = buildDoguOverview(resource, {});
  assert.deepEqual(result.present.map(row => row.name), ["mysql"]);
  assert.deepEqual(result.absent.map(row => [row.name, row.found, row.state]), [["swaggerui", false, "Unknown"]]);
});

test("does not assume health when the resource has no health check", () => {
  const result = buildDoguOverview(blueprint, {nodes: [node("mysql", "Healthy", {health: undefined})]});
  assert.equal(result.present[0].found, true);
  assert.equal(result.present[0].state, "Unknown");
});

test("updates health when a newer tree arrives", () => {
  const healthy = buildDoguOverview(blueprint, {nodes: [node("swaggerui")]});
  const degraded = buildDoguOverview(blueprint, {nodes: [node("swaggerui", "Degraded")]});
  assert.equal(healthy.present[1].state, "Healthy");
  assert.equal(degraded.present[1].state, "Degraded");
});

test("handles empty Blueprints and identifies inline masks", () => {
  assert.deepEqual(buildDoguOverview({}).present, []);
  assert.equal(buildDoguOverview({...blueprint, spec: {...blueprint.spec, blueprintMask: {manifest: {}}}}).hasMask, true);
});

test("uses the ArgoCD base path and application namespace", () => {
  const url = resourceTreeUrl({metadata: {name: "ces blueprint", namespace: "custom-argocd"}}, "https://example.com/argo/");
  assert.equal(url, "https://example.com/argo/api/v1/applications/ces%20blueprint/resource-tree?appNamespace=custom-argocd");
});
