# Auto Repair Shop

Monolithic layered backend for an auto repair shop management system. Manages service orders (SO), customers, vehicles, parts/stock, and administrative operations.

**Stack:** Go · Gin · PostgreSQL · JWT · Swagger

## Domain

- **SO (Service Order)** — Service order, the central aggregate
- **Customer** — Customer identified by CPF (individual) or CNPJ (company)
- **Vehicle** — Vehicle (plate, brand, model, year), linked to a Customer
- **Service** — Billable service (name, description, unit price)
- **Part/Supply** — Part or supply with stock quantity and unit price
- **Budget** — Budget, auto-calculated from services and parts on a SO
- **Stock** — Stock/inventory of Parts

## SO Status Lifecycle

```
NEW → RECEIVED → IN_DIAGNOSIS → AWAITING_APPROVAL → IN_PROGRESS → COMPLETED → DELIVERED
```

`REJECTED` is a terminal state reached when the customer rejects the budget (from `AWAITING_APPROVAL`).

## Work Status Lifecycle (within a Service Order)

Each work item linked to a service order follows its own lifecycle:

```
AWAITING_START → IN_PROGRESS → COMPLETED
                     ↓               ↓
                 CANCELLED       CANCELLED  (can cancel from any non-terminal status)
```

| Status          | Description                                          |
|-----------------|------------------------------------------------------|
| `AWAITING_START`| Work linked to the SO, not yet started               |
| `IN_PROGRESS`   | Work is being executed by the mechanic               |
| `COMPLETED`     | Work was successfully finished                        |
| `CANCELLED`     | Work was cancelled (terminal state)                  |

Every status change is recorded as an immutable entry in `work_service_order_status_history`, preserving the full audit trail.

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
- Make

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
- Para customers criados automaticamente, a senha padrão (CPF/CNPJ) também é imediatamente hasheada antes de persistir.
- Bcrypt já aplica salt de forma segura; não é necessário gerir salt manualmente.

### Adding New Migrations

When adding schema changes, create migration files following the naming convention:
```
NNNNNN_description.up.sql
NNNNNN_description.down.sql
```

Where `NNNNNN` is a sequential 6-digit number (e.g., `000002_add_customer_status.up.sql`).

## Environment Variables

| Variable            | Description                                      |
|---------------------|--------------------------------------------------|
| `PORT`              | Server port (default: 8080)                      |
| `POSTGRES_USER`     | PostgreSQL username                              |
| `POSTGRES_PASSWORD` | PostgreSQL password                              |
| `POSTGRES_DB`       | PostgreSQL database name                          |
| `POSTGRES_HOST`     | PostgreSQL host (`db` in Docker, `localhost` local) |
| `POSTGRES_PORT`     | PostgreSQL port (default: 5432)                  |
| `JWT_SECRET`        | Secret key for JWT signing                        |
| `JWT_EXPIRES_IN`    | Token expiration duration (e.g. `24h`)           |
| `BCRYPT_COST`       | bcrypt hashing cost factor                        |
| `LOG_LEVEL`         | Application log level (default: `INFO`)           |

## API Endpoints

### Public

| Method | Path                 | Description         |
|--------|----------------------|---------------------|
| POST   | `/v1/auth/login`     | Get JWT             |
| GET    | `/v1/so/:id/status`  | Customer SO tracking  |

### Customers

| Method | Path                  | Roles               |
|--------|-----------------------|----------------------|
| POST   | `/v1/customers`       | ADMIN, ATTENDANT     |
| GET    | `/v1/customers`       | ADMIN, ATTENDANT     |
| GET    | `/v1/customers/:id`   | ADMIN, ATTENDANT     |
| PUT    | `/v1/customers/:id`   | ADMIN, ATTENDANT     |
| DELETE | `/v1/customers/:id`   | ADMIN                |

### Service Orders (SO)

| Method | Path                                          | Roles                               |
|--------|-----------------------------------------------|-------------------------------------|
| POST   | `/v1/service-order`                           | ADMIN, ATTENDANT                    |
| GET    | `/v1/service-order`                           | ADMIN, ATTENDANT, MECHANIC          |
| GET    | `/v1/service-order/:id`                       | ADMIN, ATTENDANT, MECHANIC          |
| PUT    | `/v1/service-order/:id/start-diagnosis`       | ADMIN, ATTENDANT, MECHANIC          |
| PUT    | `/v1/service-order/:id/send`                  | ADMIN, ATTENDANT, MECHANIC          |
| PUT    | `/v1/service-order/:id/accept`                | CUSTOMER (own SO only)              |
| PUT    | `/v1/service-order/:id/reject`                | CUSTOMER (own SO only)              |
| PUT    | `/v1/service-order/:id/deliver`               | ADMIN, ATTENDANT                    |
| PUT    | `/v1/service-order/:id/cancel`                | ADMIN, ATTENDANT                    |
| GET    | `/v1/service-order/:id/history`               | ADMIN, ATTENDANT, MECHANIC          |

### Work Status Transitions (within a SO)

| Method | Path                                                | Roles                      |
|--------|-----------------------------------------------------|----------------------------|
| PUT    | `/v1/service-order/:id/work/:workId/next`           | ADMIN, ATTENDANT, MECHANIC |
| PUT    | `/v1/service-order/:id/work/:workId/cancel`         | ADMIN, ATTENDANT, MECHANIC |

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
| ATTENDANT  | Customers, vehicles, SO, budgets, reports            |
| MECHANIC   | SO queries, add services/parts, status transitions   |
| CUSTOMER   | Own SO tracking and budget approve/reject only       |

## License

This project is part of the FIAP 15SOAT postgraduate program.

## Members

- Caetano Agostinho de Freitas - RM371006
- Emanuel Jesus Santos - RM371184
- Diogo Estevão Ferreira - RM371059
- Giusier Ferreira Soares - RM371064
- Guilherme Ferreira Santos - RM374002
