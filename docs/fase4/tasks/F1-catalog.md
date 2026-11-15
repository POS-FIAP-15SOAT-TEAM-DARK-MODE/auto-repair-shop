# Ficha F1 — Catalog (`auto-repair-shop-catalog-service`, :8082, PG `catalog`)

Pré-requisitos: F0 concluída; ler `_baseline.md`, `contracts/{resources,messaging,internal-rest,ddl}.md`. `<svc>=catalog-service`, `<db>=catalog`, secret `auto-repair-shop-<env>/catalog`.

## T1 — Baseline
Executar B1–B3 e B6–B7 de `_baseline.md`. Com mensageria (consumer) → compose com `ministack` + `aws-init`.

## T2 — Código de domínio (mover do monólito `$SRC`)
- `cp -r $SRC/internal/work $SRC/internal/supply` para `internal/` (inclui testes e mocks).
- `internal/pkg/auth/roles.go` (novo): copiar de `$SRC/internal/auth/domain/auth.go` **apenas** `type Role string`, constantes `ADMIN/ATTENDANT/MECHANIC/CUSTOMER` e os grupos `AttendantRoles`, `MechanicRoles`, `CustomerRoles`, `AttendantAndMechanicRoles`.
- `internal/app/middleware/auth.go`: trocar o import `internal/auth/domain` por `internal/pkg/auth` (tipo `auth.Role`). Ajustar `*_test.go` do middleware.
- Verificar: `grep -rn 'internal/auth/domain\|internal/customer\|internal/vehicle\|internal/service_order' --include=*.go .` **vazio**.
- `internal/app/container/handlers.go`: `HandlersWrapper{WorkHandler, SupplyHandler, ReservationHandler}`; `bootstrap.go`: criar só esses handlers; remover `metrics.NewStatusDurationCollector`.
- `internal/app/routing/routing.go`: manter `/ping`, `/metrics`, swagger, `/v1/works*` (GET atendente+mecânico; POST/PUT/DELETE atendente) e `/v1/supplies*` (atendente+mecânico) **idênticos ao original**; adicionar o grupo `/internal` (T3).

## T3 — Endpoints internos e reservas (código novo, mesmo padrão em camadas `domain/ interfaces/ adapters/ repository/ controller.go service.go di.go`)
- Middleware `InternalAuth()` (novo, em `internal/app/middleware/internal_auth.go`): compara header `X-Internal-Token` com `INTERNAL_TOKEN` (`crypto/subtle`); falha → `401` no formato de erro padrão. Teste unitário.
- `GET /internal/works/:id` → `{id,name,unit_price (string),active}` (`status`→`active`); `404` se inexistente.
- Pacote `internal/reservation` (novo):
  - `Reserve(ctx, serviceOrderID, supplyID, qty)`: valida `qty>0` (`422`), supply existe (`404`), aplica regra do `ddl.md` (upsert quando `RELEASED`; `409` se já `RESERVED` ou sem estoque), decrementa estoque com `UPDATE supply SET stock_quantity = stock_quantity - $1, version = version + 1 WHERE id=$2 AND stock_quantity >= $1` (0 linhas → `409`), tudo em **uma** transação (`uow`). Retorna `{reservation_id, supply_id, name, unit_price}` (`201`).
  - `Release(ctx, reservationID)`: `RESERVED`→`RELEASED` + devolve estoque na mesma transação; já `RELEASED` → `204` sem efeito; inexistente → `404`.
  - `ReleaseAllByServiceOrder(ctx, serviceOrderID)`: idem para todas as `RESERVED` da OS.
  - Rotas: `POST /internal/reservations` (`201`), `DELETE /internal/reservations/:id` (`204`).
- Repositórios Postgres **e** in-memory (para testes unitários), como os demais domínios.

## T4 — Consumer SQS (`internal/pkg/messaging`, versão mínima)
- `envelope.go`: struct `Envelope` (campos de `messaging.md`) + `Decode([]byte)`.
- `consumer.go`: loop com `ReceiveMessage` (long poll 20 s, `MaxNumberOfMessages=10`), `Handler` registrado por `type`; para cada mensagem: abrir transação → `INSERT INTO processed_message(message_id) ON CONFLICT DO NOTHING` (0 linhas → duplicada: `DeleteMessage` e seguir) → executar handler **na mesma transação** → commit → `DeleteMessage`. Erro → não deleta (volta à fila → DLQ). Tipo desconhecido → log e delete.
- Handler `ReleaseStock` → `ReleaseAllByServiceOrder(env.service_order_id)`.
- Cliente SQS atrás de uma interface (`ReceiveMessage`, `DeleteMessage`) para testar com fake; construtor lê `AWS_REGION`/`AWS_ENDPOINT_URL` (`aws-sdk-go-v2`, `BaseEndpoint` quando definido).
- Ligado no `bootstrap` como goroutine com `context` cancelado no shutdown; **desabilitado se `CATALOG_QUEUE_URL` vazio** (permite rodar sem MiniStack).

## T5 — Migrations e seed
- `migrations/000001_init_schema.{up,down}.sql` = seção **catalog** de `contracts/ddl.md` (+ `processed_message`).
- `internal/app/seed.go`: copiar `insertWorks`/`insertSupplies` de `$SRC/internal/app/seed.go`, **mantendo nomes/preços/quantidades** e trocando `uuid.NewString()` pelos IDs fixos de `resources.md` (works `1111…0001..0005`, supplies `2222…`); `isAlreadySeeded` passa a checar `COUNT(*) FROM "work"`; só roda se `SEED_ENABLED=true`.

## T6 — Dev token, testes e docs
- `cmd/devtoken/main.go` (novo): imprime um JWT HS256 com claims `user_id`, `roles` (flag `-role ATTENDANT`) usando `JWT_SECRET` e o mesmo gerador de `internal/pkg/auth`; alvo `make token ROLE=ATTENDANT`.
- Testes unitários: mover os existentes; criar para `reservation` (service com repos in-memory: reservar, sem estoque, upsert após liberar, liberar idempotente), `InternalAuth`, consumer (fake SQS: duplicada ignorada, tipo desconhecido, erro de handler não deleta).
- `internal/integration/` (`//go:build integration`): subir app com Postgres real (+ MiniStack): fluxo `POST /internal/reservations` (201) → estoque diminui → segunda reserva sem estoque (409) → `ReleaseStock` publicado **2×** no SQS → estoque restaurado **uma vez**; `GET /v1/works` com token ATTENDANT (200) e sem token (401).
- `docs/swagger.yaml`: paths `/v1/works*` e `/v1/supplies*` do original. README (propósito, endpoints, mensagens consumidas, como subir, cobertura), ADR `0002-stock-reservation-semantics.md`.

## T7 — Entrega
B8 de `_baseline.md`.

## Aceite (todos verificados e reportados com a saída do comando)
- [ ] `go build ./... && go vet ./...` ok; `LOG_LEVEL=PANIC go test ./... --race` ok; **cobertura total ≥ 80%** (`go tool cover -func=coverage.out | tail -1`)
- [ ] `docker compose up -d --build`: `GET :8082/ping` = 200; `make token ROLE=ATTENDANT` + `GET /v1/works` lista os 5 works do seed (`SEED_ENABLED=true`)
- [ ] `POST /internal/reservations` com `X-Internal-Token` (201/409/401/404/422 conforme `internal-rest.md`)
- [ ] `ReleaseStock` duplicado restaura o estoque uma única vez (teste de integração)
- [ ] `kubectl kustomize` dos 3 overlays ok; nenhum nome genérico (`app-config`…) restante
- [ ] PR aberto para `develop`; jobs `pre-commit`, `build`, `unit-tests`, `integration-tests`, `coverage` verdes (`sonarqube` conforme secret)
