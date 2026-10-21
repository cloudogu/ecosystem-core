import {eventSourceKey, loadDoguEvents, recentEvents, EVENT_LIMIT} from "./events.js";

const React = window.React;
const h = React.createElement;

export function DoguEvents({application, resources = [], doguName}) {
  const [open, setOpen] = React.useState(false);
  const [warningsOnly, setWarningsOnly] = React.useState(true);
  const [snapshot, setSnapshot] = React.useState({byResource: {}, time: null});
  const [loading, setLoading] = React.useState(false);
  const previous = React.useRef(snapshot);
  const inputs = React.useRef({application, resources});
  inputs.current = {application, resources};
  const reload = React.useRef(() => {});
  const identity = JSON.stringify([application?.metadata?.namespace, application?.metadata?.name, resources.map(eventSourceKey).sort()]);

  React.useEffect(() => {
    if (!open || resources.length === 0) return;
    let active = true;
    let busy = false;
    let controller;
    const refresh = async () => {
      if (!active || busy) return;
      busy = true;
      setLoading(true);
      controller = new AbortController();
      try {
        const result = await loadDoguEvents({...inputs.current, previous: previous.current, baseURI: document.baseURI, signal: controller.signal});
        if (active) {
          previous.current = result;
          setSnapshot(result);
        }
      } finally {
        busy = false;
        if (active) setLoading(false);
      }
    };
    // Abort is expected when closing or changing the Dogu; prevent an unhandled rejection.
    const run = () => refresh().catch(error => {
      if (active) setSnapshot(value => ({...value, error: error.message}));
    });
    reload.current = run;
    run();
    const timer = window.setInterval(run, 15000);
    return () => { active = false; window.clearInterval(timer); controller?.abort(); };
  }, [open, identity]);

  // A recreated Pod or Dogu must never inherit old events from an earlier UID.
  const sources = resources.map(node => snapshot.byResource[eventSourceKey(node)]).filter(Boolean);
  const errors = sources.map(source => source.error).filter(Boolean);
  const items = sources.flatMap(source => source.items);
  const visible = recentEvents(items, {warningsOnly});
  const total = recentEvents(items, {warningsOnly, limit: Infinity}).length;
  const warnings = recentEvents(items, {limit: Infinity}).length;
  const pending = resources.some(node => !snapshot.byResource[eventSourceKey(node)]);
  const refreshed = snapshot.time && !pending;
  return h("details", {className: "ces-dogu-tab__events", "aria-label": `Ereignisse für ${doguName}`, open, onToggle: event => setOpen(event.currentTarget.open)},
    h("summary", null, `Warnungen & Ereignisse${refreshed ? ` (${warnings} Warnungen)` : ""}`),
    open && h("div", null,
      h("p", {className: "ces-dogu-tab__muted"}, "Dogu und zugeordnete Workloads/PVCs · letzte 24 Stunden · neueste zuerst, maximal 20 Einträge. Kubernetes bewahrt Ereignisse nur begrenzt auf; keine Ereignisse bedeuten nicht automatisch einen gesunden Zustand."),
      h("div", {className: "ces-dogu-tab__toolbar"},
        h("label", null, h("input", {type: "checkbox", checked: warningsOnly, onChange: event => setWarningsOnly(event.target.checked)}), " Nur Warnungen"),
        h("button", {type: "button", className: "argo-button argo-button--base", disabled: loading || resources.length === 0, onClick: () => reload.current()}, loading ? "Ereignisse werden geladen …" : "Ereignisse aktualisieren")),
      resources.length === 0 && h("p", null, "Keine Ressource mit UID im ArgoCD-Baum verfügbar; Ereignisse können nicht sicher zugeordnet werden."),
      (errors.length > 0 || snapshot.error) && h("div", {className: "ces-dogu-tab__notice", role: "alert"},
        h("strong", null, "Ereignisdaten sind unvollständig oder veraltet."),
        errors.map(message => h("p", {key: message}, message)), snapshot.error && h("p", null, snapshot.error)),
      visible.length > 0 && h("ol", {className: "ces-dogu-tab__event-list", "aria-label": `Ereignisliste für ${doguName}`}, visible.map(item => h("li", {key: item.id},
        h("div", null, h("strong", {className: item.type === "Warning" ? "ces-dogu-tab__error" : undefined}, `${item.type} · ${item.reason}`), ` · ${item.resource} · ${item.count}×`),
        h("div", {className: "ces-dogu-tab__muted"}, item.timestamp == null ? "Zeitpunkt unbekannt" : new Date(item.timestamp).toLocaleString("de-DE")),
        h("p", null, item.message)))),
      refreshed && !loading && total === 0 && errors.length === 0 && !snapshot.error && h("p", null, warningsOnly ? "Keine aufbewahrten Warnungen im ausgewählten Zeitraum gefunden. Normale Ereignisse lassen sich über den Filter einblenden." : "Keine aufbewahrten Ereignisse im ausgewählten Zeitraum gefunden."),
      total > EVENT_LIMIT && h("p", {className: "ces-dogu-tab__muted"}, `${EVENT_LIMIT} von ${total} Ereignissen angezeigt.`),
      refreshed && h("p", {className: "ces-dogu-tab__muted", "aria-live": "polite"}, `Letzte Ereignisabfrage: ${snapshot.time.toLocaleTimeString("de-DE")} · Aktualisierung alle 15 Sekunden, solange geöffnet.`)));
}
