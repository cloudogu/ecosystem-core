import {resourceKey} from "./argo-api.js";

export const HEALTH_STATES = new Set(["Healthy", "Degraded", "Progressing", "Suspended", "Missing", "Unknown"]);

export function condition(resource, type) {
  return resource?.status?.conditions?.find(entry => entry.type === type);
}

export function currentCondition(resource, type) {
  const entry = condition(resource, type);
  if (entry?.observedGeneration != null && resource?.metadata?.generation != null && entry.observedGeneration < resource.metadata.generation) return null;
  return entry;
}

export function versionComparison(desired, installed, error) {
  if (error) return {state: "Unknown", text: "Daten veraltet / nicht verfügbar"};
  if (!desired || !installed) return {state: "Unknown", text: "Nicht vergleichbar"};
  return desired === installed ? {state: "Healthy", text: "Übereinstimmung"} : {state: "Degraded", text: "Abweichung"};
}

export function operatingState(resource) {
  if (!resource) return ["Unbekannt"];
  const flags = [];
  if (resource.spec?.stopped === true) flags.push("Gestoppt");
  if (resource.spec?.supportMode === true || currentCondition(resource, "supportMode")?.status === "True") flags.push("Supportmodus");
  if (resource.spec?.pauseReconciliation === true || currentCondition(resource, "pauseReconciliation")?.status === "True") flags.push("Reconciliation pausiert");
  const ready = currentCondition(resource, "ready");
  if (ready) flags.push(ready.status === "True" ? "Bereit" : ready.status === "False" ? "Nicht bereit" : "Bereitschaft unbekannt");
  else flags.push(resource.status?.status || "Bereitschaft unbekannt");
  return flags;
}

export function deploymentReplicas(detail) {
  const resource = detail?.resource;
  if (!resource) return null;
  const desired = resource.spec?.replicas ?? 1;
  const available = resource.status?.availableReplicas ?? 0;
  const stale = Boolean(detail.error) || (resource.status?.observedGeneration != null && resource.metadata?.generation != null && resource.status.observedGeneration < resource.metadata.generation);
  return {name: resource.metadata.name, desired, available, stale, text: `${available}/${desired}${stale ? " (veraltet)" : ""}`};
}

export function buildBlueprintStatus(resource, error) {
  const conditions = resource?.status?.conditions || [];
  const is = (type, value) => currentCondition(resource, type)?.status === value;
  let phase = "Status unbekannt";
  if (resource?.spec?.stopped) phase = "Gestoppt";
  else if (is("LastApplySucceeded", "False")) phase = "Letzte Anwendung fehlgeschlagen";
  else if (is("Valid", "False")) phase = "Ungültig";
  else if (is("Executable", "False")) phase = "Nicht ausführbar";
  else if (is("Completed", "True")) phase = "Abgeschlossen";
  else if (is("EcosystemHealthy", "False")) phase = "Wartet auf gesunde Dogus";
  else if (is("Executable", "True")) phase = "Noch nicht abgeschlossen";
  const completed = condition(resource, "Completed");
  const lastApply = condition(resource, "LastApplySucceeded");
  return {phase, error, stopped: resource?.spec?.stopped === true, ignoreDoguHealth: resource?.spec?.ignoreDoguHealth === true,
    lastTransition: lastApply?.lastTransitionTime || completed?.lastTransitionTime,
    conditions: conditions.map(entry => ({...entry, stale: entry.observedGeneration != null && resource?.metadata?.generation != null && entry.observedGeneration < resource.metadata.generation}))};
}

export function buildComponentOverview(components = [], details = {}) {
  return components.map(({node, application}) => {
    const detail = details[resourceKey(node)];
    const resource = detail?.resource;
    let state = HEALTH_STATES.has(node.health?.status) ? node.health.status : "Unknown";
    let source = "ArgoCD";
    if (state === "Unknown" && resource) {
      state = resource.status?.health === "available" ? "Healthy" : resource.status?.health === "unavailable" ? "Degraded" : "Unknown";
      source = "Component-Status";
    }
    if (detail?.error) state = "Unknown";
    const desired = resource?.spec?.version;
    const installed = resource?.status?.installedVersion;
    return {name: node.name, desired: desired || "—", installed: installed || "—", state, source,
      comparison: versionComparison(desired, installed, detail?.error), status: resource?.status?.status || "Unbekannt",
      application: application?.metadata?.name || "—", message: detail?.error || node.health?.message || ""};
  }).sort((a, b) => a.name.localeCompare(b.name));
}
