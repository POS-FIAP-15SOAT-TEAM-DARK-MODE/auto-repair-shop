# Workflow

## Conventional Commits

Format: `<type>(<scope>): <short description>`, with a body and footers when they add information.

Types: `feat` · `fix` · `refactor` · `test` · `chore` · `docs` · `perf` · `style` · `revert` · `ci`

Scopes: `gateway` · `routing` · `proxy` · `observability` · `k8s` · `ci` · `docs` · `local` · `e2e`

Rules:
- Subject at most 72 characters, lowercase, imperative mood, no trailing period.
- Breaking changes use `!` (`refactor!: remove ...`) and a `BREAKING CHANGE:` footer.
- Small commits by theme. A routing change and its test go in the same commit.

## Pre-commit hooks

Configured in `.pre-commit-config.yaml`:

- `trailing-whitespace`, `end-of-file-fixer`, `check-yaml` (multi-document k8s files), `check-added-large-files`, `check-merge-conflict`
- `golangci-lint` (pinned version)
- `go fmt ./...`, `go vet ./...`, `go build`, `go unit tests`
- `check-ai-rules`: enforces `.ai/` as the only rule source (`.ai/hooks/check-ai-rules.sh`)

## Development commands

```bash
make run          # run the gateway with go run
make test         # unit tests with the race detector
make coverage     # coverage profile, filtered total and HTML report
make check        # fmt, vet and tests
make docker-build # image from the Dockerfile
```

Kubernetes manifests are validated with `kubectl kustomize k8s/manifests/overlays/<lab|stg|prd>`.
