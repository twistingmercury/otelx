.PHONY: build test clean help local

GIT_COMMIT := $(shell git rev-parse --short=8 HEAD 2>/dev/null || echo "unknown")
GIT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nAvailable targets:\n"} /^[a-zA-Z0-9_-]+:.*##/ { printf "  %-12s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

local: ## Build locally using workspace
	 go build -o .out/uuid_api cmd/main/main.go

run: ## Run locally for dev purposes
	go run cmd/main/main.go

build: ## Build Docker image
	./build/build.sh
	docker system prune -f

clean: ## Remove built images and containers
	docker compose down --rmi local 2>/dev/null || true
	docker rmi uuid-api:latest 2>/dev/null || true
