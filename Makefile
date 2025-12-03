# Set these to the desired values
ARTIFACT_ID=ecosystem-core
ARTIFACT_ID_DEFAULT_CONFIG=${ARTIFACT_ID}-default-config

VERSION=2.0.2
GOTAG?=1.25.1

ADDITIONAL_CLEAN=clean_charts
MAKEFILES_VERSION=10.2.1

IMAGE=cloudogu/${ARTIFACT_ID_DEFAULT_CONFIG}:${VERSION}
IMAGE_DEV?=$(CES_REGISTRY_HOST)$(CES_REGISTRY_NAMESPACE)/$(ARTIFACT_ID_DEFAULT_CONFIG)/$(GIT_BRANCH)

IMAGE_IMPORT_TARGET=images-import
K8S_COMPONENT_SOURCE_VALUES = ${HELM_SOURCE_DIR}/values.yaml
K8S_COMPONENT_TARGET_VALUES = ${HELM_TARGET_DIR}/values.yaml
HELM_PRE_GENERATE_TARGETS = helm-values-update-image-version
HELM_POST_GENERATE_TARGETS = helm-values-replace-image-repo template-log-level template-image-pull-policy
COMPONENT_CRD_CHART_REF ?= oci://registry.cloudogu.com/k8s/k8s-component-operator-crd
COMPONENT_CRD_VERSION ?= 1.10.0

include build/make/variables.mk
include build/make/self-update.mk
include build/make/dependencies-gomod.mk
include build/make/build.mk
include build/make/mocks.mk
include build/make/test-common.mk
include build/make/test-unit.mk
include build/make/static-analysis.mk
include build/make/clean.mk
include build/make/k8s-component.mk
include build/make/release.mk

test-default-config: $(GO_JUNIT_REPORT)
	@echo "Compiling default-config..."
	cd ${WORKDIR}/default-config && $(GO_ENV_VARS) go mod vendor
	cd ${WORKDIR}/default-config && $(GO_ENV_VARS) go build $(GO_BUILD_FLAGS)
	@echo "Compiling default-config..."
	cd ${WORKDIR}/default-config && $(GO_ENV_VARS) go test -v -coverprofile=target/coverage.out -json ./... | $(GO_JUNIT_REPORT) > target/unit-tests.xml

.PHONY: mocks
mocks: ${MOCKERY_BIN} ## target is used to generate mocks for all interfaces in a project.
	cd ${WORKDIR}/default-config && ${MOCKERY_BIN}
	@echo "Mocks successfully created."

.PHONY: install-component-crd
install-component-crd: ## Installs the k8s-component-operator-crd Helm chart from OCI in version ${COMPONENT_CRD_VERSION}
	@echo "Installing Component-CRD with Helm: ${COMPONENT_CRD_CHART_REF} (${COMPONENT_CRD_VERSION}) into namespace ${NAMESPACE}"
	@${BINARY_HELM} upgrade --install "k8s-component-operator-crd" "${COMPONENT_CRD_CHART_REF}" \
		--version "${COMPONENT_CRD_VERSION}" \
		--namespace="${NAMESPACE}" --kube-context="${KUBE_CONTEXT_NAME}"

.PHONY: uninstall-component-crd
uninstall-component-crd: ## Installs the k8s-component-operator-crd Helm chart from OCI in version ${COMPONENT_CRD_VERSION}
	@echo "Unnstalling Component-CRD with Helm"
	@${BINARY_HELM} uninstall "k8s-component-operator-crd" \
		--namespace="${NAMESPACE}" --kube-context="${KUBE_CONTEXT_NAME}"

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

.PHONY: template-log-level
template-log-level: $(BINARY_YQ)
	@if [ -n "${LOG_LEVEL}" ]; then \
		echo "Setting LOG_LEVEL env in deployment to ${LOG_LEVEL}!"; \
		$(BINARY_YQ) -i e ".k8s-component-operator.manager.env.logLevel=\"${LOG_LEVEL}\"" ${K8S_COMPONENT_TARGET_VALUES}; \
		$(BINARY_YQ) -i e ".defaultConfig.env.logLevel=\"${LOG_LEVEL}\"" ${K8S_COMPONENT_TARGET_VALUES}; \
	else \
		echo "LOG_LEVEL not set; skipping log level templating."; \
	fi

.PHONY: docker-build
docker-build: check-docker-credentials check-k8s-image-env-var ${BINARY_YQ} ## Overwrite docker-build from k8s.mk to build from subdir
	@echo "Building docker image $(IMAGE) in directory $(IMAGE_DIR)..."
	@DOCKER_BUILDKIT=1 docker build $(IMAGE_DIR) -t $(IMAGE)

.PHONY: images-import
images-import: ## import images from ces-importer and
	@echo "Import default config"
	@make image-import \
		IMAGE_DIR=./default-config \
		IMAGE=${ARTIFACT_ID_DEFAULT_CONFIG}:${VERSION} \
		IMAGE_DEV_VERSION=$(IMAGE_DEV):${VERSION}

.PHONY: helm-values-update-image-version
helm-values-update-image-version: $(BINARY_YQ)
	@echo "Updating the image version in source values.yaml to ${VERSION}..."
	@$(BINARY_YQ) -i e ".defaultConfig.image.tag = \"${VERSION}\"" ${K8S_COMPONENT_SOURCE_VALUES}

.PHONY: helm-values-replace-image-repo
helm-values-replace-image-repo: $(BINARY_YQ)
	@if [[ "${STAGE}" == "development" ]]; then \
		echo "Setting dev image repo in target values.yaml!" ;\
		echo "Component target values: ${IMAGE_DEV}" ;\
		REGISTRY=$$(echo "${IMAGE_DEV}" | sed 's|\([^/]*\)/.*|\1|') ;\
		MAIN_REPOSITORY=$$(echo "${IMAGE_DEV}" | sed 's|^[^/]*/||; s|:.*$$||') ;\
		echo "Registry: $$REGISTRY" ;\
		echo "Main Repository: $$MAIN_REPOSITORY" ;\
		$(BINARY_YQ) -i e ".defaultConfig.image.registry=\"$$REGISTRY\"" ${K8S_COMPONENT_TARGET_VALUES} ;\
		$(BINARY_YQ) -i e ".defaultConfig.image.repository=\"$$MAIN_REPOSITORY\"" ${K8S_COMPONENT_TARGET_VALUES} ;\
	fi

.PHONY: template-image-pull-policy
template-image-pull-policy: $(BINARY_YQ)
	@if [[ "${STAGE}" == "development" ]]; then \
		echo "Setting pull policy to always!" ; \
		$(BINARY_YQ) -i e ".defaultConfig.imagePullPolicy=\"Always\"" "${K8S_COMPONENT_TARGET_VALUES}" ; \
	fi

clean_charts:
	rm -rf ${HELM_SOURCE_DIR}/charts


.PHONY: ecosystem-core-release
ecosystem-core-release: ## Interactively starts the release workflow for ecosystem-core
	@echo "Starting git flow release..."
	@build/make/release.sh ecosystem-core