# Architecture

The repository is an HTTP reverse proxy. Requests come in, a routing table decides the upstream, and the request is forwarded. The gateway has no business logic, no database and no secrets.

## Packages and their allowed imports

| Package | Responsibility | Framework allowed |
|---|---|---|
| `internal/gateway/match` | Method and path matching against the ordered table | none (standard library only) |
| `internal/gateway/config` | Environment and routing table loading and validation | none |
| `internal/gateway/proxy` | `httputil.ReverseProxy` per upstream, timeouts, 502/504 mapping | `net/http` only |
| `internal/gateway/handlers` | Gateway endpoints, `NoRoute` proxying, router | Gin |
| `internal/app/middleware` | Request id and logging, panic recovery, metrics | Gin |
| `internal/app/bootstrap` | Wiring only: config, matcher, proxy, handlers, server | all of the above |
| `internal/pkg/*` | env, logger, metrics, lifecycle, error envelope | none beyond logging and metrics libraries |
| `config` | Embeds `routes.yaml` | `embed` |

## Rules

- Routing decisions happen in `match` and nowhere else. Handlers call the matcher and do not inspect paths themselves.
- Gin registers only the gateway's own endpoints. Proxied routes are never registered in Gin (see `api.md`).
- The proxy never parses or rewrites bodies. It only adds `X-Forwarded-*` and the request id.
- `bootstrap` contains no decisions: a new behaviour belongs to the package that owns it.
- Shared helpers in `internal/pkg` stay small and free of gateway policy.
