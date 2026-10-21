import {resourceKey, resourceEventsUrl, requestJSON, mapLimited} from "./argo-api.js";

const EVENT_KINDS = new Set(["k8s.cloudogu.com/Dogu", "apps/Deployment", "apps/ReplicaSet", "apps/StatefulSet", "batch/Job", "/Pod", "/PersistentVolumeClaim"]);
export const EVENT_WINDOW_MS = 24 * 60 * 60 * 1000;
export const EVENT_LIMIT = 20;
export const eventSourceKey = node => JSON.stringify([resourceKey(node), node.uid]);

/** Follow real ownership, including UIDs, rather than matching Pod name prefixes. */
export function doguEventResources(dogu, nodes = []) {
  if (!dogu?.uid) return [];
  const owned = new Map([[eventSourceKey(dogu), dogu]]);
  let changed = true;
  while (changed) {
    changed = false;
    for (const node of nodes) {
      if (!node.uid || node.namespace !== dogu.namespace || owned.has(eventSourceKey(node))) continue;
      if (node.parentRefs?.some(parent => [...owned.values()].some(owner => parent.uid === owner.uid && resourceKey(parent) === resourceKey(owner)))) {
        owned.set(eventSourceKey(node), node);
        changed = true;
      }
    }
  }
  return [...owned.values()].filter(node => EVENT_KINDS.has(`${node.group || ""}/${node.kind}`));
}

export function normalizeResourceEvents(items, node) {
  const normalized = new Map();
  for (const event of items) {
    const ref = event?.involvedObject;
    if (!ref || ref.uid !== node.uid || ref.kind !== node.kind || ref.name !== node.name || ref.namespace !== node.namespace) continue;
    if (ref.apiVersion && (ref.apiVersion.includes("/") ? ref.apiVersion.split("/")[0] : "") !== (node.group || "")) continue;
    const dates = [event.series?.lastObservedTime, event.lastTimestamp, event.eventTime, event.firstTimestamp, event.metadata?.creationTimestamp]
      .map(value => typeof value === "string" ? Date.parse(value) : NaN).filter(Number.isFinite);
    const timestamp = dates.length ? Math.max(...dates) : null;
    const id = event.metadata?.uid || JSON.stringify([node.uid, event.type, event.reason, event.message, event.firstTimestamp]);
    const count = Math.max(1, Number(event.series?.count) || Number(event.count) || 1);
    const item = {id, resource: `${node.kind}/${node.name}`, resourceUID: node.uid, timestamp, count,
      type: event.type || "Unbekannt", reason: event.reason || "—", message: event.message || "—"};
    const existing = normalized.get(id);
    if (!existing || (item.timestamp ?? -Infinity) >= (existing.timestamp ?? -Infinity)) normalized.set(id, item);
  }
  return [...normalized.values()];
}

export function recentEvents(items = [], {now = Date.now(), warningsOnly = true, limit = EVENT_LIMIT} = {}) {
  return items.filter(event => (!warningsOnly || event.type === "Warning") && (event.timestamp == null || event.timestamp >= now - EVENT_WINDOW_MS))
    .sort((a, b) => (b.timestamp ?? -Infinity) - (a.timestamp ?? -Infinity) || a.id.localeCompare(b.id)).slice(0, limit);
}

export async function loadDoguEvents({application, resources, previous = {}, baseURI, signal, fetcher = fetch}) {
  const byResource = {};
  const unique = new Map(resources.filter(node => node.uid).map(node => [eventSourceKey(node), node]));
  await mapLimited([...unique.values()], async node => {
    if (signal?.aborted) throw new DOMException("Aborted", "AbortError");
    const key = eventSourceKey(node);
    try {
      const body = await requestJSON(resourceEventsUrl(application, node, baseURI), signal, fetcher);
      if (!body || typeof body !== "object" || Array.isArray(body) || (body.items != null && !Array.isArray(body.items))) throw new Error("ArgoCD hat keine gültige Ereignisliste geliefert.");
      byResource[key] = {items: normalizeResourceEvents(body.items || [], node), error: null};
    } catch (error) {
      if (signal?.aborted) throw error;
      byResource[key] = {items: previous.byResource?.[key]?.items || [], error: `${node.kind}/${node.name}: ${error.message}`};
    }
  }, 4);
  if (signal?.aborted) throw new DOMException("Aborted", "AbortError");
  return {byResource, time: new Date()};
}
