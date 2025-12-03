# Installation via ArgoCD (`ecosystem-core`)

Neben der manuellen Installation des `ecosystem-core` mit Hilfe des Helm-Charts, kann auch ArgoCD als GitOps-Tool 
für die Installation verwendet werden.

## Voraussetzungen
- Laufender Kubernetes-Cluster
- `kubectl` CLI konfiguriert und mit dem Cluster verbunden
- Cluster-Admin-Rechte

### ArgoCD-Instanz
Falls keine ArgoCD-Instanz bereits im Cluster läuft, kann diese wie folgt installiert werden:

```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

Für den Zugriff auf das ArgoCD-Dashboard muss eine separate Ingress-Ressource für den Port 443 eingerichtet werden. 
Alternativ kann ein Zugriff temporär über ein Port-Forwarding eingerichtet werden, bei dem das ArgoCD-Dashboard anschließend über 
`https://localhost:8080` erreichbar ist.

```bash
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

Für den initialen `admin` User muss das erstellte Password ausgelesen werden:

```bash
kubectl -n argocd get secret argocd-initial-admin-secret   -o jsonpath="{.data.password}" | base64 -d && echo
```

### Secrets
Damit ArgoCD auf die Helm-Charts zugreifen kann, wird ein Secret für die Cloudogu Helm-Registry benötigt. Über kubectl lässt es sich wie folgt
ausbringen:

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

Ist das Repository für die GitOps-Ressourcen privat, muss hierfür separat ein Secret für das Repository angelegt werden:

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

Das komplette CES-MN kann mittels ArgoCD über mehrere sogenannte Sync-Waves in Form von Applikationen ausgebracht werden:

- **k8s-component-operator-crd** (sync-wave -1): installiert die Component CRD als Voraussetzung für den Component-Operator
- **ecosystem-core** (sync-wave 0): installiert den Component-Operator mit allen notwendigen Komponenten und erstellt eine Default-Konfiguration
- **blueprint** (sync-wave 1): installiert ein Blueprint mit einer nutzerspezfischen Konfiguration sowie alle  gewünschten Dogus

Darüber hinaus könnten weitere Sync-Waves mit Priorität -1 verwendet werden, um Zertifkate oder Secrets im Cluster bereitzustellen.

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

Alle oben genannten Applikationen bzw. das Blueprint werden zusammen über eine übergeordnete Applikation ausgebracht, die im folgenden
als `ces-mn` betitelt wird: 

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

Alle Bestandteile der Applikation finden sich im Repositry `repoURL` im Branch/Tag/Commit `targetRevision` unter dem Pfad `path`.
