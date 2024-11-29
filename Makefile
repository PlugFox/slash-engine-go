SHELL :=/bin/bash -e -o pipefail
PWD   := $(shell pwd)

BUILD_MODE?=c-shared
OUTPUT_DIR?=output
GO_BINARY?=go
BINDING_NAME?=librsa_bridge
BINDING_FILE?=$(BINDING_NAME).so
BINDING_ARGS?=
BINDING_OUTPUT?=$(OUTPUT_DIR)/binding
EXTRA_LD_FLAGS?=

.DEFAULT_GOAL := all
.PHONY: all
all: ## build pipeline
all: mod inst gen build spell lint test

.PHONY: precommit
precommit: ## validate the branch before commit
precommit: all vuln

.PHONY: ci
ci: ## CI build pipeline
ci: lint-reports test vuln precommit diff

.PHONY: help
help:
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: out
out: ## create out directory
	@mkdir -p out

.PHONY: git-hooks
git-hooks: ## install git hooks
	@git config --local core.hooksPath .githooks/

.PHONY: deps
deps: ## Downloads the dependencies
	@go mod download

.PHONY: run
run: fmt ## Run the app
	@go run ./cmd/main.go

.PHONY: test-build
test-build: ## Tests whether the code compiles
	@go build -o /dev/null ./...

.PHONY: clean
clean: ## remove files created during build pipeline
	$(call print-target)
	@rm -rf dist bin out build output
	@rm -f coverage.*
	@rm -f '"$(shell go env GOCACHE)/../golangci-lint"'
	@go clean -i -cache -testcache -modcache -fuzzcache -x

.PHONY: mod
mod: ## go mod tidy, cleans up go.mod and go.sum
	$(call print-target)
	@go mod tidy
	@cd tools && go mod tidy

.PHONY: fmt
fmt: ## Formats all code with go fmt
	@go fmt ./...

.PHONY: inst
inst: ## go install tools
	$(call print-target)
	@cd tools && go install $(shell cd tools && go list -e -f '{{ join .Imports " " }}' -tags=tools)

.PHONY: get
get: ## get and update dependencies
	$(call print-target)
	@go get -u ./...

.PHONY: gen
gen: ## go generate
	$(call print-target)
	@go generate ./...

.PHONY: build
build: ## goreleaser build
	$(call print-target)
	@goreleaser build --clean --single-target --snapshot

.PHONY: spell
spell: ## misspell
	$(call print-target)
	@misspell -error -locale=US -w **.md

.PHONY: lint
lint: fmt deps ## Lints all code with golangci-lint
	$(call print-target)
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix

.PHONY: lint-reports
lint-reports: out deps ## Lint reports
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./... --out-format checkstyle | tee "$(@)"

.PHONY: vuln
vuln: ## govulncheck
	$(call print-target)
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

.PHONY: test
test: ## go test
	$(call print-target)
	@go test -race -covermode=atomic -coverprofile=coverage.out -coverpkg=./... ./...
	@go tool cover -html=coverage.out -o coverage.html

.PHONY: diff
diff: ## git diff
	$(call print-target)
	@git diff --exit-code
	@RES=$$(git status --porcelain) ; if [ -n "$$RES" ]; then echo $$RES && exit 1 ; fi

.PHONY: binding
binding: deps ## build the binding
	mkdir -p $(BINDING_OUTPUT)
	$(GO_BINARY) build -ldflags="-w -s $(EXTRA_LD_FLAGS)" -o $(BINDING_OUTPUT)/$(BINDING_FILE) -buildmode=$(BUILD_MODE) $(BINDING_ARGS) main.go

define print-target
    @printf "Executing target: \033[36m$@\033[0m\n"
endef

#include Makefile.android
#include Makefile.ios
#include Makefile.darwin
#include Makefile.linux
#include Makefile.windows
#include Makefile.gomobile
#include Makefile.wasm
