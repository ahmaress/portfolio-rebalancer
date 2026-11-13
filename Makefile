.PHONY: help run build test test-unit test-integration lint fmt vet cover up down infra token clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

run: ## Run the API locally (needs `make infra` and a .env)
	go run ./cmd/api

build: ## Build server and token binaries into ./bin
	CGO_ENABLED=0 go build -trimpath -o bin/server ./cmd/api
	CGO_ENABLED=0 go build -trimpath -o bin/token ./cmd/token

test: ## Run all tests (integration tests skip if Elasticsearch is unreachable)
	go test -count=1 ./...

test-unit: ## Run unit tests only
	go test -count=1 -short ./...

test-integration: ## Run Elasticsearch integration tests
	go test -count=1 -v ./internal/infrastructure/elasticsearch/...

cover: ## Run unit tests with a coverage report
	go test -short -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

lint: ## Run golangci-lint
	golangci-lint run ./...

fmt: ## Format code
	gofmt -w .

vet: ## Run go vet
	go vet ./...
