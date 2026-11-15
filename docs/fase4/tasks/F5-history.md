# Ficha F5 — History & Reporting (`auto-repair-shop-history-service`, :8085, DynamoDB) — serviço novo

Pré-requisitos: onda 2 (PR do Billing disponível); ler `_baseline.md` (B9), `contracts/{resources,messaging,ddl}.md` — **o modelo de itens DynamoDB está na seção `history` de `ddl.md`**. `<svc>=history-service`, secret `auto-repair-shop-<env>/history` (só `JWT_SECRET`). **Sem Postgres**: não copiar `db.go`, `uow`, `internal/pkg/db`, `migrations/`; sem serviço `db`/`migrate` no compose; o job `integration-tests` usa só o MiniStack.

## T1 — Baseline e mensageria (sem outbox)
- B1–B3, B6–B7 adaptados (sem Postgres; `docker.yml` sem "Sync migrations"). `internal/pkg/auth/roles.go` como no F1/T2.
- Do `messaging` do Billing copiar **só** `envelope.go`, `sqs.go`, `config.go`: o History **não publica** (sem outbox/relay/SNS). O consumer é novo (sem Postgres), ver T3.
- Dependências: `aws-sdk-go-v2/service/dynamodb` (+ `attributevalue`).

## T2 — Camada DynamoDB (`internal/store`)
- Interface `Store` (`WriteStatusEvent`, `WriteWorkEvent`, `ListByServiceOrder`, `ListReport`) atrás de uma interface fina do cliente DynamoDB (`TransactWriteItems`, `Query`, `GetItem`) para testar com fake e com MiniStack.
- Implementação exatamente como `ddl.md` (seção `history`): chaves `PK`/`SK`, itens `S#…`, `W#…`, `R#…`, `REPORT#WORK/W#…`, `MSG#…/DEDUPE`. **Toda escrita é um `TransactWriteItems`** com o dedupe (`attribute_not_exists(PK)`); `TransactionCanceledException` por condição do dedupe = mensagem duplicada (descartar e apagar da fila); outro motivo = erro (mensagem volta à fila).
- ⚠️ **Verificar primeiro, antes de construir o resto**: o MiniStack suporta `TransactWriteItems` (teste mínimo no CI). Se **não** suportar, **PARAR e reportar** — o desenho alternativo (itens idempotentes por `message_id` + marcador condicional) é decisão do dono dos contratos.
- Tabela local: `local/init-aws.sh` cria `auto-repair-shop-local-history` (`aws dynamodb create-table --billing-mode PAY_PER_REQUEST`, chaves `PK`/`SK` do tipo S); env `DYNAMODB_TABLE`.

## T3 — Consumer (`internal/consumer`)
Loop SQS (long poll; `HISTORY_QUEUE_URL`) com decode do envelope; roteia por `type`:
- `ServiceOrderStatusChanged` → item `S#`: `previous_status` (omitir se `null`), `new_status`, `reason` (se houver), `created_at = occurred_at`.
- `ExecutionWorkStatusChanged` → item `W#`; se `new == IN_PROGRESS` grava também o marcador `R#<work_id>` (`started_at = occurred_at`); se `previous == IN_PROGRESS && new == COMPLETED`: ler o marcador (`ConsistentRead`) e, existindo, somar `duração = occurred_at − started_at` ao agregado `REPORT#WORK/W#<work_id>` (`ADD count :1, total_seconds :d`, `SET work_name`) **na mesma transação**; marcador ausente → grava só o evento e loga.
- Qualquer outro tipo (`BudgetGenerated`, `PaymentApproved`, …) → descartar com log e apagar da fila. Payload inválido → erro (DLQ após 5 tentativas).
Handler ligado como goroutine no bootstrap com shutdown por `context`; desabilitado se `HISTORY_QUEUE_URL` vazio.

## T4 — API REST (formatos idênticos ao monólito — `$SRC/internal/service_order_history/adapters/so_history_read.go`)
- `GET /v1/service-order/:id/history` (Attendant, Mechanic, Admin): `Query PK=SO#<id>` com `begins_with(SK,"S#")` (ordem ascendente) e `begins_with(SK,"W#")`. Resposta: array de `{"previousStatus":"","newStatus":"…","createdAt":"…","workTransitions":[{"workId":"…","status":[{"previousStatus":"…","newStatus":"…","createdAt":"…"}]}]}`. `previousStatus` ausente → `""`. **`workTransitions` anexado somente à entrada cujo `newStatus == "IN_PROGRESS"`** (a primeira), agrupado por `work_id` na ordem da primeira aparição, estados em ordem cronológica; omitido se não há. OS sem itens → `200 []`.
- `GET /v1/reports/average-execution-time?work_id=a&work_id=b` (Attendant, Admin): `Query PK=REPORT#WORK`; filtrar pelos `work_id` informados (nenhum = todos); resposta array de `{"workId","workName","averageExecutionTimeHours"}` ordenado por `workName` asc, `averageExecutionTimeHours = total_seconds / count / 3600` (float).
- Erros no formato padrão (`{"code","errors"}`); `:id` vazio → `422`.

## T5 — Config e k8s
Variáveis: `PORT`, `JWT_SECRET`, `AWS_REGION`, `AWS_ENDPOINT_URL`, `HISTORY_QUEUE_URL`, `DYNAMODB_TABLE`. `docker.yml`: `Resolve messaging endpoints` (fila `history-queue`) e `DYNAMODB_TABLE=auto-repair-shop-${ENV_NAME}-history` no ConfigMap. Overlays sem `migrate-job`/`rbac`/initContainer (B6).
`local/init-aws.sh`: tópicos `so-events`, `billing-events`, `execution-events`; fila `history-queue` (+DLQ, policy) assinada nos **três** tópicos (espelha o Terraform; o consumer descarta o que não usa); a tabela DynamoDB.

## T6 — Testes (cobertura total ≥ 80%; ≥ 90% em `internal/store` e `internal/consumer`)
- Unitários com fake do cliente DynamoDB: montagem de chaves/itens, agrupamento e ordenação do histórico, anexo de `workTransitions` só no `IN_PROGRESS`, média do relatório (incluindo filtro por `work_id`), marcador ausente, duplicada, tipo desconhecido.
- Contrato: `testdata/contracts/ServiceOrderStatusChanged.json` e `ExecutionWorkStatusChanged.json` (exemplos de `messaging.md`, com e sem `previous`).
- Integração (`-tags=integration`, MiniStack): publicar a sequência `NEW → … → IN_PROGRESS`, work `AWAITING_START → IN_PROGRESS → COMPLETED` → `GET history` com o formato esperado e `GET report` com a média correta; a mesma mensagem 2× não duplica itens nem soma o agregado duas vezes; **`TransactWriteItems` funcionando no MiniStack**.

## T7 — Docs e entrega
Swagger (2 endpoints), README (modelo de itens, tipos tratados, limitação: mensagens fora de ordem), ADR `0002-history-read-model-dynamodb.md`, limpar `.ai/rules/*`. Entrega B8.

## Aceite
- [ ] `go build`, `go vet` (+`-tags=integration`), `go test ./... --race`; cobertura ≥ 80%
- [ ] CI: `pre-commit`, `build`, `unit-tests`, `integration-tests`, `coverage` verdes (Sonar conforme token)
- [ ] Formato de `GET /history` idêntico ao do monólito (teste comparando com JSON de referência extraído do monólito)
- [ ] Idempotência comprovada (mesma mensagem 2×); relatório correto
- [ ] 3 overlays kustomize ok; PR aberto para `develop`
