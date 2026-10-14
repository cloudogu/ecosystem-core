import {buildDoguOverview, resourceTreeUrl} from "./dogus.js";

const React = window.React;
const h = React.createElement;
const POLL_INTERVAL_MS = 15000;

const styles = `
.ces-dogu-tab { padding: 20px; font: inherit; }
.ces-dogu-tab h3 { margin: 0 0 6px; font-size: 20px; }
.ces-dogu-tab p { margin: 8px 0; }
.ces-dogu-tab__toolbar { display: flex; align-items: center; justify-content: space-between; gap: 16px; flex-wrap: wrap; margin-bottom: 16px; }
.ces-dogu-tab__muted { opacity: .75; font-size: 13px; }
.ces-dogu-tab__summary { display: flex; gap: 12px; flex-wrap: wrap; margin: 12px 0; }
.ces-dogu-tab__table-wrapper { overflow-x: auto; }
.ces-dogu-tab table { width: 100%; border-collapse: collapse; text-align: left; }
.ces-dogu-tab th, .ces-dogu-tab td { padding: 12px 10px; border-bottom: 1px solid rgba(128,128,128,.3); vertical-align: top; }
.ces-dogu-tab th { font-size: 12px; text-transform: uppercase; opacity: .75; }
.ces-dogu-tab__badge { display: inline-flex; align-items: center; gap: 6px; white-space: nowrap; font-weight: 600; }
.ces-dogu-tab__dot { display: inline-block; width: 9px; height: 9px; border-radius: 50%; background: #7f8c8d; }
.ces-dogu-tab__dot--Healthy { background: #18a36b; }
.ces-dogu-tab__dot--Degraded, .ces-dogu-tab__dot--Missing { background: #e65454; }
.ces-dogu-tab__dot--Progressing { background: #e5a126; }
.ces-dogu-tab__dot--Suspended { background: #6386ca; }
.ces-dogu-tab__notice { padding: 12px; border: 1px solid rgba(229,161,38,.5); border-radius: 4px; margin-bottom: 12px; }
.ces-dogu-tab__error { color: #d34848; }
.ces-dogu-tab details { margin-top: 24px; }
.ces-dogu-tab summary { cursor: pointer; margin-bottom: 12px; }
.ces-dogu-tab button { cursor: pointer; }
`;

function Badge({state}) {
  return h("span", {className: "ces-dogu-tab__badge"},
    h("span", {className: `ces-dogu-tab__dot ces-dogu-tab__dot--${state}`, "aria-hidden": true}), state);
}

function DoguTable({rows, absent = false}) {
  return h("div", {className: "ces-dogu-tab__table-wrapper"},
    h("table", {"aria-label": absent ? "Zur Entfernung vorgesehene Dogus" : "Dogu-Liste"},
      h("thead", null, h("tr", null,
        h("th", {scope: "col"}, "Dogu"),
        h("th", {scope: "col"}, "Version im Blueprint"),
        h("th", {scope: "col"}, absent ? "Ressource" : "Health"),
        h("th", {scope: "col"}, "Meldung"))),
      h("tbody", null, rows.map(row => h("tr", {key: row.qualifiedName},
        h("td", null, h("strong", null, row.name),
          h("div", {className: "ces-dogu-tab__muted"}, row.qualifiedName)),
        h("td", null, row.version),
        h("td", null, absent ? (row.found ? "Noch vorhanden" : "Nicht im Ressourcenbaum") : h(Badge, {state: row.state})),
        h("td", null, row.message))))));
}

function DoguOverview({resource, application, tree}) {
  const [snapshot, setSnapshot] = React.useState({tree: tree || {}, time: null, error: null});
  const [loading, setLoading] = React.useState(false);
  const reload = React.useRef(() => {});
  const appName = application?.metadata?.name;
  const appNamespace = application?.metadata?.namespace;
  const blueprintNamespace = resource?.metadata?.namespace;
  const blueprintName = resource?.metadata?.name;

  React.useEffect(() => {
    setSnapshot({tree: tree || {}, time: new Date(), error: null});
  }, [tree, appName, appNamespace, blueprintNamespace, blueprintName]);

  React.useEffect(() => {
    let active = true;
    let controller;
    let busy = false;
    const refresh = async () => {
      if (busy || !active) return;
      busy = true;
      setLoading(true);
      controller = new AbortController();
      const timeout = window.setTimeout(() => controller.abort(), 10000);
      try {
        const url = resourceTreeUrl(application, document.baseURI);
        const response = await fetch(url, {credentials: "same-origin", headers: {Accept: "application/json"}, signal: controller.signal});
        if (!response.ok) {
          throw new Error(response.status === 401 ? "Bitte erneut bei ArgoCD anmelden." :
            response.status === 403 ? "Keine Berechtigung zum Lesen dieser Application." : `Die Datenabfrage ist fehlgeschlagen (HTTP ${response.status}).`);
        }
        const data = await response.json();
        if (!data || (!Array.isArray(data.nodes) && !Array.isArray(data.orphanedNodes))) {
          throw new Error("ArgoCD hat keinen gültigen Ressourcenbaum geliefert.");
        }
        if (active) setSnapshot({tree: data, time: new Date(), error: null});
      } catch (error) {
        if (active) setSnapshot(previous => ({...previous, error: error.name === "AbortError" ? "Die Datenabfrage hat zu lange gedauert." : error.message}));
      } finally {
        window.clearTimeout(timeout);
        busy = false;
        if (active) setLoading(false);
      }
    };
    reload.current = refresh;
    refresh();
    const interval = window.setInterval(refresh, POLL_INTERVAL_MS);
    return () => {
      active = false;
      window.clearInterval(interval);
      controller?.abort();
    };
  }, [appName, appNamespace, blueprintNamespace, blueprintName]);

  const overview = buildDoguOverview(resource, snapshot.tree);
  const missingData = overview.present.filter(row => !row.found).length;
  return h("section", {className: "ces-dogu-tab", "aria-label": "Dogus des Blueprints"},
    h("div", {className: "ces-dogu-tab__toolbar"},
      h("div", null, h("h3", null, "Dogus"),
        h("div", {className: "ces-dogu-tab__muted"}, `${blueprintNamespace || ""}/${blueprintName || ""} · ${overview.present.length} Dogus im Blueprint`)),
      h("button", {type: "button", className: "argo-button argo-button--base", disabled: loading, onClick: () => reload.current()}, loading ? "Aktualisierung läuft …" : "Aktualisieren")),
    snapshot.error && h("div", {className: "ces-dogu-tab__notice", role: "alert"},
      h("strong", {className: "ces-dogu-tab__error"}, snapshot.error),
      h("p", null, "Die angezeigten Werte können veraltet sein.")),
    overview.hasMask && h("p", {className: "ces-dogu-tab__muted"}, "Die Liste stammt aus dem Blueprint. Eine Blueprint-Maske kann den tatsächlich angewendeten Zustand verändern."),
    h("div", {className: "ces-dogu-tab__summary", "aria-label": "Dogu-Health Übersicht"},
      Object.entries(overview.counts).map(([state, count]) => h("span", {key: state}, h(Badge, {state}), `: ${count}`))),
    missingData > 0 && h("p", {className: "ces-dogu-tab__notice"}, `Für ${missingData} Dogus sind keine Ressourcendaten verfügbar. Das bedeutet nicht automatisch, dass sie im Cluster fehlen.`),
    overview.present.length ? h(DoguTable, {rows: overview.present}) : h("p", null, "Im Blueprint sind keine Dogus zur Installation aufgeführt."),
    overview.absent.length > 0 && h("details", null,
      h("summary", null, `Zur Entfernung vorgesehen (${overview.absent.length})`),
      h(DoguTable, {rows: overview.absent, absent: true})),
    h("p", {className: "ces-dogu-tab__muted", "aria-live": "polite"},
      snapshot.time ? `Letzte Datenaktualisierung: ${snapshot.time.toLocaleTimeString("de-DE")} · Aktualisierung alle 15 Sekunden` : "Warte auf Ressourcendaten …"));
}

if (!document.getElementById("ces-dogu-tab-styles")) {
  const style = document.createElement("style");
  style.id = "ces-dogu-tab-styles";
  style.textContent = styles;
  document.head.appendChild(style);
}

window.extensionsAPI.registerResourceExtension(DoguOverview, "k8s.cloudogu.com", "Blueprint", "Dogus");
