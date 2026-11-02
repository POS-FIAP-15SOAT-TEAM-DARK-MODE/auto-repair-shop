# Go Conventions

## Stack

Go 1.26 · Gin (own endpoints and middlewares only) · `net/http/httputil.ReverseProxy` · `gopkg.in/yaml.v3` · Zap (JSON logs) · `prometheus/client_golang` · Testify · `google/uuid` (request ids).

## Idioms

- Errors are returned, wrapped with `%w` where the caller may test them, and never include configured URLs or credentials.
- Configuration is read once, at startup. Invalid configuration is an error returned from `bootstrap`, which `main` logs before exiting.
- Packages under `internal/gateway/` do not share state through globals. Dependencies are passed to constructors (`handlers.New`, `proxy.New`, `match.New`).
- Context values carry the request-scoped logger (`internal/pkg/logger`) and nothing else.
- `go vet`, `gofmt` and golangci-lint must be clean. Fix errcheck findings (for example `resp.Body.Close()`) instead of ignoring them.

## Dependencies

Keep the module small. A new dependency needs a reason that the standard library cannot cover in a few lines. Run `go mod tidy` after removing code.
