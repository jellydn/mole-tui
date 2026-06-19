SHELL := /bin/sh
.SHELLFLAGS := -eu -c

# Build version. Override with `make build VERSION=v0.1.0`.
VERSION ?= dev

# Local install target; respects $GOBIN / $GOPATH/bin.
GOBIN ?= $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell go env GOPATH)/bin
endif

BIN_DIR := bin
BIN := $(BIN_DIR)/mole-tui

# Version is injected at link time.
LDFLAGS := -X main.version=$(VERSION)

.PHONY: help
help: ## Show this help.
	@awk 'BEGIN {FS = ":.*##"; printf "Targets:\n"} /^[a-zA-Z0-9_.-]+:.*##/ { printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build ./bin/mole-tui for the host OS/arch.
	@mkdir -p $(BIN_DIR)
	go build -ldflags '$(LDFLAGS)' -trimpath -o $(BIN) ./cmd/mole-tui

.PHONY: install
install: ## go install ./cmd/mole-tui to $(GOBIN).
	go install -ldflags '$(LDFLAGS)' -trimpath ./cmd/mole-tui

.PHONY: test
test: ## Run all unit tests.
	go test ./...

.PHONY: vet
vet: ## Run go vet on all packages.
	go vet ./...

.PHONY: fmt
fmt: ## Check gofmt (fails if any file is unformatted).
	@gofmt -l -s . | tee /tmp/mole-tui-fmt.out
	@test ! -s /tmp/mole-tui-fmt.out
	@rm -f /tmp/mole-tui-fmt.out
