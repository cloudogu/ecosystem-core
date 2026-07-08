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
| `use-lop-idp`                | `boolean` | Enables the LOP-IDP stack. Default: `false`. See [LOP-IDP stack](#lop-idp-stack-use-lop-idp).                             |

## Component Operator Configuration (`k8s-component-operator`)

The **Component Operator** manages the installation and lifecycle management of the components.

Note: Configuring the `image` section is rarely needed as the component-operator image is defined in the `Chart.yaml` of ecosystem-core.
You can leave the `image` section out to use the defaults.

### Example
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
    valuesConfigRef:
      name: "configMapName"
      key: "configMapKey"    
```

| Field             | Type      | Description                                                                                                     |
|-------------------|-----------|-----------------------------------------------------------------------------------------------------------------|
| `disabled`        | `boolean` | Deactivates the component (default: `false`)                                                                    |
| `version`         | `string`  | Component version (e.g., Docker or Helm tag). By specifying “latest” the newest available version will be used. |
| `helmNamespace`   | `string`  | Namespace used by the component for Helm operations (default: `k8s`)                                            |
| `deployNamespace` | `string`  | Target namespace where the component is installed (default: component namespace)                                |
| `mainLogLevel`    | `string`  | Log level for the component (`debug`, `info`, `warn`, `error`)                                                  |
| `valuesObject`    | `string`  | YAML block for overwriting default values                                                                       |
| `valuesConfigRef` | `object`  | Specifies a reference to a ConfigMap and a key contained therein to override default values.                    |

## LOP-IDP stack (`use-lop-idp`)

Setting `use-lop-idp: true` activates the identity provider stack for authentication registration.
The following changes are applied automatically:

- Components `k8s-auth-registration-crd`, `lop-idp` and `postfix` are enabled.
- `k8s-dogu-operator` is configured with:
  ```yaml
  controllerManager:
    env:
      authRegistrationEnabled: true
      disablePostfixDependencyCheck: true
  ```
- `k8s-blueprint-operator` is configured with:
  ```yaml
  manager:
    env:
      authRegistrationEnabled: true
      disablePostfixDependencyCheck: true
  ```
- `k8s-service-discovery` is updated to version `6.1.0` and configured with:
  ```yaml
  exposition:
    discoverExpositionCR: true
  ```
  This translates the Exposition CRs created by the LOP-IDP sub-components into routes. Without this configuration, all paths become unreachable from outside the cluster after upgrading to `lop-idp` >= 1.2.0 (404).
- `k8s-exposition-crd` (1.0.0) is installed, providing the Exposition CRD that `k8s-service-discovery` reconciles when `discoverExpositionCR` is enabled.

When using `use-lop-idp`, the following additional values must be configured:

```yaml
defaultConfig:
  env:
    initialDomain: "your.domain.com"   # required: must be known at install time
    initialFQDN: "your.fqdn.com"       # required if enableFqdnApplier is false
```

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
Warning: If you disable the monitoring stack you also have to disable the k8s-ces-control component!

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

## Default config job (`defaultConfig`)

The default config job runs once after install and upgrade to write initial values into the CES global config and dogu configs.

```yaml
defaultConfig:
  env:
    enableFqdnApplier: false
    initialFQDN: ""
    initialDomain: ""
```

| Field                   | Type      | Description                                                                                                                            |
|-------------------------|-----------|----------------------------------------------------------------------------------------------------------------------------------------|
| `env.enableFqdnApplier` | `boolean` | Polls for the LoadBalancer IP and writes it as `fqdn` into the global config. Has no effect if `initialFQDN` is set. Default: `false`. |
| `env.initialFQDN`       | `string`  | Sets the initial `fqdn` in the global config. Takes precedence over `enableFqdnApplier`. Required when using `use-lop-idp`.            |
| `env.initialDomain`     | `string`  | Sets the initial `domain` in the global config. Required when using `use-lop-idp`.                                                     |

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
