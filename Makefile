DOCKER ?= docker
VERIFY_IMAGE ?= otelx-ci:local

.PHONY: build docker-build unit-test analyze help

default: help

build: ## Build and validate the project in Docker
	IS_LOCAL=1 build/build.sh

analyze: ## Run analysis in Docker
	goimports -w .
	golangci-lint run
	govulncheck ./...
	gosec -quiet -exclude-dir=tests ./...

unit-test: ## Run unit tests in Docker
	go vet ./...
	CGO_ENABLED=1 go test -race ./...

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nAvailable targets:\n"} /^[a-zA-Z0-9_-]+:.*##/ { printf "  %-12s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
