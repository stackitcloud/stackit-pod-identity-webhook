# Image URL to use all building/pushing image targets
REGISTRY ?= ghcr.io
REPO ?= stackitcloud/stackit-pod-identity-webhook
VERSION ?= $(shell git describe --dirty --tags --match='v*' 2>/dev/null || git rev-parse --short HEAD)
PUSH ?= false

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit immediately on error, unset variables, and pipe failures.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: verify

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
verify: verify-fmt verify-modules check

##@ Development

.PHONY: skaffold-dev
skaffold-dev: ## Run skaffold dev with cert-manager
	skaffold dev -p cert-manager

##@ Build

.PHONY: build
build: fmt ## Build manager binary.
	go build -o bin/manager ./cmd/stackit-pod-identity-webhook/main.go

.PHONY: image
image: ## Builds the image using ko
	KO_DOCKER_REPO=$(REGISTRY)/$(REPO) \
	go tool ko build --push=$(PUSH) \
	--image-label org.opencontainers.image.source="https://github.com/stackitcloud/stackit-pod-identity-webhook" \
	--sbom none -t $(VERSION) \
	--bare \
	--platform linux/amd64,linux/arm64 \
	./cmd/stackit-pod-identity-webhook

.PHONY: clean
clean: ## Clean binaries
	rm -rf bin/
