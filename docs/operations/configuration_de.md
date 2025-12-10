# Konfiguration (`ecosystem-core`)

Ecosystem-Core ist ein Helm-Chart, das die Kernkomponenten (Operatoren) installiert, die für die Ausführung des [Cloudogu-Ecosystem](https://platform.cloudogu.com/en/info/cloudogu-ecosystem/) auf Kubernetes erforderlich sind.
Es funktioniert eigenständig oder über GitOps-Tools wie [Argo CD](https://argoproj.github.io/cd/).

Die Konfiguration erfolgt über die Datei `values.yaml`.

## Globale Einstellungen

| Feld                         | Typ       | Beschreibung                                                                                                                                 |
|------------------------------|-----------|----------------------------------------------------------------------------------------------------------------------------------------------|
| `skipPreconditionValidation` | `boolean` | Überspringt die [Prüfung von Voraussetzungen](./preparation_de.md) (z. B. in lokalen Entwicklungsumgebungen oder ArgoCD). Standard: `false`. |
| `loadbalancer-annotations`   | `object`  | Schreibt die übergebenen Key-Value Pairs als Annotation in den LoadBalancer-Service des Ecosystems.                                          |

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

Aktiviert und verwaltet den **Monitoring-Stacks** und dessen Komponenten.

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
