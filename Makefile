-include .env
export

GO ?= go
BIN ?= bin/gateway

.PHONY: run build test coverage vet fmt check docker-build

run: ## Run the gateway on $$PORT (default 8080)
	$(GO) run ./cmd/service

build: ## Compile the gateway binary
	CGO_ENABLED=0 $(GO) build -trimpath -o $(BIN) ./cmd/service

test: ## Unit tests with the race detector
	LOG_LEVEL=PANIC $(GO) test ./... --race -count=1

coverage: ## Coverage profile (generated mocks excluded) and HTML report
	LOG_LEVEL=PANIC $(GO) test ./... -coverprofile=coverage.out
	grep -v '/mocks/' coverage.out > coverage.filtered.out
	$(GO) tool cover -func=coverage.filtered.out | tail -1
	$(GO) tool cover -html=coverage.out

vet: ## go vet
	$(GO) vet ./...

fmt: ## Fail if any file is not gofmt-formatted
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

check: fmt vet test ## Everything CI runs locally, except the linter

docker-build: ## Build the gateway image
	docker build -t auto-repair-shop:dev .
