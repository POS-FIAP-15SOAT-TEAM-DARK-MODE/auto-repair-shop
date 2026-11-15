# Testing

## Coverage gates (enforced by the `coverage` CI job)

- Total coverage at least 80 %, with generated mocks excluded.
- `internal/gateway/match` at least 90 %.

## Style

- Unit tests with Testify, table-driven where the cases are a list.
- Handler and proxy tests run against a real `httptest.Server`. `httptest.ResponseRecorder` lacks `CloseNotify`, which `httputil.ReverseProxy` requires, so recorder-based tests of the proxy path would panic.
- Upstreams are fake `httptest` servers that record what they received (path, method, request id, headers). Tests assert on what reached the upstream, not on internal state.
- Environment-driven tests use `t.Setenv`. An empty `<NAME>_UPSTREAM` means "default" (or "unset" for the lambda), which tests must account for.
- Every test runs under `-race` (`make test`).

## Required cases for a routing change

- The new or changed line in `config/routes.yaml` has a row in `routes_table_test.go`.
- The matcher's precedence is covered when a new route overlaps an existing one.
- Upstream failures (502, 504) keep their fixed messages.
