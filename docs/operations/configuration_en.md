# Configuration (`ecosystem-core`)

Ecosystem-core is a Helm chart that installs the core components (operators) required to run
the [Cloudogu Ecosystem](https://platform.cloudogu.com/en/info/cloudogu-ecosystem/) on Kubernetes.
It works standalone or via GitOps tools such as [Argo CD](https://argoproj.github.io/cd/).

Configuration is done via the `values.yaml` file.

## Global settings

| Field                        | Type      | Description                                                                                                                |
|------------------------------|-----------|----------------------------------------------------------------------------------------------------------------------------|
| `skipPreconditionValidation` | `boolean` | Skips the [precondition check](./preparation_en.md) (e.g., in local development environments or ArgoCD). Default: `false`. |
| `loadbalancer-annotations`   | `object`  | Writes the provided key value pairs in the annotations of the loadbalancer service of the Ecosystems.                      |

## Component Operator Configuration (`k8s-component-operator`)

The **Component Operator** manages the installation and lifecycle management of the components.

### Example
```yaml
k8s-component-operator:
  manager:
    image:
      registry: registry.cloudogu.com
      repository: k8s-component-operator
      tag: 1.0.0
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

| Field                     | Type      | Description                                  |
|---------------------------|-----------|----------------------------------------------|
| `image.registry`          | `string`  | Registry for the operator image              |
| `image.repository`        | `string`  | Repository name                              |
| `image.tag`               | `string`  | Tag/version                                  |
| `env.logLevel`            | `string`  | Log level (`debug`, `info`, `warn`, `error`) |
| `resourceLimits.memory`   | `string`  | Maximum allowed memory consumption           |
| `resourceRequests.cpu`    | `string`  | Minimum requested CPU resources              |
| `resourceRequests.memory` | `string`  | Minimum requested memory                     |
| `networkPolicies.enabled` | `boolean` | Enables network policies (default: `true`)   |

## Components (`components`)

Individual components of the CES can be defined under `components`.
Each entry corresponds to a component and uses a standardized schema.

### Example
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
```

| Field             | Type      | Description                                                                                                     |
|-------------------|-----------|-----------------------------------------------------------------------------------------------------------------|
| `disabled`        | `boolean` | Deactivates the component (default: `false`)                                                                    |
| `version`         | `string`  | Component version (e.g., Docker or Helm tag). By specifying “latest” the newest available version will be used. |
| `helmNamespace`   | `string`  | Namespace used by the component for Helm operations (default: `k8s`)                                            |
| `deployNamespace` | `string`  | Target namespace where the component is installed (default: component namespace)                                |
| `mainLogLevel`    | `string`  | Log level for the component (`debug`, `info`, `warn`, `error`)                                                  |
| `valuesObject`    | `string`  | YAML block for overwriting default values                                                                       |

## Backup components (`backup`)

Enables and manages the **backup stack** and its components.

```yaml
backup:
  enabled: true
  components:
    ...
```

| Field        | Type      | Description                                             |
|--------------|-----------|---------------------------------------------------------|
| `enabled`    | `boolean` | Enables the backup stack                                |
| `components` | `map`     | List of backup components, structured like `components` |

## Monitoring components (`monitoring`)

Enables and manages the **monitoring stack** and its components.

```yaml
monitoring:
  enabled: true
  components:
    ...
```

| Field        | Type      | Description                                                 |
|--------------|-----------|-------------------------------------------------------------|
| `enabled`    | `boolean` | Enables the monitoring stack                                |
| `components` | `map`     | List of monitoring components, structured like `components` |

## Cleanup job (`cleanup`)

Before deletion (`helm uninstall`), a cleanup job is executed that deletes all components before the component operator
is deleted.

```yaml
cleanup:
  timeoutSeconds: 300
  image:
    registry: docker.io
    repository: alpine/kubectl
    tag: "1.34.0"
```

| Field              | Type      | Description                         |
|--------------------|-----------|-------------------------------------|
| `timeoutSeconds`   | `integer` | Maximum runtime in seconds          |
| `image.registry`   | `string`  | Registry for the cleanup image      |
| `image.repository` | `string`  | Repository  for the cleanup image   |
| `image.tag`        | `string`  | Tag/version   for the cleanup image |
