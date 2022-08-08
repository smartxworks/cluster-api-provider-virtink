
# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.23

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./api/... ./controllers/... -coverprofile cover.out

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go

.PHONY: docker-build
docker-build: test ## Build docker image with the manager.
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GINKGO ?= $(LOCALBIN)/ginkgo
KIND ?= $(LOCALBIN)/kind
KUBECTL ?= $(LOCALBIN)/kubectl
CMCTL ?= $(LOCALBIN)/cmctl

## Tool Versions
KUSTOMIZE_VERSION ?= v4.5.7
CONTROLLER_TOOLS_VERSION ?= v0.8.0
GINKGO_VERSION ?= v1.16.5

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	rm -rf $(KUSTOMIZE)
	curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: ginkgo
ginkgo: $(GINKGO)
$(GINKGO): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install github.com/onsi/ginkgo/ginkgo@$(GINKGO_VERSION)

.PHONY: kind
kind: $(KIND)
$(KIND): $(LOCALBIN)
	curl -sLo $(KIND) https://kind.sigs.k8s.io/dl/v0.14.0/kind-linux-amd64 && chmod +x $(KIND)

.PHONY: kubectl
kubectl: $(KUBECTL)
$(KUBECTL): $(LOCALBIN)
	curl -sLo $(KUBECTL) https://dl.k8s.io/release/v1.24.0/bin/linux/amd64/kubectl && chmod +x $(KUBECTL)

.PHONY: cmctl
cmctl: $(CMCTL)
$(CMCTL): $(LOCALBIN)
	curl -sLo cmctl.tar.gz https://github.com/cert-manager/cert-manager/releases/download/v1.8.2/cmctl-linux-amd64.tar.gz
	tar xzf cmctl.tar.gz -C $(LOCALBIN)
	rm -rf cmctl.tar.gz

.PHONY: e2e-image
e2e-image:
	docker buildx build -t docker.io/smartxworks/capch-controller:e2e .

REPO_ROOT := $(shell pwd )
E2E_CLUSTER_TEMPLATE_DIR ?= $(REPO_ROOT)/test/e2e/data/infrastructure-virtink

.PHONY: e2e-cluster-templates-v1alpha1
e2e-cluster-templates-v1alpha1: $(KUSTOMIZE) ## Generate cluster templates for v1beta1
	$(KUSTOMIZE) build $(E2E_CLUSTER_TEMPLATE_DIR)/v1alpha1/cluster-template-internal --load-restrictor LoadRestrictionsNone > $(E2E_CLUSTER_TEMPLATE_DIR)/v1alpha1/cluster-template-internal.yaml
	$(KUSTOMIZE) build $(E2E_CLUSTER_TEMPLATE_DIR)/v1alpha1/cluster-template --load-restrictor LoadRestrictionsNone > $(E2E_CLUSTER_TEMPLATE_DIR)/v1alpha1/cluster-template.yaml

SKIP_RESOURCE_CLEANUP ?= false
CERT_MANAGER_MANIFEST ?= https://github.com/cert-manager/cert-manager/releases/download/v1.8.2/cert-manager.yaml
VIRTINK_MANIFEST ?= https://github.com/smartxworks/virtink/releases/download/v0.8.0/virtink.yaml	
E2E_KIND_CLUSTER_NAME ?= capch-e2e-$(shell date "+%Y-%m-%d-%H-%M-%S")
E2E_KIND_CLUSTER_KUBECONFIG := /tmp/$(E2E_KIND_CLUSTER_NAME).kubeconfig

.PHONY: e2e
e2e: kind e2e-image kubectl cmctl kustomize ginkgo e2e-cluster-templates-v1alpha1
	echo "e2e kind cluster: $(E2E_KIND_CLUSTER_NAME)"

	$(KIND) create cluster --config test/e2e/config/kind/config.yaml --name $(E2E_KIND_CLUSTER_NAME) --kubeconfig $(E2E_KIND_CLUSTER_KUBECONFIG)
	$(KIND) load docker-image --name $(E2E_KIND_CLUSTER_NAME) docker.io/smartxworks/capch-controller:e2e

	KUBECONFIG=$(E2E_KIND_CLUSTER_KUBECONFIG) $(KUBECTL) apply -f test/e2e/data/cni/calico/calico.yaml
	KUBECONFIG=$(E2E_KIND_CLUSTER_KUBECONFIG) $(KUBECTL) apply -f $(CERT_MANAGER_MANIFEST)
	KUBECONFIG=$(E2E_KIND_CLUSTER_KUBECONFIG) $(CMCTL) check api --wait=10m
	KUBECONFIG=$(E2E_KIND_CLUSTER_KUBECONFIG) $(KUBECTL) apply -f $(VIRTINK_MANIFEST)
	KUBECONFIG=$(E2E_KIND_CLUSTER_KUBECONFIG) $(KUBECTL) wait -n virtink-system deployment virt-controller --for condition=Available --timeout -1s

	PATH=$(LOCALBIN):$(PATH) KUBECONFIG=$(E2E_KIND_CLUSTER_KUBECONFIG) $(GINKGO) -v -trace -tags=e2e ./test/e2e -- \
		-e2e.artifacts-folder="$(REPO_ROOT)/_artifacts" \
		-e2e.config="$(REPO_ROOT)/test/e2e/config/virtink.yaml" \
		-e2e.skip-resource-cleanup=$(SKIP_RESOURCE_CLEANUP)

	$(KIND) delete cluster --name $(E2E_KIND_CLUSTER_NAME)
