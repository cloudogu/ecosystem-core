# Preparing to install `ecosystem-core`

To successfully install the Helm chart **`ecosystem-core`**, various Kubernetes secrets and config maps must be created.
These contain the access data for Dogu, container, and Helm registries.

## Prerequisites

- Access to the Kubernetes cluster (`kubectl` must be configured)
- A set Kubernetes namespace (`$NAMESPACE`)
- Access data for the registries (username, password, email if necessary)

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
| ------------- |-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
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
| --------------------- |-------------------------------------------------------------------------------------------------------------------------------|
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
  --from-literal=insecureTls="false"  \
  --namespace="${NAMESPACE}"
```

| Field           | Description                                                                                                                               |
| --------------- |-------------------------------------------------------------------------------------------------------------------------------------------|
| **endpoint**    | Hostname or address of the Helm registry. Example: `registry.cloudogu.com`.                                                               |
| **schema**      | The protocol/schema used to communicate with the registry. Typical values: `oci` (for OCI-compliant Helm repositories) or `https`.        |
| **plainHttp**   | Specifies whether unencrypted HTTP connections are allowed. Default: `false` (HTTPS is used).                                             |
| **insecureTls** | Determines whether insecure TLS certificates should be accepted. Default: `false`. If `true`, self-signed certificates are also accepted. |
| **namespace**   | The Kubernetes namespace in which the ConfigMap is created. The Component Operator can only access the ConfigMap within this namespace.   |


#### Secret

```bash
kubectl create secret generic component-operator-helm-registry \
  --from-literal=config.json='{"auths": {"'registry.cloudogu.com'": {"auth": "'$(echo -n "${HELM_REGISTRY_USERNAME}:${HELM_REGISTRY_PASSWORD}" | base64)'"}}}' \
  --namespace="${NAMESPACE}"
```

| Field                     | Description                                                                                                                    |
| ------------------------- |--------------------------------------------------------------------------------------------------------------------------------|
| **auths**                 | Object containing the authentication information for one or more registries.                                                   |
| **registry.cloudogu.com** | Hostname of the Helm registry to which the credentials apply.                                                                  |
| **auth**                  | Base64-encoded string of `username:password`. Example: `ZGVtbzpwYXNzd29ydA==` corresponds to `demo:password`.                  |
| **namespace**             | The Kubernetes namespace in which the secret is created. The component operator can only use the secret within this namespace. |


---

## Summary

- **Dogu Registry Secret** → `k8s-dogu-operator-dogu-registry`
- **Container Registry Secret** → `ces-container-registries`
- **Helm Registry ConfigMap & Secret** → `component-operator-helm-repository`, `component-operator-helm-registry`

These resources must be created in the desired namespace **before installing `ecosystem-core`**.