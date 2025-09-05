#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

# this function will be sourced from release.sh and be called from release_functions.sh
update_versions_modify_files() {
  #newReleaseVersion="${1}"
  valuesYAML=k8s/helm/values.yaml
  componentPatchTplYAML=k8s/helm/component-patch-tpl.yaml

  chartLockYAML=k8s/helm/Chart.lock

  # Extract component-operator chart
  local componentOperatorVersion
  componentOperatorVersion=$(./.bin/yq '.dependencies[] | select(.name=="k8s-component-operator").version' < ${chartLockYAML})

  ./.bin/yq -i ".k8s-component-operator.manager.image.tag = \"${componentOperatorVersion}\"" "${valuesYAML}"
  ./.bin/yq -i ".values.images.componentOperator |= sub(\":(([0-9]+)\.([0-9]+)\.([0-9]+)((?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))|(?:\+[0-9A-Za-z-]+))?)\", \":${componentOperatorVersion}\")" "${componentPatchTplYAML}"

  local kubectlVersion
  kubectlVersion=$(./.bin/yq '.cleanup.image.tag' < ${valuesYAML})
  ./.bin/yq -i ".values.images.kubectl |= sub(\":(([0-9]+)\.([0-9]+)\.([0-9]+)((?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))|(?:\+[0-9A-Za-z-]+))?)\", \":${kubectlVersion}\")" "${componentPatchTplYAML}"
}

update_versions_stage_modified_files() {
  valuesYAML=k8s/helm/values.yaml
  componentPatchTplYAML=k8s/helm/component-patch-tpl.yaml

  git add "${valuesYAML}" "${componentPatchTplYAML}"
}
