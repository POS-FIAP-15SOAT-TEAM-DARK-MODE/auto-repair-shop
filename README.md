# Auto Repair Shop

Backend for an auto repair shop management system, organized by domain following a clean/hexagonal architecture. Manages service orders (SO), customers, vehicles, parts/stock, and administrative operations.

**Stack:** Go · Gin · PostgreSQL · JWT · Swagger

> This is one of 4 repositories that make up the system (Tech Challenge Fase
> 3). This repo holds the application; infrastructure lives in its own repos:
> - [auto-repair-shop-infra-k8s](https://github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-infra-k8s) — Kubernetes cluster (EKS/VPC/add-ons)
> - [auto-repair-shop-infra-db](https://github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-infra-db) — managed database (RDS)

## Domain

- **SO (Service Order)** — Service order, the central aggregate
- **Customer** — Customer identified by CPF (individual) or CNPJ (company)
- **Vehicle** — Vehicle (plate, brand, model, year), linked to a Customer
- **Work** — Billable service item (name, description, unit price), exposed under `/v1/works`
- **Part/Supply** — Part or supply with stock quantity and unit price
- **Budget** — Budget, auto-calculated from services and parts on a SO
- **Stock** — Stock/inventory of Parts

## SO Status Lifecycle

```
NEW → RECEIVED → IN_DIAGNOSIS → AWAITING_APPROVAL → IN_PROGRESS → COMPLETED → DELIVERED
  └──────────────────────────────────────────────────────────────────────────→ CANCELLED
                               AWAITING_APPROVAL → REJECTED
```

`REJECTED` is a terminal state reached when the customer rejects the budget (from `AWAITING_APPROVAL`).
`CANCELLED` is a terminal state available from cancelable statuses.

## Work Status Lifecycle (within a Service Order)

Each work item linked to a service order follows its own lifecycle:

```
AWAITING_START → IN_PROGRESS → COMPLETED
                     ↓               ↓
                 CANCELLED       CANCELLED  (can cancel from any non-terminal status)
```

| Status           | Description                            |
|------------------|----------------------------------------|
| `AWAITING_START` | Work linked to the SO, not yet started |
| `IN_PROGRESS`    | Work is being executed by the mechanic |
| `COMPLETED`      | Work was successfully finished         |
| `CANCELLED`      | Work was cancelled (terminal state)    |

Every status change is recorded as an immutable entry in `work_service_order_status_history`, preserving the full audit trail.

## Project Structure

Each business domain is a self-contained package under `internal/`, wired together
by `internal/app`. Shared, domain-agnostic helpers live in `internal/pkg`.

```
cmd/
  service/              # Application entrypoint
internal/
  app/                  # Application wiring & HTTP layer
    bootstrap/          # Startup sequence
    container/          # Handler & middleware containers
    middleware/         # Auth, logger, recovery
    routing/            # Route definitions
    server.go, db.go, seed.go, ...
  auth/                 # Domain package (auth, customer, supply, vehicle,
  customer/             #   work, service_order, service_order_history)
  supply/               #   Each contains:
  vehicle/              #     domain/      — models & errors (framework-free)
  work/                 #     interfaces/  — service/repo/controller ports + mocks
  service_order/        #     adapters/    — request/response DTOs
  service_order_history/#     repository/  — postgres + in-memory implementations
  ...                   #     controller.go, service.go, di.go
  integration/          # Integration test suite
  pkg/                  # Shared utilities (db, uow, logger, web, auth, env, ...)
```

## Architecture

The same Go workload runs in three interchangeable ways: **(1)** locally with just
Docker Compose (no Kubernetes), **(2)** on a local **Kind** cluster that simulates
the full Kubernetes behavior end-to-end (no AWS needed), and **(3)** on **AWS
(EKS + RDS)** provisioned by Terraform. Paths (2) and (3) apply the *same*
Kustomize manifests — only the overlay changes.

### Application components

```mermaid
flowchart LR
    client([HTTP client])

    subgraph ns["Kubernetes namespace: auto-repair-shop"]
        app["Deployment<br/>auto-repair-shop (Go / Gin)<br/>probes on /ping"]
        svc["Service<br/>NodePort local · LoadBalancer lab · ClusterIP+Ingress stg/prd"]
        cm[["ConfigMap<br/>DB host/port, non-sensitive config"]]
        sec[["Secret<br/>POSTGRES_PASSWORD · JWT_SECRET"]]
        hpa["HPA<br/>1→5 · 70% CPU / 80% mem"]
        job["Job: db-migrate<br/>waits for DB → golang-migrate up"]

        svc --> app
        cm -.envFrom.-> app
        sec -.envFrom.-> app
        hpa -.scales.-> app
    end

    db[("PostgreSQL")]
    client -->|":8080 REST · JWT · Swagger"| svc
    app -->|"SQL :5432"| db
    job -->|"applies schema"| db
```

### Provisioned infrastructure (Terraform — AWS)

> **This repository no longer contains the Terraform for AWS.** As part of
> Tech Challenge Fase 3's 4-repository requirement, provisioning moved to two
> sibling repos:
> - [**auto-repair-shop-infra-k8s**](https://github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-infra-k8s) — `bootstrap` (state bucket), `shared` (ECR, GitHub OIDC, IAM roles), `aws` (VPC, EKS), `addons` (ALB Controller, metrics-server, External Secrets).
> - [**auto-repair-shop-infra-db**](https://github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-infra-db) — RDS PostgreSQL + Secrets Manager, placed into the VPC provisioned by `infra-k8s` (read via `terraform_remote_state`).
>
> This repo only builds and deploys the application **onto** infra those repos
> already provisioned — see [Deploy the app](#deploy-the-app) below.

The diagram below is still the accurate deployed topology — it is now just
split across three repositories instead of one:

```mermaid
flowchart TB
    gha["GitHub Actions"]
    inet([Internet])

    subgraph aws["AWS account"]
        iam["IAM roles"]
        ecr["ECR<br/>app images"]
        sm["Secrets Manager<br/>app secret"]

        subgraph vpc["VPC"]
            subgraph pub["public subnets"]
                igw["IGW / NAT"]
                alb["ALB (public)"]
            end
            subgraph priv["private subnets"]
                eks["EKS managed node group<br/>app pods"]
                rds[("RDS PostgreSQL<br/>multi-AZ in prd")]
            end
            alb --> eks
            eks -->|":5432 · SG: EKS nodes only"| rds
        end

        sm -->|"External Secrets Operator"| eks
        ecr -.image pull.-> eks
    end

    gha -->|"OIDC (no static keys)"| iam
    inet --> alb
```

Locally, the `overlays/local` Kustomize overlay stands in for all of this: an
in-cluster PostgreSQL replaces RDS, a NodePort replaces the ALB, and a static
Secret replaces Secrets Manager — so Kind reproduces the AWS topology with zero
cloud dependency.

### Deploy flow (GitHub Actions)

This repo's CI/CD only builds and deploys the app image — Terraform
plan/apply for the cluster and the database run in their own workflows, in
the [infra-k8s](https://github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-infra-k8s)
and [infra-db](https://github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-infra-db)
repos.

```mermaid
flowchart TB
    pr["Pull Request"] --> ci["ci.yml<br/>lint · build · unit · integration · SonarCloud"]

    dev["push develop"] --> docker
    main["push main"] --> docker
    docker["docker.yml"] --> stg{{"develop → STG"}}
    docker --> prd{{"main → PRD"}}
    stg --> steps
    prd --> steps
    steps["build image → push ECR → kubectl apply -k overlay<br/>→ set image (rollout) → wait for rollout<br/>(AWS auth via OIDC · migrations synced as ConfigMap)"]
```

## Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- Make

## Contributing

To contribute to the project you should install pre-commit:

```bash
# 1. Make sure you have Python installed
python3 --version # or (on Windows) python --version

# 1.1 If you do not have Python installed:
brew install python # or (on Windows) download from https://www.python.org/downloads/

# 2. Install pre-commit using pip
pip install pre-commit

# 3. Install the pre-commit hooks for this repository
pre-commit install
```

## Getting Started

There are three ways to run this project — pick the one that fits what you want to do:

| # | Path | Needs | Use when | Jump to |
|---|---|---|---|---|
| 1 | **Local (Docker Compose)** | Docker only, **no Kubernetes/infra** | Run the app + DB fast, for development | [Setup with Docker](#2-setup-with-docker) |
| 2 | **Kubernetes on Kind** | Docker only (toolbox) | Exercise the *full* k8s workload locally (Deployment, HPA, migrate Job, Service) — a faithful stand-in for AWS | [Full environment on Kind](#bring-up-the-full-environment-locally-docker-only) |
| 3 | **AWS (real deploy)** | Cloud infra already provisioned by [infra-k8s](https://github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-infra-k8s) + [infra-db](https://github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-infra-db) | Deploy the app for real, onto an existing EKS cluster + RDS | [AWS (stg / prd)](#aws-stg--prd--driven-from-github) |

> Paths 2 and 3 deploy the **same** Kustomize manifests — Kind locally reproduces the
> AWS topology, so you can validate the entire Kubernetes behavior without any cloud cost.

### 1) Environment Variables

Create a local `.env` file before running the project:

```bash
cp .env.example .env
```

### 2) Setup with Docker

```bash
# Start all services (app, migrations, and database)
make docker-up
```

API URL: `http://localhost:8080`

### 3) Local Development (app outside Docker)

Use this flow when you want to run only the database in Docker and the app directly with Go:

```bash
# Start only PostgreSQL
# (in a separate terminal, from repository root)
docker-compose up -d db

# Install migration tool (first time only)
make migrate-install

# Run all pending migrations
make migrate-up

# Start the application locally
make run
```

### Useful Commands

```bash
# Run tests
make test

# Generate coverage report
make coverage

# Stop Docker containers
make docker-down
```

The server starts on port `${PORT:-8080}` (default: 8080).

## API Collection (Swagger UI)

Once the app is running locally (`make docker-up` or `make k8s-up`), the interactive
API collection is served by the built-in Swagger UI — use it to browse and execute
every endpoint directly against the running server:

**➡️ http://localhost:8080/swagger/index.html**

The UI loads the OpenAPI specification from `/swagger.yaml` (source: `docs/swagger.yaml`)
and lets you fire requests without any external tool (Postman/Insomnia not required).

## Database Migrations

Migrations are managed using `golang-migrate` and stored in the `/migrations` directory. Each migration consists of `.up.sql` (apply) and `.down.sql` (rollback) files.

### Migration Commands

```bash
# Install golang-migrate CLI (first time only)
make migrate-install

# Apply all pending migrations
make migrate-up

# Rollback the last applied migration
make migrate-down

# Check current migration version
make migrate-status
```

When adding schema changes, create migration files following the naming convention:
```
NNNNNN_description.up.sql
NNNNNN_description.down.sql
```

Where `NNNNNN` is a sequential 6-digit number (e.g., `000002_add_customer_status.up.sql`).

## Infrastructure (`k8s/`)

`k8s/` now holds only the app's **workload** — the Kubernetes cluster and the
database are provisioned by the sibling
[infra-k8s](https://github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-infra-k8s)
and [infra-db](https://github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-infra-db)
repos (Terraform config lives there, not here).

- **`k8s/manifests/`** — the app's Kubernetes workload (Deployment, Service,
  ConfigMap, Secret, HPA, migration Job) as Kustomize overlays.
- **`k8s/kind/`** — local Kind cluster config, for the Docker-only full-stack
  local environment below (never touches AWS).
- **`k8s/toolbox/`** — Docker-only runner (`kind`/`kubectl` in a container) that
  builds the app image and deploys it onto the local Kind cluster.

```
k8s/
├── manifests/
│   ├── base/        # Deployment, Service, ConfigMap, HPA, migrate Job
│   └── overlays/
│       ├── local/   # + static Secret + in-cluster Postgres, NodePort, dev values
│       ├── lab/     # AWS Academy Learner Lab: LabRole, public ELB, static Secret
│       ├── stg/     # RDS host, ECR image, ALB Ingress, External Secrets, HPA 2-6
│       └── prd/     # same as stg with prod values (HPA 3-10)
├── kind/            # local Kind cluster config
└── toolbox/         # Docker-only runner (kind/kubectl in a container)
```

### Bring up the full environment locally (Docker-only)

**The only requirement on your machine is Docker.** `kind` and `kubectl` run
inside a toolbox container that drives the host Docker daemon
(Docker-outside-of-Docker) — nothing else is installed. Three commands:

```bash
make k8s-up      # Kind cluster -> build+load image -> deploy -> smoke -> metrics-server (HPA)
make k8s-down    # tear everything down (delete the Kind cluster)
make k8s-shell   # shell inside the toolbox for anything else
```

After `make k8s-up` the app is reachable from the host:

```bash
curl http://localhost:8080/ping   # -> {"message":"pong"}
```

Anything else is plain `kubectl`/`kind` from the toolbox shell:

```bash
make k8s-shell
# then, inside the container:
kubectl -n auto-repair-shop get pods,svc,job,hpa
kubectl -n auto-repair-shop logs -l app=auto-repair-shop -f
```

How it works: with the Docker socket mounted, Kind and the app image land on the
host daemon. The toolbox joins the `kind` network and reaches the cluster by the
control-plane container name (works on macOS, where `--network host` is
unreliable), while the `30080 -> localhost:8080` mapping lets you hit the app
from the host. Locally Postgres runs in-cluster (stand-in for RDS), so there is
no AWS dependency at all.

### AWS (stg / prd) — driven from GitHub

No AWS tooling is needed on your machine; the app is deployed by GitHub
Actions in **this** repo, but it needs an already-provisioned cluster and
database. Provisioning (bootstrap, VPC/EKS, add-ons, RDS) lives entirely in
the two infra repos — follow their READMEs first:

1. [**auto-repair-shop-infra-k8s**](https://github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-infra-k8s#readme) —
   one-time bootstrap (state bucket, GitHub OIDC, ECR), then `aws` (VPC+EKS)
   and `addons` (ALB Controller, metrics-server, External Secrets) per
   environment (`stg`/`prd`).
2. [**auto-repair-shop-infra-db**](https://github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-infra-db#readme) —
   RDS PostgreSQL, applied after step 1 (it reads the VPC from infra-k8s's
   state).

**GitHub config in *this* repo** (Settings → Environments `STG` and `PRD`):

| Variable | Source |
|---|---|
| `AWS_REGION` | `us-east-1` |
| `AWS_DEPLOY_ROLE_ARN` | infra-k8s `shared` output `deploy_role_arns` (STG / PRD) |
| `EKS_CLUSTER_NAME` | infra-k8s `aws` output `cluster_name` (per workspace) |
| `RDS_HOST` | infra-db output `db_host` (per workspace) |

### Deploy the app

Push to `develop` (→ STG) or `main` (→ PRD): the Docker workflow builds,
pushes to ECR and applies the matching Kustomize overlay onto the cluster
provisioned by the infra repos.

#### Restricted accounts (AWS Academy Learner Lab)

On a Learner Lab account, the *provisioning* side (IAM/OIDC skipped, `LabRole`
reused, no ALB Controller/External Secrets) is handled entirely in the two
infra repos — see their READMEs. On the **deploy** side (this repo),
`docker.yml` auto-detects lab mode from `aws sts get-caller-identity` (a
`voclabs`/`LabRole` caller) and, when active:
- applies the `lab` overlay (public ELB, `kubectl port-forward`-friendly)
  instead of the ALB `Ingress`;
- creates the app `Secret` directly from Secrets Manager (no External Secrets
  Operator in lab mode) instead of relying on the synced one.

Add secrets `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` / `AWS_SESSION_TOKEN`
to **this** repo too (the lab's temporary credentials — refresh them each
session, they expire) so `docker.yml` can authenticate. `AWS_AUTH_MODE` /
`MANAGE_IAM` / `EXECUTION_ROLE_ARN` are still honoured as explicit repo
variable overrides if you need them.

**Recovery after a lab restart** — if the lab stops/starts, nodes cycle and
CoreDNS can be stranded on a dead node, breaking DNS. Reschedule it and restart
the app:

```bash
kubectl -n kube-system rollout restart deploy/coredns
kubectl -n auto-repair-shop rollout restart deploy/auto-repair-shop
```

**Tear down** the cluster/database — see the *Tear down* sections in the
[infra-k8s](https://github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-infra-k8s#readme)
and [infra-db](https://github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-infra-db#readme)
READMEs — delete the app's `LoadBalancer`/`Ingress` Service first from this
side if one was created, so it doesn't block VPC deletion:

```bash
aws eks update-kubeconfig --region us-east-1 --name auto-repair-shop-stg-eks
kubectl -n auto-repair-shop delete svc auto-repair-shop --ignore-not-found
```

### Accessing the app on AWS

The app is exposed by an **ALB**, created automatically by the AWS Load Balancer
Controller from the `Ingress` (the controller is installed by the `addons`
state). Terraform provisions the VPC routing (route tables, IGW, NAT) and installs
the controller; the ALB itself is created at deploy time from the Ingress.

```bash
make k8s-shell        # or use your own kubectl against the cluster
aws eks update-kubeconfig --region us-east-1 --name auto-repair-shop-stg-eks
kubectl -n auto-repair-shop get ingress auto-repair-shop
# ADDRESS = k8s-autorepa-....elb.amazonaws.com  <- the ALB DNS name
```

The Ingress has a `host:` rule (`auto-repair-shop-stg.example.com`), so the ALB
only routes requests carrying that Host header. To actually reach it either:
- point a DNS record (Route 53) for that host at the ALB DNS name, then browse
  `http://auto-repair-shop-stg.example.com/ping`; or
- for a quick test, send the header: `curl -H 'Host: auto-repair-shop-stg.example.com' http://<alb-dns>/ping`.

> **Notes:** the static Secret in `k8s/manifests/overlays/local/secret.yaml` holds
> dev-only values and is used only by Kind. On AWS the `stg`/`prd` overlays get
> `POSTGRES_PASSWORD` / `JWT_SECRET` from Secrets Manager via the External Secrets
> Operator — the values are generated and mirrored into Secrets Manager by
> `auto-repair-shop-infra-db`'s Terraform, real values are never committed here.
> The `.github/workflows/docker.yml` `publish`/`deploy` jobs are fully implemented
> (build → push to ECR → `kubectl apply -k` → image rollout).

## Environment Variables

| Variable                  | Description                                         |
|---------------------------|-----------------------------------------------------|
| `SONARQUBE_USERNAME`      | SonarQube username                                  |
| `SONARQUBE_PASSWORD`      | SonarQube password                                  |
| `SONARQUBE_PROJECT_TOKEN` | SonarQube project token                             |
| `POSTGRES_USER`           | PostgreSQL username                                 |
| `POSTGRES_PASSWORD`       | PostgreSQL password                                 |
| `POSTGRES_DB`             | PostgreSQL database name                            |
| `POSTGRES_HOST`           | PostgreSQL host (`db` in Docker, `localhost` local) |
| `POSTGRES_PORT`           | PostgreSQL port (default: 5432)                     |
| `JWT_SECRET`              | Secret key for JWT signing                          |
| `JWT_EXPIRES_IN`          | Token expiration duration (e.g. `24h`)              |
| `BCRYPT_COST`             | bcrypt hashing cost factor (min 12 in production)   |

## API Endpoints

### Public

| Method | Path                           | Description          |
|--------|--------------------------------|----------------------|
| POST   | `/v1/auth/login`               | Get JWT              |
| GET    | `/v1/service-order/:id/status` | Customer SO tracking |

### Users

| Method | Path                  | Roles |
|--------|-----------------------|-------|
| POST   | `/v1/auth/register`   | ADMIN |
| PATCH  | `/v1/users/:id/role`  | ADMIN |

### Customers

| Method | Path                | Roles            |
|--------|---------------------|------------------|
| POST   | `/v1/customers`     | ADMIN, ATTENDANT |
| GET    | `/v1/customers`     | ADMIN, ATTENDANT |
| GET    | `/v1/customers/:id` | ADMIN, ATTENDANT |
| PUT    | `/v1/customers/:id` | ADMIN, ATTENDANT |
| DELETE | `/v1/customers/:id` | ADMIN, ATTENDANT |

### Works

| Method | Path            | Roles                      |
|--------|-----------------|----------------------------|
| GET    | `/v1/works`     | ADMIN, ATTENDANT, MECHANIC |
| POST   | `/v1/works`     | ADMIN, ATTENDANT           |
| PUT    | `/v1/works/:id` | ADMIN, ATTENDANT           |
| DELETE | `/v1/works/:id` | ADMIN, ATTENDANT           |

### Vehicles

| Method | Path                         | Roles                      |
|--------|------------------------------|----------------------------|
| GET    | `/v1/vehicles`               | ADMIN, ATTENDANT, MECHANIC |
| POST   | `/v1/vehicles`               | ADMIN, ATTENDANT           |
| PUT    | `/v1/vehicles/:id`           | ADMIN, ATTENDANT           |
| DELETE | `/v1/vehicles/:id`           | ADMIN, ATTENDANT           |

`GET /v1/vehicles` accepts optional `customerId`, `plate`, `page` and `pageSize` query filters.

### Supplies

| Method | Path               | Roles                      |
|--------|--------------------|----------------------------|
| GET    | `/v1/supplies`     | ADMIN, ATTENDANT, MECHANIC |
| POST   | `/v1/supplies`     | ADMIN, ATTENDANT, MECHANIC |
| PUT    | `/v1/supplies/:id` | ADMIN, ATTENDANT, MECHANIC |
| DELETE | `/v1/supplies/:id` | ADMIN, ATTENDANT, MECHANIC |

### Service Orders (SO)

| Method | Path                                       | Roles                      |
|--------|--------------------------------------------|----------------------------|
| POST   | `/v1/service-order`                        | ADMIN, ATTENDANT           |
| GET    | `/v1/service-order`                        | ADMIN, ATTENDANT, MECHANIC |
| GET    | `/v1/service-order/:id`                    | ADMIN, ATTENDANT, MECHANIC |
| PUT    | `/v1/service-order/:id/received`           | ADMIN, ATTENDANT           |
| PUT    | `/v1/service-order/:id/start-diagnosis`    | ADMIN, MECHANIC            |
| PUT    | `/v1/service-order/:id/send`               | ADMIN, MECHANIC            |
| PUT    | `/v1/service-order/:id/accept`             | CUSTOMER (own SO only)     |
| PUT    | `/v1/service-order/:id/reject`             | CUSTOMER (own SO only)     |
| PUT    | `/v1/service-order/:id/finish`             | ADMIN, MECHANIC            |
| PUT    | `/v1/service-order/:id/deliver`            | ADMIN, ATTENDANT           |
| PUT    | `/v1/service-order/:id/cancel`             | ADMIN, ATTENDANT, MECHANIC |
| GET    | `/v1/service-order/:id/history`            | ADMIN, ATTENDANT, MECHANIC |
| GET    | `/v1/service-order/:id/works`              | ADMIN, ATTENDANT, MECHANIC |
| POST   | `/v1/service-order/:id/works`              | ADMIN, ATTENDANT           |
| DELETE | `/v1/service-order/:id/works/:serviceId`   | ADMIN, ATTENDANT           |
| GET    | `/v1/service-order/:id/supplies`           | ADMIN, ATTENDANT, MECHANIC |
| POST   | `/v1/service-order/:id/supplies`           | ADMIN, MECHANIC            |
| DELETE | `/v1/service-order/:id/supplies/:supplyId` | ADMIN, MECHANIC            |

### Work Status Transitions (within a SO)

| Method | Path                                        | Roles            |
|--------|---------------------------------------------|------------------|
| PUT    | `/v1/service-order/:id/work/:workId/next`   | ADMIN, MECHANIC  |
| PUT    | `/v1/service-order/:id/work/:workId/cancel` | ADMIN, MECHANIC  |

### Admin

| Method | Path                                 | Roles            |
|--------|--------------------------------------|------------------|
| POST   | `/v1/auth/register`                  | ADMIN            |
| PATCH  | `/v1/users/:id/role`                 | ADMIN            |

### Reports

| Method | Path                                 | Roles            |
|--------|--------------------------------------|------------------|
| GET    | `/v1/reports/average-execution-time` | ADMIN, ATTENDANT |

## RBAC Roles

| Role       | Description                                                                  |
|------------|------------------------------------------------------------------------------|
| ADMIN      | Full access, including user management and role updates                      |
| ATTENDANT  | Customers, vehicles, works, service orders, supplies, and execution reports  |
| MECHANIC   | Service order execution flow, supplies, and work status transitions          |
| CUSTOMER   | Own service order status tracking and budget approve/reject only             |

## Password Encryption

User passwords are protected using bcrypt (`golang.org/x/crypto/bcrypt`).

Key details:
- **Type:** one-way hash (not reversible). The output includes an internal salt and metadata.
- **Implementation:** `bcrypt.GenerateFromPassword` is used when creating or updating passwords; `bcrypt.CompareHashAndPassword` is used for validation.
- **Cost factor:** controlled by the `BCRYPT_COST` environment variable (see Environment Variables). A minimum cost of 12 is recommended in production — increase it according to your infrastructure capacity.

Notes:
- For automatically created customers, the default password (CPF/CNPJ) is immediately hashed before being persisted.
- bcrypt applies a salt securely by default; there is no need to manage salts manually.

## License

This project is part of the FIAP 15SOAT postgraduate program.

## Members

- Caetano Agostinho de Freitas - RM371006
- Emanuel Jesus Santos - RM371184
- Diogo Estevão Ferreira - RM371059
- Giusier Ferreira Soares - RM371064
- Guilherme Ferreira Santos - RM374002
