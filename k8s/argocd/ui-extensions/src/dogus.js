import {doguEventResources} from "./events.js";
import {resourceKey, treeNodes} from "./argo-api.js";
import {HEALTH_STATES, versionComparison, operatingState, deploymentReplicas} from "./status.js";
export {resourceTreeUrl} from "./argo-api.js";

export function buildDoguOverview(blueprint, tree = {}, details = {}) {
  const namespace = blueprint?.metadata?.namespace;
  const allNodes = treeNodes(tree);
  const nodes = new Map(allNodes.filter(node => node.group === "k8s.cloudogu.com" && node.kind === "Dogu" && node.namespace === namespace).map(node => [node.name, node]));
  const present = [];
  const absent = [];
  for (const entry of blueprint?.spec?.blueprint?.dogus || []) {
    if (typeof entry.name !== "string" || !entry.name.trim()) continue;
    const name = entry.name.split("/").at(-1);
    const node = nodes.get(name);
    const detail = node && details[resourceKey(node)];
    const resource = detail?.resource;
    const installedVersion = resource?.status?.installedVersion;
    const deployments = node ? allNodes.filter(candidate => candidate.group === "apps" && candidate.kind === "Deployment" && candidate.namespace === namespace &&
      candidate.parentRefs?.some(parent => resourceKey(parent) === resourceKey(node) && (!parent.uid || parent.uid === node.uid))) : [];
    let comparison = versionComparison(entry.version, installedVersion, detail?.error);
    if (!detail?.error && resource?.spec?.version && entry.version && resource.spec.version !== entry.version && comparison.state !== "Degraded") {
      comparison = {state: "Degraded", text: "CR-Ziel weicht ab"};
    }
    const row = {
      name, qualifiedName: entry.name, version: entry.version || "—", installedVersion: installedVersion || "—",
      specVersion: resource?.spec?.version || "—", comparison,
      operations: operatingState(resource), detailError: detail?.error,
      replicas: deployments.map(deployment => deploymentReplicas(details[resourceKey(deployment)]) || {name: deployment.name, text: "Unbekannt"}),
      state: HEALTH_STATES.has(node?.health?.status) ? node.health.status : "Unknown",
      message: node?.health?.message || (node ? "Keine Health-Bewertung verfügbar." : "Keine Ressourcendaten im ArgoCD-Baum verfügbar."),
      found: Boolean(node), eventResources: doguEventResources(node, allNodes),
    };
    (entry.absent === true ? absent : present).push(row);
  }
  const byName = (left, right) => left.name.localeCompare(right.name);
  present.sort(byName);
  absent.sort(byName);
  return {present, absent, counts: present.reduce((counts, row) => ({...counts, [row.state]: (counts[row.state] || 0) + 1}), {}),
    hasMask: Boolean(blueprint?.spec?.blueprintMask), deviations: present.filter(row => row.comparison.state === "Degraded").length};
}
