# GitHub Copilot Instructions — Sistema Integrado de Oficina Mecânica

## Project Overview
Monolithic layered backend for an auto repair shop. Manages service orders (OS), clients, vehicles, parts/stock, and administrative operations. Database: PostgreSQL. Auth: JWT. API: RESTful + Swagger/OpenAPI.

## Domain Language
- **OS / Ordem de Serviço** — service order, the central aggregate of the system
- **Cliente** — client, uniquely identified by CPF (individual) or CNPJ (company)
- **Veiculo** — vehicle with placa, marca, modelo, ano; always linked to a Cliente
- **Servico** — billable service (oil change, alignment, etc.) with name, description, unit price
- **Peca / Insumo** — part or supply with stock quantity and unit price
- **Orcamento** — budget auto-calculated from services and parts on an OS
- **Estoque** — stock/inventory control for Pecas
- **User** — system user account; stores nome, e-mail, password hash, and roles
- **Customer** — client-specific data (CPF/CNPJ, razão social, telefone), linked 1:1 to a User

## Architecture
Layered monolith: **Controller → Service → Repository**
- Controllers: parse HTTP input, validate request, call service, return response — no business logic
- Services: all business logic (status transitions, budget calc, stock checks, validations)
- Repositories: all DB access — no raw queries outside this layer
- Domain models: plain objects, no framework dependencies

## OS Status Lifecycle
Transitions must follow this exact order — never skip, never reverse:
```
RECEBIDA → EM_DIAGNOSTICO → AGUARDANDO_APROVACAO → EM_EXECUCAO → FINALIZADA → ENTREGUE
```
`RECUSADA` is a terminal state set when the client rejects the budget (only from `AGUARDANDO_APROVACAO`).
Any out-of-order transition must throw a domain error → HTTP 422.

## Critical Business Rules

### User & Customer Split
- Creating a Cliente automatically creates a linked User with role `CLIENTE`
- User entity stores: nome, e-mail, password hash, roles
- Customer entity stores: CPF/CNPJ, razão social, telefone, FK to User
- Default password = CPF (individual) or CNPJ (company), hashed with bcrypt
- Deleting a Cliente must also delete or deactivate the associated User

### RBAC — Role-Based Access Control
Four roles: `ADMIN`, `ATENDENTE`, `MECANICO`, `CLIENTE`
- Users may hold multiple roles simultaneously (many-to-many via `USER_ROLE`), **except** `ADMIN` which is exclusive
- **ADMIN**: unrestricted access to all endpoints, including user management and role assignment
- **ATENDENTE**: clients, vehicles, OS, budgets, reports — no user management
- **MECANICO**: read-only OS, add services/parts to OS, OS status transitions only
- **CLIENTE**: own OS tracking + own budget approve/reject only; authenticated by e-mail + password; access filtered by `customerId` in JWT
- Only `ADMIN` can create other `ADMIN` users
- Only `ADMIN` can change any user's roles
- Out-of-scope access → HTTP 403

### Stock
- Always decrement stock inside a DB transaction when associating a Peca to an OS
- Check `quantidade > 0` before decrementing; throw `OUT_OF_STOCK` error if zero
- Use `SELECT FOR UPDATE` to prevent race conditions on concurrent decrements

### Budget
- `total = Σ servico.valor_unitario + Σ (peca.valor_unitario × quantidade)`
- Recalculate automatically whenever services or parts change on the OS
- Budget is immutable once OS status reaches `EM_EXECUCAO`

### Input Validation
- CPF: validate format `###.###.###-##` AND check digit algorithm
- CNPJ: validate format `##.###.###/####-##` AND check digit algorithm
- Placa: accept old format `ABC-1234` and Mercosul format `ABC1D23`
- Monetary values: `NUMERIC(10,2)` — never `float` or `double`
- Validate all input at controller layer before passing to service

### Security
- All admin endpoints require `Authorization: Bearer <JWT>` — return 401 if absent or invalid
- Public endpoints (no auth): `POST /auth/login`, `GET /os/:id/status`
- Cliente endpoints require JWT with role `CLIENTE`; access scoped to `customerId` in JWT payload
- Hash passwords with bcrypt, minimum cost factor 12
- Never log or expose CPF, CNPJ, passwords, or tokens in any response or log
- JWT expiry must be configurable via environment variable
- JWT payload must include: `roles` (array) and `customerId` (when applicable) — middleware must not make extra DB calls for authorization

## API Endpoints
```
POST   /auth/login                  public
GET    /os/:id/status               public — client tracking

POST   /clientes                    ADMIN, ATENDENTE
GET    /clientes                    ADMIN, ATENDENTE
GET    /clientes/:id                ADMIN, ATENDENTE
PUT    /clientes/:id                ADMIN, ATENDENTE
DELETE /clientes/:id                ADMIN

(same CRUD pattern for /veiculos, /servicos, /pecas)

POST   /os                          ADMIN, ATENDENTE — creates OS, sets RECEBIDA, calculates budget
GET    /os                          ADMIN, ATENDENTE, MECANICO — filter by ?status=&cliente_id=&veiculo_id=
GET    /os/:id                      ADMIN, ATENDENTE, MECANICO — full detail: services, parts, values, history
PATCH  /os/:id/status               ADMIN, ATENDENTE, MECANICO — advance status
POST   /os/:id/approve              CLIENTE — own OS only → advances to EM_EXECUCAO
POST   /os/:id/reject               CLIENTE — own OS only → sets RECUSADA
GET    /relatorios/tempo-medio      ADMIN, ATENDENTE — average service execution time

POST   /usuarios                    ADMIN only — create user
PUT    /usuarios/:id/roles          ADMIN only — update user roles
```

## Database Conventions
- Monetary columns: `NUMERIC(10,2)` always
- OS status column: ENUM or CHECK constraint; valid values: RECEBIDA, EM_DIAGNOSTICO, AGUARDANDO_APROVACAO, EM_EXECUCAO, FINALIZADA, ENTREGUE, RECUSADA
- Required foreign keys: veiculo→cliente, os→cliente, os→veiculo, os_servico→(os, servico), os_peca→(os, peca), customer→user
- Many-to-many: `USER_ROLE` table linking user↔role
- Required indexes: `customer.cpf`, `customer.cnpj`, `veiculo.placa`, `os.status`, `os.cliente_id`, `user.email`
- IDs: UUID; timestamps: UTC ISO 8601
- Schema changes only via versioned migration files — never alter schema manually

## Testing Requirements
- Critical domains (OS, estoque, orcamento): **90% minimum coverage**
- All other domains: **80% minimum coverage**
- Unit test: status transitions (valid + invalid paths), budget calculation, CPF/CNPJ/placa validators, stock decrement (success, out-of-stock, concurrent), RBAC role checks (403 for out-of-scope)
- Integration test: full OS creation flow, each status transition, approve/reject budget (CLIENTE role), public status endpoint (no auth), 401 on admin endpoints without token, 403 on wrong role, CLIENTE filtered by customerId
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
`os`, `cliente`, `veiculo`, `servico`, `peca`, `estoque`, `orcamento`, `auth`, `usuario`, `infra`, `db`, `api`

### Rules
- Subject line max 72 characters, lowercase, no period at end
- Use imperative mood: "add feature" not "added feature"
- Breaking changes must include `BREAKING CHANGE:` footer or `!` after type: `feat(auth)!: change JWT payload structure`
- Reference issues in footer: `Closes #123`
