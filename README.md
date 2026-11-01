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

Local end-to-end run with every service and the messaging stack is part of the local ecosystem planned for this repository (see the phase 4 plan). Until it lands, the gateway is validated with its unit and HTTP tests.

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
```
