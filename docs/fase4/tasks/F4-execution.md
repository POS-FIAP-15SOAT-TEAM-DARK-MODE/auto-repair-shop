# Ficha F4 — Execution (`auto-repair-shop-execution-service`, :8084, PG `execution`) — serviço novo

Pré-requisitos: onda 2 (PR do Billing disponível); ler `_baseline.md` (**inclui a seção B9 — lições da onda 2**), `contracts/{resources,messaging,ddl}.md`. `<svc>=execution-service`, `<db>=execution`, secret `auto-repair-shop-<env>/execution`.

## T1 — Baseline e mensageria
- B1–B3 e B6–B7 (compose com `db`, `migrate`, `ministack`, `aws-init`). `internal/pkg/auth/roles.go` como no F1/T2.
- **Copiar `internal/pkg/messaging/*` do Billing** (envelope, outbox, relay, consumer, sns, sqs, config e testes) — ver B9 para o caminho; trocar apenas o module path. O compose cria `execution-queue` (+DLQ), assinada em `saga-commands` com filtro `target=execution`, e o tópico `execution-events`.

## T2 — Domínio (camadas simples: portas como interfaces Go no pacote; repositórios Postgres e in-memory em `repository/`; testes `sqlmock` para cobertura)
- `Task{ID, ServiceOrderID, SagaID, Status, CreatedAt, StartedAt, CompletedAt, Works}`; `Work{ID, TaskID, WorkID, Name, Status}`.
- Máquina do work: copiar de `$SRC/internal/service_order/domain/work_service_order_history.go` a lógica de `NextStatus` (`AWAITING_START → IN_PROGRESS → COMPLETED`; `CANCELLED` terminal; `ErrWorkServiceOrderAlreadyCompleted/AlreadyCancelled/InvalidStatus`) e os erros `ErrWorkServiceOrderNotFound`. Cancelar: de `AWAITING_START`/`IN_PROGRESS`; `COMPLETED`/`CANCELLED` → erro.
- Erros novos: `ErrTaskNotFound` (404), `ErrTaskNotActive` (409), `ErrWorksNotFinished` (409). **Copiar o mapeamento erro→HTTP de `$SRC/internal/pkg/web/errors.go`** para os erros de work já existentes e acrescentar os novos.

## T3 — Handlers de comando (cada um: 1 transação = dedupe + efeito + histórico + `Enqueue` + commit; propagar `saga_id`, `service_order_id`, `request_id` do envelope)
| Comando | Efeito | Eventos (`execution-events`) |
|---|---|---|
| `EnqueueExecution` | payload inválido (sem works, `work_id` repetido/vazio) → **não cria tarefa** e emite `ExecutionFailed{task_id:"", reason:"INVALID_REQUEST"}`. Tarefa já existe para a OS: `QUEUED`/`IN_PROGRESS` → só dedupe/log (nada a fazer); `COMPLETED`/`CANCELLED` → `ExecutionFailed{reason:"INVALID_REQUEST"}`. Senão cria `execution_task` (`QUEUED`, grava `saga_id`) + `execution_work` (`AWAITING_START`) + 1 linha de `execution_work_history` por work (`previous` NULL) | 1× `ExecutionWorkStatusChanged{work_id,work_name,previous:null,new:"AWAITING_START"}` por work, depois `ExecutionQueued{task_id}` |
| `CancelExecution` | tarefa `QUEUED`/`IN_PROGRESS`: todo work não terminal → `CANCELLED` (+ histórico) e tarefa → `CANCELLED`; sem tarefa ou já terminal → só dedupe/log | `ExecutionWorkStatusChanged` por work afetado (`new:"CANCELLED"`); **sem** `ExecutionFailed` (quem cancelou foi o SO) |

## T4 — API REST (papéis de `internal/pkg/auth`; propagar `X-Request-Id`; eventos usam o `saga_id` da tarefa)
| Rota | Papéis | Regra / resposta |
|---|---|---|
| `GET /v1/execution/queue` | Attendant, Mechanic, Admin | `200 {"items":[{"taskId","serviceOrderId","status","createdAt","works":[{"workId","name","status"}]}]}` com tarefas `QUEUED`/`IN_PROGRESS`, `created_at` crescente |
| `PUT /v1/service-order/:id/work/:workId/next` | Mechanic, Admin | tarefa da OS (`404` se não há) ativa (`409` se não); work da tarefa (`404`); `NextStatus` (`409` se já concluído/cancelado). Se o work vai a `IN_PROGRESS` e a tarefa está `QUEUED` → tarefa `IN_PROGRESS` + `started_at`. Grava histórico e emite `ExecutionWorkStatusChanged`. `204` |
| `PUT /v1/service-order/:id/work/:workId/cancel` | Mechanic, Admin | idem acima com cancelamento. `204` |
| `PUT /v1/service-order/:id/finish` | Mechanic, Admin | tarefa ativa (`409`); **todos** os works `COMPLETED`/`CANCELLED` (senão `409 ErrWorksNotFinished`). Se **todos** estiverem `CANCELLED` → tarefa `CANCELLED` + `ExecutionFailed{task_id,reason:"ALL_WORKS_CANCELLED"}`; senão tarefa `COMPLETED` (+`completed_at`) + `ExecutionCompleted{task_id}`. `204` |
Nota de comportamento: o monólito não exigia works concluídos para finalizar a OS; a exigência é nova (decisão do spec) — documentar no README.

## T5 — Migrations, config, k8s
- `migrations/000001_init_schema` = seção **execution** de `ddl.md` + `processed_message` + `outbox`. Sem seed.
- Variáveis (`resources.md`): `PORT`, `POSTGRES_*`, `JWT_SECRET`, `AWS_REGION`, `AWS_ENDPOINT_URL`, `EXECUTION_QUEUE_URL`, `EXECUTION_EVENTS_TOPIC_ARN`. No `docker.yml`: passo `Resolve messaging endpoints` (B4) com fila `execution-queue` e tópico `execution-events`.

## T6 — Testes (cobertura total ≥ 80%; **≥ 90%** em `task` e `work`; gate sem mocks)
- Unitários: ciclo do work (todas as transições, inclusive inválidas), `finish` (todos concluídos, mistura concluídos/cancelados, todos cancelados, pendentes), handlers `EnqueueExecution` (válido, inválido, duplicado em cada estado) e `CancelExecution`, saga_id propagado, rotas com papéis (401/403).
- Contrato: `testdata/contracts/<Type>.json` para `EnqueueExecution`, `CancelExecution`, `ExecutionQueued`, `ExecutionWorkStatusChanged` (com `previous:null` e com valor), `ExecutionCompleted`, `ExecutionFailed` (exemplos de `messaging.md`).
- Integração (`-tags=integration`, Postgres + MiniStack): publicar `EnqueueExecution` → eventos na fila de teste assinada em `execution-events`; mesma mensagem 2× ⇒ 1 tarefa e 1 `ExecutionQueued`; fluxo REST `next`/`next`/`finish` → `ExecutionCompleted` com o `saga_id` do comando; `CancelExecution`.

## T7 — Docs e entrega
Swagger (queue, next, cancel, finish), README (comandos/eventos, ciclo, a exigência nova do `finish`), ADR `0002-execution-owns-work-progress.md`, limpar exemplos do monólito em `.ai/rules/*`. Entrega B8.

## Aceite
- [ ] `go build`, `go vet` (+`-tags=integration`), `go test ./... --race`; cobertura total ≥ 80% (≥ 90% em `task`/`work`)
- [ ] `docker compose config -q` ok; CI: `pre-commit`, `build`, `unit-tests`, `integration-tests`, `coverage` verdes
- [ ] Mesma mensagem 2× ⇒ 1 efeito; `saga_id` do `EnqueueExecution` presente em todos os eventos posteriores (inclusive os disparados por REST)
- [ ] 3 overlays kustomize ok; nomes `execution-service` em todos os objetos; PR aberto para `develop`
