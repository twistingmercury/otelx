.PHONY: build test clean help local

GIT_COMMIT := $(shell git rev-parse --short=8 HEAD 2>/dev/null || echo "unknown")
GIT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nAvailable targets:\n"} /^[a-zA-Z0-9_-]+:.*##/ { printf "  %-12s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

local: ## Build locally using workspace
	 go build -o .out/gingonic-decorator cmd/main/main.go

# A local collector without TLS; set OTEL_EXPORTER_OTLP_ENDPOINT to point elsewhere.
OTEL_EXPORTER_OTLP_ENDPOINT ?= http://localhost:4317
export OTEL_EXPORTER_OTLP_ENDPOINT

run: ## Run locally for dev purposes
	go run cmd/main/main.go

build: ## Build Docker image
	docker compose build

clean: ## Remove built images and containers
	docker compose down --rmi local
