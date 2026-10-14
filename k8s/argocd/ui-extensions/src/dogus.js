const DOGU_GROUP = "k8s.cloudogu.com";
const HEALTH_STATES = new Set(["Healthy", "Degraded", "Progressing", "Suspended", "Missing", "Unknown"]);

export function buildDoguOverview(blueprint, tree = {}) {
  const namespace = blueprint?.metadata?.namespace;
  const nodes = new Map();
  // Prefer regular application nodes when a resource occurs in both lists.
  for (const node of [...(tree.orphanedNodes || []), ...(tree.nodes || [])]) {
    if (node.kind === "Dogu" && node.group === DOGU_GROUP && node.namespace === namespace) {
      nodes.set(node.name, node);
    }
  }

  const present = [];
  const absent = [];
  const entries = blueprint?.spec?.blueprint?.dogus || [];
  for (const entry of entries) {
    if (typeof entry.name !== "string" || !entry.name.trim()) continue;
    const name = entry.name.split("/").at(-1);
    const node = nodes.get(name);
    const state = HEALTH_STATES.has(node?.health?.status) ? node.health.status : "Unknown";
    const row = {
      name,
      qualifiedName: entry.name,
      version: entry.version || "—",
      state,
      message: node?.health?.message || (node ? "Keine Health-Bewertung verfügbar." : "Keine Ressourcendaten im ArgoCD-Baum verfügbar."),
      found: Boolean(node),
    };
    (entry.absent === true ? absent : present).push(row);
  }
  const byName = (left, right) => left.name.localeCompare(right.name);
  present.sort(byName);
  absent.sort(byName);
  return {
    present,
    absent,
    counts: present.reduce((counts, row) => ({...counts, [row.state]: (counts[row.state] || 0) + 1}), {}),
    hasMask: Boolean(blueprint?.spec?.blueprintMask),
  };
}

export function resourceTreeUrl(application, baseURI) {
  const name = application?.metadata?.name;
  if (!name) throw new Error("Die ArgoCD-Application ist nicht verfügbar.");
  const url = new URL(`api/v1/applications/${encodeURIComponent(name)}/resource-tree`, baseURI);
  url.searchParams.set("appNamespace", application.metadata.namespace || "argocd");
  return url.toString();
}
