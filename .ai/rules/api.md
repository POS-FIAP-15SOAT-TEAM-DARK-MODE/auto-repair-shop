# API

The gateway exposes its own endpoints and proxies the business API of the services.

## Own endpoints

| Method and path | Behaviour |
|---|---|
| `GET /ping` | `200 {"message":"pong"}` |
| `GET /ready` | `200 {"message":"ready"}`, or `503 {"code":503,"errors":["upstreams not ready"],"upstreams":[names]}` |
| `GET /metrics` | Prometheus exposition |
| `GET /swagger.yaml` | Gateway OpenAPI document (`docs/swagger.yaml`) |
| `GET /swagger/index.html` | Static index page (`docs/index.html`) |
| `GET /docs/<service>/swagger.yaml` | Proxied OpenAPI document of one documented service |

Any other path or method goes through the routing table. Unmatched or blocked requests answer `404 {"code":404,"errors":["route not found"]}`.

## Error envelope

Every error the gateway writes uses `internal/pkg/web`: `{"code": <status>, "errors": [<message>...]}`. Upstream responses are not rewritten, so the services' own envelopes reach the client as they are.

| Status | Message | When |
|---|---|---|
| 404 | `route not found` | no route matched, or a blocked route |
| 502 | `upstream unavailable` | connection refused, DNS failure, reset |
| 504 | `upstream timeout` | no response headers within `UPSTREAM_TIMEOUT` |
| 500 | `internal server error` | a panic in the gateway (logged with stack on the server) |

## Adding an endpoint to the gateway

Gateway endpoints are rare. Before adding one, check that it cannot be a route in `config/routes.yaml`. If it must be one, register it explicitly in `internal/gateway/handlers/router.go`, document it in `docs/swagger.yaml`, and add a handler test.
