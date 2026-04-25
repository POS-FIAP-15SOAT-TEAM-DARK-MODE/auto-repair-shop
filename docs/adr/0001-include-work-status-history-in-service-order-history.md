# ADR 0001: Include per-work status history in the Service Order history endpoint

- Status: Accepted
- Date: 2026-04-25
- Deciders: backend team
- Scope: `GET /v1/service-order/{id}/history`

## Context

`GET /v1/service-order/{id}/history` currently returns a paginated list of status transitions read from `service_order_status_history` (one row per service-order status change).

Each service order also has zero or more works attached to it (`service_order_work`), and the project tracks per-work status transitions in the dedicated table `work_service_order_status_history`. Per the existing business rule, work transitions are produced when the service order enters `IN_PROGRESS`.

Consumers of the history endpoint (the SO detail screen) need to render the SO timeline and the per-work timelines side by side. Today they would have to call a second endpoint to get the work data, and one does not yet exist.

## Decision

Augment the existing response with a non-paginated `works` field placed directly after `items`. Each entry carries the work id and the full per-work timeline. The current status is intentionally omitted: it is trivially derivable from the last entry of `history` and was redundant on the wire. The work `name` is also omitted to keep this endpoint focused on status data; consumers that need the catalog name resolve it from the `work` catalog or from the SO detail endpoint.

The existing paginated `items` array, which represents the SO-level transitions, is preserved unchanged. Pagination still applies only to `items`. Works with no transitions yet appear as entries with an empty `history` array.

```json
{
  "items": [ ... SO transitions ... ],
  "works": [
    {
      "work_id": "wrk_123",
      "history": [
        { "previous_status": "PENDING", "new_status": "IN_PROGRESS", "created_at": "..." }
      ]
    },
    {
      "work_id": "wrk_456",
      "history": []
    }
  ],
  "total_items": 5, "total_pages": 1, "page": 1, "page_size": 10
}
```

The repository gains a single read-only method `WorkTimelineByServiceOrderID(ctx, serviceOrderID)` that runs one SQL query joining `service_order_work` LEFT JOIN `work_service_order_status_history`. The `work` catalog is no longer joined since the response no longer carries the work `name`. The service composes the query concurrently with the existing count/search using `errgroup.WithContext` so a failure in any branch cancels the in-flight queries.

## Considered alternatives

- **Option A (chosen) — augmented response.** Adds a sibling `works` field. No breaking change to the existing `items` shape. Mixes paginated and non-paginated data in one envelope, which is acceptable because the number of works on a single SO is bounded.
- **Option B — unified chronological timeline.** Replace `items` with a single chronologically-sorted, polymorphic stream interleaving SO and work transitions. Best for an audit-log UI but requires a breaking change to the response shape and weaker typing on the wire.
- **Option C — nested under SO transitions.** Anchor work transitions under the matching SO transition (likely the `IN_PROGRESS` one). The "association" is fuzzy because the schema does not store an FK from `work_service_order_status_history` to a specific `service_order_status_history` row, so this would be heuristic and would drift over time without a migration adding the FK.
- **Option D — separate sub-resource endpoint.** Expose `GET /v1/service-order/{id}/works/history` as its own paginated endpoint. Cleanest REST design but requires the consumer to make two calls to render the combined view, and does not satisfy the "same response" requirement.

## Consequences

### Positive

- No breaking change to the existing `items[]` shape; existing UI clients continue to work.
- Single round-trip for the consumer; the SO detail screen renders both timelines from one API call.
- Strongly typed, easy to extend (e.g., re-add per-work `name` or `current_status` later if a consumer needs them).
- Reuses existing index `idx_status_history_work_service_order_id` — no DB migration required.

### Negative / trade-offs

- The response envelope mixes paginated and non-paginated data.
- Payload size grows by `O(works × transitions per work)` per request; works on an SO are bounded but extreme cases are not paginated.
- Adds one extra repository round-trip; mitigated by running it concurrently with the existing count/search via `errgroup.WithContext`.

### Follow-ups (out of scope)

- If we later need to know which SO transition triggered which work transition, add `service_order_status_history_id` as a FK on `work_service_order_status_history` via a new migration and revisit Option C.
- If the `works` list ever needs pagination, introduce Option D as a separate dedicated endpoint rather than paginating this field in place.
