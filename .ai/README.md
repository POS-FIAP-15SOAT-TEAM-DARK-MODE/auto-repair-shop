# AI Rules — Master Index

Single source of truth for all AI agent rules in this project.
**Never duplicate content** in agent pointer files — those files only contain a routing table and 3 always-apply rules.

## Folder Map

```
.ai/
  rules/               # topic-scoped rule files — open on demand
    architecture.md        packages, allowed imports, what the gateway is not
    routing.md             routing table semantics, precedence, invariants
    api.md                 gateway endpoints, error envelope, adding an endpoint
    go-conventions.md      Go idioms, dependencies
    security.md            credentials, logs, blocked paths, error leakage
    observability.md       JSON logs, metrics labels, health endpoints
    testing.md             coverage gates, httptest rules, routing test rows
    workflow.md            Conventional Commits, pre-commit hooks, dev commands
  commands/            # reusable playbooks (also Claude Code slash commands via symlinks)
    add-route.md           expose a new service endpoint through the gateway
  hooks/               # pre-commit hook scripts
    check-ai-rules.sh      enforces .ai/ as the only rule source
```

## Topic Index

| Topic | File | Key contents |
|---|---|---|
| Architecture | `rules/architecture.md` | Package responsibilities, import rules |
| Routing | `rules/routing.md` | Table order, patterns, blocked and optional routes |
| API | `rules/api.md` | Own endpoints, error envelope, status messages |
| Go Conventions | `rules/go-conventions.md` | Stack, idioms, dependency policy |
| Security | `rules/security.md` | Credential forwarding, log redaction, error leakage |
| Observability | `rules/observability.md` | Log fields, metric labels, probes |
| Testing | `rules/testing.md` | Coverage gates, httptest, routing rows |
| Workflow | `rules/workflow.md` | Conventional Commits, pre-commit, dev commands |

## Task Routing — Required Reading by Task Type

| Task type | Files to read |
|---|---|
| New route or routing change | `routing.md` · `architecture.md` · `testing.md` |
| New gateway endpoint | `api.md` · `architecture.md` · `security.md` |
| Proxy or timeout behaviour | `architecture.md` · `security.md` · `observability.md` |
| Tests | `testing.md` · `architecture.md` |
| Logging or observability | `observability.md` · `security.md` |
| Git workflow | `workflow.md` |
| Expose a service endpoint (command) | `commands/add-route.md` |
