# Contrato — Mensageria (SNS + SQS)

Nomes de tópicos/filas: ver `resources.md`. Filas **Standard**, DLQ por fila (`maxReceiveCount=5`).

## Assinaturas

| Fila | Tópico | Filter policy |
|---|---|---|
| `billing-queue` | `saga-commands` | `{"target":["billing"]}` |
| `execution-queue` | `saga-commands` | `{"target":["execution"]}` |
| `catalog-queue` | `saga-commands` | `{"target":["catalog"]}` |
| `so-saga-replies` | `billing-events`, `execution-events` | — (SO ignora tipos que não usa) |
| `history-queue` | `so-events`, `billing-events`, `execution-events` | — |

Publicar **com raw message delivery ligado** nas assinaturas SNS→SQS (o corpo da fila é o envelope JSON, sem o invólucro do SNS). Comandos carregam o atributo de mensagem SNS `target`; todas as mensagens carregam o atributo `type`.

## Envelope (todo corpo de mensagem)

```json
{
  "message_id": "b7a1c6f0-2f64-4f6e-9a49-0d5a0b1e6c11",
  "saga_id": "0f3c9a2e-6a1d-4a53-8a6e-5d3b1f2c7e90",
  "type": "GenerateBudget",
  "occurred_at": "2026-11-01T12:00:00Z",
  "service_order_id": "01JD8Z4R3Q8M2V7K5N9XWABCDE",
  "request_id": "2c5e1f0a-9d77-4b2c-a0c1-3f8e6b7d1a22",
  "payload": {}
}
```

Regras: `message_id` é UUID e é a chave de idempotência (consumidor grava em `processed_message` antes/junto do efeito); `service_order_id` é a ULID da OS; `occurred_at` em UTC (RFC 3339); `payload` varia por tipo (abaixo). Consumidor ignora tipos desconhecidos (descarta com log) e **nunca** loga CPF/CNPJ.

## Comandos (SO → serviço, tópico `saga-commands`)

| `type` | `target` | `payload` |
|---|---|---|
| `GenerateBudget` | billing | `{"customer_user_id":"3b123072-5ad5-4df8-8d3b-3c2770f11b2a","works":[{"work_id":"…","name":"Troca de óleo","unit_price":"150.00"}],"supplies":[{"supply_id":"…","name":"Filtro","quantity":2,"unit_price":"35.50"}]}` |
| `CancelBudget` | billing | `{}` |
| `RequestPayment` | billing | `{"amount":"221.00"}` |
| `RefundPayment` | billing | `{}` (pagamento `PENDING` → `CANCELLED`; `APPROVED` → `REFUNDED`) |
| `EnqueueExecution` | execution | `{"works":[{"work_id":"…","name":"Troca de óleo"}]}` |
| `CancelExecution` | execution | `{}` |
| `ReleaseStock` | catalog | `{}` (libera todas as reservas da OS) |

Valores monetários são **strings decimais** (`"221.00"`), nunca número JSON.

## Eventos

| `type` | Tópico | Publicador | `payload` |
|---|---|---|---|
| `BudgetGenerated` | `billing-events` | Billing | `{"budget_id":"…","total_amount":"221.00"}` |
| `BudgetCancelled` | `billing-events` | Billing | `{"budget_id":"…"}` |
| `PaymentApproved` | `billing-events` | Billing | `{"payment_id":"…","amount":"221.00"}` |
| `PaymentFailed` | `billing-events` | Billing | `{"payment_id":"…","amount":"221.00","reason":"PAYMENT_REJECTED"}` |
| `PaymentRefunded` | `billing-events` | Billing | `{"payment_id":"…","amount":"221.00"}` |
| `ExecutionQueued` | `execution-events` | Execution | `{"task_id":"…"}` |
| `ExecutionWorkStatusChanged` | `execution-events` | Execution | `{"work_id":"…","previous":"AWAITING_START","new":"IN_PROGRESS"}` |
| `ExecutionCompleted` | `execution-events` | Execution | `{"task_id":"…"}` |
| `ExecutionFailed` | `execution-events` | Execution | `{"task_id":"…","reason":"ALL_WORKS_CANCELLED"}` |
| `ServiceOrderStatusChanged` | `so-events` | Service Order | `{"previous":"IN_DIAGNOSIS","new":"AWAITING_APPROVAL","reason":null,"total_amount":"221.00"}` |

`reason` de `ServiceOrderStatusChanged` quando `new=REJECTED`: `BUDGET_REJECTED` ou `PAYMENT_REJECTED`; quando `new=CANCELLED`: `USER_CANCELLED`, `EXECUTION_FAILED` ou `SAGA_TIMEOUT`; nos demais casos `null`. `ExecutionFailed.reason`: `INVALID_REQUEST` ou `ALL_WORKS_CANCELLED`.

## Garantias exigidas

- **Publicador:** *outbox* — estado e mensagem gravados na mesma transação; um relay lê `outbox` e publica no SNS; marca como enviada só após sucesso.
- **Consumidor:** dedupe por `message_id`; efeito idempotente; falha → mensagem volta à fila → DLQ após 5 tentativas; mensagem fora de ordem (ex.: `PaymentApproved` sem saga em `PAYMENT`) é ignorada com log (compensações podem ter ocorrido).
- **Ordem** não é garantida (fila Standard): a máquina de estados da saga é quem valida a sequência.
- **Fixtures:** cada mensagem acima deve virar um arquivo `testdata/contracts/<type>.json` nos dois lados (publicador e consumidor), com o exemplo desta tabela.
