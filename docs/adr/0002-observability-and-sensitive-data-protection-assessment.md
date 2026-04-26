# ADR 0002 — Observability and sensitive-data protection assessment

- Status: Accepted
- Date: 2026-04-26
- Deciders: Auto Repair Shop Team
- Tags: observability, security, logs
- Related issues: #236, #223

## Context

A repository review was requested to verify compliance with two non-functional requirements:

- #236 — structured JSON logs for service-order creation, status transitions, and validation errors, including timestamp, level, and context.
- #223 — ensure CPF, CNPJ, and passwords do not appear in application logs or API error messages.

The codebase uses Go + Gin + Zap, with a global JSON logger and an HTTP middleware for request tracing.

The most relevant components are:
- Central logger: `internal/pkg/logger/logger.go`
- Request logging middleware: `internal/infra/http/middleware/logger.go`
- Main service-order flow: `internal/services/service_order/service.go`
- Service-order status-history persistence: `internal/infra/repository/service_order/postgres.go`
- HTTP error mapping: `internal/pkg/web/errors.go`
- Handlers with payload debug logging: `internal/infra/handler/*`
- Repository AI guidance files: `CLAUDE.md` and `.github/copilot-instructions.md`

## Decision

1. Consider the previously identified gaps for issues #236 and #223 as implemented in the current code.
2. Formalize the adopted standards:
   - structured service-order events in the service layer;
   - centralized sensitive-data sanitization in the logger package.
3. Keep the rule that raw payload/domain-structure logging must be avoided when sensitive data may be present.

## Consequences

### Positive

- Standardized observability across the service-order lifecycle.
- Reduced risk of sensitive-data leakage through logs.
- Better operational traceability for creation, validation, and status transitions.

### Negative

- New logs must continuously follow the same event and field conventions, or consistency will degrade.
- Future changes in handlers and repositories can reintroduce sensitive-data exposure if raw payload logging returns.

## Alternatives considered

- **A. Keep only ad hoc logs.** Lower effort, but no stable event taxonomy and weaker traceability.
- **B. Sanitize at call sites only.** More localized control, but high chance of missing sensitive fields and inconsistent behavior.
- **C. Remove most logs in critical flows.** Reduces leakage risk, but harms debugging and auditability.
- **D. Introduce strict schema-level masking only.** Useful at storage boundaries, but insufficient for in-process log payloads.

## Notes

- API error mapping was kept as implemented, without direct exposure of password, CPF, or CNPJ in default messages.
- Recommended follow-up: add automated logging/sanitization tests and extend the `zap.Any` review to non-critical modules.
