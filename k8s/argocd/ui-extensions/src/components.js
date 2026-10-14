const React = window.React;
const h = React.createElement;

export const styles = `
.ces-dogu-tab { padding: 20px; font: inherit; }
.ces-dogu-tab h3 { margin: 0 0 6px; font-size: 20px; }
.ces-dogu-tab h4 { margin: 24px 0 8px; font-size: 17px; }
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
.ces-dogu-tab details { margin-top: 12px; }
.ces-dogu-tab summary { cursor: pointer; margin-bottom: 12px; }
.ces-dogu-tab button { cursor: pointer; }
.ces-dogu-tab__card { padding: 16px; border: 1px solid rgba(128,128,128,.3); border-radius: 4px; }
.ces-dogu-tab__line { margin: 4px 0; }
`;

export function Badge({state, text = state}) {
  return h("span", {className: "ces-dogu-tab__badge"},
    h("span", {className: `ces-dogu-tab__dot ces-dogu-tab__dot--${state}`, "aria-hidden": true}), text);
}

function table(label, headers, rows) {
  return h("div", {className: "ces-dogu-tab__table-wrapper"}, h("table", {"aria-label": label},
    h("thead", null, h("tr", null, headers.map(title => h("th", {key: title, scope: "col"}, title)))), h("tbody", null, rows)));
}

function date(value) {
  const parsed = value && new Date(value);
  return parsed && !Number.isNaN(parsed.getTime()) ? parsed.toLocaleString("de-DE") : "Unbekannt";
}

export function BlueprintStatus({status}) {
  const labels = {Valid: "Gültig", Executable: "Ausführbar", EcosystemHealthy: "Dogu-Gesundheit bei Ausführung", Completed: "Abgeschlossen", LastApplySucceeded: "Letzte Anwendung erfolgreich"};
  return h("section", {className: "ces-dogu-tab__card", "aria-label": "Blueprint-Status"},
    h("h3", null, "Blueprint-Status"), h("strong", null, status.phase),
    h("p", {className: "ces-dogu-tab__muted"}, `Letzter Zustandswechsel der Anwendung: ${date(status.lastTransition)}`),
    h("p", null, `Anwendung gestoppt: ${status.stopped ? "Ja" : "Nein"} · Dogu-Health ignorieren: ${status.ignoreDoguHealth ? "Ja" : "Nein"}`),
    status.error && h("p", {className: "ces-dogu-tab__error"}, `Blueprint-Daten veraltet / nicht verfügbar: ${status.error}`),
    h("p", {className: "ces-dogu-tab__muted"}, "Dieser Status beschreibt die Blueprint-Ausführung. Die aktuelle Dogu-Gesundheit steht in der Tabelle darunter."),
    h("details", null, h("summary", null, `Bedingungen (${status.conditions.length})`),
      table("Blueprint-Bedingungen", ["Bedingung", "Zustand", "Meldung / Grund", "Zustandswechsel"], status.conditions.map(entry => h("tr", {key: entry.type},
        h("td", null, labels[entry.type] || entry.type),
        h("td", null, entry.stale ? "Veraltet" : entry.status === "True" ? "Ja" : entry.status === "False" ? "Nein" : "Unbekannt"),
        h("td", null, entry.message || entry.reason || "—"), h("td", null, date(entry.lastTransitionTime)))))));
}

function Versions({desired, installed, comparison, specVersion}) {
  return h("div", null,
    h("div", {className: "ces-dogu-tab__line"}, `Soll: ${desired}`),
    h("div", {className: "ces-dogu-tab__line"}, `Installiert: ${installed}`),
    specVersion && specVersion !== "—" && specVersion !== desired && h("div", {className: "ces-dogu-tab__line"}, `Ziel der Dogu-CR: ${specVersion}`),
    h(Badge, {state: comparison.state, text: comparison.text}));
}

export function DoguTable({rows, absent = false}) {
  return table(absent ? "Zur Entfernung vorgesehene Dogus" : "Dogu-Liste",
    ["Dogu", "Versionen", absent ? "Ressource" : "Health", "Betriebszustand", "Meldung"],
    rows.map(row => h("tr", {key: row.qualifiedName},
      h("td", null, h("strong", null, row.name), h("div", {className: "ces-dogu-tab__muted"}, row.qualifiedName)),
      h("td", null, absent ? `Installiert: ${row.installedVersion}` : h(Versions, {desired: row.version, installed: row.installedVersion, comparison: row.comparison, specVersion: row.specVersion})),
      h("td", null, absent ? (row.found ? "Noch vorhanden" : "Nicht im Ressourcenbaum") : h(Badge, {state: row.state})),
      h("td", null, row.operations.map(text => h("div", {key: text, className: "ces-dogu-tab__line"}, text)),
        row.replicas.map(replica => h("div", {key: replica.name, className: "ces-dogu-tab__muted"}, `Verfügbare Replikate: ${replica.text}`))),
      h("td", null, row.message, row.detailError && h("p", {className: "ces-dogu-tab__error"}, `Betriebs-/Versionsdaten veraltet / nicht verfügbar: ${row.detailError}`)))));
}

export function ComponentTable({rows, namespace}) {
  return h("section", {"aria-label": "CES-Komponenten"}, h("h4", null, `Komponenten (${rows.length})`),
    h("p", {className: "ces-dogu-tab__muted"}, `Komponenten im CES-Namespace ${namespace || "—"}, einschließlich anderer lesbarer Applications mit demselben Zielcluster und Namespace. Sie gehören nicht zur Dogu-Liste des Blueprints.`),
    rows.length ? table("Komponenten-Liste", ["Komponente", "Versionen", "Health", "Betriebszustand", "Application"], rows.map(row => h("tr", {key: row.name},
      h("td", null, row.name), h("td", null, h(Versions, {desired: row.desired, installed: row.installed, comparison: row.comparison})),
      h("td", null, h(Badge, {state: row.state}), h("div", {className: "ces-dogu-tab__muted"}, row.source), row.message && h("p", null, row.message)),
      h("td", null, row.status), h("td", null, row.application)))) : h("p", null, "Keine Komponentendaten in den lesbaren Ressourcenbäumen verfügbar."));
}
