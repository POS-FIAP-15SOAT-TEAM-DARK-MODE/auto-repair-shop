# Auto Repair Shop

Monolithic layered backend for an auto repair shop management system. Manages service orders (SO), clients, vehicles, parts/stock, and administrative operations.

**Stack:** Go · Gin · PostgreSQL · JWT · Swagger

## Domain

- **SO (Service Order)** — Service order, the central aggregate
- **Client** — Client identified by CPF (individual) or CNPJ (company)
- **Vehicle** — Vehicle (plate, brand, model, year), linked to a Client
- **Service** — Billable service (name, description, unit price)
- **Part/Supply** — Part or supply with stock quantity and unit price
- **Budget** — Budget, auto-calculated from services and parts on a SO
- **Stock** — Stock/inventory of Parts

## SO Status Lifecycle

```
RECEIVED → IN_DIAGNOSIS → AWAITING_APPROVAL → IN_PROGRESS → COMPLETED → DELIVERED
```

`REJECTED` is a terminal state reached when the client rejects the budget (from `AWAITING_APPROVAL`).

## Project Structure

```
cmd/
  service/          # Application entrypoint
internal/
  domain/           # Domain models (framework-free)
  infra/
    container/      # Dependency injection container
    factory/        # Container factory
    handler/        # HTTP handlers
  routing/          # Route definitions
  services/         # Business logic layer
```

## Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- PostgreSQL (provided via Docker Compose)

## Contributing
To contribute with the project you should install pre-commit:

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

### Setup with Docker

```bash
# 1. Start the database
make docker-up

# 2. Install migration tool (first time only)
make migrate-install

# 3. Run all pending migrations
make migrate-up

# 4. Start the application locally
make run
```

### Local Development

```bash
# Run the application locally
make run

# Run tests
make test

# Generate coverage report
make coverage

# Stop Docker containers
make docker-down
```

The server starts on port `${PORT:-8080}` (default: 8080).

## Swagger UI

You can access the service Swagger UI in your browser to view and test the available endpoints:

http://localhost:8080/swagger/index.html

The UI loads the OpenAPI specification from `/swagger.yaml` and lets you execute requests directly against the running server.

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

### Metodo de Criptografia
Senhas de usuários são protegidas usando bcrypt (golang.org/x/crypto/bcrypt).

Detalhes principais:
- Tipo: hash one‑way (não é reversível). O resultado inclui salt interno e metadados.
- Implementação: usamos `bcrypt.GenerateFromPassword` ao criar/atualizar senhas e `bcrypt.CompareHashAndPassword` para validação.
- Fator de custo: controlado pela variável de ambiente `BCRYPT_COST` (ver seção Environment Variables). Recomenda‑se um custo mínimo de 12 em produção — aumente conforme a capacidade da infra.

Notas:
- Para clientes criados automaticamente, a senha padrão (CPF/CNPJ) também é imediatamente hasheada antes de persistir.
- Bcrypt já aplica salt de forma segura; não é necessário gerir salt manualmente.

### Adding New Migrations

When adding schema changes, create migration files following the naming convention:
```
NNNNNN_description.up.sql
NNNNNN_description.down.sql
```

Where `NNNNNN` is a sequential 6-digit number (e.g., `000002_add_customer_status.up.sql`).

## Environment Variables

| Variable       | Description                  |
|----------------|------------------------------|
| `PORT`         | Server port (default: 8080)  |
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET`   | Secret key for JWT signing   |
| `JWT_EXPIRY`   | Token expiration duration    |
| `BCRYPT_COST`  | bcrypt hashing cost factor   |
| `POSTGRES_USER` | PostgreSQL username (for migrations) |
| `POSTGRES_PASSWORD` | PostgreSQL password (for migrations) |

## API Endpoints

### Public

| Method | Path                 | Description         |
|--------|----------------------|---------------------|
| POST   | `/v1/auth/login`     | Get JWT             |
| GET    | `/v1/so/:id/status`  | Client SO tracking  |

### Clients

| Method | Path               | Roles               |
|--------|--------------------|----------------------|
| POST   | `/v1/clients`      | ADMIN, ATTENDANT     |
| GET    | `/v1/clients`      | ADMIN, ATTENDANT     |
| GET    | `/v1/clients/:id`  | ADMIN, ATTENDANT     |
| PUT    | `/v1/clients/:id`  | ADMIN, ATTENDANT     |
| DELETE | `/v1/clients/:id`  | ADMIN                |

### Service Orders (SO)

| Method | Path                    | Roles                          |
|--------|-------------------------|--------------------------------|
| POST   | `/v1/so`                | ADMIN, ATTENDANT               |
| GET    | `/v1/so`                | ADMIN, ATTENDANT, MECHANIC     |
| GET    | `/v1/so/:id`            | ADMIN, ATTENDANT, MECHANIC     |
| PATCH  | `/v1/so/:id/status`     | ADMIN, ATTENDANT, MECHANIC     |
| POST   | `/v1/so/:id/approve`    | CLIENT (own SO only)           |
| POST   | `/v1/so/:id/reject`     | CLIENT (own SO only)           |

### Admin

| Method | Path                           | Roles              |
|--------|--------------------------------|---------------------|
| POST   | `/v1/users`                    | ADMIN               |
| PUT    | `/v1/users/:id/roles`          | ADMIN               |
| GET    | `/v1/reports/average-time`     | ADMIN, ATTENDANT    |

## RBAC Roles

| Role       | Description                                          |
|------------|------------------------------------------------------|
| ADMIN      | Unrestricted access (exclusive, cannot combine)      |
| ATTENDANT  | Clients, vehicles, SO, budgets, reports              |
| MECHANIC   | SO queries, add services/parts, status transitions   |
| CLIENT     | Own SO tracking and budget approve/reject only       |

## License

This project is part of the FIAP 15SOAT postgraduate program.

## Members

- Caetano Agostinho de Freitas - RM371006
- Emanuel Jesus Santos - RM371184
- Diogo Estevão Ferreira - RM371059
- Giusier Ferreira Soares - RM371064
- Guilherme Ferreira Santos - RM374002
