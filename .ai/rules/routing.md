# Routing

The routing table is the only domain of the gateway. It lives in `config/routes.yaml` and is embedded in the binary.

## Table semantics

- Routes are tried in order; the first one whose methods and path match wins. Order is priority: put the most specific route first.
- `methods: ["*"]` accepts any method. Method comparison is case-insensitive.
- Patterns: literal segments, `:name` for exactly one non-empty segment, and a trailing `*` on the last segment for any suffix. `*` anywhere else is an error.
- `block: true` answers 404 without proxying. `/internal/*` must always be blocked, first in the table.
- `upstream` names are fixed: `customer`, `catalog`, `billing`, `execution`, `history`, `service_order`, `lambda`. Each one reads `<NAME>_UPSTREAM`.
- `optional: true` drops the route when its upstream is not configured. Without `optional`, an unconfigured upstream stops the gateway at startup.

## Invariants

- Every line of the table has a test in `internal/gateway/config/routes_table_test.go` that asserts the chosen upstream. Adding a line means adding a row.
- The `/docs/<service>/swagger.yaml` endpoint serves only the names in `config.Services`, never the lambda.
- Paths containing `.` or `..` segments never match.
- Ownership of an endpoint is decided by the service that implements it. The gateway never guesses; a new endpoint is a new line plus a test.
