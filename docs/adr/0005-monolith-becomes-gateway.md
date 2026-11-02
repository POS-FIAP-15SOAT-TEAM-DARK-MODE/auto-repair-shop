# ADR 0005 — The monolith repository becomes the API gateway

- Status: Accepted
- Date: 2026-11-01
- Tags: architecture, gateway, routing, microservices

## Context

The monolith split into six services, each in its own repository, with its own database, CI and deploy. Clients still expect one API under one host. The AWS API Gateway in front of the cluster already routes `POST /auth/customer-login` to the login Lambda, and it forwards everything else to the host that this repository publishes to SSM (`/auto-repair-shop/<env>/app-backend-host`, ADR 0004).

So the public entry point must:

- send each `/v1/...` request to the service that owns it;
- never expose `/internal/*` (service-to-service REST);
- keep the existing error format and the request id;
- not hold any domain logic, database or secret.

## Decision

This repository keeps its name and its deploy target and becomes the **gateway**: a reverse proxy written in Go with Gin and `net/http/httputil.ReverseProxy`.

- **Routing table.** `config/routes.yaml` is embedded in the binary and read in order: the first route whose method and path match wins. Patterns are literal segments, `:param` for one segment, and a trailing `*` for any suffix. Upstreams come from `<NAME>_UPSTREAM` environment variables, with in-cluster defaults.
- **Own matcher.** Routes are resolved by `internal/gateway/match`, not by Gin's router. Gin only registers the gateway's own endpoints (`/ping`, `/ready`, `/metrics`, `/swagger*`, `/docs/*`). Every other request goes to `NoRoute` and then to the matcher. This avoids Gin's wildcard conflicts between `:id` and `*` and keeps the rules testable without HTTP.
- **Blocked paths.** `/internal/*` answers 404 and is never forwarded.
- **No authentication in the gateway.** It does not validate JWTs and does not touch a database. Each service validates its own tokens, and the request's `Authorization` header is forwarded unchanged. The gateway therefore holds no secrets and needs no migration or RBAC in Kubernetes.
- **Failure mapping.** An upstream that does not answer its headers within `UPSTREAM_TIMEOUT` gives 504. An unreachable upstream gives 502. Both bodies are fixed messages (`upstream timeout`, `upstream unavailable`), and the cause is logged on the server only. Upstream responses, including 4xx and 5xx, pass through unchanged.
- **Request id.** `X-Request-Id` is accepted when it is short and made of safe characters, otherwise generated. It is returned to the client, forwarded to the upstream and written in every log line.
- **Readiness.** `/ready` checks each configured upstream's `/ping` in parallel and answers 503 with the names of the failing ones, never their URLs.
- **Documentation.** `/docs/<service>/swagger.yaml` proxies each service's OpenAPI document, and `/swagger/index.html` links to them.

### Alternatives considered

- **Routing only in the AWS API Gateway, one integration per route.** Keeps the gateway out of the cluster, but every route change becomes an infrastructure change, the lab account cannot manage that API definition, and the internal-path block would depend on a second place to keep in sync. The AWS API Gateway stays in front of this service and does not replace it.
- **Ingress-NGINX or Envoy/Kong.** Capable, but adds a new component to operate and to learn for a service whose whole job is a short table of prefixes. The matcher in this repository is a single small package with full test coverage.
- **Gin's `router.Any` with one group per upstream.** Conflicts between parameters and wildcards on the same prefix made the routing order hard to reason about. A flat ordered table is easier to read in review.

## Consequences

- The routing table is the single place to change where an endpoint lives. A new route is a line in `config/routes.yaml` plus a test row.
- A malformed table or an invalid upstream stops the gateway at startup instead of misrouting requests.
- The gateway is a single point in the request path. It is stateless and scales horizontally (HPA in `k8s/manifests/base`), and it has no database to fail.
- Clients see one host and one request id, but a request crosses two hops, so latency includes the proxy. Timeouts are bounded by `UPSTREAM_TIMEOUT`.
- The gateway does not know which tokens are valid. A request with an invalid token reaches the service and is rejected there with 401, which is the behaviour the end-to-end scenarios check.
- The monolith code leaves this repository. Its last state stays available under the tag `monolith-v1`.
