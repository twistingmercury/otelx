ROOT_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
SRC_DIR := $(ROOT_DIR)src

DOCKER ?= docker
MODULES := src src/middleware/gin
VERIFY_IMAGE ?= otelx-ci:local

.PHONY: build docker-build unit-test analyze help

default: help

build: ## Build and validate the project in Docker
	IS_LOCAL=1 build/build.sh

analyze: ## Run formatters, linters, and security scanners
	cd "$(SRC_DIR)" && goimports -w .
	cd "$(SRC_DIR)" && golangci-lint run
	cd "$(SRC_DIR)" && govulncheck ./...
	cd "$(SRC_DIR)" && gosec -quiet -exclude-dir=tests ./...

unit-test: ## Run unit tests in Docker
	for m in $(MODULES); do (cd $$m && go vet ./... && CGO_ENABLED=1 go test -race ./...) || exit 1; done

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nAvailable targets:\n"} /^[a-zA-Z0-9_-]+:.*##/ { printf "  %-12s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
