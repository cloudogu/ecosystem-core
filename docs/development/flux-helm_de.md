## Flux als Komponente für Helm

Wir haben entschieden, dass wir Flux für die Ausbringung von Helm-Charts verwenden möchten. Dies betrifft
vor allem den Dogu-Operator für die Ausbringung von Dogus in V3. Um Flux auch anderen Komponenten zur Verfügung zu stellen,
wird Flux als zentraler Bestandteil der Plattform über den Ecosystem-Core verwaltet. Hierfür wurde das [Community-Chart](https://github.com/fluxcd-community/helm-charts)
verwendet und für die LOP angepasst. Flux kann über das Flag `flux.enabled` für die LOP aktiviert werden.

### Flux wird intern für Helm genutzt

Flux soll grundsätzlich für die Ausbringung von Helm-Charts verwendet werden. Hierfür wird der [Source-Controller](https://fluxcd.io/flux/components/source/) für 
die Synchronisation der Helm-Charts sowie der [Helm-Controller](https://fluxcd.io/flux/components/helm/) für die Ausbringung
im Cluster benötigt. Alle weiteren Controller wurden für den momentanen Anwendungszweck deaktiviert.

### CRDs

Die CRDs von Flux werden über das Helm-Chart installiert bzw. aktualisiert und sind somit an den Lebenszyklus des Helm-Charts
gekoppelt. Wird das Chart gelöscht, sei es aufgrund einer Migration oder eines Fehlers, werden damit auch die CRDs mit ihren CRs gelöscht
und damit auch alle Dogus. Um diesen Umstand zu vermeiden, werden die CRDs mit der Annotation `"helm.sh/resource-policy": keep`
installiert. Mithilfe der Annotation bleiben Ressourcen bestehen, selbst wenn ihr dazugehöriges Helm-Chart deinstalliert wird.

### Sharding
Flux unterstützt nativ [Sharding](https://fluxcd.io/flux/installation/configuration/sharding/) wodurch es möglich ist 
mehrere Flux-Operatoren/Controller parallel im Cluster betreiben zu können. Mit aktivierten Sharding überwachen die jeweiligen
Controller nur die für sie relevanten Ressourcen. Sowohl der Source- als auch Helm-Controller sind für den Sharding-Key
`sharding.fluxcd.io/key=ces` konfiguriert. D.h. es werden nur Ressourcen überwacht, die als Label den Key `sharding.fluxcd.io/key=ces`
tragen. 

Mit Nutzung eines eigenen Sharding-Keys ermöglichen wir es Kunden weiterhin ihre eigenen Flux-Komponenten zu nutzen, ohne Seiteneffekte
mit unserer Plattform zu haben. 

### Begrenzung des Namespaces

Per Default werden von Flux alle Namespaces des Clusters überwacht. Da wir aktuell nur auf einem Namespace operieren, wurde
der Namespace auf den Namespace des Ecosystem-Core begrenzt, in den auch Flux ausgebracht wird.

### Monitoring

Grundsätzlich bietet Flux die Möglichkeit Prometheus mithilfe eines `PodMonitor` anzubinden und Metriken der einzelne Controller
zu exportieren. Dies setzt voraus, dass die `PodMonitor`-CRD im Cluster existiert, was zum Zeitpunkt der Installation des Ecosystem-Core
nicht gegeben ist. Prometheus selbst wird bei uns erst später als Komponente mit ihren CRDs installiert. Aus diesem Grund 
wird das Monitoring zum jetzigen Zeitpunkt **deaktiviert**. 

### NetworkPolicies

Das Helm-Chart von Flux ist darauf ausgelegt in einem separaten Namespace ausgebracht zu werden. Unter Berücksichtigung 
dieser Annahme sind die NetworkPolices, die das Helm-Chart mit ausbringt sehr "breit" aufgestellt und **adressieren alle
Pods** des Namespaces, was bei uns zu Seiteneffekten führe würde. Aus diesem Grund wurden die NetworkPolicies des Charts **deaktiviert**.

Da der Helm-Controller jedoch Zugriff auf Source-Controller braucht, um das Helm-Chart lokal zu laden, bringt der Ecosystem-Core
hierfür eine eigene NetworkPolicy `flux-source-controller-allow-artifact-access` (`flux-network-policies.yaml`) aus.
Neben dem Helm-Controller ist darin auch der Dogu-Operator freigeschaltet, da dieser die Artefakte ebenfalls direkt abruft.

### Retries für fehlgeschlagene Helm-Operationen

Treten Fehler während einer Installation oder Upgrade im Helm-Controller auf, greift die konfigurierte Strategie des `HelmRelease`.
Diese ist standardmäßig auf eine Remediate (`RemediateOnFailure`) mit einem `rollback` zwischen den Retries ausgelegt. Sind die Retries
ausgeschöpft, bleibt das `HelmRelease` im fehlerhaften Zustand und kann ohne manuelles Eingreifen nicht repariert werden. 

Um ein Self-Healing zu ermöglichen wird im Helm-Controller das Feature-Gate `DefaultToRetryOnFailure` aktiviert. Mit dem 
aktivierten Feature wird im `HelmRelease` definierten Intervall `.retryInterval` ein stetiger Retry ausgeführt. Das Intervall ist
per Default auf **5 Minuten** gesetzt. 
