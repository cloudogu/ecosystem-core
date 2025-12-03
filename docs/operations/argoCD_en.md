# Installation via ArgoCD (`ecosystem-core`)

In addition to manually installing `ecosystem-core` using the Helm chart, ArgoCD can also be used as a GitOps tool
for installation.

## Requirements
- Running Kubernetes cluster
- `kubectl` CLI configured and connected to the cluster
- Cluster admin rights

### ArgoCD instance
If no ArgoCD instance is already running in the cluster, it can be installed as follows:

```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

To access the ArgoCD dashboard, a separate ingress resource must be set up for port 443.
Alternatively, temporary access can be set up via port forwarding, which makes the ArgoCD dashboard accessible via
`https://localhost:8080`.

```bash
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

The password created for the initial `admin` user needs to be read:

```bash
kubectl -n argocd get secret argocd-initial-admin-secret   -o jsonpath="{.data.password}" | base64 -d && echo
```

### Secrets
In order for ArgoCD to access the Helm charts, a secret for the Cloudogu Helm registry is required. It can be deployed via kubectl as follows:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cloudogu-oci-registry-k8s
  namespace: argocd
  labels:
    argocd.argoproj.io/secret-type: repository
stringData:
  type: "helm"
  name: "Cloudogu-Registry"
  url: "registry.cloudogu.com/k8s"
  enableOCI: "true"
  username: "USERNAME"
  password: "TOKEN"
```

If the repository for the GitOps resources is private, a separate secret must be created for the repository:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: scm-repo-secret
  namespace: argocd
  labels:
    # Tells Argo CD this secret defines a repository
    argocd.argoproj.io/secret-type: repository
type: Opaque
stringData:
  url: "<Link to repository>"
  username: "my_username"
  password: "API_TOKEN"
```

## Deployment

The complete CES-MN can be deployed via ArgoCD using several so-called sync waves in the form of applications:

- **k8s-component-operator-crd** (sync wave -1): installs the Component CRD as a prerequisite for the Component Operator
- **ecosystem-core** (sync wave 0): installs the Component Operator with all necessary components and creates a default configuration
- **blueprint** (sync wave 1): installs a blueprint with a user-specific configuration and all desired Dogus
  
In addition, further sync waves with priority -1 could be used to provide certificates or secrets in the cluster.

### k8s-component-operator-crd

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: k8s-component-operator-crd
  namespace: argocd
  annotations:
    argocd.argoproj.io/sync-wave: "-1"
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: default
  source:
    repoURL: registry.cloudogu.com/k8s
    chart: k8s-component-operator-crd
    targetRevision: "1.10.0"
    path: "."                   # required for Helm OCI sources
  destination:
    server: https://kubernetes.default.svc
    namespace: ecosystem
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
      - PruneLast=true          # safer deletion
```

### ecosystem-core

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: ecosystem-core
  namespace: argocd
  annotations:
    argocd.argoproj.io/sync-wave: "0"
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: default
  source:
    repoURL: registry.cloudogu.com/k8s
    chart: ecosystem-core
    targetRevision: "0.2.2"
    path: "."  # required for Helm OCI sources
    helm:
      valuesObject:
        skipPreconditionValidation: true
  destination:
    server: https://kubernetes.default.svc
    namespace: ecosystem
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
      - PruneLast=true          # safer deletion
```

### blueprint

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: blueprint
  namespace: argocd
  annotations:
    argocd.argoproj.io/sync-wave: "1"
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: default
  source:
    repoURL: registry.cloudogu.com/k8s
    chart: k8s-blueprint
    targetRevision: "1.1.1"
    path: "."  # required for Helm OCI sources
    # You can override the default configuration here.
    # See https://github.com/cloudogu/k8s-blueprint-helm/blob/develop/k8s/helm/values.yaml for all available options.
    #helm:
    #  valuesObject:
    #    spec:
    #      blueprint:
    #        official/cas: "7.2.7-5"
  destination:
    server: https://kubernetes.default.svc
    namespace: ecosystem
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
      - PruneLast=true          # safer deletion
```

### CES-MN

All of the above applications and the blueprint are deployed together via a higher-level application, which will be referred to as `ces-mn` in the following:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: "ces-mn"
  namespace: argocd
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: default
  source:
    repoURL: https://example.local/repo/ces
    targetRevision: main               # branch, tag, or commit
    path: apps/ces-mn                  # folder with manifests
    directory:
      recurse: true                    # pick up nested folders
  destination:
    server: https://kubernetes.default.svc
    namespace: ecosystem
  syncPolicy:
    automated:
      prune: true                      # delete drifted resources
      selfHeal: true                   # fix out-of-band changes
```

All components of the application can be found in the repository `repoURL` in the branch/tag/commit `targetRevision` under the path `path`.
