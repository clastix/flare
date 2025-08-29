# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

VERSION ?= latest

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KO				?= $(LOCALBIN)/ko
OAPI_CODEGEN	?= $(LOCALBIN)/oapi-codegen
CONTROLLER_GEN	?= $(LOCALBIN)/controller-gen
YQ				?= $(LOCALBIN)/yq

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Binary
.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) CGO_ENABLED=0 go install -ldflags="-s -w" sigs.k8s.io/controller-tools/cmd/controller-gen@v0.16.1

.PHONY: yq
yq: $(YQ) ## Download yq locally if necessary.
$(YQ): $(LOCALBIN)
	test -s $(LOCALBIN)/yq || GOBIN=$(LOCALBIN) CGO_ENABLED=0 go install -ldflags="-s -w" github.com/mikefarah/yq/v4@v4.44.2

.PHONY: oapi-codegen
oapi-codegen: $(OAPI_CODEGEN) ## Download ko locally if necessary.
$(OAPI_CODEGEN): $(LOCALBIN)
	test -s $(LOCALBIN)/oapi-codegen || GOBIN=$(LOCALBIN) go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.0

.PHONY: ko
ko: $(KO) ## Download ko locally if necessary.
$(KO): $(LOCALBIN)
	test -s $(LOCALBIN)/ko || GOBIN=$(LOCALBIN) CGO_ENABLED=0 go install -ldflags="-s -w" github.com/google/ko@v0.14.1

##@ Development

.PHONY: build-server
build-server: $(KO)
	KO_DOCKER_REPO=docker.io/clastix/flare-server $(KO) build ./cmd/server --sbom=none --bare --tags=$(VERSION) --local=true --push=false

.PHONY: build-operator
build-operator: $(KO)
	KO_DOCKER_REPO=docker.io/clastix/flare-operator $(KO) build ./cmd/operator --sbom=none --bare --tags=$(VERSION) --local=true --push=false

.PHONY: build
build: build-server build-operator

.PHONY: load
load: build
	kind load docker-image clastix/flare-server:latest --name=fluidos-consumer-1
	kind load docker-image clastix/flare-operator:latest --name=fluidos-consumer-1

.PHONY: gogenerate
gogenerate: controller-gen ## Generate deep-copy methods for Custom Resource types.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: api
api: oapi-codegen ## Generate web server Go files from OAPIv3 specification.
	$(LOCALBIN)/oapi-codegen -package=api -generate=types,server -o internal/api/api.gen.go specification.yaml

.PHONY: crds
crds: yq controller-gen ## Generate Custom Resource Definition YAML from source code for the Helm Chart.
	$(CONTROLLER_GEN) crd:allowDangerousTypes=true webhook paths="./..." output:stdout | $(YQ) 'select(documentIndex == 0) | .spec' > ./charts/flare/hack/flare.clastix.io_intent_spec.yaml

.PHONY: rbac
rbac: controller-gen yq
	$(CONTROLLER_GEN) rbac:roleName=flare-role paths="./..." output:stdout | $(YQ) '.rules' > ./charts/flare/hack/clusterrole.yaml

.PHONY: generate
generate: gogenerate api crds

.PHONY: install
install: rbac crds ## Install the Helm Chart into the Kubernetes cluster.
	helm upgrade --install flare charts/flare --namespace=flare-system --create-namespace \
	--set "server.replicas=1" \
	--set "server.image.pullPolicy=IfNotPresent" \
	--set "operator.replicas=1" \
	--set "operator.image.pullPolicy=IfNotPresent"