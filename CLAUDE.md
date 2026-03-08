# Integrated Auto Repair Shop System — Claude Code Rules

## Project Overview
Monolithic layered backend for an auto repair shop. Manages service orders, customers, vehicles, parts/stock, and administrative operations. Database: PostgreSQL. Auth: JWT. API: RESTful + Swagger.

## Domain Language (Ubiquitous Language)
- **Service Order (SO)** — service order, the central aggregate
- **Customer** — customer, identified by CPF (individual) or CNPJ (company)
- **Vehicle** — vehicle (license plate, brand, model, year), always linked to a Customer
- **Service** — a billable service offered by the shop (name, description, unit price)
- **Part (Peca)** — part or supply with stock quantity and unit price
- **Budget** — budget, auto-calculated from services and parts on a service order
- **Stock** — stock/inventory of Parts
- **User** — system user account; holds name, email, password hash, and roles
- **Customer** — customer-specific data (CPF/CNPJ, company name, phone), linked 1:1 to a User

## Service Order Status Lifecycle
Statuses must only advance in this exact order — never skip, never go back:
```
RECEIVED → IN_DIAGNOSIS → AWAITING_APPROVAL → IN_PROGRESS → COMPLETED → DELIVERED
```
- `REJECTED` is a terminal state reached when the customer rejects the budget (from `AWAITING_APPROVAL`)
- Any attempt to transition out of order must throw a domain error (HTTP 422)

## Architecture Rules
- **Controller layer**: parse and validate HTTP input, call service, return response — zero business logic
- **Service layer**: all business logic, status transitions, budget calculation, stock checks
- **Repository layer**: all database queries — no raw SQL outside repositories
- **Domain models**: framework-free plain objects/classes

## Critical Business Rules to Always Enforce

### User & Customer Split
- When creating a Customer, automatically create a linked User with role `CLIENT`
- The User entity stores: name, email, password hash, roles
- The Customer entity stores: CPF/CNPJ, company name, phone, and a FK to User
- Default password for the auto-created user = CPF (individual) or CNPJ (company), hashed with bcrypt
- Deleting a Customer must also delete or deactivate the associated User

### RBAC — Role-Based Access Control
Four roles exist: `ADMIN`, `ATTENDANT`, `MECHANIC`, `CLIENT`
- A user may hold multiple roles simultaneously (many-to-many via `USER_ROLE` table), **except** `ADMIN`, which is exclusive and cannot be combined with other roles
- **ADMIN**: unrestricted access to all admin endpoints, including user management and role assignment
- **ATTENDANT**: access to customers, vehicles, service orders, budgets, and reports — no user management
- **MECHANIC**: read-only service order queries, adding services/parts to service orders, status transitions only
- **CLIENT**: access only to own service order status tracking and own budget approve/reject, authenticated by email + password, filtered by `customerId` embedded in JWT
- Only `ADMIN` can create other `ADMIN` users
- Only `ADMIN` can change any user's roles
- Out-of-scope access attempts must return HTTP 403

### Stock
- Decrement stock inside a DB transaction when adding a Part (Peca) to a service order (OS)
- Check `quantity > 0` before decrementing; throw `OUT_OF_STOCK` domain error if zero
- Use `SELECT FOR UPDATE` (pessimistic lock) to prevent race conditions

### Budget
- `total = Σ service.unit_price + Σ (part.unit_price × quantity)`
- Auto-recalculate whenever services or parts change on the service order
- Budget becomes immutable once service order status reaches `IN_PROGRESS`

### Validations
- CPF: validate format `###.###.###-##` AND check digit algorithm
- CNPJ: validate format `##.###.###/####-##` AND check digit algorithm
- License plate: support old format `ABC-1234` and Mercosul `ABC1D23`
- Monetary values: stored as `NUMERIC(10,2)` — never `float` or `double`

### Security
- All admin endpoints require `Authorization: Bearer <JWT>` — return 401 if missing/invalid
- Public endpoints (no auth required): `POST /auth/login`, `GET /service-orders/:id/status`
- CLIENT endpoints require JWT with role `CLIENT`; access filtered by `customerId` in JWT payload
- Passwords hashed with bcrypt (min cost factor 12)
- Never log or expose CPF, CNPJ, passwords, or JWT tokens in any response or log entry
- JWT expiry configurable via environment variable
- JWT payload must include: `roles` (array of all user roles) and `customerId` (when applicable) — authorization middleware must not make extra DB calls

## API Endpoints Reference
```
POST   /auth/login                    — get JWT (public)
GET    /service-orders/:id/status     — customer service order tracking (public)
POST   /customers                     — create customer (ADMIN, ATTENDANT)
GET    /customers                     — list customers (ADMIN, ATTENDANT)
GET    /customers/:id                 — get customer (ADMIN, ATTENDANT)
PUT    /customers/:id                 — update customer (ADMIN, ATTENDANT)
DELETE /customers/:id                 — delete customer (ADMIN)
(same CRUD for /vehicles, /services, /parts)
POST   /service-orders                — create service order (ADMIN, ATTENDANT)
GET    /service-orders                — list service orders (?status=&customer_id=&vehicle_id=) (ADMIN, ATTENDANT, MECHANIC)
GET    /service-orders/:id            — full service order detail (ADMIN, ATTENDANT, MECHANIC)
PATCH  /service-orders/:id/status     — advance service order status (ADMIN, ATTENDANT, MECHANIC)
POST   /service-orders/:id/approve    — approve budget (CLIENT — own service order only)
POST   /service-orders/:id/reject     — reject budget (CLIENT — own service order only)
GET    /reports/average-time          — average service execution time (ADMIN, ATTENDANT)
POST   /users                         — create user (ADMIN only)
PUT    /users/:id/roles               — update user roles (ADMIN only)
```

## Database Conventions
- All monetary columns: `NUMERIC(10,2)`
- Service order status: ENUM or CHECK constraint with the 7 valid values
- Foreign keys required: vehicle→customer, service_order→customer, service_order→vehicle, service_order_service→service_order+service, service_order_part→service_order+part, customer→user
- Many-to-many: `USER_ROLE` table linking user↔role
- Indexes required on: `customer.cpf`, `customer.cnpj`, `vehicle.plate`, `service_order.status`, `service_order.customer_id`, `user.email`
- IDs: UUID
- Timestamps: UTC, ISO 8601
- Schema changes only via versioned migration files

## Testing Requirements
- Critical domains (service orders, stock, budget): minimum **90% coverage**
- All other domains: minimum **80% coverage**
- Always write unit tests for: status transitions (valid + invalid), budget calculation, CPF/CNPJ/license plate validation, stock decrement (success + out-of-stock + concurrency), RBAC role checks (403 for out-of-scope)
- Always write integration tests for: service order creation flow, status progression, budget approve/reject, public status endpoint, auth guard on admin endpoints, CLIENT role filtering by customerId
- Test pattern: Arrange / Act / Assert
- Tests must be independent — no shared mutable state
- Integration tests use a real test DB, rolled back after each test

## Error Response Format
```json
{ "error": "Human-readable message", "code": "MACHINE_READABLE_CODE" }
```
HTTP codes: 400 (validation), 401 (auth), 403 (forbidden), 404 (not found), 409 (conflict), 422 (business rule violation), 500 (unexpected)

## Logging
- Structured JSON logs with fields: `timestamp`, `level`, `operation`, `entity_id`
- Always log: service order creation, every status transition, stock decrement, validation errors
- Never log: CPF, CNPJ, passwords, tokens
- Levels: INFO (normal ops), WARN (business rule violations), ERROR (unexpected failures)

## Infrastructure
- `Dockerfile` for reproducible build
- `docker-compose.yml` orchestrating app + PostgreSQL
- Environment variables for: DATABASE_URL, JWT_SECRET, JWT_EXPIRY, BCRYPT_COST
- App must start cleanly with only Docker — no manual external dependencies
