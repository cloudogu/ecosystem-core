import {buildDoguOverview} from "./dogus.js";
import {loadOverviewSnapshot} from "./argo-api.js";
import {buildBlueprintStatus, buildComponentOverview} from "./status.js";
import {Badge, BlueprintStatus, DoguTable, ComponentTable, styles} from "./components.js";

const React = window.React;
const h = React.createElement;
const POLL_INTERVAL_MS = 15000;

function DoguOverview({resource, application, tree}) {
  const initial = () => ({blueprint: resource, tree: tree || {}, details: {}, components: [], warnings: [], time: null, error: null});
  const [snapshot, setSnapshot] = React.useState(initial);
  const [loading, setLoading] = React.useState(false);
  const latest = React.useRef(snapshot);
  const inputs = React.useRef({resource, application});
  inputs.current = {resource, application};
  const reload = React.useRef(() => {});
  const appName = application?.metadata?.name;
  const appNamespace = application?.metadata?.namespace;
  const blueprintNamespace = resource?.metadata?.namespace;
  const blueprintName = resource?.metadata?.name;

  React.useEffect(() => {
    const start = initial();
    latest.current = start;
    setSnapshot(start);
  }, [appName, appNamespace, blueprintNamespace, blueprintName]);

  React.useEffect(() => {
    // Parent tree updates can update health immediately without dropping detail data.
    setSnapshot(previous => {
      const updated = {...previous, tree: tree || previous.tree};
      latest.current = updated;
      return updated;
    });
  }, [tree]);

  React.useEffect(() => {
    let active = true;
    let controller;
    let busy = false;
    const refresh = async () => {
      if (busy || !active) return;
      busy = true;
      setLoading(true);
      controller = new AbortController();
      try {
        const current = inputs.current;
        const updated = await loadOverviewSnapshot({blueprint: current.resource, application: current.application,
          previous: latest.current, baseURI: document.baseURI, signal: controller.signal});
        if (active) {
          latest.current = updated;
          setSnapshot(updated);
        }
      } catch (error) {
        if (active) setSnapshot(previous => {
          const updated = {...previous, error: error.name === "AbortError" ? "Die Datenabfrage wurde abgebrochen." : error.message};
          latest.current = updated;
          return updated;
        });
      } finally {
        busy = false;
        if (active) setLoading(false);
      }
    };
    reload.current = refresh;
    refresh();
    const interval = window.setInterval(refresh, POLL_INTERVAL_MS);
    return () => { active = false; window.clearInterval(interval); controller?.abort(); };
  }, [appName, appNamespace, blueprintNamespace, blueprintName]);

  const overview = buildDoguOverview(snapshot.blueprint, snapshot.tree, snapshot.details);
  const status = buildBlueprintStatus(snapshot.blueprint, snapshot.blueprintError);
  const components = buildComponentOverview(snapshot.components, snapshot.details);
  const missingData = overview.present.filter(row => !row.found).length;
  return h("section", {className: "ces-dogu-tab", "aria-label": "CES-Übersicht des Blueprints"},
    h("div", {className: "ces-dogu-tab__toolbar"},
      h("div", null, h("h3", null, "CES-Übersicht"), h("div", {className: "ces-dogu-tab__muted"}, `${blueprintNamespace || ""}/${blueprintName || ""}`)),
      h("button", {type: "button", className: "argo-button argo-button--base", disabled: loading, onClick: () => reload.current()}, loading ? "Aktualisierung läuft …" : "Aktualisieren")),
    snapshot.error && h("div", {className: "ces-dogu-tab__notice", role: "alert"},
      h("strong", {className: "ces-dogu-tab__error"}, snapshot.error), h("p", null, "Die angezeigten Werte können veraltet sein.")),
    snapshot.warnings.length > 0 && h("div", {className: "ces-dogu-tab__notice", role: "status"}, snapshot.warnings.map(message => h("p", {key: message}, message))),
    h(BlueprintStatus, {status}),
    h("h4", null, `Dogus (${overview.present.length})`),
    h("p", {className: "ces-dogu-tab__muted"}, "Soll-Versionen stammen aus der Blueprint-Deklaration, installierte Versionen aus dem Dogu-Status. Replikate bedeuten verfügbar / gewünscht."),
    overview.hasMask && h("p", {className: "ces-dogu-tab__notice"}, "Eine Blueprint-Maske ist gesetzt. Die Soll-Versionen zeigen die Deklaration ohne Maskenauswertung; angezeigte Abweichungen können durch die Maske beabsichtigt sein."),
    h("div", {className: "ces-dogu-tab__summary", "aria-label": "Dogu-Health Übersicht"},
      Object.entries(overview.counts).map(([state, count]) => h("span", {key: state}, h(Badge, {state}), `: ${count}`)),
      h("span", null, `Versionsabweichungen: ${overview.deviations}`)),
    missingData > 0 && h("p", {className: "ces-dogu-tab__notice"}, `Für ${missingData} Dogus sind keine Ressourcendaten verfügbar. Das bedeutet nicht automatisch, dass sie im Cluster fehlen.`),
    overview.present.length ? h(DoguTable, {rows: overview.present}) : h("p", null, "Im Blueprint sind keine Dogus zur Installation aufgeführt."),
    overview.absent.length > 0 && h("details", null, h("summary", null, `Zur Entfernung vorgesehen (${overview.absent.length})`), h(DoguTable, {rows: overview.absent, absent: true})),
    h(ComponentTable, {rows: components, namespace: blueprintNamespace}),
    h("p", {className: "ces-dogu-tab__muted", "aria-live": "polite"}, snapshot.time ?
      `Letzte Datenaktualisierung: ${snapshot.time.toLocaleTimeString("de-DE")} · Aktualisierung alle 15 Sekunden` : "Warte auf Ressourcendaten …"));
}

if (!document.getElementById("ces-dogu-tab-styles")) {
  const style = document.createElement("style");
  style.id = "ces-dogu-tab-styles";
  style.textContent = styles;
  document.head.appendChild(style);
}

window.extensionsAPI.registerResourceExtension(DoguOverview, "k8s.cloudogu.com", "Blueprint", "Dogus");
