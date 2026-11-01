# Observability

## Logs

- One JSON line per request, written when the response completes: `method`, `path`, `ip`, `upstream`, `status`, `duration_ms`, plus `request_id` on the logger.
- Levels: `INFO` for requests and readiness, `WARN` for an upstream that failed `/ready`, `ERROR` for proxy failures and recovered panics. `LOG_LEVEL=PANIC` silences tests.
- Proxy failures log the cause with `upstream` and `status`; the client gets the fixed message.
- Never log the query string or request bodies (see `security.md`).

## Metrics

- `http_request_duration_seconds` histogram with labels `method`, `route`, `upstream`, `status`.
- `route` is the route template from the table (`/v1/works*`), `/docs/*path` for documentation, or `unmatched`. Never the raw path: raw paths would explode the cardinality.
- `upstream` is the target name, `gateway` for the gateway's own endpoints, or `none` when nothing was proxied.

## Health

- `/ping` is liveness and readiness of the gateway process; the Deployment probes use it.
- `/ready` checks the upstreams and is not used as a Kubernetes probe, so one slow service does not restart the gateway.
