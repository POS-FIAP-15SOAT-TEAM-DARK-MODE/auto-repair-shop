# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# Integrated Auto Repair Shop System — Claude Code Rules

## Project Overview
Monolithic layered backend for an auto repair shop. Manages service orders, customers, vehicles, works (billable services), supplies (parts/stock), and administrative operations. Database: PostgreSQL. Auth: JWT. API: RESTful + Swagger.

**Tech stack:** Go 1.26 · Gin · PostgreSQL 15 · `lib/pq` (raw SQL) · `golang-migrate` · JWT (`golang-jwt/jwt/v5`) · Zap logger · Testify · Mockery v2 · `shopspring/decimal` (monetary values) · `google/uuid` (most IDs) · `oklog/ulid/v2` (ServiceOrder, ServiceOrderHistory IDs)

## Development Commands

```bash
make run              # go run cmd/service/main.go (port 8080)
make test             # run all tests with race detector (LOG_LEVEL=PANIC suppresses noise)
make coverage         # generate coverage.out and open HTML report
make docker-up        # docker-compose up --build (app + PostgreSQL)
make docker-down      # docker-compose down
make mockgen          # regenerate all mocks via go generate ./...
make migrate-install  # install golang-migrate CLI (once per machine)
make migrate-up       # apply pending migrations
make migrate-down     # roll back last migration
make migrate-status   # show current migration version
```

Run a single test or package:
```bash
go test ./internal/domain/... -run TestCustomer_Validate -v
go test ./internal/services/customer/... -v --race
```

Swagger UI: `http://localhost:8080/swagger/index.html` (when running).

## Code Architecture

Layered hexagonal architecture — no framework leakage into inner layers:

```
HTTP → Handler (infra/handler/*) → Service (internal/services/*) → Repository (infra/repository/*) → PostgreSQL
                                        ↕
                              Domain (internal/domain/*)
```

- **`internal/domain/`** — Framework-free structs, validation methods, and service/repository interfaces. Mocks generated to `domain/mocks/` via `//go:generate` directives.
- **`internal/services/`** — All business logic, status transitions, stock checks. Services receive `uow.Executor` and repository interfaces via constructor.
- **`internal/infra/repository/`** — All raw SQL queries via `lib/pq`. Repositories are stateless (`type repo struct{}`); they extract the DB connection from `context` (see below). Every domain has both `postgres.go` and `in_memory.go` implementations.
- **`internal/infra/handler/`** — Gin handlers: parse/validate input, call service, return response. Zero business logic.
- **`internal/infra/http/`** — `HandlersWrapper` struct and `Middlewares` map used by the router.
- **`internal/infra/factory/http.go`** — Single wiring file: creates DB, repositories, services, and handlers; passes them into `HandlersWrapper`. This is the DI root.
- **`internal/infra/db/postgres/`** — PostgreSQL client, transactional UoW, and context helpers.
- **`internal/infra/db/seed/`** — Seeds initial data on every startup.
- **`internal/pkg/`** — Shared utilities: `auth/` (JWT), `uow/` (Unit of Work interface), `logger/` (Zap), `env/` (env vars), `web/` (HTTP response helpers), `db/` (QueryBuilder for dynamic SQL).
- **`migrations/`** — Versioned SQL files (`NNNNNN_description.up.sql` / `.down.sql`). All schema changes go through new migration files.
- **`cmd/service/main.go`** — Entrypoint: calls `factory.HTTPServer()`, wraps in lifecycle manager, starts.

### Repository DB Connection Pattern

Repositories do **not** hold a `*sql.DB`. The connection is injected via context by the UoW:

- **Mutating operations** (inside a transaction): call `postgres.GetTransaction(ctx)` → returns the `*sql.Tx` stored in context by the UoW `OnStart` hook.
- **Read-only operations** (outside a transaction): call `postgres.GetOneTimeTransaction(ctx)` → opens a fresh `*sql.DB` connection directly.

Services wrap all repository calls in `uow.Execute(ctx, steps...)`. The UoW begins a transaction, injects it into context, runs all steps, then commits or rolls back.

### Dynamic SQL

Use `db.QueryBuilder(baseQuery)` from `internal/pkg/db/` for queries with optional filters. Call `.Add("column =", value)`, `.OrderBy(field, db.ASC)`, `.AddPagination(limit, offset)`, then `.Build()` to get the final query string and args slice.

## Domain Language (Ubiquitous Language)

| Code name | Business concept |
|---|---|
| `Work` | Billable service offered by the shop (name, description, unit price, ACTIVE/INACTIVE) |
| `Supply` | Part/supply with stock quantity, unit price, and optimistic-lock `Version` |
| `ServiceOrder` | Central aggregate; links Customer + Vehicle + []Work + []Supply |
| `ServiceOrderHistory` | Immutable record of every status transition on a ServiceOrder |
| `Customer` | Identified by CPF (individual) or CNPJ (company); linked 1:1 to a User |
| `User` | System account: name, email, bcrypt-hashed password, roles |
| `Vehicle` | License plate + brand/model/year, always linked to a Customer |

Roles: `ADMIN`, `ATTENDANT`, `MECHANIC`, `CUSTOMER` (role constants live in `internal/domain/roles.go`).

## Service Order Status Lifecycle

```
NEW → (works and supplies added here) → ... (future transitions TBD)
```

Currently the only enforced lifecycle rule is that `AddWorks`, `RemoveWork`, `AddSupplies`, and `RemoveSupply` all require the service order to be in `NEW` status, returning `ErrServiceOrderNotNew` otherwise.

Full status constants: `NEW`, `RECEIVED`, `IN_DIAGNOSIS`, `AWAITING_APPROVAL`, `IN_PROGRESS`, `COMPLETED`, `DELIVERED`. `REJECTED` is the terminal state for customer-rejected budgets (from `AWAITING_APPROVAL`).

## Stock Decrement Pattern

Stock is decremented **atomically** with a conditional UPDATE:
```sql
UPDATE supply SET stock_quantity = stock_quantity - $amount, updated_at = NOW()
WHERE id = $id AND stock_quantity >= $amount
```
If `RowsAffected == 0`, the repository returns `ErrSupplyOutOfStock`. There is no `SELECT FOR UPDATE`; the conditional UPDATE itself prevents overselling. Stock is restored on `RemoveSupply` via a matching `RestoreStock` call, all within the same transaction.

## Architecture Rules

- **Handler layer**: parse and validate HTTP input, call service, return response — zero business logic.
- **Service layer**: all business logic, status enforcement, stock checks.
- **Repository layer**: all SQL — no queries outside repositories.
- **Domain models**: framework-free; validation logic lives on the struct methods.

## Critical Business Rules to Always Enforce

### User & Customer Split
- Creating a Customer automatically creates a linked User with role `CUSTOMER`.
- The User stores: name, email, password hash, roles.
- The Customer stores: CPF/CNPJ, company name, phone, FK to User.
- Default password = CPF (individual) or CNPJ (company), bcrypt-hashed (min cost 12).
- Deleting a Customer must also delete/deactivate the associated User.

### RBAC — Role-Based Access Control
Roles: `ADMIN`, `ATTENDANT`, `MECHANIC`, `CUSTOMER`
- `ADMIN` is exclusive — cannot be combined with other roles.
- `ADMIN`: unrestricted access including user management and role assignment.
- `ATTENDANT`: customers, vehicles, service orders, works, supplies — no user management.
- `MECHANIC`: read service orders; add/remove works and supplies; status transitions.
- `CUSTOMER`: own service order status and budget approve/reject only, filtered by `customerId` in JWT.
- Role groups are defined in `domain/roles.go`: `AttendantRoles`, `MechanicRoles`, `CustomerRoles`, `AttendantAndMechanicRoles`.

### Budget
- `total = Σ work.unit_price + Σ (supply.unit_price × quantity)`
- Auto-recalculate whenever works or supplies change on the service order.
- Budget becomes immutable once status reaches `IN_PROGRESS`.

### Validations
- CPF: format `###.###.###-##` AND check digit algorithm.
- CNPJ: format `##.###.###/####-##` AND check digit algorithm.
- License plate: old format `ABC-1234` and Mercosul `ABC1D23`.
- Monetary values: `NUMERIC(10,2)` — never `float` or `double`.

### Security
- All protected endpoints require `Authorization: Bearer <JWT>` — return 401 if missing/invalid.
- Public endpoints: `POST /v1/auth/login`, and the ping endpoint.
- JWT payload must include: `roles` and `customerId` (when applicable) — auth middleware must not make extra DB calls.
- Passwords: bcrypt, min cost factor 12.
- Never log or expose CPF, CNPJ, passwords, or JWT tokens.

## API Endpoints Reference (all under `/v1/`)

```
POST   /auth/register                              — create user (ADMIN)
POST   /auth/login                                 — get JWT (public)
POST   /customers                                  — create (ADMIN, ATTENDANT)
GET    /customers/:id                              — get by ID (ADMIN, ATTENDANT)
GET    /customers                                  — get by document (ADMIN, ATTENDANT)
PUT    /customers/:id                              — update (ADMIN, ATTENDANT)
DELETE /customers/:id                              — delete (ADMIN, ATTENDANT)
POST   /works                                      — create (ADMIN, ATTENDANT)
GET    /works                                      — list (ADMIN, ATTENDANT, MECHANIC)
PUT    /works/:id                                  — update (ADMIN, ATTENDANT, MECHANIC)
DELETE /works/:id                                  — delete (ADMIN, ATTENDANT, MECHANIC)
POST   /vehicles                                   — create (ADMIN, ATTENDANT, MECHANIC)
GET    /vehicles                                   — find by license plate (ADMIN, ATTENDANT, MECHANIC)
PUT    /vehicles/:id                               — update (ADMIN, ATTENDANT, MECHANIC)
DELETE /vehicles/:id                               — delete (ADMIN, ATTENDANT, MECHANIC)
GET    /vehicles/:customerId                       — list by customer (ADMIN, ATTENDANT)
POST   /supplies                                   — create (ADMIN, ATTENDANT, MECHANIC)
GET    /supplies                                   — list (ADMIN, ATTENDANT, MECHANIC)
PUT    /supplies/:id                               — update (ADMIN, ATTENDANT, MECHANIC)
POST   /service-order                              — create (ADMIN, ATTENDANT)
GET    /service-order/:id/history                  — status history (ADMIN, ATTENDANT)
GET    /service-order/:id/services                 — list works (ADMIN, ATTENDANT)
POST   /service-order/:id/services                 — add works (ADMIN, ATTENDANT)
DELETE /service-order/:id/services/:serviceId      — remove work (ADMIN, ATTENDANT)
GET    /service-order/:id/supplies                 — list supplies (ADMIN, ATTENDANT)
POST   /service-order/:id/supplies                 — add supplies + decrement stock (ADMIN, ATTENDANT)
DELETE /service-order/:id/supplies/:supplyId       — remove supply + restore stock (ADMIN, ATTENDANT)
```

## Database Conventions
- All monetary columns: `NUMERIC(10,2)`
- ServiceOrder status: CHECK constraint with the 7 valid values
- IDs: UUID for most entities; ULID for `service_order` and `service_order_history`
- Timestamps: UTC, ISO 8601
- Junction tables: `service_order_work` (SO ↔ Work), `service_order_supply` (SO ↔ Supply)
- Schema changes only via versioned migration files

## Testing Requirements
- Critical domains (service orders, stock, budget): minimum **90% coverage**
- All other domains: minimum **80% coverage**
- Test pattern: Arrange / Act / Assert; tests must be independent.
- All repositories have in-memory implementations for unit tests; integration tests use a real test DB rolled back after each test.
- Use the in-memory UoW (`infra/db/in_memory`) for service-layer unit tests.

## Error Response Format
```json
{ "error": "Human-readable message", "code": "MACHINE_READABLE_CODE" }
```
HTTP codes: 400 (validation), 401 (auth), 403 (forbidden), 404 (not found), 409 (conflict), 422 (business rule violation), 500 (unexpected)

## Logging
- Structured JSON via Zap with standardized service-order events implemented in the service layer.
- Issue #236 implemented with explicit events: `service_order.created`, `service_order.status_transition`, `service_order.validation_failed`.
- Issue #223 implemented with centralized sanitization/redaction in logger (`internal/pkg/logger/logger.go`) for sensitive keys and DSN passwords.
- Avoid payload logging with `zap.Any` for domain/request objects when fields can include sensitive data.
- Never log or expose CPF, CNPJ, passwords, or tokens.
- Preferred levels: INFO (business events), WARN (rule violations), ERROR (unexpected failures).
- Read operations should prioritize error logging and keep success logs minimal.

## Infrastructure
- `Dockerfile` + `docker-compose.yml` (app + PostgreSQL)
- Env vars: `DATABASE_URL`, `JWT_SECRET`, `JWT_EXPIRES_IN`, `BCRYPT_COST`, `PORT` (default 8080)
- Makefile reads from `.env` and exports all vars automatically
