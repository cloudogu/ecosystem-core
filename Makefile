# Set these to the desired values
ARTIFACT_ID=ecosystem-core
VERSION=0.1.0

MAKEFILES_VERSION=10.2.1

include build/make/variables.mk
include build/make/self-update.mk
include build/make/clean.mk
include build/make/k8s-component.mk

##@ registry-configs
.PHONY: registry-configs
registry-configs: dogu-registry-config container-registry-config helm-registry-config ## Creates the secrets for all registries

##@ Dogu-Registry-Config
.PHONY: dogu-registry-config
dogu-registry-config: ## Creates a secret for the dogu registry
	@echo "Creating Dogu-Registry Secret!"
	@kubectl create secret generic k8s-dogu-operator-dogu-registry \
		--from-literal=endpoint=${DOGU_REGISTRY_URL} \
		--from-literal=urlschema=${DOGU_REGISTRY_URL_SCHEMA} \
		--from-literal=username=${DOGU_REGISTRY_USERNAME} \
		--from-literal=password=$(shell echo ${DOGU_REGISTRY_PASSWORD} | base64 -d) \
		--namespace="${NAMESPACE}" --context="${KUBE_CONTEXT_NAME}"

##@ Container-Registry-Config
.PHONY: container-registry-config
container-registry-config: ## Creates a secret for the container registry
	@echo "Creating Container-Registry Secret!"
	@kubectl create secret docker-registry ces-container-registries \
		--docker-server="${DOCKER_REGISTRY_URL}" \
		--docker-username="${DOCKER_REGISTRY_USERNAME}" \
		--docker-password="$(shell echo ${DOCKER_REGISTRY_PASSWORD} | base64 -d)" \
		--docker-email="${DOCKER_REGISTRY_EMAIL}" \
		--namespace="$${NAMESPACE}" \
		--namespace="${NAMESPACE}" --context="${KUBE_CONTEXT_NAME}"

##@ Helm-Registry-Config
.PHONY: helm-registry-config
helm-registry-config: ## Creates a configMap and a secret for the helm registry
	@echo "Creating Helm-Repo Configmap & Secret!"
	@kubectl create configmap component-operator-helm-repository \
		--from-literal=endpoint=${HELM_REGISTRY_HOST} \
		--from-literal=schema=${HELM_REGISTRY_SCHEMA} \
		--from-literal=plainHttp=${HELM_REGISTRY_PLAIN_HTTP} \
		--from-literal=insecureTls=${HELM_REGISTRY_INSECURE_TLS} \
		--namespace="${NAMESPACE}" --context="${KUBE_CONTEXT_NAME}"
	@kubectl create secret generic component-operator-helm-registry \
		--from-literal=config.json='{"auths": {"${HELM_REGISTRY_HOST}": {"auth": "$(shell echo -n "${HELM_REGISTRY_USERNAME}:$(shell echo ${HELM_REGISTRY_PASSWORD} | base64 -d)" | base64)"}}}' \
		--namespace="${NAMESPACE}" --context="${KUBE_CONTEXT_NAME}"
