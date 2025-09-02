# Set these to the desired values
ARTIFACT_ID=ecosystem-core
VERSION=0.1.0

MAKEFILES_VERSION=10.2.1

include build/make/variables.mk
include build/make/self-update.mk
include build/make/clean.mk
include build/make/k8s-component.mk
