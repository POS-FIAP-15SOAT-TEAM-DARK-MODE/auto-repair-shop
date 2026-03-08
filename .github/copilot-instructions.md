# GitHub Copilot Instructions — Integrated Auto Repair Shop System

## Project Overview
Monolithic layered backend for an auto repair shop. Manages service orders, customers, vehicles, parts/stock, and administrative operations. Database: PostgreSQL. Auth: JWT. API: RESTful + Swagger/OpenAPI.

## Domain Language
- **Service Order (SO)** — service order, the central aggregate of the system
- **Customer** — customer, uniquely identified by CPF (individual) or CNPJ (company)
- **Vehicle** — vehicle with license plate, brand, model, year; always linked to a Customer
- **Service** — billable service (oil change, alignment, etc.) with name, description, unit price
- **Part (Peca)** — part or supply with stock quantity and unit price
- **Budget** — budget auto-calculated from services and parts on a service order
- **Stock** — stock/inventory control for Parts
- **User** — system user account; stores name, email, password hash, and roles
- **Customer** — customer-specific data (CPF/CNPJ, company name, phone), linked 1:1 to a User

## Architecture
Layered monolith: **Controller → Service → Repository**
- Controllers: parse HTTP input, validate request, call service, return response — no business logic
- Services: all business logic (status transitions, budget calc, stock checks, validations)
- Repositories: all DB access — no raw queries outside this layer
- Domain models: plain objects, no framework dependencies

## Service Order Status Lifecycle
Transitions must follow this exact order — never skip, never reverse:
```
RECEIVED → IN_DIAGNOSIS → AWAITING_APPROVAL → IN_PROGRESS → COMPLETED → DELIVERED
```
`REJECTED` is a terminal state set when the customer rejects the budget (only from `AWAITING_APPROVAL`).
Any out-of-order transition must throw a domain error → HTTP 422.

## Critical Business Rules

### User & Customer Split
- Creating a Customer automatically creates a linked User with role `CLIENT`
- User entity stores: name, email, password hash, roles
- Customer entity stores: CPF/CNPJ, company name, phone, FK to User
- Default password = CPF (individual) or CNPJ (company), hashed with bcrypt
- Deleting a Customer must also delete or deactivate the associated User

### RBAC — Role-Based Access Control
Four roles: `ADMIN`, `ATTENDANT`, `MECHANIC`, `CLIENT`
- Users may hold multiple roles simultaneously (many-to-many via `USER_ROLE`), **except** `ADMIN` which is exclusive
- **ADMIN**: unrestricted access to all endpoints, including user management and role assignment
- **ATTENDANT**: customers, vehicles, service orders, budgets, reports — no user management
- **MECHANIC**: read-only service orders, add services/parts to service orders, status transitions only
- **CLIENT**: own service order tracking + own budget approve/reject only; authenticated by email + password; access filtered by `customerId` in JWT
- Only `ADMIN` can create other `ADMIN` users
- Only `ADMIN` can change any user's roles
- Out-of-scope access → HTTP 403

### Stock
- Always decrement stock inside a DB transaction when associating a Part (Peca) to a service order (OS)
- Check `quantity > 0` before decrementing; throw `OUT_OF_STOCK` error if zero
- Use `SELECT FOR UPDATE` to prevent race conditions on concurrent decrements

### Budget
- `total = Σ service.unit_price + Σ (part.unit_price × quantity)`
- Recalculate automatically whenever services or parts change on the service order
- Budget is immutable once service order status reaches `IN_PROGRESS`

### Input Validation
- CPF: validate format `###.###.###-##` AND check digit algorithm
- CNPJ: validate format `##.###.###/####-##` AND check digit algorithm
- License plate: accept old format `ABC-1234` and Mercosul format `ABC1D23`
- Monetary values: `NUMERIC(10,2)` — never `float` or `double`
- Validate all input at controller layer before passing to service

### Security
- All admin endpoints require `Authorization: Bearer <JWT>` — return 401 if absent or invalid
- Public endpoints (no auth): `POST /auth/login`, `GET /service-orders/:id/status`
- CLIENT endpoints require JWT with role `CLIENT`; access scoped to `customerId` in JWT payload
- Hash passwords with bcrypt, minimum cost factor 12
- Never log or expose CPF, CNPJ, passwords, or tokens in any response or log
- JWT expiry must be configurable via environment variable
- JWT payload must include: `roles` (array) and `customerId` (when applicable) — middleware must not make extra DB calls for authorization

## API Endpoints
```
POST   /auth/login                        public
GET    /service-orders/:id/status         public — customer tracking

POST   /customers                         ADMIN, ATTENDANT
GET    /customers                         ADMIN, ATTENDANT
GET    /customers/:id                     ADMIN, ATTENDANT
PUT    /customers/:id                     ADMIN, ATTENDANT
DELETE /customers/:id                     ADMIN

(same CRUD pattern for /vehicles, /services, /parts)

POST   /service-orders                    ADMIN, ATTENDANT — creates service order, sets RECEIVED, calculates budget
GET    /service-orders                    ADMIN, ATTENDANT, MECHANIC — filter by ?status=&customer_id=&vehicle_id=
GET    /service-orders/:id                ADMIN, ATTENDANT, MECHANIC — full detail: services, parts, values, history
PATCH  /service-orders/:id/status         ADMIN, ATTENDANT, MECHANIC — advance status
POST   /service-orders/:id/approve        CLIENT — own service order only → advances to IN_PROGRESS
POST   /service-orders/:id/reject         CLIENT — own service order only → sets REJECTED
GET    /reports/average-time              ADMIN, ATTENDANT — average service execution time

POST   /users                             ADMIN only — create user
PUT    /users/:id/roles                   ADMIN only — update user roles
```

## Database Conventions
- Monetary columns: `NUMERIC(10,2)` always
- Service order status column: ENUM or CHECK constraint; valid values: RECEIVED, IN_DIAGNOSIS, AWAITING_APPROVAL, IN_PROGRESS, COMPLETED, DELIVERED, REJECTED
- Required foreign keys: vehicle→customer, service_order→customer, service_order→vehicle, service_order_service→(service_order, service), service_order_part→(service_order, part), customer→user
- Many-to-many: `USER_ROLE` table linking user↔role
- Required indexes: `customer.cpf`, `customer.cnpj`, `vehicle.plate`, `service_order.status`, `service_order.customer_id`, `user.email`
- IDs: UUID; timestamps: UTC ISO 8601
- Schema changes only via versioned migration files — never alter schema manually

## Testing Requirements
- Critical domains (service orders, stock, budget): **90% minimum coverage**
- All other domains: **80% minimum coverage**
- Unit test: status transitions (valid + invalid paths), budget calculation, CPF/CNPJ/license plate validators, stock decrement (success, out-of-stock, concurrent), RBAC role checks (403 for out-of-scope)
- Integration test: full service order creation flow, each status transition, approve/reject budget (CLIENT role), public status endpoint (no auth), 401 on admin endpoints without token, 403 on wrong role, CLIENT filtered by customerId
- Pattern: Arrange / Act / Assert; tests must be independent (no shared mutable state)
- Integration tests run against a real test DB, rolled back after each test

## Error Response Format
```json
{ "error": "Human-readable message", "code": "MACHINE_READABLE_CODE" }
```
- 400 validation error, 401 missing/invalid auth, 403 forbidden, 404 not found, 409 conflict, 422 business rule violation, 500 unexpected

## Logging
- Structured JSON logs: `{ timestamp, level, operation, entity_id }`
- Log: OS creation, every status transition, stock decrements, validation errors
- Never log: CPF, CNPJ, passwords, tokens (mask or omit entirely)
- Levels: INFO (normal), WARN (business violations), ERROR (unexpected)

## Infrastructure
- `Dockerfile` for reproducible builds
- `docker-compose.yml` orchestrating app + PostgreSQL
- Required env vars: `DATABASE_URL`, `JWT_SECRET`, `JWT_EXPIRY`, `BCRYPT_COST`
- App must boot with Docker only — no manual external setup

## Commit Messages — Conventional Commits
All commits must follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Format
```
<type>(<scope>): <short description>

[optional body]

[optional footer(s)]
```

### Types
- `feat` — new feature
- `fix` — bug fix
- `refactor` — code change that neither fixes a bug nor adds a feature
- `test` — adding or updating tests
- `chore` — build, tooling, dependency updates, CI changes
- `docs` — documentation only
- `perf` — performance improvement
- `style` — formatting, missing semicolons, etc (no logic change)
- `revert` — revert a previous commit

### Scopes (domain-aligned)
`service-order`, `customer`, `vehicle`, `service`, `part`, `stock`, `budget`, `auth`, `user`, `infra`, `db`, `api`

### Rules
- Subject line max 72 characters, lowercase, no period at end
- Use imperative mood: "add feature" not "added feature"
- Breaking changes must include `BREAKING CHANGE:` footer or `!` after type: `feat(auth)!: change JWT payload structure`
- Reference issues in footer: `Closes #123`
