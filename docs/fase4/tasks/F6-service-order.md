# Ficha F6 — Service Order + orquestrador (`auto-repair-shop-service-order-service`, :8086, PG `service_order`)

Pré-requisitos: onda 2 (Billing); ler `_baseline.md` (B9), `contracts/{resources,messaging,internal-rest,ddl}.md` e `SPEC.md` §3.1–3.2 (**máquina de estados e saga**). `<svc>=service-order-service`, `<db>=service_order`, secret `auto-repair-shop-<env>/service-order`. É o serviço mais complexo: **nenhuma regra nova de negócio além das descritas aqui; em dúvida, parar e reportar.**

## T1 — Baseline e mensageria
- B1–B3, B6–B7; `internal/pkg/auth/roles.go` como no F1/T2 e `GetClaimsFromHeader` (já em `internal/pkg/auth/jwt.go`).
- **Copiar `internal/pkg/messaging/*` do Billing** (B9) e **adaptar o consumer**: depois do `BeginTx`, anexar a transação ao `context` (`uow.WithTx`, helper exportado a ser adicionado ao `uow` copiado do Catalog — ver `joinableUoW` em `auto-repair-shop-catalog-service/internal/pkg/uow/postgres.go`) e chamar o handler com `(ctx, tx, env)`; assim os repositórios baseados em `uow`/`GetTransaction(ctx)` funcionam dentro da transação do consumer.
- Compose: `db`, `migrate`, `ministack`, `aws-init` (tópicos `saga-commands`, `so-events`, `billing-events`, `execution-events`; fila `so-saga-replies` assinada em `billing-events` e `execution-events`; **mais** filas de teste `test-billing-commands`, `test-execution-commands`, `test-catalog-commands` assinadas em `saga-commands` com os filtros `target`, para os testes lerem os comandos) e stubs (`profiles: ["stubs"]`) de Customer e Catalog em `local/stubs/` (WireMock com os JSON de `internal-rest.md`).

## T2 — Domínio e persistência (mover/adaptar de `$SRC/internal/service_order`)
- Copiar `domain/service_order.go` (status, `IsCancelable`, erros), `domain/service_order_errors.go`, `domain/notifier.go`, `adapters/*` (DTOs), `controller.go`, `mappers.go`, `repository/*` (Postgres, in-memory, queries) e testes. **Remover** tudo de `wso_history*`, `work_service_order_history`, `AverageExecutionTimeInHours`, `NextWork`, `CancelWork`, `Finish`-por-works (agora Execution/History) e dependências de `customer`, `vehicle`, `work`, `supply` (`grep` por `internal/(customer|vehicle|work|supply|auth)` deve ficar **vazio**).
- `ServiceOrder` agora guarda o **snapshot** (`ddl.md`, seção `service_order`): `CustomerSnapshot{ID, UserID, Name, Email, Type, Document, CompanyName, Phone}`, `VehicleSnapshot{ID, LicensePlate, Brand, Model, Year}`; `Works`/`Supplies` com os campos do snapshot (`name`, `description`, `unit_price`, `version`, `quantity`, `reservation_id`). Repositórios gravam/leem as colunas novas.
- **Preço**: `total_amount = Σ work.unit_price + Σ supply.unit_price × quantity`, recalculado **de forma síncrona na mesma transação** a cada inclusão/remoção (substitui o `reviewOSPricing` assíncrono).
- **Formatos de resposta idênticos ao monólito** (`adapters/so_read.go`): `SOResponse`, `PaginatedSOResponse`, `SODetailResponse` com `customer{id,name,email,type,document,companyName,phone}` e `vehicle{id,licensePlate,brandModel:"<Brand> - <Model>",year}`, `works[{id,name,description,price,status}]` (`status`: constante `"ACTIVE"` — o SO só aceita works ativos), `supplies[{id,name,description,price,version,amount}]` (`amount` = quantity). Listagem: mesmos filtros (`status`, `customerId`, `vehicleId`, `page`, `pageSize`, `sort_by`, `sort_order`).

## T3 — Clientes REST (pacotes `internal/clients/customer` e `internal/clients/catalog`)
Interfaces + implementação `net/http` (timeout 3 s, header `X-Internal-Token`, propaga `X-Request-Id`; **sem retry em POST**, 1 retry em GET) conforme `internal-rest.md`: `GetCustomer`, `GetVehicle`, `GetWork`, `Reserve`, `ReleaseReservation`. Mapeamento: `404` → `ErrCustomerNotFound`/`ErrVehicleNotFound`/`ErrWorkNotFound`/`ErrSupplyNotFound`; `409` → `ErrSupplyOutOfStock`; `422` → erro de validação; timeout/5xx/conexão → `ErrUpstreamUnavailable` (HTTP **502**). Testes com `httptest`.

## T4 — Fluxos locais (sem saga) — regras de status **iguais ao monólito**
| Operação | Regra |
|---|---|
| `POST /v1/service-order` | busca cliente e veículo (veículo deve ser do cliente, senão `ErrInvalidCustomerProperty`); resolve works (ativos) e **reserva** supplies no Catalog; gera ULID **antes** de reservar; grava OS `NEW` + links + snapshot em **uma** transação; se a gravação falhar, libera as reservas feitas (`DELETE`, melhor esforço). Publica `ServiceOrderStatusChanged{previous:null,new:"NEW"}` |
| `POST …/:id/works`, `DELETE …/works/:serviceId` | só `NEW` (`ErrServiceOrderNotNew`); works via Catalog |
| `POST …/:id/supplies`, `DELETE …/supplies/:supplyId` | só `IN_DIAGNOSIS` (`ErrServiceOrderNotInDiagnosis`); `POST` reserva no Catalog; `DELETE` chama `DELETE /internal/reservations/{reservation_id}` e remove o link |
| `PUT …/received` | `NEW → RECEIVED` |
| `PUT …/start-diagnosis` | `RECEIVED → IN_DIAGNOSIS` |
| `PUT …/deliver` | `COMPLETED → DELIVERED` |
| `GET …`, `/:id`, `/:id/works`, `/:id/supplies`, `/:id/status` (público) | como hoje, lendo só o banco local |
Toda mudança de status: grava, publica `ServiceOrderStatusChanged` (outbox; `reason` quando aplicável), incrementa `service_order_status_transitions_total`, e chama o `StatusNotifier` (log, com e-mail mascarado) **depois do commit**, melhor esforço (falha só conta em `service_order_notification_failures_total`).

## T5 — Saga (`internal/saga`) — regras exatas
Estado em `saga_instance` (`ddl.md`). Todos os comandos usam o `saga_id` = `saga_instance.id`; toda transição abaixo é **uma transação** (estado da OS + `saga_instance` + `outbox` + `processed_message` quando vier de mensagem). Mensagem cujo `saga_id` não existe, ou cuja saga não está `RUNNING` ou não está no `step` esperado, é **ignorada com log** (e apagada da fila).

| Gatilho | Guarda | Efeito |
|---|---|---|
| `PUT …/:id/send` (204) | OS `IN_DIAGNOSIS`, ≥ 1 work (`ErrEmptyServicesList`), sem saga `RUNNING` (`409`) | cria saga `step=BUDGET`, `budget_requested=true`, `deadline=now+SAGA_STEP_TIMEOUT`; outbox `GenerateBudget{customer_user_id, works[…], supplies[…]}` (alvo `billing`). **Status continua `IN_DIAGNOSIS`** |
| `BudgetGenerated` | step `BUDGET` | `total_amount`, `quote_sent_at=now`; OS `IN_DIAGNOSIS → AWAITING_APPROVAL`; saga `step=AWAITING_CUSTOMER`, `deadline=NULL` |
| `PUT …/:id/accept` (204) | JWT `user_id == customer_user_id` (senão `ErrInvalidCustomerProperty`); OS `AWAITING_APPROVAL`; saga `AWAITING_CUSTOMER` (senão `409`) | OS → **`IN_PROGRESS` imediato**; saga `step=PAYMENT`, `payment_requested=true`, `deadline=now+SAGA_PAYMENT_TIMEOUT`; outbox `RequestPayment{amount=total_amount}` |
| `PUT …/:id/reject` (204) | mesmas guardas do accept | OS → `REJECTED` (`reason=BUDGET_REJECTED`); compensação: `CancelBudget` + `ReleaseStock` (alvo `catalog`); saga `COMPENSATED` |
| `PaymentApproved` | step `PAYMENT` | `payment_approved=true`; saga `step=EXECUTION_QUEUE`, `execution_enqueued=true`, `deadline=now+SAGA_STEP_TIMEOUT`; outbox `EnqueueExecution{works[{work_id,name}]}` (alvo `execution`) |
| `PaymentFailed` | step `PAYMENT` | OS `IN_PROGRESS → REJECTED` (`reason=PAYMENT_REJECTED`; **transição permitida só pela saga**); compensação `CancelBudget` + `ReleaseStock`; saga `COMPENSATED` |
| `ExecutionQueued` | step `EXECUTION_QUEUE` | saga `step=EXECUTION`, `deadline=NULL` |
| `ExecutionCompleted` | step `EXECUTION` | OS `IN_PROGRESS → COMPLETED`, `completed_at=now`; saga `COMPLETED`, `step=DONE` |
| `ExecutionFailed` | step `EXECUTION_QUEUE` ou `EXECUTION` | OS → `CANCELLED` (`reason=EXECUTION_FAILED`); compensação `RefundPayment` + `CancelBudget` + `ReleaseStock`; saga `COMPENSATED` |
| `PUT …/:id/cancel` (204) | `IsCancelable()` (senão `ErrServiceOrderNotCancelable`) | OS → `CANCELLED` (`reason=USER_CANCELLED`); com saga `RUNNING`: compensa conforme as flags — `CancelExecution` (se `execution_enqueued`), `RefundPayment` (se `payment_requested`), `CancelBudget` (se `budget_requested`) —, e **sempre** `ReleaseStock` se a OS tem supplies; saga `COMPENSATED`. Sem saga: só `ReleaseStock` quando houver supplies |
| **Job de prazos** (a cada 30 s) | sagas `RUNNING` com `deadline < now` | `step=BUDGET`: saga `FAILED` e envia `CancelBudget` (idempotente); OS **continua `IN_DIAGNOSIS`** (novo `send` permitido). `PAYMENT` ou `EXECUTION_QUEUE`: compensa como no `cancel`, OS → `CANCELLED` (`reason=SAGA_TIMEOUT`) |
Compensações são **fire-and-forget**: o SO marca a saga `COMPENSATED` ao publicar os comandos e apenas loga `BudgetCancelled`/`PaymentRefunded` recebidos. A transição `IN_PROGRESS → REJECTED` e `IN_PROGRESS → CANCELLED` precisam ser aceitas pelo domínio; `REJECTED` continua não cancelável.

## T6 — Consumer, relay e endpoint interno
- Consumer em `SO_SAGA_REPLIES_QUEUE_URL` com os 6 tipos tratados (`BudgetGenerated`, `PaymentApproved`, `PaymentFailed`, `ExecutionQueued`, `ExecutionCompleted`, `ExecutionFailed`); `BudgetCancelled`, `PaymentRefunded`, `ExecutionWorkStatusChanged` → log e apagar. Relay publica `saga-commands` (atributos `type` e `target`) e `so-events`.
- `GET /internal/service-orders/exists?customer_id=|vehicle_id=` (`X-Internal-Token`; `InternalAuth` copiado do F1/T3) → `{"exists":bool}`; `422` se não houver exatamente um parâmetro.

## T7 — Migrations, config, k8s
`migrations/000001_init_schema` = seção **service_order** de `ddl.md` + `processed_message` + `outbox`. Variáveis (`resources.md`): `PORT`, `POSTGRES_*`, `JWT_SECRET`, `INTERNAL_TOKEN`, `CUSTOMER_URL`, `CATALOG_URL`, `AWS_REGION`, `AWS_ENDPOINT_URL`, `SO_SAGA_REPLIES_QUEUE_URL`, `SAGA_COMMANDS_TOPIC_ARN`, `SO_EVENTS_TOPIC_ARN`, `SAGA_STEP_TIMEOUT` (default 2m), `SAGA_PAYMENT_TIMEOUT` (default 24h). Overlays: `CUSTOMER_URL=http://customer-service:8081`, `CATALOG_URL=http://catalog-service:8082`. `docker.yml`: `Resolve messaging endpoints` (fila `so-saga-replies`; tópicos `saga-commands`, `so-events`) + secret `service-order`.

## T8 — Testes (cobertura total ≥ 80%; **≥ 90%** em `internal/saga` e `internal/service_order/domain`; gate sem mocks)
- Unitários: cada linha da tabela do T5 (caminho feliz **e** guardas violadas), idempotência por `message_id`, mensagem fora de ordem/step errado ignorada, `cancel` em cada status e em cada step da saga (verificar exatamente quais comandos saem), job de prazos (BUDGET vs PAYMENT), snapshots e preço síncrono, clientes REST (404/409/5xx/timeout), formatos de resposta.
- Contrato: `testdata/contracts/*.json` para os 7 comandos publicados/consumidos pelo SO e os 6 eventos recebidos (exemplos de `messaging.md`, com `customer_user_id` no `GenerateBudget`).
- Integração (`-tags=integration`, Postgres + MiniStack, Customer/Catalog simulados com `httptest`): o teste **faz o papel de Billing e Execution** lendo os comandos nas filas `test-*-commands` e publicando as respostas em `billing-events`/`execution-events`: (1) caminho feliz `send → BudgetGenerated → accept → PaymentApproved → ExecutionQueued → ExecutionCompleted` com os status esperados em cada passo (inclusive **`IN_PROGRESS` logo após o accept**); (2) `PaymentFailed` → `REJECTED` (`PAYMENT_REJECTED`) e comandos `CancelBudget`+`ReleaseStock`; (3) `ExecutionFailed` → `CANCELLED` e `RefundPayment`+`CancelBudget`+`ReleaseStock`; (4) `reject` → `REJECTED` (`BUDGET_REJECTED`) + `ReleaseStock`; (5) mesma mensagem 2× ⇒ 1 efeito; (6) cada mudança de status publica `ServiceOrderStatusChanged` em `so-events`.

## T9 — Docs e entrega
Swagger (todas as rotas públicas da OS; `accept`/`reject`/`send` com 409 documentado), README (máquina de estados, tabela da saga, snapshots, comandos/eventos), ADRs `0002-orchestrated-saga.md` e `0003-service-order-snapshots.md`, limpar `.ai/rules/*`. Entrega B8.

## Aceite
- [ ] `go build`, `go vet` (+`-tags=integration`), `go test ./... --race`; cobertura total ≥ 80% (≥ 90% em saga/domain)
- [ ] CI: `pre-commit`, `build`, `unit-tests`, `integration-tests`, `coverage` verdes (Sonar conforme token)
- [ ] Os 6 cenários de integração do T8 passam; `accept` deixa a OS em `IN_PROGRESS` **antes** de qualquer resposta do Billing
- [ ] Respostas de `GET` com o mesmo formato do monólito (testes com JSON de referência)
- [ ] `grep -rn 'internal/\(customer\|vehicle\|work\|supply\|auth\)' --include=*.go .` vazio; 3 overlays kustomize ok; PR aberto para `develop`
