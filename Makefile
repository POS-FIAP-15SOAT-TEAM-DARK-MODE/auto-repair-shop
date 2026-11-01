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

# ---- Local ecosystem (docker compose) and end-to-end tests ----------------------------------
# Needs Docker. The services are cloned into services/ (see scripts/clone-services.sh).
COMPOSE ?= docker compose
E2E_BASE_URL ?= http://localhost:8080

clone: ## Clone the service repositories (or fast-forward the existing clones)
	sh scripts/clone-services.sh

build-all: ## Build the images one at a time (building them in parallel exhausts a small Docker VM)
	for s in customer-service catalog-service billing-service execution-service history-service service-order-service lambda lambda-adapter gateway; do $(COMPOSE) build $$s || exit 1; done

up-all: ## Clone what is missing, build and start the whole ecosystem, wait for /ping and /ready
	sh scripts/clone-services.sh --if-missing
	$(MAKE) build-all
	$(COMPOSE) up -d
	sh scripts/wait-ready.sh 180

down-all: ## Stop the ecosystem and remove its volumes
	$(COMPOSE) down -v

logs: ## Follow the logs of the ecosystem
	$(COMPOSE) logs -f --tail=100

e2e: ## Run the BDD end-to-end suite (godog) against E2E_BASE_URL
	cd e2e && E2E_BASE_URL=$(E2E_BASE_URL) go test ./... -count=1 -v

.PHONY: clone build-all up-all down-all logs e2e
