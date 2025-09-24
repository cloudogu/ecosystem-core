# ecosystem-core

Ecosystem-Core is a Helm chart that installs the core components (operators) required to run the [Cloudogu Ecosystem](https://platform.cloudogu.com/en/info/cloudogu-ecosystem/)
on Kubernetes.
It works standalone or via GitOps tools like [Argo CD](https://argoproj.github.io/cd/).

## Prerequisites
- A Kubernetes cluster with cluster-admin privileges (tested with recent LTS releases).
- `kubectl` and `helm` v3.8+ installed.
- A namespace: The chart is namespace-scoped. You can install to any namespace (e.g., ecosystem).

## Validations / preconditions

The "Component" CustomResourceDefinition (CRD) must be installed in the cluster.
This is required by the `k8s-component-operator` to manage component objects.

This chart can fail fast when required Secrets/ConfigMaps are missing. We use Helm’s `lookup` during install to 
verify they exist and (optionally) contain specific keys. 

More information about the required CRDs, Secrets and ConfigMaps can be found [here](docs/operations/preparation_en.md).

To simplify the creation of secrets and ConfigMaps, there is a make target that can be used in conjunction with a .env file:

- `cp .env.template .env`
- Provide all information needed for Cloudogu's dogu registry, docker registry and helm registry
- `make install-component-crd`
- `make registry-configs`


## Troubleshooting
- **helm template fails**: When no cluster is available set skipPreconditionValidation=true in `values.yaml` for dry runs.
- **Missing keys in Secret/ConfigMap**: ensure keys are in .data key.
- **OCI pull errors**: verify helm registry login to Cloudogu's helm registry.


---
## What is the Cloudogu EcoSystem?
The Cloudogu EcoSystem is an open platform, which lets you choose how and where your team creates great software. Each service or tool is delivered as a Dogu, a Docker container. Each Dogu can easily be integrated in your environment just by pulling it from our registry.

We have a growing number of ready-to-use Dogus, e.g. SCM-Manager, Jenkins, Nexus Repository, SonarQube, Redmine and many more. Every Dogu can be tailored to your specific needs. Take advantage of a central authentication service, a dynamic navigation, that lets you easily switch between the web UIs and a smart configuration magic, which automatically detects and responds to dependencies between Dogus.

The Cloudogu EcoSystem is open source and it runs either on-premises or in the cloud. The Cloudogu EcoSystem is developed by Cloudogu GmbH under [AGPL-3.0-only](https://spdx.org/licenses/AGPL-3.0-only.html).

## License
Copyright © 2020 - present Cloudogu GmbH
This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, version 3.
This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License for more details.
You should have received a copy of the GNU Affero General Public License along with this program. If not, see https://www.gnu.org/licenses/.
See [LICENSE](LICENSE) for details.


---
MADE WITH :heart:&nbsp;FOR DEV ADDICTS. [Legal notice / Imprint](https://cloudogu.com/en/imprint/?mtm_campaign=ecosystem&mtm_kwd=imprint&mtm_source=github&mtm_medium=link)
