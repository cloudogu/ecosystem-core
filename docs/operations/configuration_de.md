# Konfiguration (`ecosystem-core`)

Ecosystem-Core ist ein Helm-Chart, das die Kernkomponenten (Operatoren) installiert, die für die Ausführung des [Cloudogu-Ecosystem](https://platform.cloudogu.com/en/info/cloudogu-ecosystem/) auf Kubernetes erforderlich sind.
Es funktioniert eigenständig oder über GitOps-Tools wie [Argo CD](https://argoproj.github.io/cd/).

Die Konfiguration erfolgt über die Datei `values.yaml`.

## Globale Einstellungen

| Feld                         | Typ       | Beschreibung                                                                                                                                 |
|------------------------------|-----------|----------------------------------------------------------------------------------------------------------------------------------------------|
| `skipPreconditionValidation` | `boolean` | Überspringt die [Prüfung von Voraussetzungen](./preparation_de.md) (z. B. in lokalen Entwicklungsumgebungen oder ArgoCD). Standard: `false`. |
| `loadbalancer-annotations`   | `object`  | Schreibt die übergebenen Key-Value Pairs als Annotation in den LoadBalancer-Service des Ecosystems.                                          |
| `use-lop-idp`                | `boolean` | Aktiviert den LOP-IDP-Stack. Standard: `false`. Siehe [LOP-IDP-Stack](#lop-idp-stack-use-lop-idp).                                          |

## Component-Operator-Konfiguration (`k8s-component-operator`)

Der **Component Operator** verwaltet die Installation und das Lifecycle-Management der Komponenten.

Anmerkung: Es ist selten nötig, den `image`-Bereich zu ändern, da das Component-Operator-Image durch die `Chart.yaml` von ecosystem-core festgelegt wird.
Der Bereich kann normalerweise auskommentiert bleiben.

### Beispiel
```yaml
k8s-component-operator:
  manager:
#    image:
#      registry: docker.io
#      repository: cloudogu/k8s-component-operator
#      tag: 1.12.0
    env:
      logLevel: info
    resourceLimits:
      memory: 512Mi
    resourceRequests:
      cpu: 100m
      memory: 256Mi
    networkPolicies:
      enabled: true
```

| Feld                      | Typ       | Beschreibung                                   |
|---------------------------|-----------|------------------------------------------------|
| `image.registry`          | `string`  | Registry für das Operator-Image                |
| `image.repository`        | `string`  | Repository-Name                                |
| `image.tag`               | `string`  | Tag/Version                                    |
| `env.logLevel`            | `string`  | Log-Level (`debug`, `info`, `warn`, `error`)   |
| `resourceLimits.memory`   | `string`  | Maximal erlaubter Speicherverbrauch            |
| `resourceRequests.cpu`    | `string`  | Minimal angeforderte CPU-Ressourcen            |
| `resourceRequests.memory` | `string`  | Minimal angeforderter Speicher                 |
| `networkPolicies.enabled` | `boolean` | Aktiviert Netzwerk-Policies (Standard: `true`) |

## Komponenten (`components`)

Unter `components` können einzelne Komponenten des CES definiert werden.  
Jeder Eintrag entspricht einer Komponente und nutzt ein standardisiertes Schema.

### Beispiel
```yaml
components:
  my-component:
    disabled: false
    version: "1.2.3"
    helmNamespace: "k8s"
    deployNamespace: "my-namespace"
    mainLogLevel: "info"
    valuesObject:
      replicaCount: 2
      service:
        name: "myService"
    valuesConfigRef:
      name: "configMapName"
      key: "configMapKey"    
```

| Feld              | Typ       | Beschreibung                                                                                                                  |
|-------------------|-----------|-------------------------------------------------------------------------------------------------------------------------------|
| `disabled`        | `boolean` | Deaktiviert die Komponente (Standard: `false`)                                                                                |
| `version`         | `string`  | Version der Komponente (z. B. Docker- oder Helm-Tag). Durch Angabe von "latest" wird die neuste verfügbare Version verwendet. |
| `helmNamespace`   | `string`  | Namespace, den die Komponente für Helm-Operationen nutzt (Standard: `k8s`)                                                    |
| `deployNamespace` | `string`  | Ziel-Namespace, in den die Komponente installiert wird (Standard: Namespace der Komponente)                                   |
| `mainLogLevel`    | `string`  | Log-Level für die Komponente (`debug`, `info`, `warn`, `error`)                                                               |
| `valuesObject`    | `object`  | YAML-Block zum Überschreiben von Standardwerten                                                                               |
| `valuesConfigRef` | `object`  | Angabe einer Referenz auf eine ConfigMap und einen darin enthaltenen Key zum Überschreiben von Standardwerten.                |

## LOP-IDP-Stack (`use-lop-idp`)

Mit `use-lop-idp: true` wird der Identity-Provider-Stack für die Authentifizierungsregistrierung aktiviert.
Folgende Änderungen werden automatisch vorgenommen:

- Die Komponenten `k8s-auth-registration-crd`, `lop-idp` und `postfix` werden aktiviert.
- `k8s-dogu-operator` wird konfiguriert mit:
  ```yaml
  controllerManager:
    env:
      authRegistrationEnabled: true
      disablePostfixDependencyCheck: true
  ```
- `k8s-blueprint-operator` wird konfiguriert mit:
  ```yaml
  manager:
    env:
      authRegistrationEnabled: true
      disablePostfixDependencyCheck: true
  ```
- `k8s-service-discovery` wird auf Version `6.1.0` aktualisiert und konfiguriert mit:
  ```yaml
  exposition:
    discoverExpositionCR: true
  ```
  Dadurch werden die von den LOP-IDP-Sub-Komponenten erzeugten Exposition-CRs in Routen übersetzt. Ohne diese Konfiguration sind nach einem Upgrade auf `lop-idp` >= 1.2.0 alle Pfade von außen nicht erreichbar (404).
- `k8s-exposition-crd` (1.0.0) wird installiert; es stellt die Exposition-CRD bereit, die `k8s-service-discovery` bei aktiviertem `discoverExpositionCR` verarbeitet.

Bei Verwendung von `use-lop-idp` müssen zusätzlich folgende Werte konfiguriert werden:

```yaml
defaultConfig:
  env:
    initialDomain: "your.domain.com"   # erforderlich: muss zur Installationszeit bekannt sein
    initialFQDN: "your.fqdn.com"       # erforderlich, wenn enableFqdnApplier false ist
```

## Backup-Komponenten (`backup`)

Aktiviert und verwaltet den **Backup-Stack** und dessen Komponenten.

```yaml
backup:
  enabled: true
  components:
    ...
```

| Feld         | Typ       | Beschreibung                                                |
|--------------|-----------|-------------------------------------------------------------|
| `enabled`    | `boolean` | Aktiviert den Backup-Stack                                  |
| `components` | `map`     | Liste der Backup-Komponenten, strukturiert wie `components` |

## Monitoring-Komponenten (`monitoring`)

Aktiviert und verwaltet den **Monitoring-Stack** und dessen Komponenten.
Achtung: Wenn der Monitoring-Stack deaktiviert wird, muss auch die k8s-ces-control-Komponente deaktiviert werden!

```yaml
monitoring:
  enabled: true
  components:
    ...
```

| Feld         | Typ       | Beschreibung                                                    |
|--------------|-----------|-----------------------------------------------------------------|
| `enabled`    | `boolean` | Aktiviert den Monitoring-Stack                                  |
| `components` | `map`     | Liste der Monitoring-Komponenten, strukturiert wie `components` |

## Default-Config-Job (`defaultConfig`)

Der Default-Config-Job läuft einmalig nach Installation und Upgrade und schreibt initiale Werte in die globale CES-Konfiguration und die Dogu-Konfigurationen.

```yaml
defaultConfig:
  env:
    enableFqdnApplier: false
    initialFQDN: ""
    initialDomain: ""
```

| Feld                    | Typ       | Beschreibung                                                                                                                                                      |
|-------------------------|-----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `env.enableFqdnApplier` | `boolean` | Wartet auf die LoadBalancer-IP und schreibt sie als `fqdn` in die globale Konfiguration. Hat keine Auswirkung, wenn `initialFQDN` gesetzt ist. Standard: `false`. |
| `env.initialFQDN`       | `string`  | Setzt die initiale `fqdn` in der globalen Konfiguration. Hat Vorrang vor `enableFqdnApplier`. Erforderlich bei Verwendung von `use-lop-idp`.                      |
| `env.initialDomain`     | `string`  | Setzt die initiale `domain` in der globalen Konfiguration. Erforderlich bei Verwendung von `use-lop-idp`.                                                         |

## Cleanup-Job (`cleanup`)

Vor dem Löschen (`helm uninstall`) wird ein Cleanup-Job ausgeführt, der alle Komponenten löscht bevor der Component-Operator gelöscht wird. 

```yaml
cleanup:
  timeoutSeconds: 300
  image:
    registry: docker.io
    repository: alpine/kubectl
    tag: "1.34.0"
```

| Feld               | Typ       | Beschreibung                        |
|--------------------|-----------|-------------------------------------|
| `timeoutSeconds`   | `integer` | Maximale Laufzeit in Sekunden       |
| `image.registry`   | `string`  | Registry für das Cleanup-Image      |
| `image.repository` | `string`  | Repository  für das Cleanup-Image   |
| `image.tag`        | `string`  | Tag/Version   für das Cleanup-Image |
