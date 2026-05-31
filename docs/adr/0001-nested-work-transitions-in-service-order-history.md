# ADR 0001 — Nested work transitions in the service-order history endpoint

- Status: Accepted
- Date: 2026-04-25
- Deciders: Giusier F.
- Tags: api, service-order, history

## Context

The endpoint `GET /v1/service-order/{id}/history` returns all rows from `service_order_status_history` for a given SO. Each row records a single status transition.

The system also persists, on a separate table `work_service_order_status_history`, one row per status transition of a Work attached to an SO. By business rule, a Work progresses (its status changes) only when the parent SO transitions to `IN_PROGRESS`. Both tables are written within the same database transaction, but the schema does **not** model a foreign key between them — only `service_order_id`, `work_id`, and `created_at` are stored.

Consumers of the history endpoint (front-end timeline view, audit reports) need to see, for each SO transition, which Work status changes are relevant.

The endpoint is intentionally non-paginated: a service order's lifecycle is bounded (~8 transitions in `service_order_status_history`).

## Decision

Work transitions are **nested** inside the SO transition whose `new_status` is `IN_PROGRESS`. The `work_transitions` field is an array of groups, each containing:
- `work_id` — the UUID of the work
- `status` — the full status history for that work, ordered by `created_at` ASC

All other SO transitions carry no `work_transitions` field (omitted from the JSON).

The service layer:
1. Fetches all SO transitions via `Search`.
2. Skips the work-transition lookup entirely when no SO transition has `NewStatus == IN_PROGRESS`.
3. Otherwise, fetches all work transitions for the SO via `SearchWorkTransitions`, groups them by `work_id` preserving insertion order, and attaches the grouped result to the `IN_PROGRESS` entry.

## Consequences

### Positive

- No schema change, no data migration.
- Response root stays a simple `{ items: [...] }` envelope.
- Grouping by `work_id` makes it trivial for the client to render a per-work timeline.
- Work transition lookup is entirely skipped when the SO has not yet reached `IN_PROGRESS`.
- Two repository round-trips per request at most (`Search` + `SearchWorkTransitions`); both are bounded by the SO lifecycle.

### Negative

- All work transitions are attached to a single SO entry (`IN_PROGRESS`). If a future requirement needs to show work transitions on other SO statuses, this design must be revisited.
- No direct FK between `service_order_status_history` and `work_service_order_status_history`; the attachment relies on the business rule that work only progresses at `IN_PROGRESS`. If this invariant changes, the grouping logic will silently miss transitions.

## Alternatives considered

- **A. Sibling block.** Return a top-level `works` array alongside `items`. Simpler, but does not express the "with this SO transition" coupling that the UI needs.
- **B. Time-bucket heuristic.** Distribute work transitions across SO entries by matching `created_at` windows. More flexible but complex and fragile — the business rule already pins work transitions to `IN_PROGRESS`.
- **C. FK column.** Add `service_order_status_history_id` to `work_service_order_status_history`. Most correct semantically, but requires a schema migration. Deferred — can supersede this ADR if needed.
- **D. Separate sub-resource.** Expose `GET /v1/service-order/{id}/works/history` independently. Does not satisfy the requirement of returning both sets in one response.

## Notes

- This decision affects only the read path. The write path for `work_service_order_status_history` is out of scope.
- `work_transitions` is omitted (not an empty array) on SO entries that are not `IN_PROGRESS`, per the `omitempty` JSON tag.
