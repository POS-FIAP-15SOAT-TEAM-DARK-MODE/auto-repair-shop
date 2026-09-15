# ADR 0004 — Observability stack: self-hosted Prometheus/Grafana/Loki

- Status: Accepted
- Date: 2026-08-23
- Deciders: Auto Repair Shop Team
- Tags: observability, metrics, logs, tech-challenge

## Context

The Fase 3 Tech Challenge requires monitoring API latency, Kubernetes CPU/memory
consumption, healthchecks/uptime, alerts for service-order processing failures,
structured JSON logs with request correlation, and dashboards for daily
service-order volume, average execution time per status, and integration
errors — explicitly citing "ferramentas como Datadog ou New Relic" as the
expected integration.

Both Datadog and New Relic are SaaS-only: the only thing that runs in-cluster
is a lightweight collector agent, and the actual dashboard UI is always
hosted on the vendor's own domain, behind that vendor's account/login — there
is no self-hosted, git-owned instance of either product.

This is a short-lived AWS Academy Learner Lab environment used for study and
for the tech-challenge demo video, not a production account: signing up for
and depending on a third-party SaaS account adds friction (account creation,
API keys as another secret to rotate alongside the Lab's own expiring
credentials) without a matching benefit, since nothing here needs to outlive
the Lab session.

The app already emits structured JSON logs (Zap) with a `request_id` field
threaded through `internal/pkg/logger` and `internal/app/middleware/logger.go`
(see [[0002-observability-and-sensitive-data-protection-assessment]]) — the
request-correlation requirement was already satisfied before this decision;
what was missing was metrics and a place to query the logs.

## Decision

Adopt a self-hosted stack instead, all provisioned as Terraform/Helm in
`auto-repair-shop-infra-k8s`'s `addons` state:

- **kube-prometheus-stack** (Prometheus + Grafana + node-exporter +
  kube-state-metrics) for infra-level CPU/memory (satisfies "Consumo de
  recursos do Kubernetes").
- **loki-stack** (Loki + Promtail) for centralized, queryable logs — the
  app's existing structured JSON logs become searchable in Grafana's Explore
  view, filterable by `request_id`.
- The app exposes its own `/metrics` (Prometheus text format) via
  `internal/pkg/metrics` + a `Metrics()` Gin middleware, and ships a
  `ServiceMonitor` (as a kustomize Component, not `base/` — see Notes) so
  Prometheus discovers it. Counters/histograms are instrumented at the
  service layer's single existing chokepoints (`Create`,
  `emitStatusTransition`), not scattered across handlers:
  - `http_request_duration_seconds` (latency by method/route/status)
  - `service_orders_created_total` (daily volume)
  - `service_order_status_transitions_total{status}` (volume by status)
  - `service_order_notification_failures_total` (integration failures)
  - `service_order_status_avg_duration_hours{status}` — a custom
    `prometheus.Collector` backed by a new read-only query
    (`AverageDurationByStatusInHours`, self-joining
    `service_order_status_history` on itself) satisfying "tempo médio de
    execução por status".
- A custom Grafana dashboard ("Auto Repair Shop — App Metrics"), provisioned
  as a labeled ConfigMap in `infra-k8s`, covers all of the above in one
  place.

## Consequences

### Positive

- No third-party account, API key, or billing dependency — everything lives
  in the cluster and in git (Terraform/Helm/JSON), reviewable in the same PR
  flow as the rest of the infra.
- Metrics/logs/dashboards are reproducible from a clean Lab account by
  re-running `terraform apply`, with no manual "log into Datadog and
  configure X" step.
- `/metrics` and the ServiceMonitor are portable to any Prometheus Operator
  install, not tied to this specific cluster.

### Negative

- Deviates from the tech challenge's literal example tools (Datadog/New
  Relic) — the requirement text allows "escolha livre" for the monitoring
  tool, but this should be called out explicitly when presenting the
  solution, since it is the more visible interpretation gap in this ADR.
- No hosted alerting/on-call integration (Alertmanager is explicitly
  disabled — see the `infra-k8s` addons ADR) and no APM/distributed tracing,
  both of which a SaaS APM product would have given for free.
- Prometheus retention is 6h and Loki has no persistent storage (no EBS CSI
  driver installed) — this stack does not survive a pod restart with its
  history intact, which is fine for a live demo but not for
  longer-term analysis.

## Alternatives considered

- **A. Datadog.** Matches the requirement text exactly, has hosted APM/logs/
  dashboards in one product with a real shareable URL and proper
  auth — the cleanest fit for the video demo. Rejected for now due to the
  SaaS account dependency described above; still the recommended path if a
  team account is set up later (agent Helm chart install is a small,
  additive change on top of this ADR, not a rewrite).
- **B. New Relic.** Same trade-offs as Datadog.
- **C. AWS CloudWatch Container Insights.** Native to the AWS account
  already in use, IAM-based access (no separate login), but does not cover
  APM/custom-metric ergonomics as well as Prometheus, and would still need a
  separate answer for logs correlation. Used instead for the RDS side (see
  the `infra-db` ADR), where it is a strictly better fit since RDS already
  publishes those metrics to CloudWatch natively.

## Notes

- The `ServiceMonitor` lives in `k8s/manifests/components/observability/` as
  a kustomize Component, included only by the `lab`, `stg` and `prd`
  overlays — the `local` (Kind) overlay has no Prometheus Operator CRDs
  installed, so including it in `base/` would break `kubectl apply` there.
- See the sibling ADRs in `auto-repair-shop-infra-k8s` (Grafana exposure,
  Loki/Promtail, Prometheus ServiceMonitor selector) and
  `auto-repair-shop-infra-db` (CloudWatch dashboard for RDS) for the
  infra-side half of this decision.
