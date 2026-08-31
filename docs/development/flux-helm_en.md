## Flux as a component for Helm

We decided that we want to use Flux for deploying Helm charts. This concerns mainly the Dogu operator 
for deploying Dogus in V3. In order to make Flux available to other components as well,
Flux is managed as a central part of the platform via the Ecosystem-Core. For this, the [community chart](https://github.com/fluxcd-community/helm-charts)
was used and adapted for the LOP. Flux can be enabled for the LOP via the `flux.enabled` flag.

### Flux is used internally for Helm

Flux should generally be used for deploying Helm charts. For this, the [Source Controller](https://fluxcd.io/flux/components/source/) is needed for
synchronizing the Helm charts, and the [Helm Controller](https://fluxcd.io/flux/components/helm/) is needed for deployment
in the cluster. All other controllers have been disabled for the current use case.

### CRDs

Flux's CRDs are installed and updated via the Helm chart and are thus coupled to the lifecycle of the Helm chart.
If the chart is deleted, whether due to a migration or an error, the CRDs and their CRs are deleted along with it,
and therefore all Dogus as well. To avoid this situation, the CRDs are installed with the annotation `"helm.sh/resource-policy": keep`.
With the help of this annotation, resources are retained even if their associated Helm chart is uninstalled.

### Sharding
Flux natively supports [sharding](https://fluxcd.io/flux/installation/configuration/sharding/), which makes it possible
to run multiple Flux operators/controllers in parallel in the cluster. With sharding enabled, the respective
controllers only watch the resources relevant to them. Both the Source Controller and the Helm Controller are configured for the sharding key
`sharding.fluxcd.io/key=ces`. This means that only resources are watched that carry the key `sharding.fluxcd.io/key=ces`
as a label.

By using our own sharding key, we still allow customers to use their own Flux components without side effects
with our platform.

### Limiting the namespace

By default, Flux watches all namespaces of the cluster. Since we currently operate on only one namespace,
the namespace was limited to the namespace of the Ecosystem-Core, into which Flux is also deployed.

### Monitoring

In general, Flux offers the ability to connect Prometheus using a `PodMonitor` and to export metrics of the individual controllers.
This requires that the `PodMonitor` CRD exists in the cluster, which is not the case at the time of installation of the Ecosystem-Core.
Prometheus itself is installed later as a component with its CRDs. For this reason,
monitoring is **disabled** at the moment.

### NetworkPolicies

Flux's Helm chart is designed to be deployed in a separate namespace. Under this assumption,
the NetworkPolicies that the Helm chart deploys are set up very "broadly" and **address all
pods** of the namespace, which would lead to side effects in our case. For this reason, the NetworkPolicies of the chart have been **disabled**.

However, since the Helm Controller needs access to the Source Controller in order to load the Helm chart locally, the Ecosystem-Core
deploys its own NetworkPolicy `flux-source-controller-allow-artifact-access` (`flux-network-policies.yaml`) for this purpose.
Besides the Helm Controller, the Dogu Operator is granted access as well, since it also fetches the artifacts directly.

### Retries for failed Helm operations

If errors occur during an installation or upgrade in the Helm Controller, the configured strategy of the `HelmRelease` takes effect.
By default, this is set to remediate (`RemediateOnFailure`) with a `rollback` between the retries. Once the retries
are exhausted, the `HelmRelease` remains in the faulty state and cannot be repaired without manual intervention.

To enable self-healing, the feature gate `DefaultToRetryOnFailure` is enabled in the Helm Controller. With this
feature enabled, a continuous retry is performed at the interval `.retryInterval` defined in the `HelmRelease`. The interval is
set to **5 minutes** by default.
