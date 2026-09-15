DATABASE_URL ?= postgres://feature_flags:feature_flags@localhost:5432/feature_flags?sslmode=disable
REDIS_URL ?= redis://localhost:6379/0
HTTP_ADDR ?= :8080
MIGRATE := go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.0

.PHONY: help deps-up deps-down deps-logs deps-reset migrate-up migrate-down run build generate test test-cover e2e vulncheck

help:
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-16s %s\n", $$1, $$2}'

deps-up: ## Start PostgreSQL and Redis
	docker compose up -d

deps-down: ## Stop PostgreSQL and Redis
	docker compose down

deps-logs: ## Follow dependency logs
	docker compose logs -f

deps-reset: ## Delete all local PostgreSQL and Redis data
	docker compose down -v

migrate-up: ## Apply all PostgreSQL migrations
	$(MIGRATE) -path db/migrations -database "$(DATABASE_URL)" up

migrate-down: ## Revert the latest PostgreSQL migration
	$(MIGRATE) -path db/migrations -database "$(DATABASE_URL)" down 1

run: ## Run the API locally
	HTTP_ADDR="$(HTTP_ADDR)" DATABASE_URL="$(DATABASE_URL)" REDIS_URL="$(REDIS_URL)" go run ./cmd/api

build: ## Build the API binary
	go build ./cmd/api

generate: ## Regenerate mocks
	go generate ./...

test: ## Run Go tests (requires Docker for Testcontainers)
	go test ./...

test-cover: ## Run Go tests with coverage (requires Docker for Testcontainers)
	go test -cover ./...

e2e: ## Run the HTTP end-to-end suite
	sh ./scripts/run-e2e.sh

vulncheck: ## Scan reachable Go vulnerabilities
	go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...
