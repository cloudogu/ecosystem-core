# Vorbereitung für die Installation von `ecosystem-core`

Um das Helm-Chart **`ecosystem-core`** erfolgreich zu installieren, müssen verschiedene Kubernetes-Secrets und ConfigMaps erstellt werden. 
Diese enthalten die Zugangsdaten zu Dogu-, Container- und Helm-Registries.

## Voraussetzungen

- Die "Component"-CustomResourceDefinition (CRD) muss im Cluster installiert sein.
  Diese wird vom `k8s-component-operator` benötigt, um Komponenten-Objekte zu verwalten.
- Zugriff auf das Kubernetes-Cluster (`kubectl` muss konfiguriert sein)
- Ein gesetztes Kubernetes-Namespace (`$NAMESPACE`)
- Zugangsdaten zu den Registries (Benutzername, Passwort, ggf. E-Mail)
- Bereistellung eines TLS-Zertifikats (falls gewünscht)

### Component-CRD

Damit Component-CRs angelegt werden können, muss die zugehörige CustomResourceDefinition (CRD) bereits im Cluster registriert sein.
Installieren Sie die CRD über das veröffentlichte Helm-Chart aus dem OCI-Repository.
```bash
helm upgrade --install k8s-component-operator-crd \
  oci://registry.cloudogu.com/k8s/k8s-component-operator-crd \
  --version 1.10.0 \
  --namespace <namespace>
```

Verifizieren Sie die Installation:
```bash
kubectl get crd components.k8s.cloudogu.com
```
Die Ausgabe sollte die CRD `components.k8s.cloudogu.com` zeigen.

### Dogu-Registry Secret

Dieses Secret enthält die Zugangsdaten zur **Dogu-Registry**.

```bash
kubectl create secret generic k8s-dogu-operator-dogu-registry \
  --from-literal=endpoint="https://dogu.cloudogu.com/api/v2/dogus" \
  --from-literal=urlschema="default" \
  --from-literal=username="${DOGU_REGISTRY_USERNAME}" \
  --from-literal=password="${DOGU_REGISTRY_PASSWORD}" \
  --namespace="${NAMESPACE}"
```

| Feld          | Beschreibung                                                                                                                                                            |
| ------------- |-------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **endpoint**  | Die vollständige URL des Dogu-Registry-Endpunkts. Beispiel: `https://dogu.cloudogu.com/api/v2/dogus`. Über diesen Endpunkt ruft der Operator Informationen und Dogus ab. |
| **urlschema** | Gibt das verwendete Schema für die Registry an. In der Regel wird hier `default` genutzt. Für dateibasierte Dogu-Registries (z.B. Nexus) muss `index` verwendet werden. |
| **username**  | Der Benutzername für die Authentifizierung an der Registry.                                    |
| **password**  | Das Passwort des Benutzers, passend zum oben angegebenen `username`. Mit diesen Zugangsdaten authentifiziert sich der Operator an der Registry.                         |
| **namespace** | Das Kubernetes-Namespace, in dem das Secret erstellt wird. Das Secret steht dann nur in diesem Namespace zur Verfügung.                                                 |


### Container-Registry Secret

Dieses Secret enthält die Zugangsdaten zur **Container-Registry** im Docker-Registry-Format.

```bash
kubectl create secret docker-registry ces-container-registries \
  --docker-server="registry.cloudogu.com" \
  --docker-username="${DOCKER_REGISTRY_USERNAME}" \
  --docker-password="${DOCKER_REGISTRY_PASSWORD}" \
  --docker-email="${DOCKER_REGISTRY_EMAIL}" \
  --namespace="${NAMESPACE}"
```

| Feld                  | Beschreibung                                                                                                                        |
| --------------------- |-------------------------------------------------------------------------------------------------------------------------------------|
| **--docker-server**   | Die URL der Container-Registry. Beispiel: `registry.cloudogu.com`. Hier ruft Kubernetes die Container-Images ab.                    |
| **--docker-username** | Der Benutzername für die Authentifizierung an der Registry.                                                                         |
| **--docker-password** | Das Passwort des oben angegebenen Benutzers. Mit diesen Zugangsdaten authentifiziert sich Kubernetes an der Registry.               |
| **--docker-email**    | Eine E-Mail-Adresse, die dem Registry-Account zugeordnet ist. Manche Registries benötigen dieses Feld für Authentifizierungszwecke. |
| **--namespace**       | Das Kubernetes-Namespace, in dem das Secret erstellt wird.                                                                          |


### Helm-Registry ConfigMap & Secret

Zusätzlich zur Authentifizierung muss eine ConfigMap und ein Secret für die **Helm-Registry** erstellt werden.

#### ConfigMap

```bash
kubectl create configmap component-operator-helm-repository \
  --from-literal=endpoint="registry.cloudogu.com" \
  --from-literal=schema="oci" \
  --from-literal=plainHttp="false" \
  --from-literal=insecureTls="false"  \
  --namespace="${NAMESPACE}"
```

| Feld            | Beschreibung                                                                                                                                              |
| --------------- |-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| **endpoint**    | Hostname oder Adresse der Helm-Registry. Beispiel: `registry.cloudogu.com`.                                                                               |
| **schema**      | Das Protokoll/Schema, das für die Kommunikation mit der Registry verwendet wird. Typische Werte: `oci` (für OCI-konforme Helm-Repositories) oder `https`. |
| **plainHttp**   | Gibt an, ob unverschlüsselte HTTP-Verbindungen erlaubt sind. Standard: `false` (es wird HTTPS genutzt).                                                   |
| **insecureTls** | Bestimmt, ob unsichere TLS-Zertifikate akzeptiert werden sollen. Standard: `false`. Wenn `true`, werden auch selbstsignierte Zertifikate akzeptiert.      |
| **namespace**   | Das Kubernetes-Namespace, in dem die ConfigMap erstellt wird. Der Component-Operator kann nur innerhalb dieses Namespaces auf die ConfigMap zugreifen.    |


#### Secret

```bash
kubectl create secret generic component-operator-helm-registry \
  --from-literal=config.json='{"auths": {"'registry.cloudogu.com'": {"auth": "'$(echo -n "${HELM_REGISTRY_USERNAME}:${HELM_REGISTRY_PASSWORD}" | base64)'"}}}' \
  --namespace="${NAMESPACE}"
```

| Feld                      | Beschreibung                                                                                                                                 |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| **auths**                 | Objekt, das die Authentifizierungsinformationen für eine oder mehrere Registries enthält.                                                    |
| **registry.cloudogu.com** | Hostname der Helm-Registry, für die die Zugangsdaten gelten.                                                                                 |
| **auth**                  | Base64-kodierte Zeichenkette aus `username:password`. Beispiel: `ZGVtbzpwYXNzd29ydA==` entspricht `demo:password`.                           |
| **namespace**             | Das Kubernetes-Namespace, in dem das Secret erstellt wird. Der Component-Operator kann das Secret nur innerhalb dieses Namespaces verwenden. |

### Zertifikat

Die Kommunikation mit den Web-Applikationen, die über das Ecosystem betrieben werden, ist grundsätzlich über TLS verschlüsselt.
Hierfür wird ein entsprechendes TLS-Zertifikat benötigt, dass im Cluster zentral hinterlegt wird. Ist kein Zertifikat im Cluster hinterlegt,
wird ein selbst-signiertes Zertifikat erzeugt und bereitgestellt.

#### Bereitstellung eines externen Zertifikats

Soll für CES-MN ein eigenes externes Zertifikat verwendet werden, sollte diese vor der Installation des `ecosystem-core`
im Cluster bereitgestellt werden und den [Vorgaben von Kubernetes](https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets) entsprechen. 
Das Zertifikat muss als Secret mit dem Namen `ecosystem-certificate` im entsprechenden Namespace der Dogus erstellt werden:

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

