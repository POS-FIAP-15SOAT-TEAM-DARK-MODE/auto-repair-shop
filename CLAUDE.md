# Sistema Integrado de Oficina Mecânica — Claude Code Rules

## Project Overview
Monolithic layered backend for an auto repair shop. Manages service orders (OS), clients, vehicles, parts/stock, and administrative operations. Database: PostgreSQL. Auth: JWT. API: RESTful + Swagger.

## Domain Language (Ubiquitous Language)
- **OS / Ordem de Serviço** — service order, the central aggregate
- **Cliente** — client, identified by CPF (individual) or CNPJ (company)
- **Veiculo** — vehicle (placa, marca, modelo, ano), always linked to a Cliente
- **Servico** — a billable service offered by the shop (name, description, unit price)
- **Peca / Insumo** — part or supply with stock quantity and unit price
- **Orcamento** — budget, auto-calculated from services and parts on an OS
- **Estoque** — stock/inventory of Pecas
- **User** — system user account; holds nome, e-mail, password hash, and roles
- **Customer** — client-specific data (CPF/CNPJ, razão social, telefone), linked 1:1 to a User

## OS Status Lifecycle
Statuses must only advance in this exact order — never skip, never go back:
```
RECEBIDA → EM_DIAGNOSTICO → AGUARDANDO_APROVACAO → EM_EXECUCAO → FINALIZADA → ENTREGUE
```
- `RECUSADA` is a terminal state reached when the client rejects the budget (from `AGUARDANDO_APROVACAO`)
- Any attempt to transition out of order must throw a domain error (HTTP 422)

## Architecture Rules
- **Controller layer**: parse and validate HTTP input, call service, return response — zero business logic
- **Service layer**: all business logic, status transitions, budget calculation, stock checks
- **Repository layer**: all database queries — no raw SQL outside repositories
- **Domain models**: framework-free plain objects/classes

## Critical Business Rules to Always Enforce

### User & Customer Split
- When creating a Cliente, automatically create a linked User with role `CLIENTE`
- The User entity stores: nome, e-mail, password hash, roles
- The Customer entity stores: CPF/CNPJ, razão social, telefone, and a FK to User
- Default password for the auto-created user = CPF (individual) or CNPJ (company), hashed with bcrypt
- Deleting a Cliente must also delete or deactivate the associated User

### RBAC — Role-Based Access Control
Four roles exist: `ADMIN`, `ATENDENTE`, `MECANICO`, `CLIENTE`
- A user may hold multiple roles simultaneously (many-to-many via `USER_ROLE` table), **except** `ADMIN`, which is exclusive and cannot be combined with other roles
- **ADMIN**: unrestricted access to all admin endpoints, including user management and role assignment
- **ATENDENTE**: access to clients, vehicles, OS, budgets, and reports — no user management
- **MECANICO**: read-only OS queries, adding services/parts to OS, OS status transitions only
- **CLIENTE**: access only to own OS status tracking and own budget approve/reject, authenticated by e-mail + password, filtered by `customerId` embedded in JWT
- Only `ADMIN` can create other `ADMIN` users
- Only `ADMIN` can change any user's roles
- Out-of-scope access attempts must return HTTP 403

### Stock (Estoque)
- Decrement stock inside a DB transaction when adding a Peca to an OS
- Check `quantidade > 0` before decrementing; throw `OUT_OF_STOCK` domain error if zero
- Use `SELECT FOR UPDATE` (pessimistic lock) to prevent race conditions

### Budget (Orçamento)
- `total = Σ servico.valor_unitario + Σ (peca.valor_unitario × quantidade)`
- Auto-recalculate whenever services or parts change on the OS
- Budget becomes immutable once OS status reaches `EM_EXECUCAO`

### Validations
- CPF: validate format `###.###.###-##` AND check digit algorithm
- CNPJ: validate format `##.###.###/####-##` AND check digit algorithm
- Placa: support old format `ABC-1234` and Mercosul `ABC1D23`
- Monetary values: stored as `NUMERIC(10,2)` — never `float` or `double`

### Security
- All admin endpoints require `Authorization: Bearer <JWT>` — return 401 if missing/invalid
- Public endpoints (no auth required): `POST /auth/login`, `GET /os/:id/status`
- Cliente endpoints require JWT with role `CLIENTE`; access filtered by `customerId` in JWT payload
- Passwords hashed with bcrypt (min cost factor 12)
- Never log or expose CPF, CNPJ, passwords, or JWT tokens in any response or log entry
- JWT expiry configurable via environment variable
- JWT payload must include: `roles` (array of all user roles) and `customerId` (when applicable) — authorization middleware must not make extra DB calls

## API Endpoints Reference
```
POST   /auth/login                    — get JWT (public)
GET    /os/:id/status                 — client OS tracking (public)
POST   /clientes                      — create client (ADMIN, ATENDENTE)
GET    /clientes                      — list clients (ADMIN, ATENDENTE)
GET    /clientes/:id                  — get client (ADMIN, ATENDENTE)
PUT    /clientes/:id                  — update client (ADMIN, ATENDENTE)
DELETE /clientes/:id                  — delete client (ADMIN)
(same CRUD for /veiculos, /servicos, /pecas)
POST   /os                            — create OS (ADMIN, ATENDENTE)
GET    /os                            — list OS (?status=&cliente_id=&veiculo_id=) (ADMIN, ATENDENTE, MECANICO)
GET    /os/:id                        — full OS detail (ADMIN, ATENDENTE, MECANICO)
PATCH  /os/:id/status                 — advance OS status (ADMIN, ATENDENTE, MECANICO)
POST   /os/:id/approve                — approve budget (CLIENTE — own OS only)
POST   /os/:id/reject                 — reject budget (CLIENTE — own OS only)
GET    /relatorios/tempo-medio        — average service execution time (ADMIN, ATENDENTE)
POST   /usuarios                      — create user (ADMIN only)
PUT    /usuarios/:id/roles            — update user roles (ADMIN only)
```

## Database Conventions
- All monetary columns: `NUMERIC(10,2)`
- OS status: ENUM or CHECK constraint with the 7 valid values
- Foreign keys required: veiculo→cliente, os→cliente, os→veiculo, os_servico→os+servico, os_peca→os+peca, customer→user
- Many-to-many: `USER_ROLE` table linking user↔role
- Indexes required on: `customer.cpf`, `customer.cnpj`, `veiculo.placa`, `os.status`, `os.cliente_id`, `user.email`
- IDs: UUID
- Timestamps: UTC, ISO 8601
- Schema changes only via versioned migration files

## Testing Requirements
- Critical domains (OS, estoque, orcamento): minimum **90% coverage**
- All other domains: minimum **80% coverage**
- Always write unit tests for: status transitions (valid + invalid), budget calculation, CPF/CNPJ/placa validation, stock decrement (success + out-of-stock + concurrency), RBAC role checks (403 for out-of-scope)
- Always write integration tests for: OS creation flow, status progression, budget approve/reject, public status endpoint, auth guard on admin endpoints, CLIENTE role filtering by customerId
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
- Always log: OS creation, every status transition, stock decrement, validation errors
- Never log: CPF, CNPJ, passwords, tokens
- Levels: INFO (normal ops), WARN (business rule violations), ERROR (unexpected failures)

## Infrastructure
- `Dockerfile` for reproducible build
- `docker-compose.yml` orchestrating app + PostgreSQL
- Environment variables for: DATABASE_URL, JWT_SECRET, JWT_EXPIRY, BCRYPT_COST
- App must start cleanly with only Docker — no manual external dependencies
