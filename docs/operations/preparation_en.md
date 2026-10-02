# Preparing to install `ecosystem-core`

To successfully install the Helm chart **`ecosystem-core`**, various Kubernetes secrets and config maps must be created.
These contain the access data for Dogu, container, and Helm registries.

## Prerequisites

- The "Component" CustomResourceDefinition (CRD) must be installed in the cluster. This is required by the
  `k8s-component-operator` to manage component objects.
- Access to the Kubernetes cluster (`kubectl` must be configured)
- A set Kubernetes namespace (`$NAMESPACE`)
- Access data for the registries (username, password, email if necessary)

### Component CRD

In order to create component CRs, the corresponding CustomResourceDefinition (CRD) must already be registered in the
cluster. Install the CRD using the published Helm chart from the OCI repository.

```bash
helm upgrade --install k8s-component-operator-crd \
  oci://registry.cloudogu.com/k8s/k8s-component-operator-crd \
  --version 1.10.0 \
  --namespace <namespace>
```

Verify the installation:

```bash
kubectl get crd components.k8s.cloudogu.com
```

The output should show the CRD `components.k8s.cloudogu.com`.

### Dogu Registry Secret

This secret contains the access data for the **Dogu Registry**.

```bash
kubectl create secret generic k8s-dogu-operator-dogu-registry \
  --from-literal=endpoint="https://dogu.cloudogu.com/api/v2/dogus" \
  --from-literal=urlschema="default" \
  --from-literal=username="${DOGU_REGISTRY_USERNAME}" \
  --from-literal=password="${DOGU_REGISTRY_PASSWORD}" \
  --namespace="${NAMESPACE}"
```

| Field         | Description                                                                                                                                                           |
|---------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **endpoint**  | The complete URL of the Dogu registry endpoint. Example: `https://dogu.cloudogu.com/api/v2/dogus`. The operator uses this endpoint to retrieve information and Dogus. |
| **urlschema** | Specifies the schema used for the registry. Usually, `default` is used here. For file-based Dogu registries (e.g., Nexus), `index` must be used.                      |
| **username**  | The username for authentication at the registry.                                                                                                                      |
| **password**  | The user's password, matching the `username` specified above. The operator uses this access data to authenticate to the registry.                                     |
| **namespace** | The Kubernetes namespace in which the secret is created. The secret is then only available in this namespace.                                                         |

### Container Registry Secret

This secret contains the access data for the **container registry** in Docker registry format.

```bash
kubectl create secret docker-registry ces-container-registries \
  --docker-server="registry.cloudogu.com" \
  --docker-username="${DOCKER_REGISTRY_USERNAME}" \
  --docker-password="${DOCKER_REGISTRY_PASSWORD}" \
  --docker-email="${DOCKER_REGISTRY_EMAIL}" \
  --namespace="${NAMESPACE}"
```

| Field                 | Description                                                                                                                   |
|-----------------------|-------------------------------------------------------------------------------------------------------------------------------|
| **--docker-server**   | The URL of the container registry. Example: `registry.cloudogu.com`. This is where Kubernetes retrieves the container images. |
| **--docker-username** | The username for authentication at the registry.                                                                              |
| **--docker-password** | The password for the user specified above. Kubernetes uses this credential to authenticate to the registry.                   |
| **--docker-email**    | An email address associated with the registry account. Some registries require this field for authentication purposes.        |
| **--namespace**       | The Kubernetes namespace in which the secret is created.                                                                      |

### Helm Registry ConfigMap & Secret

In addition to authentication, a ConfigMap and a secret must be created for the **Helm registry**.

#### ConfigMap

```bash
kubectl create configmap component-operator-helm-repository \
  --from-literal=endpoint="registry.cloudogu.com" \
  --from-literal=schema="oci" \
  --from-literal=plainHttp="false" \
  --from-literal=insecureTls="false" \
  --namespace="${NAMESPACE}"
```

| Field           | Description                                                                                                                               |
|-----------------|-------------------------------------------------------------------------------------------------------------------------------------------|
| **endpoint**    | Hostname or address of the Helm registry. Example: `registry.cloudogu.com`.                                                               |
| **schema**      | The protocol/schema used to communicate with the registry. Typical values: `oci` (for OCI-compliant Helm repositories) or `https`.        |
| **plainHttp**   | Specifies whether unencrypted HTTP connections are allowed. Default: `false` (HTTPS is used).                                             |
| **insecureTls** | Determines whether insecure TLS certificates should be accepted. Default: `false`. If `true`, self-signed certificates are also accepted. |
| **namespace**   | The Kubernetes namespace in which the ConfigMap is created. The Component Operator can only access the ConfigMap within this namespace.   |

#### Secret

```bash
kubectl create secret generic component-operator-helm-registry \
  --from-literal=config.json='{"auths": {"'registry.cloudogu.com'": {"auth": "'$(
    echo -n "${HELM_REGISTRY_USERNAME}:${HELM_REGISTRY_PASSWORD}" | base64
  )'"}}}' \
  --namespace="${NAMESPACE}"
```

| Field                     | Description                                                                                                                    |
|---------------------------|--------------------------------------------------------------------------------------------------------------------------------|
| **auths**                 | Object containing the authentication information for one or more registries.                                                   |
| **registry.cloudogu.com** | Hostname of the Helm registry to which the credentials apply.                                                                  |
| **auth**                  | Base64-encoded string of `username:password`. Example: `ZGVtbzpwYXNzd29ydA==` corresponds to `demo:password`.                  |
| **namespace**             | The Kubernetes namespace in which the secret is created. The component operator can only use the secret within this namespace. |

### Certificate

Communication with web applications operated via the ecosystem is always encrypted using TLS. This requires a
corresponding TLS certificate, which is stored centrally in the cluster. If no certificate is stored in the cluster, a
self-signed certificate is generated and provided.

#### Provisioning an external certificate

If you want to use your own external certificate for CES-MN, it should be provided in the cluster before installing the
`ecosystem-core`
and should comply with
the [Kubernetes specifications](https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets). The certificate
must be created as a secret with the name `ecosystem-certificate` in the corresponding Dogu namespace:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ecosystem-certificate
  namespace: ecosystem
type: kubernetes.io/tls
data:
  # values are base64 encoded, which obscures them but does NOT provide
  # any useful level of confidentiality
  # Replace the following values with your own base64-encoded certificate and key.
  tls.crt: "REPLACE_WITH_BASE64_CERT"
  tls.key: "REPLACE_WITH_BASE64_KEY"
```

## Dogu API v3 via Flux

[Flux](https://fluxcd.io/) provides operators and associated CRDs capable of automatically managing Helm releases.
The [Dogu Operator](https://github.com/cloudogu/k8s-dogu-operator/) utilizes these to install Dogus in accordance with
the Dogu API v3.

Here, `flux` configures the installation routine, while `flux2` contains the configuration for the actual sub-chart.

Additional values (which are not provided by `ecosystem-core` for standard operation within the Cloudogu EcoSystem) are
listed in the [Flux Helm chart description](https://github.com/fluxcd-community/helm-charts/tree/main/charts/flux2).

```yaml
flux:
  enabled: true
flux2:
  installCRDs: true
  crds:
    annotations:
      "helm.sh/resource-policy": keep
  helmController:
    create: true
    container:
      additionalArgs:
        - "--feature-gates=DefaultToRetryOnFailure=true"
        - "--watch-label-selector=sharding.fluxcd.io/key=ces"
  imageAutomationController:
    create: false
  imageReflectionController:
    create: false
  kustomizeController:
    create: false
  notificationController:
    create: false
  sourceController:
    create: true
    container:
      additionalArgs:
        - "--watch-label-selector=sharding.fluxcd.io/key=ces"
  sourceWatcher:
    create: false
  policies:
    create: false
  watchAllNamespaces: false
  imagePullSecrets:
    - name: ces-container-registries
  prometheus:
    podMonitor:
      create: false
```

| Field                                             | Type             | Description                                                                                                                                                                                       |
|---------------------------------------------------|------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `flux.enabled`                                    | `bool`           | Controls the installation of Flux for Dogu v3 in general. Default `false`.                                                                                                                        |
| `flux2.installCRDs`                               | `bool`           | Controls the automatic installation of the required Flux CRDs. Default `true`.                                                                                                                    |
| `flux2.crds.annotations.helm.sh/resource-policy`  | `string`         | Controls the Helm resource deletion policy. The value `keep` allows CRDs – and thus dogus – to persist in the cluster even if Flux is (unintentionally) deleted. Default `keep`.                  |
| `flux2.helmController.container.additionalArgs`   | List of `string` | Allows adding extra arguments to the Flux HelmRelease Controller. Defaults: <br/>`- "--feature-gates=DefaultToRetryOnFailure=true"` <br/> `- "--watch-label-selector=sharding.fluxcd.io/key=ces"` |
| `flux2.imageAutomationController.create`          | `bool`           | Controls the installation of the Flux Image Automation Controller. Default `false`.                                                                                                               |
| `flux2.imageReflectionController.create`          | `bool`           | Controls the installation of the Flux Image Reflection Controller. Default `false`.                                                                                                               |
| `flux2.kustomizeController.create`                | `bool`           | Controls the installation of the Flux Kustomize Controller. Default `false`.                                                                                                                      |
| `flux2.notificationController.create`             | `bool`           | Controls the installation of the Flux Notification Controller. Default `false`.                                                                                                                   |
| `flux2.sourceController.create`                   | `bool`           | Controls the installation of the Flux Source Controller. Default `true`.                                                                                                                          |
| `flux2.sourceController.container.additionalArgs` | `string`         | Allows the source controller to be augmented with additional arguments. Default: <br/>`- "--watch-label-selector=sharding.fluxcd.io/key=ces"`                                                     |
| `flux2.sourceWatcher.create`                      | `bool`           | Controls the installation of the Flux source-watcher controller. Default `false`.                                                                                                                 |
| `flux2.policies.create`                           | `bool`           | Controls the installation of ingress and egress NetworkPolicies for Flux. Default `false`.                                                                                                        |
| `flux2.watchAllNamespaces`                        | `bool`           | Controls whether Flux operators should watch all cluster namespaces (instead of just one). Default `false`.                                                                                       |
| `flux2.imagePullSecrets[].name`                   | `string`         | Allows for the specification of credentials for various OCI registries. Default `ces-container-registries`.                                                                                       |
| `flux2.prometheus.podMonitor.create`              | `bool`           | Enables the Prometheus podMonitor endpoint. Default `false`.                                                                                                                                      |                                      