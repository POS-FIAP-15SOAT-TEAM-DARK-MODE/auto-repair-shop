# ADR 0004 — Publish the app's LoadBalancer hostname to SSM for cross-repo handoff

- Status: Accepted
- Date: 2026-08-22
- Deciders: Giusier F.
- Tags: infra, cross-repo, api-gateway

## Context

`auto-repair-shop-infra-k8s`'s new API Gateway (issue #5) needs to proxy non-Lambda routes to this app's public endpoint. In Learner Lab mode the app is exposed via a Kubernetes `LoadBalancer` Service (`k8s/manifests/overlays/lab/patch-service-loadbalancer.yaml`), which provisions a Classic ELB through Kubernetes' in-tree AWS cloud provider — not through any Terraform `apply`.

This means there is no Terraform state, in this repo or any sibling infra repo, that owns that ELB. `terraform_remote_state` — the mechanism already used to hand off `db_host`, `app_secret_arn`, `function_arn`, etc. between the 4 repos — has nothing to read here, because nothing in Terraform created the resource. A raw AWS API lookup by tag was also considered and rejected: Classic ELBs get a random Kubernetes-generated name, and Terraform's `aws_elb` data source only supports lookup by exact name (unlike `aws_lb` for ALB/NLB, which supports tag filtering).

## Decision

The `deploy` job in `docker.yml`, once it resolves the LoadBalancer's hostname (existing "Publish app URL" step), also writes it to AWS Systems Manager Parameter Store at `/auto-repair-shop/<env>/app-backend-host` (String, `--overwrite`). `auto-repair-shop-infra-k8s`'s `gateway` Terraform state reads it back via an `aws_ssm_parameter` data source.

The `deploy-stg`/`deploy-prd` IAM roles (created in `infra-k8s`'s `shared` state, only when `manage_iam = true`) were granted `ssm:PutParameter` scoped to that parameter path. Under Learner Lab's `manage_iam = false` mode, this is a no-op since `LabRole` already has broad SSM access.

## Consequences

### Positive

- No manually-set, easily-stale GitHub repo variable (the earlier working version of the gateway PR required a human to copy the ELB hostname out of a deploy summary and paste it into a repo variable by hand — this fully automates that handoff).
- Every deploy self-updates the parameter, so a re-run of the gateway's `terraform apply` after any app redeploy (which can get a new ELB, e.g. after Service recreation) picks up the current hostname automatically.
- SSM Parameter Store is free at this tier and needs no additional infrastructure.

### Negative

- Introduces an implicit *deploy-order* dependency: the `gateway` layer's first `apply` will fail (parameter not found) until the app has deployed at least once. Documented in `infra-k8s`'s README, not enforced by tooling.
- The parameter is a plain `String`, not tied to any resource lifecycle — if the app's Kubernetes Service is deleted without a subsequent deploy, the parameter goes stale until the next deploy overwrites it. Acceptable at this deployment's current scale; would want a cleanup/expiry story for anything longer-lived.

## Alternatives considered

- **A. Manual GitHub repo variable (`APP_BACKEND_HOST`).** The original implementation. Works, but requires a human step on every ELB recreation — the exact kind of drift-prone manual handoff this ADR exists to remove.
- **B. Migrate the Service to an NLB via annotation**, then have the gateway look it up via `data "aws_lb"` with a tag filter (works for ALB/NLB, not Classic ELB). Rejected for now: it's a real change to the app's Kubernetes manifests, and doesn't obviously improve on SSM's simplicity for this project's scope. Worth reconsidering if the AWS Load Balancer Controller add-on becomes available (it's currently skipped under Learner Lab's IRSA restrictions).
- **C. Terraform-manage the Kubernetes Service itself** (via the `kubernetes`/`kubectl` Terraform provider from `infra-k8s`), making the ELB a real Terraform-owned resource with a real `terraform_remote_state` output. Rejected as too large a change for this decision's scope — the app's Kubernetes manifests are deliberately applied via plain `kubectl` in this repo's own CI, not Terraform, and changing that ownership model is a separate, bigger architectural decision.

## Notes

- Companion decision in `auto-repair-shop-infra-k8s`'s ADR on API Gateway topology — that ADR covers *why* the gateway proxies to this hostname at all (public HTTP proxy vs. private VPC Link); this ADR covers only *how* the hostname itself gets from Kubernetes into Terraform's hands.
