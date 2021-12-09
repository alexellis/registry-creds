
# Image URL to use all building/pushing image targets
TAG?=latest

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"
export DOCKER_CLI_EXPERIMENTAL=enabled

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: controller

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build controller binary
controller: generate fmt vet
	go build -o bin/controller main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/controller && kustomize edit set image controller=ghcr.io/ghcr.io/alexellis/registry-creds:$(TAG)
	kustomize build config/default | kubectl apply -f -

.PHONY: shrinkwrap
shrinkwrap:
	cd config/default && \
	kustomize edit set image ghcr.io/ghcr.io/alexellis/registry-creds:$(TAG) && \
	kustomize build > ../../manifest.yaml
# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=registry-creds-role paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: docker-build # Build the docker image
docker-build:
	@docker buildx create --use --name=multiarch --node=multiarch && \
	docker buildx build \
		--output "type=docker,push=false" \
		--tag ghcr.io/alexellis/registry-creds:$(TAG) \
		.

.PHONY: docker-publish # Push the docker image to the remote registry
docker-publish:
	@docker buildx create --use --name=multiarch --node=multiarch && \
	docker buildx build \
		--platform linux/amd64,linux/arm64,linux/arm/v7 \
		--output "type=image,push=true" \
		--tag ghcr.io/alexellis/registry-creds:$(TAG) .

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
