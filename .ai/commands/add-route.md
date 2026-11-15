# Playbook: Add a Route

Use this playbook when a new endpoint of a service must be reachable through the gateway.

## 1. Decide the upstream and the position

- Find the service that owns the endpoint. The gateway never decides ownership.
- Find the most specific existing line in `config/routes.yaml` that overlaps. The new line goes above the broader one.
- If the endpoint is internal (`/internal/*`) it must stay blocked. Do not add a route.

## 2. Edit the table

- Add one line with `methods`, `path` and `upstream`. Use `:name` for path parameters and a trailing `*` only for prefixes.
- If the upstream is optional in some environments, use `optional: true`.

## 3. Test

- Add a row to `TestEmbeddedRoutingTable` in `internal/gateway/config/routes_table_test.go` with a representative request and the expected upstream.
- Update `require.Len(t, routes, N)` in `TestRoutesFile_CompilesAndIsWellFormed` when the number of lines changes.
- Run `make test` and `make coverage`. The matcher must stay at 90 % or more.

## 4. Document

- Add the endpoint to the routes table in `README.md`.
- If the service publishes a new swagger document, nothing changes here: `/docs/<service>/swagger.yaml` already proxies it.

## 5. Verify

- `go vet ./...`, `make fmt`, and `kubectl kustomize` of the three overlays when the Kubernetes config changed.
