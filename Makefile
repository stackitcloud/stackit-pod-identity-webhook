# Image URL to use all building/pushing image targets
REGISTRY                    ?= ghcr.io
REPOSITORY                  := $(REGISTRY)/stackitcloud/stackit-pod-identity-webhook
IS_DEV                      ?= true
ifeq ($(IS_DEV),true)
REPO_POSTFIX                := -dev
endif
REPO_ROOT                   := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
HACK_DIR                    := $(REPO_ROOT)/hack
export VERSION                     := $(shell git describe --tag --always --dirty)

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit immediately on error, unset variables, and pipe failures.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: verify

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: modules
modules: ## Runs go mod to ensure modules are up to date.
	go mod tidy

.PHONY: lint
lint: ## Run golangci-lint against code.
	go tool golangci-lint run ./...

.PHONY: test
test: ## Run tests.
	KUBEBUILDER_ASSETS="$(shell go tool setup-envtest use 1.34.x -p path)" ./hack/test.sh ./pkg/... ./test/...

.PHONY: check
check: lint test ## Check everything (lint + test).

.PHONY: verify-fmt
verify-fmt: fmt ## Verify go code is formatted.
	@if !(git diff --quiet HEAD); then \
		echo "unformatted files detected, please run 'make fmt'"; exit 1; \
	fi

.PHONY: verify-modules
verify-modules: modules ## Verify go module files are up to date.
	@if !(git diff --quiet HEAD -- go.sum go.mod); then \
		echo "go module files are out of date, please run 'make modules'"; exit 1; \
	fi

.PHONY: verify
verify: verify-fmt verify-modules check ## Runs formatting checks, module tidying, linting, and all tests.

.PHONY: skaffold-dev
skaffold-dev: ## Run skaffold dev with cert-manager
	skaffold dev -p cert-manager

.PHONY: kind-up
kind-up: ## Create kind cluster.
	kind create cluster

.PHONY: kind-down
kind-down: ## Delete kind cluster.
	kind delete cluster

##@ Build

export PUSH ?= false

.PHONY: build
build: fmt ## Build manager binary.
	go build -o bin/manager ./cmd/stackit-pod-identity-webhook/main.go

.PHONY: image
image: ## Builds the image using ko
	KO_DOCKER_REPO=$(REPOSITORY)$(REPO_POSTFIX) \
	go tool ko build --push=$(PUSH) \
	--image-label org.opencontainers.image.source="https://github.com/stackitcloud/stackit-pod-identity-webhook" \
	--sbom none -t $(VERSION) \
	--bare \
	--platform linux/amd64,linux/arm64 \
	./cmd/stackit-pod-identity-webhook \
  | tee images.txt

.PHONY: artifacts-only
artifacts-only: $(HELM) $(YQ)
	hack/push-artifacts.sh images.txt

.PHONY: artifacts
artifacts: image artifacts-only

.PHONY: clean
clean: ## Clean binaries
	rm -rf bin/
