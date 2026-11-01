# Ficha F3 — Billing (`auto-repair-shop-billing-service`, :8083, PG `billing`) — serviço novo

Pré-requisitos: F0; ler `_baseline.md`, `contracts/{resources,messaging,ddl}.md`. `<svc>=billing-service`, `<db>=billing`, secret `auto-repair-shop-<env>/billing`. **Este serviço cria o pacote `internal/pkg/messaging` completo** que Execution, History e Service Order vão copiar — manter pequeno, sem abstrações extras.

## T1 — Baseline
B1–B3 e B6–B7. Não há código de domínio para copiar (só `internal/pkg/auth/{jwt.go,roles.go}` — `roles.go` como na ficha F1/T2). Compose com `ministack` + `aws-init`: cria `billing-queue` (+DLQ), assinada em `saga-commands` com filtro `target=billing`, e o tópico `billing-events`.

## T2 — Pacote `internal/pkg/messaging`
- `envelope.go`: `Envelope{MessageID, SagaID, Type, OccurredAt, ServiceOrderID, RequestID, Payload json.RawMessage}` com tags JSON de `messaging.md`; `NewEnvelope(type, sagaID, soID, requestID, payload)` gera `message_id` (UUID v4) e `occurred_at` UTC.
- `outbox.go`: `Enqueue(ctx, tx, topic, target, env Envelope)` → `INSERT INTO outbox` **na transação do chamador**.
- `relay.go`: a cada 1 s, em transação: `SELECT … FROM outbox WHERE published_at IS NULL ORDER BY id LIMIT 20 FOR UPDATE SKIP LOCKED`; publica cada linha no SNS (`TopicArn` resolvido por mapa `topic lógico → ARN` vindo de env; atributos de mensagem `type` e, se houver, `target`); `UPDATE … SET published_at = now()` só após sucesso; erro → mantém e tenta de novo (log).
- `consumer.go`: idêntico ao descrito em F1/T4 (dedupe + handler na mesma transação + delete após commit), com `RegisterHandler(type, func(ctx, tx *sql.Tx, env Envelope) error)`.
- `sns.go`/`sqs.go`: wrappers finos sobre `aws-sdk-go-v2` atrás de interfaces (`Publish`, `ReceiveMessage`, `DeleteMessage`) para teste com fakes; `config.go` lê `AWS_REGION`/`AWS_ENDPOINT_URL`.
- Ligar relay e consumer no `bootstrap` (goroutines, shutdown por `context`); desabilitar se `BILLING_QUEUE_URL`/`BILLING_EVENTS_TOPIC_ARN` vazios.

## T3 — Domínio e regras (camadas `domain/ interfaces/ adapters/ repository/ service.go controller.go di.go`)
**Pacote `budget`:** `Budget{ID, ServiceOrderID, CustomerUserID, Status, TotalAmount, Items}`; total = `Σ work.unit_price + Σ supply.unit_price × quantity` com `shopspring/decimal` (duas casas; **nunca float**).
**Pacote `payment`:** `Payment{…}`; interface `PaymentProvider { Charge(ctx, p Payment) (Outcome, error); Refund(ctx, p Payment) error }`, `Outcome ∈ {Approved, Rejected, Pending}`. Implementação `internal.Provider` guiada por `PAYMENT_MODE`: `auto-approve`→Approved, `auto-reject`→Rejected, `manual`→Pending (default `manual`). Registrar `provider='internal'` na linha.
**Handlers de comando** (cada um: 1 transação = dedupe + efeito + `Enqueue` + commit):

| Comando | Efeito | Eventos (`billing-events`) |
|---|---|---|
| `GenerateBudget` | valida payload (≥ 1 work; preços ≥ 0; `customer_user_id` presente); se já existe orçamento `GENERATED` para a OS → só dedupe (nada a fazer); senão cria `budget` + `budget_item`s | `BudgetGenerated{budget_id,total_amount}` |
| `CancelBudget` | `GENERATED`→`CANCELLED`; sem orçamento/já cancelado → nada | `BudgetCancelled{budget_id}` (só quando mudou) |
| `RequestPayment` | exige orçamento `GENERATED` e `amount == total_amount` (senão log + descarta); cria `payment` `PENDING`; chama `Charge` | Approved→`PaymentApproved{payment_id,amount}`; Rejected→`PaymentFailed{payment_id,amount,reason:"PAYMENT_REJECTED"}`; Pending→nenhum |
| `RefundPayment` | `PENDING`→`CANCELLED` (sem evento); `APPROVED`→`REFUNDED` (+`Refund`); outros → nada | `PaymentRefunded{payment_id,amount}` (só em `APPROVED→REFUNDED`) |

Eventos e `saga_id`/`service_order_id`/`request_id` do envelope de entrada são **propagados** para o envelope de saída. Payload inválido → erro (mensagem volta à fila e vai à DLQ).

## T4 — API REST (papéis de `internal/pkg/auth`)
| Rota | Papéis | Resposta |
|---|---|---|
| `GET /v1/budgets/:serviceOrderId` | Attendant, Mechanic, Admin **e Customer dono** (`JWT user_id == budget.customer_user_id`; outro customer → `404`) | `200 {id,service_order_id,status,total_amount,items:[{kind,ref_id,name,quantity,unit_price}],created_at}`; `404` se não há orçamento `GENERATED` |
| `GET /v1/payments/:serviceOrderId` | Attendant, Admin | `200 {id,budget_id,service_order_id,amount,status,provider,created_at}` (pagamento mais recente da OS); `404` |
| `POST /v1/payments/:id/confirm` | Attendant, Admin | `204`; `409` se não `PENDING`; `404`. Efeito = `PaymentApproved` (mesma transação + outbox) |
| `POST /v1/payments/:id/reject` | Attendant, Admin | `204`; `409`; `404`. Efeito = `REJECTED` + `PaymentFailed` |

Valores monetários em JSON como **string decimal**. Sem rotas `/internal` neste serviço.

## T5 — Migrations, seed, config
- `migrations/000001_init_schema` = seção **billing** de `ddl.md` + `processed_message` + `outbox` (comuns). **Sem seed.**
- Overlays: `PAYMENT_MODE` — `lab`: `auto-approve`; `stg`: `auto-approve`; `prd`: `manual` (ConfigMap). `.env.example` default `manual`.
- Variáveis: `PORT`, `POSTGRES_*`, `JWT_SECRET`, `AWS_REGION`, `AWS_ENDPOINT_URL`, `BILLING_QUEUE_URL`, `BILLING_EVENTS_TOPIC_ARN`, `PAYMENT_MODE`.

## T6 — Testes
- Unitários (cobertura ≥ 80%; **≥ 90%** em `budget` e `payment`): cálculo do total (só works, só supplies, misto, arredondamento), todas as transições do pagamento (inclusive inválidas e repetidas), cada handler de comando com repositórios in-memory e `Enqueue` fake, os 3 modos do provider, outbox/relay (publica e marca; falha de publicação mantém), consumer (duplicada, tipo desconhecido, erro → não deleta).
- **Contrato:** para cada mensagem consumida/publicada, ler `testdata/contracts/<Type>.json` (exemplos de `messaging.md`) e validar decode/encode; criar os arquivos com os exemplos do contrato.
- Integração (`-tags=integration`, Postgres + MiniStack reais): publicar `GenerateBudget` na fila → `BudgetGenerated` aparece em fila de teste assinada em `billing-events`; fluxo `RequestPayment` nos 3 modos; mesma mensagem 2× não duplica (1 orçamento/1 pagamento/1 evento); `confirm`/`reject` via REST.

## T7 — Docs e entrega
Swagger dos 4 endpoints; README (propósito, comandos/eventos, modos de pagamento, como adicionar o Mercado Pago depois = nova implementação de `PaymentProvider`); ADR `0002-billing-without-payment-gateway.md`. Entrega B8.

## Aceite
- [ ] `go build`, `go vet`, `go test ./... --race` ok; cobertura total ≥ 80%
- [ ] `docker compose up`: `GET :8083/ping` = 200; com `aws-cli` (ou script `local/send-command.sh`) publicar `GenerateBudget` → orçamento criado e `BudgetGenerated` na fila de teste; `RequestPayment` em `auto-reject` → `PaymentFailed`
- [ ] Mesma mensagem 2× ⇒ 1 efeito (teste de integração)
- [ ] `testdata/contracts/*.json` presentes e usados
- [ ] 3 overlays kustomize ok; PR aberto para `develop` com jobs verdes
