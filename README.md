# Auto Repair Shop — API Gateway

The single public entry point of the Auto Repair Shop. It is a reverse proxy written in Go (Gin and `net/http/httputil.ReverseProxy`): it sends each request to the service that owns it, and nothing else. It validates no token, reads no database and holds no secret. Each service validates its own tokens.

- Port: `8080` · Namespace `auto-repair-shop` · Kubernetes name `auto-repair-shop`
- Decisions: [`docs/adr/`](docs/adr) (ADR 0005 explains why this repository is the gateway)
- Agent rules: [`.ai/`](.ai/README.md) · [`CLAUDE.md`](CLAUDE.md) · [`AGENTS.md`](AGENTS.md)

## Architecture

```
client ─► AWS API Gateway ─┬─ POST /auth/customer-login ─► lambda-auth
                           └─ everything else ─► auto-repair-shop (this gateway, :8080)
                                                   │ reverse proxy, ClusterIP upstreams
     ┌──────────┬──────────┬──────────┬───────────┬─────────┬─────────────┐
  customer   catalog    billing   execution    history  service-order
  :8081      :8082      :8083     :8084        :8085    :8086
```

## Routes

The routing table is [`config/routes.yaml`](config/routes.yaml), embedded in the binary. Routes are tried in order and the first match (method and path) wins.

| Method | Path | Upstream |
|---|---|---|
| any | `/internal/*` | blocked, 404 |
| PUT | `/v1/service-order/:id/work/:workId/next`, `.../cancel`, `/v1/service-order/:id/finish` | `execution` |
| GET | `/v1/execution/*` | `execution` |
| GET | `/v1/service-order/:id/history`, `/v1/reports/*` | `history` |
| any | `/v1/auth/*`, `/v1/users/*`, `/v1/customers*`, `/v1/vehicles*` | `customer` |
| any | `/v1/works*`, `/v1/supplies*` | `catalog` |
| any | `/v1/budgets*`, `/v1/payments*` | `billing` |
| any | `/v1/service-order*` (the rest) | `service_order` |
| POST | `/auth/customer-login` (optional) | `lambda`, only when `LAMBDA_UPSTREAM` is set |

Patterns: a `:param` matches one non-empty segment; a trailing `*` matches any suffix. Requests that match nothing, or that match a blocked route, answer `404 {"code":404,"errors":["route not found"]}`.

## Gateway endpoints

| Endpoint | Result |
|---|---|
| `GET /ping` | `200 {"message":"pong"}` (liveness and readiness probes) |
| `GET /ready` | `200` when every configured upstream answers `/ping` within 2 s; otherwise `503` listing the failing upstream names |
| `GET /metrics` | Prometheus metrics, `http_request_duration_seconds{method,route,upstream,status}` |
| `GET /swagger.yaml` | OpenAPI document of the gateway itself |
| `GET /swagger/index.html` | Page with links to the documentation of each service |
| `GET /docs/<service>/swagger.yaml` | Proxied `swagger.yaml` of the service (`customer`, `catalog`, `billing`, `execution`, `history`, `service_order`) |

## Proxy behaviour

- Method, path, query, body and headers are forwarded, `Authorization` included. `X-Forwarded-For`, `-Host` and `-Proto` are added.
- `X-Request-Id` is kept when the client sends a valid one (up to 128 characters from `A-Z a-z 0-9 - _ . :`), otherwise generated. It is returned in the response, sent to the upstream and written in every log line.
- An upstream that does not answer its headers within `UPSTREAM_TIMEOUT` gives `504 {"code":504,"errors":["upstream timeout"]}`. An upstream that cannot be reached gives `502 {"code":502,"errors":["upstream unavailable"]}`. The cause is logged on the server and never sent to the client.
- Any response from an upstream, 4xx and 5xx included, is passed through unchanged.
- Logs are JSON, one line per request, with `request_id`, `method`, `path`, `upstream`, `status` and `duration_ms`. The query string is not logged, because it can carry a customer document.

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `8080` | Listening port |
| `LOG_LEVEL` | `INFO` | `DEBUG`, `INFO` or `PANIC` |
| `UPSTREAM_TIMEOUT` | `30s` | Wait for the upstream response headers |
| `CUSTOMER_UPSTREAM` | `http://customer-service:8081` | Base URL of the customer service |
| `CATALOG_UPSTREAM` | `http://catalog-service:8082` | Base URL of the catalog service |
| `BILLING_UPSTREAM` | `http://billing-service:8083` | Base URL of the billing service |
| `EXECUTION_UPSTREAM` | `http://execution-service:8084` | Base URL of the execution service |
| `HISTORY_UPSTREAM` | `http://history-service:8085` | Base URL of the history service |
| `SERVICE_ORDER_UPSTREAM` | `http://service-order-service:8086` | Base URL of the service order service |
| `LAMBDA_UPSTREAM` | unset | Base URL of the login Lambda adapter; unset disables `POST /auth/customer-login` |

Invalid values stop the gateway at startup, and the error never includes the configured URL.

## Run

```bash
cp .env.example .env   # optional, for make run on the host
make run               # go run ./cmd/service
make test              # unit tests with the race detector
make coverage          # coverage profile and HTML report
make docker-build      # image from the Dockerfile
```

## Local ecosystem

`docker-compose.yml` runs the gateway, the six services, the login lambda and the shared infrastructure on `auto-repair-shop-net`:

```
            host :8080
                │
            gateway ──────────────┬─ customer-service     :8081  ─┐
                │                 ├─ catalog-service      :8082   │
                │                 ├─ billing-service      :8083   │  one PostgreSQL (postgres:15-alpine),
                │                 ├─ execution-service    :8084   │  one database per service (local/postgres-init)
                │                 ├─ history-service      :8085   │  one MiniStack (SNS, SQS, DynamoDB, Secrets Manager)
                │                 ├─ service-order-service:8086   │
                │                 └─ lambda-adapter :9001 ─► lambda (RIE) ─┘
```

The service repositories are not part of this repository. `scripts/clone-services.sh` clones them into `services/` (the directory is git-ignored), on the branch given by `SERVICES_BRANCH` (default `develop`).

```bash
make up-all      # clone what is missing, docker compose up -d --build, wait for /ping and /ready
make logs        # follow the logs
make down-all    # stop and remove the volumes
make e2e         # run the BDD scenarios against http://localhost:8080 (E2E_BASE_URL)
```

Prerequisite: Docker with Compose v2. Ports used on the host: 8080 (gateway), 8081-8086 (services), 9001 (lambda adapter), 4566 (MiniStack), 5432 (PostgreSQL).

### Security of the local credentials

Every password and secret in `docker-compose.yml`, `local/postgres-init/` and `local/init-aws.sh` (for example `local-dev-only`, `local-dev-jwt-secret`, the seed passwords `attendant123` and `mechanic123`) is a **fixed local development value**. They are not secrets and must never be used outside this ecosystem. The lambda secret `auto-repair-shop-local/lambda-auth` lives only in MiniStack.

### What the ecosystem creates

- `local/postgres-init/`: one role and one database per service (`customer`, `catalog`, `billing`, `execution`, `service_order`), the read-only role `lambda_auth_ro` with its `ALTER DEFAULT PRIVILEGES` on the `customer` database, and the revoke of `CONNECT` from `PUBLIC`.
- `migrate-<service>` containers apply each relational service's `migrations/` as the service role.
- `local/init-aws.sh` (idempotent) creates on MiniStack, with the `auto-repair-shop-local-` prefix: 4 topics, 10 queues (5 queues and their 5 DLQs), 8 subscriptions (raw delivery, `target` filters on the saga commands), the table `auto-repair-shop-local-history` and the secret `auto-repair-shop-local/lambda-auth`.
- `PAYMENT_MODE=manual`: the attendant approves or rejects payments (`POST /v1/payments/{id}/confirm|reject`).

### Login lambda locally

`local/lambda/Dockerfile` builds the lambda repository's `cmd/lambda` as `bootstrap` on `public.ecr.aws/lambda/provided:al2023`, whose Runtime Interface Emulator listens on port 8080 inside the network. `local/lambda/adapter` is a standard-library HTTP adapter: it turns `POST /auth/customer-login` into an API Gateway HTTP API v2 event, invokes the emulator and returns the response. It also answers `GET /ping`, which the gateway's `/ready` uses. The gateway reaches it through `LAMBDA_UPSTREAM=http://lambda-adapter:9001`. The request body (a CPF) is never logged.

## End-to-end tests

`e2e/` is a separate Go module (`github.com/cucumber/godog`). `features/*.feature` hold the scenarios; the steps talk **only to the gateway**, through the routes of `config/routes.yaml`.

| Feature | Scenarios |
|---|---|
| `happy_path` | order received, diagnosed, budget sent, accepted, paid, executed, finished and delivered; history with work transitions; execution-time report |
| `payment_rejected` | attendant rejects the payment: order `REJECTED`, stock back to the initial value |
| `execution_failed` | every work cancelled: order `CANCELLED`, payment `REFUNDED`, stock back |
| `budget_rejected` | customer rejects the budget: order `REJECTED`, stock back |
| `customer_login` | known CPF gets a token accepted by a service (`accept`); unknown CPF gets `404` |
| `security` | `/internal/*` gives `404` through the gateway; a protected route without token gives `401` from the service |
| `idempotency` | accept and payment confirmation sent again answer `409` and do not duplicate the execution task |

Asynchronous saga results are awaited with bounded polling (20 s). The fixed identifiers are the seed IDs of the contracts (works, supply, customer A, vehicle A).

The CI job `e2e` (`.github/workflows/ci.yml`) runs the same sequence on pull requests: `make clone`, `docker compose up -d --build`, wait for `/ready`, `make e2e`, logs on failure, `docker compose down -v` always.

### Status of this change (read this first)

- **The ecosystem and the e2e suite have not been run in the environment where they were written: that environment has no Docker daemon.** What was verified there: `docker compose config -q`, `sh -n` on the scripts, `gofmt`, `go vet` for the gateway and the e2e module, `go test -run XXX` (compilation), the Go adapter build, and a check that every feature step matches a registered step definition. The `aws-init` counts were checked by reading `local/init-aws.sh` (4 topics, 10 queues, 8 subscriptions, 1 table, 1 secret).
- The e2e suite is validated by the CI job of the pull request and on the developer's machine once the service branches containing the waves 2 and 3 are merged into `develop`. Until then the job is expected to fail, because the service repositories do not provide the endpoints yet.
- **Known blocker for scenario `customer_login` (and any lambda call):** the login lambda connects to PostgreSQL with `sslmode=require` hard-coded (`cmd/lambda/main.go`). The local PostgreSQL does not serve TLS, so the lambda cannot query the database and answers `500`. Options, to be decided: enable TLS on the local PostgreSQL with a development certificate, or make the lambda's `sslmode` configurable (a change in the lambda repository). Not changed here.
- Scenario 5 of the ficha asks for the token to be accepted by `GET /v1/service-order/{id}/status`, but that endpoint is public (no JWT in the service contract). The suite therefore proves token acceptance with `PUT /v1/service-order/{id}/accept`.
- `scripts/clone-services.sh` clones over HTTPS without credentials: the service repositories must be public for the CI job (or the job needs a token).


## Quality gates

- Total coverage at least 80 % (generated mocks excluded); the route matcher at least 90 %. Both are enforced by the `coverage` CI job.
- Lint: golangci-lint in the `pre-commit` job, plus `go vet` and `gofmt`.
- Unit tests run with `-race` in the `unit-tests` job. The `sonarqube` job publishes the analysis.

## Deploy

- `k8s/manifests/base`: Deployment, ClusterIP Service, HPA and the `app-config` ConfigMap with the upstream addresses. The gateway has no secrets, no database and no migrations.
- `k8s/manifests/overlays/lab`: LoadBalancer Service (AWS Academy Learner Lab, no Ingress controller).
- `k8s/manifests/overlays/stg` and `prd`: Ingress (ALB), HPA limits and log level. `components/observability` adds the `ServiceMonitor`.
- `.github/workflows/docker.yml` builds and pushes the image, deploys the overlay and waits for the rollout. On lab it publishes the load balancer host to SSM (`/auto-repair-shop/<env>/app-backend-host`), which the AWS API Gateway reads.

## Layout

```
cmd/service                  entrypoint
config                       embedded routing table (routes.yaml)
internal/app                 HTTP server, bootstrap, Gin middlewares (logger, recovery, metrics)
internal/gateway/config      environment and routing table loading and validation
internal/gateway/match       method and path matcher (no framework dependency)
internal/gateway/proxy       httputil.ReverseProxy per upstream, timeouts, 502/504 mapping
internal/gateway/handlers    gateway endpoints, NoRoute proxying, router
internal/pkg                 env, logger, metrics, app lifecycle, web (error envelope)
docs                         gateway swagger.yaml, documentation index page (embedded)
k8s/manifests                Kustomize base and overlays
e2e                          BDD end-to-end suite (godog), own Go module
local                        ecosystem: postgres-init, init-aws.sh, lambda image and adapter
scripts                      clone-services.sh, wait-ready.sh
docker-compose.yml           local ecosystem (see "Local ecosystem")
```
