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

## Getting Started

```bash
# Run the application
make run

# Run tests
make test

# Generate coverage report
make coverage
```

The server starts on port `8080`.

## Swagger UI

You can access the service Swagger UI in your browser to view and test the available endpoints:

http://localhost:8080/swagger/index.html

The UI loads the OpenAPI specification from `/swagger.yaml` and lets you execute requests directly against the running server.

## Environment Variables

| Variable       | Description                  |
|----------------|------------------------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET`   | Secret key for JWT signing   |
| `JWT_EXPIRY`   | Token expiration duration    |
| `BCRYPT_COST`  | bcrypt hashing cost factor   |

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
