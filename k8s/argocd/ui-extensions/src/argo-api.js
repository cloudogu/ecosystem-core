export function resourceKey(node) {
  return JSON.stringify([node.group || "", node.kind, node.namespace || "", node.name]);
}

export function treeNodes(tree = {}) {
  const nodes = new Map();
  for (const node of [...(tree.orphanedNodes || []), ...(tree.nodes || [])]) nodes.set(resourceKey(node), node);
  return [...nodes.values()];
}

function applicationUrl(application, baseURI, suffix) {
  const name = application?.metadata?.name;
  if (!name) throw new Error("Die ArgoCD-Application ist nicht verfügbar.");
  const url = new URL(`api/v1/applications/${encodeURIComponent(name)}/${suffix}`, baseURI);
  url.searchParams.set("appNamespace", application.metadata.namespace || "argocd");
  return url;
}

export function resourceTreeUrl(application, baseURI) {
  return applicationUrl(application, baseURI, "resource-tree").toString();
}

export function resourceUrl(application, node, baseURI) {
  const url = applicationUrl(application, baseURI, "resource");
  for (const [key, value] of Object.entries({resourceName: node.name, namespace: node.namespace || "", group: node.group || "", kind: node.kind, version: node.version || ""})) {
    url.searchParams.set(key, value);
  }
  return url.toString();
}

export function sameDestination(left, right) {
  const a = left?.spec?.destination;
  const b = right?.spec?.destination;
  if (!a?.namespace || a.namespace !== b?.namespace) return false;
  // Do not infer that a cluster alias and an unrelated server identify the same cluster.
  return Boolean((a.server && a.server === b.server) || (!a.server && !b.server && a.name && a.name === b.name));
}

async function requestJSON(url, signal, fetcher) {
  const controller = new AbortController();
  const abort = () => controller.abort();
  signal?.addEventListener("abort", abort, {once: true});
  if (signal?.aborted) controller.abort();
  const timeout = setTimeout(abort, 10000);
  try {
    const response = await fetcher(url, {credentials: "same-origin", headers: {Accept: "application/json"}, signal: controller.signal});
    if (!response.ok) throw new Error(response.status === 401 ? "Bitte erneut bei ArgoCD anmelden." :
      response.status === 403 ? "Keine Berechtigung zum Lesen dieser Application." : `Die Datenabfrage ist fehlgeschlagen (HTTP ${response.status}).`);
    return await response.json();
  } catch (error) {
    if (controller.signal.aborted && !signal?.aborted) throw new Error("Die Datenabfrage hat zu lange gedauert.");
    throw error;
  } finally {
    clearTimeout(timeout);
    signal?.removeEventListener("abort", abort);
  }
}

async function mapLimited(items, callback, limit = 6) {
  let index = 0;
  await Promise.all(Array.from({length: Math.min(limit, items.length)}, async () => {
    while (index < items.length) {
      const item = items[index++];
      await callback(item);
    }
  }));
}

/** Read only the Blueprint, declared Dogus, their Deployments and CES Component CRs. */
export async function loadOverviewSnapshot({blueprint, application, previous = {}, baseURI, signal, fetcher = fetch}) {
  const request = url => requestJSON(url, signal, fetcher);
  const readTree = async app => {
    const result = await request(resourceTreeUrl(app, baseURI));
    if (!result || (!Array.isArray(result.nodes) && !Array.isArray(result.orphanedNodes))) throw new Error("ArgoCD hat keinen gültigen Ressourcenbaum geliefert.");
    return result;
  };
  const tree = await readTree(application);
  const namespace = blueprint?.metadata?.namespace;
  const ownNodes = treeNodes(tree);
  const details = {};
  const warnings = [];
  const jobs = new Map();
  const queue = (node, app = application) => jobs.set(resourceKey(node), {node, app});
  const bpNode = ownNodes.find(node => node.group === "k8s.cloudogu.com" && node.kind === "Blueprint" && node.namespace === namespace && node.name === blueprint?.metadata?.name);
  if (bpNode) queue(bpNode);
  else warnings.push("Der Blueprint fehlt im aktuellen Ressourcenbaum; sein Status kann veraltet sein.");

  const components = new Map();
  const addComponents = (sourceTree, app) => {
    for (const node of treeNodes(sourceTree)) {
      if (node.group === "k8s.cloudogu.com" && node.kind === "Component" && node.namespace === namespace) {
        components.set(resourceKey(node), {node, application: app});
        queue(node, app);
      }
    }
  };
  addComponents(tree, application);
  const retainComponents = (predicate, message) => {
    for (const previousComponent of previous.components || []) {
      const key = resourceKey(previousComponent.node);
      if (predicate(previousComponent.application) && !components.has(key)) {
        components.set(key, {...previousComponent, stale: true});
        details[key] = {...previous.details?.[key], error: message};
      }
    }
  };
  try {
    const apps = await request(new URL("api/v1/applications", baseURI).toString());
    if (!Array.isArray(apps?.items)) throw new Error("ArgoCD hat keine gültige Application-Liste geliefert.");
    const related = apps.items.filter(app => sameDestination(application, app) &&
      !(app.metadata?.name === application.metadata?.name && app.metadata?.namespace === application.metadata?.namespace));
    await mapLimited(related, async app => {
      try { addComponents(await readTree(app), app); }
      catch (error) {
        const message = `Komponenten aus ${app.metadata.name}: ${error.message}`;
        warnings.push(message);
        retainComponents(source => source?.metadata?.name === app.metadata.name && source?.metadata?.namespace === app.metadata.namespace, message);
      }
    }, 3);
  } catch (error) {
    const message = `Die Komponentenübersicht ist möglicherweise unvollständig: ${error.message}`;
    warnings.push(message);
    retainComponents(() => true, message);
  }

  // Fetch the current Blueprint first; changes to its declaration should update this view too.
  const readDetail = async ({node, app}) => {
    if (signal?.aborted) throw new DOMException("Aborted", "AbortError");
    const key = resourceKey(node);
    try {
      const body = await request(resourceUrl(app, node, baseURI));
      const resource = typeof body?.manifest === "string" ? JSON.parse(body.manifest) : null;
      const group = resource?.apiVersion?.includes("/") ? resource.apiVersion.split("/")[0] : "";
      if (!resource || resourceKey({group, kind: resource.kind, namespace: resource.metadata?.namespace, name: resource.metadata?.name}) !== key) {
        throw new Error("ArgoCD hat keine passende Ressource geliefert.");
      }
      details[key] = {resource, time: new Date().toISOString(), error: null};
    } catch (error) {
      if (signal?.aborted) throw error;
      // Keep values only for the same resource identity; errors remain visible next to them.
      details[key] = {...previous.details?.[key], error: error.message};
    }
  };
  if (bpNode) {
    await readDetail(jobs.get(resourceKey(bpNode)));
    jobs.delete(resourceKey(bpNode));
  }
  const blueprintDetail = bpNode && details[resourceKey(bpNode)];
  const currentBlueprint = blueprintDetail?.resource || blueprint;
  const names = new Set((currentBlueprint?.spec?.blueprint?.dogus || []).map(entry => entry.name?.split("/").at(-1)));
  const dogus = ownNodes.filter(node => node.group === "k8s.cloudogu.com" && node.kind === "Dogu" && node.namespace === namespace && names.has(node.name));
  for (const dogu of dogus) queue(dogu);
  for (const node of ownNodes) {
    if (node.group === "apps" && node.kind === "Deployment" && node.namespace === namespace &&
        node.parentRefs?.some(parent => dogus.some(dogu => resourceKey(parent) === resourceKey(dogu) && (!parent.uid || parent.uid === dogu.uid)))) queue(node);
  }
  await mapLimited([...jobs.values()], readDetail);
  if (signal?.aborted) throw new DOMException("Aborted", "AbortError");
  return {blueprint: currentBlueprint, blueprintError: blueprintDetail?.error || (!bpNode ? warnings[0] : null), tree,
    details, components: [...components.values()], warnings, time: new Date(), error: null};
}
