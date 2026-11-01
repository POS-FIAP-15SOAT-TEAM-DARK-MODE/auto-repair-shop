# SPEC — Fase 4: Microsserviços + Saga Pattern

> Fonte: `SOAT - Fase 4 - Tech challenge.pdf`. Estado atual dos repos: `PROJECTS_INDEX.md`.
> Execução em **ondas** (§5): cada microsserviço em **repositório próprio**; dentro de uma onda as fases podem rodar em paralelo, e uma onda só começa quando a anterior termina.

## 1. Decisões fechadas

| Tema | Decisão |
|---|---|
| Serviços | **6 microsserviços**, cada um com **repo próprio** (mín. exigido: 3) |
| Mensageria | **SNS + SQS** (AWS); local via **MiniStack** no Docker Compose |
| Dados | **RDS PostgreSQL**: 1 instância, 1 database por serviço. **DynamoDB**: History. (Exigência: ≥ 1 SQL e ≥ 1 NoSQL no projeto todo — atendida) |
| Saga | **Orquestrada**, orquestrador no Service Order |
| Comunicação | REST interno só onde a resposta é imediata; o resto é evento/comando SNS→SQS |
| Lambda | **`lambda-auth` mantida sem mudança de código** (§5, F9) |
| Gateway | O repo **`auto-repair-shop` (monólito) vira o gateway** (reverse proxy Go). API Gateway AWS continua à frente dele (§2, F10) |
| Billing | Sem Mercado Pago: interface `PaymentProvider` + adaptador interno (manual/fake) |
| Pagamento recusado | Reutiliza o status terminal existente **`REJECTED`** (sem status novo); o motivo (`BUDGET_REJECTED` ou `PAYMENT_REJECTED`) vai no evento `ServiceOrderStatusChanged.reason` e no History |
| Código repetido | **Duplicar sem cerimônia** (nada de módulo compartilhado, template engine ou scaffolding) |
| Local | Cada serviço sobe **sozinho** (`make up`); o ecossistema sobe junto pelo repo `auto-repair-shop` (`make up-all`) |
| Infra | Tudo para o Lab em `infra-db` e `infra-k8s`; DynamoDB/SNS/SQS disponíveis no Lab (sem spike) |
| Qualidade | Todo serviço nasce com a mesma CI/testes/Sonar do monólito + gate de cobertura ≥ 80% |

## 2. Arquitetura alvo

```
Cliente ─► AWS API Gateway ─┬─ POST /auth/customer-login ─► lambda-auth ─► DB `customer` (read-only)
                            └─ demais rotas ─► auto-repair-shop (GATEWAY, :8080, único público)
                                                  │ reverse proxy por rota (ClusterIP interno)
      ┌────────────┬─────────────┬───────────┬──────────────┬───────────┬────────────────┐
   customer     catalog       billing     execution       history    service-order
   (PG)         (PG)          (PG)        (PG)            (DynamoDB) (PG + saga)
      ▲            ▲             ▲  ▲          ▲  ▲           ▲            │ │
      └── REST ────┴── REST ─────┼──┼──────────┼──┼───────────┼── REST ────┘ │
          (SO → Customer, SO → Catalog)        └ SNS/SQS ─────┴──────────────┘
```

| Serviço | Repo | Porta | Database | Origem no monólito |
|---|---|---|---|---|
| **Gateway** | `auto-repair-shop` (existente) | 8080 | — | `app/`, `middleware/`, `pkg/` |
| **Customer & Identity** | `auto-repair-shop-customer-service` | 8081 | PG `customer` | `auth`, `customer`, `vehicle` |
| **Catalog** | `auto-repair-shop-catalog-service` | 8082 | PG `catalog` | `work`, `supply` |
| **Billing** | `auto-repair-shop-billing-service` | 8083 | PG `billing` | novo (orçamento hoje embutido na OS) |
| **Execution** | `auto-repair-shop-execution-service` | 8084 | PG `execution` | novo (progresso de works hoje embutido na OS) |
| **History & Reporting** | `auto-repair-shop-history-service` | 8085 | DynamoDB `history` | `service_order_history` |
| **Service Order** | `auto-repair-shop-service-order-service` | 8086 | PG `service_order` | `service_order` |

Por que essa divisão: `customer` já depende de `auth` (cria o `user` junto) e a lambda lê `customer`+`user_role`+`role` → ficam juntos, e a lambda não muda. `work` e `supply` são catálogos folha sem dependências. `service_order` depende de todos e é extraído por último. Billing e Execution são os domínios do PDF que hoje vivem dentro da OS. History é um *read model* alimentado por eventos (é onde entra o NoSQL).

## 3. Contratos (fonte única para todos os repos)

### 3.1 Status da OS
`NEW → RECEIVED → IN_DIAGNOSIS → AWAITING_APPROVAL → IN_PROGRESS → COMPLETED → DELIVERED`; terminais: `DELIVERED`, `REJECTED` (cliente rejeitou o orçamento **ou** o pagamento foi recusado; distinguido por `reason`), `CANCELLED`. Cancelável: qualquer status não terminal e anterior a `COMPLETED`.

Regra de negócio: **`AWAITING_APPROVAL` só existe enquanto o cliente não decidiu.** No `accept` a OS vai **imediatamente** a `IN_PROGRESS`; o pagamento e a fila de execução são processados depois, sem mudar o status. Transição nova (só pela saga): `IN_PROGRESS → REJECTED` quando o pagamento é recusado.

### 3.2 Saga `service-order-approval` (orquestrada)
Gatilho: `PUT /v1/service-order/{id}/send` (OS em `IN_DIAGNOSIS`). O OS mantém `saga_instance` (`saga_id`, `service_order_id`, `step`, `status RUNNING|COMPENSATING|COMPLETED|COMPENSATED|FAILED`, `deadline`); um job varre prazos vencidos.

| # | Comando (SO →) | Executor | Sucesso (evento) | Efeito no status | Compensação |
|---|---|---|---|---|---|
| 1 | `GenerateBudget` | Billing | `BudgetGenerated` | `AWAITING_APPROVAL` | `CancelBudget` |
| 2 | — cliente `accept`/`reject` na API do SO (**síncrono**, 204) | SO | — | `accept` → **`IN_PROGRESS` na hora**; `reject` → `REJECTED` (fim, `reason=BUDGET_REJECTED`) | no `reject`: `CancelBudget` + `ReleaseStock` |
| 3 | `RequestPayment` (gravado no outbox na mesma transação do `accept`) | Billing | `PaymentApproved` | — (segue `IN_PROGRESS`) | `RefundPayment` (se `PENDING`, apenas cancela) |
| 4 | `EnqueueExecution` | Execution | `ExecutionQueued` | — (segue `IN_PROGRESS`) | `CancelExecution` |
| 5 | — mecânico executa | Execution | `ExecutionCompleted` | `COMPLETED` | — (sem retorno) |

Notas: `send` responde 204 e a OS permanece `IN_DIAGNOSIS` até `BudgetGenerated` (aí vira `AWAITING_APPROVAL`). `accept`/`reject` retornam **409** se o orçamento ainda não foi gerado ou se já houve decisão. Enquanto o pagamento está pendente a OS está `IN_PROGRESS` mas **não existe tarefa no Execution**, então o mecânico não avança nenhum work. `deliver` é local no SO.

| Falha | Compensações (ordem inversa) | Status final |
|---|---|---|
| `GenerateBudget` falha/timeout | — | segue `IN_DIAGNOSIS` |
| `PaymentFailed` (OS já `IN_PROGRESS`) | `CancelBudget` + `ReleaseStock` | `REJECTED` (`reason=PAYMENT_REJECTED`; transição `IN_PROGRESS → REJECTED` só pela saga) |
| `EnqueueExecution` falha / `ExecutionFailed` | `RefundPayment` → `CancelBudget` → `ReleaseStock` | `CANCELLED` |
| Timeout de um passo (`SAGA_STEP_TIMEOUT`; o passo de pagamento usa `SAGA_PAYMENT_TIMEOUT`, bem maior, por causa do modo `manual`) | a partir do último passo concluído | `CANCELLED` |
| `cancel` do usuário (com ou sem saga em curso) | desfaz só o que existir: `CancelExecution` → `RefundPayment` → `CancelBudget` → `ReleaseStock` | `CANCELLED` |

Compensações são idempotentes e semânticas (cancelar/estornar), nunca "delete".

> **Detalhe dos contratos** (payloads JSON de exemplo, códigos de erro `401/404/409/422`, formato de erro, nomes com prefixo de ambiente, variáveis comuns como `LOG_LEVEL`/`AWS_REGION`, usuários de banco `<db>_app`, variáveis do gateway): `docs/fase4/contracts/{resources,messaging,internal-rest}.md` no repo `auto-repair-shop`. Este §3 é o resumo; **em divergência, `contracts/` prevalece**.

### 3.3 REST interno (rotas `/internal/*`: nunca roteadas pelo gateway; header `X-Internal-Token` = env `INTERNAL_TOKEN`)

| Chamador → Serviço | Endpoint | Resposta |
|---|---|---|
| SO → Customer | `GET /internal/customers/{id}` | `{id, user_id, type, name, document}` |
| SO → Customer | `GET /internal/vehicles/{id}` | `{id, customer_id, license_plate, brand, model, year}` |
| SO → Catalog | `GET /internal/works/{id}` | `{id, name, unit_price, active}` |
| SO → Catalog | `POST /internal/reservations` `{service_order_id, supply_id, quantity}` | `201 {reservation_id, supply_id, name, unit_price}` / `409` sem estoque |
| SO → Catalog | `DELETE /internal/reservations/{id}` | `204` (devolve estoque) |

A OS guarda **snapshots**: `customer_id`, `customer_user_id` (escopo de `accept`/`reject` pelo JWT), placa, nome/preço de cada work e supply. Nenhuma outra consulta cruzada.

### 3.4 Mensageria

Tópicos SNS: `saga-commands` (atributo `target`), `so-events`, `billing-events`, `execution-events`.
Filas SQS (cada uma com DLQ `<fila>-dlq`, redrive após 5 recebimentos), todas **Standard**:

| Fila | Dono | Assina |
|---|---|---|
| `billing-queue` | Billing | `saga-commands` com `target=billing` |
| `execution-queue` | Execution | `saga-commands` com `target=execution` |
| `catalog-queue` | Catalog | `saga-commands` com `target=catalog` |
| `so-saga-replies` | Service Order | `billing-events`, `execution-events` |
| `history-queue` | History | `so-events`, `billing-events`, `execution-events` |

| Mensagem | Tipo | Publicador → Consumidor | Payload (além do envelope) |
|---|---|---|---|
| `GenerateBudget` | cmd | SO → Billing | `works[{work_id,name,unit_price}]`, `supplies[{supply_id,name,quantity,unit_price}]` |
| `CancelBudget`, `RefundPayment` | cmd | SO → Billing | — |
| `RequestPayment` | cmd | SO → Billing | `amount` |
| `EnqueueExecution` | cmd | SO → Execution | `works[{work_id,name}]` |
| `CancelExecution` | cmd | SO → Execution | — |
| `ReleaseStock` | cmd | SO → Catalog | — (libera todas as reservas da OS) |
| `BudgetGenerated` | evt | Billing → SO, History | `budget_id`, `total_amount` |
| `BudgetCancelled` | evt | Billing → SO, History | `budget_id` |
| `PaymentApproved`, `PaymentFailed`, `PaymentRefunded` | evt | Billing → SO, History | `payment_id`, `amount`, `reason?` (só `PaymentFailed`) |
| `ExecutionQueued`, `ExecutionCompleted` | evt | Execution → SO, History | `task_id` |
| `ExecutionFailed` | evt | Execution → SO, History | `task_id`, `reason` (`INVALID_REQUEST` \| `ALL_WORKS_CANCELLED`) |
| `ExecutionWorkStatusChanged` | evt | Execution → History | `work_id`, `previous`, `new` |
| `ServiceOrderStatusChanged` | evt | SO → History | `previous`, `new`, `reason?` (`BUDGET_REJECTED`, `PAYMENT_REJECTED`, …), `total_amount?` |

**Envelope JSON:** `message_id` (UUID), `saga_id`, `type`, `occurred_at`, `service_order_id`, `request_id`, `payload`.
**Regras em todo serviço que publica:** *outbox* (estado + mensagem na mesma transação; relay publica no SNS). **Em todo que consome:** tabela `processed_message(message_id)` (dedupe) + retry/DLQ + guarda de ordem (mensagem fora de ordem é ignorada ou reenfileirada).
**AWS:** pods usam a role do node (`LabRole`, sem IRSA); local usa `AWS_ENDPOINT_URL=http://ministack:4566` com credenciais dummy — mesmo código nos dois.

### 3.5 Segurança e dados
Todos os serviços validam o JWT (HS256, mesmo `JWT_SECRET`) com o middleware atual; só Customer&Identity e a lambda **emitem** tokens. Referência entre serviços só por ID (sem FK cruzada). Dinheiro `NUMERIC(10,2)`. Nunca logar CPF/CNPJ/JWT.

## 4. Roteiro padrão de extração (vale para F1–F6)

**S1. Criar o repo.** repo **público** (a org está no plano Free e os repos atuais são públicos: proteção de branch e SonarCloud gratuito exigem isso) criado pelo script da F0, com branches `main` e `develop` (default `develop`); antes de começar, **taguear o monólito** (`git tag monolith-v1` em `auto-repair-shop`) para a baseline continuar acessível depois da limpeza.

**S2. Copiar a baseline do monólito** e ajustar nomes: `go.mod` (module path novo), `Dockerfile`, `Makefile`, `.env.example`, `.pre-commit-config.yaml`, `sonar-project.properties` (`projectKey` próprio), `.github/{workflows/ci.yml,workflows/docker.yml,dependabot.yml,pull_request_template.md,commit_message_template.txt}`, `.ai/` + `CLAUDE.md` + `AGENTS.md` + `.cursor/` + `.windsurf/` (regras ajustadas ao domínio do serviço), `docs/adr/` (ADR de branching), `k8s/manifests/` (somente `base` + overlays `lab|stg|prd`; Service **ClusterIP**; sem overlay `local`/Kind/toolbox — o compose cobre o local).

**S3. Mover o código do domínio** (`cp -r` dos pacotes de origem) e o mínimo de `internal/pkg` (`env`, `logger`, `metrics`, `web`, `json`, `id`, `uow`, `db`, `auth`) + `internal/app` (server, bootstrap, container, routing, middleware auth/logger/metrics/recovery). Trocar dependências de módulos extraídos por REST/eventos (§3). Remover o que não pertence ao serviço.

**S4. Migrations próprias** (`000001_init_schema` só com as tabelas do serviço, sem FK cruzada; `processed_message`; `outbox` quando publica).

**S5. Mensageria** (quando aplicável): `internal/pkg/messaging` com envelope, publisher SNS, relay de outbox, consumer SQS com dedupe. Nasce em **Billing** (F3); Catalog (F1) leva só o consumer simples; os demais **copiam de Billing**.

**S6. Compose independente** (`docker-compose.yml` com `profiles`): `app` + Postgres próprio (quando relacional) + `migrate` + `ministack` + `aws-init` (`local/init-aws.sh`, idempotente, cria só os tópicos/filas/tabelas do serviço). Dependências REST via `*_URL`; sem elas, `make up` sobe um **stub HTTP** (WireMock/JSON estático) para o serviço rodar isolado. Alvos: `make up | down | test | test-integration | coverage`.

**S7. CI/CD:** ligar `SONAR_TOKEN`, variáveis de ambiente do GitHub (`EKS_CLUSTER_NAME`, `*_DB_HOST`, credenciais Lab), **acrescentar os status checks obrigatórios** à proteção que a F0 já aplicou em `main` e `develop` (`pre-commit`, `build`, `unit-tests`, `integration-tests`, `sonarqube`, `coverage`), depois da primeira execução da CI. O deploy depende de F7/F8 (onda 2); nos serviços da onda 2 o `docker.yml` pode ficar validado só por `build` até a infra ficar pronta na mesma onda.

**S8. Testes:** unitários (in-memory + Mockery, como hoje) com **cobertura ≥ 80%** (90% em orçamento/estoque/saga) e gate no CI; integração (`-tags=integration`) com Postgres e MiniStack como *service containers*; teste de idempotência (mesma mensagem 2×); teste de contrato dos eventos/REST com fixtures JSON.

**S9. Docs:** README (propósito, endpoints, eventos publicados/consumidos, como subir, **evidência de cobertura**), `docs/swagger.yaml`, ADRs do serviço.

**S10. Pronto quando (DoD):** CI verde em PR → merge em `develop` → `main`; `make up` sobe o serviço isolado e um exemplo de uso funciona; cobertura ≥ 80% registrada; Sonar sem falha de Quality Gate; atualizar `PROJECTS_INDEX.md`, `CLAUDE.md`/`AGENTS.md` da raiz e o roteamento do gateway (config) com a rota nova.

**CI baseline (jobs):** `pre-commit` (golangci-lint, fmt, vet, build, unit) · `build` (`go mod download/verify`, `go build`) · `unit-tests` (`go test ./... --race`) · `integration-tests` (`-tags=integration`) · `sonarqube` (coverage.out + SonarQube Cloud) · `coverage` (falha se < 80%).

## 5. Plano faseado

### Ordem de execução (definida)

| Onda | Fases | Entrega da onda | Pré-requisito |
|---|---|---|---|
| 1 | **F0** | tag `monolith-v1`, 6 repos criados e protegidos | — |
| 2 | **F1 + F2 + F3 + F7 + F8** | Catalog, Customer & Identity, Billing (com o pacote `messaging`), `infra-db` e `infra-k8s` (mensageria, bancos, ECRs) | onda 1 |
| 3 | **F4 + F5 + F6** (+ F1b/F2b) | Execution, History, Service Order (copiam o `messaging` do Billing); F1b/F2b ampliam os contratos internos de Catalog e Customer | onda 2 |
| 4 | **F9 + F10** | lambda apontando para o database `customer`; gateway, ecossistema local e e2e | onda 3 (F9 também usa F2 e F7, já prontos) |
| 5 | **FA** | Ajustes e validação local: endurecimento de erros 5xx, validação ponta a ponta com Docker, backlog das ondas 2–4 | onda 4 |
| 6 | **F11** | deploy em STG, README, PDF, vídeo (grava o sistema já validado pela FA) | onda 5 |

Regras de convivência dentro de uma onda:
- As fases só se tocam pelos **contratos do §3**; nomes de tópicos, filas, databases e tabela DynamoDB devem ser exatamente os do §3.4 e §4/F7, porque existem em três lugares (Terraform da F8, `init-aws.sh` de cada serviço e do ecossistema).
- Na onda 2, a F3 é a dona do pacote `internal/pkg/messaging`; a F1 leva só o consumer simples.
- Escrever primeiro, nas duas pontas de cada mensagem, os testes de contrato com as fixtures JSON do §3.4 (a primeira integração real só acontece na F10).
- Até a onda 3, cada serviço é validado isoladamente (compose + MiniStack + fixtures de eventos). Como a infra (F7/F8) já existe desde a onda 2, o deploy em STG de cada serviço pode ser validado assim que o serviço fica pronto.

### F0 — Preparação
- Tag anotada `monolith-v1` no `develop` do `auto-repair-shop` (enviada ao remoto); congelar mudanças de domínio no monólito durante a quebra.
- Criar os 6 repos **públicos** (script idempotente, §F0 abaixo): README inicial, branches `main` + `develop` (default `develop`), `has_wiki=false`, environments `infra`, `STG` e `PRD` (como no monólito), proteção **igual à do monólito** — `main`: PR com 1 aprovação; `develop`: PR com 0 aprovação; `enforce_admins`, sem force-push/delete — e clone em `fases-pos/<repo>`. Os **status checks obrigatórios** só entram no S7, depois da primeira execução da CI (os nomes dos jobs só existem então).
- Secrets: o monólito usa `AWS_ACCESS_KEY_ID/SECRET_ACCESS_KEY/SESSION_TOKEN` (repo e environment `STG`) e `SONAR_TOKEN`. Valores não são legíveis por script — **você os informa** (sugestão: criar como *secrets da organização* para valerem em todos os repos); projetos no SonarCloud (`POS-FIAP-15SOAT-TEAM-DARK-MODE_<repo>`) também são criados manualmente na UI. Variáveis (`EKS_CLUSTER_NAME`, `RDS_HOST`) só existem depois da F7/F8.
- Registrar este spec e os contratos no `auto-repair-shop` em `docs/fase4/` (`SPEC.md`, `contracts/messaging.md`, `contracts/internal-rest.md`, `contracts/resources.md`, `tasks/F0.md`); é a referência dos demais repos. Fica no working tree (sem commit) até você decidir como integrá-lo.

### F1 — Catalog (`auto-repair-shop-catalog-service`, :8082, PG `catalog`) — ficha: `docs/fase4/tasks/F1-catalog.md`
- **Origem:** `internal/work`, `internal/supply`.
- **Tabelas:** `work` e `supply` como hoje (`supply.version` = lock otimista); novo `stock_reservation(id, service_order_id, supply_id, quantity, status RESERVED|RELEASED, created_at, UNIQUE(service_order_id, supply_id))`; `processed_message`.
- **Endpoints públicos:** `/v1/works` (GET atendente+mecânico; escrita atendente) · `/v1/supplies` (atendente+mecânico), mesmos contratos de hoje.
- **Internos:** `GET /internal/works/{id}`, `POST /internal/reservations` (decremento atômico `UPDATE … WHERE stock_quantity >= $n`; `409` sem estoque), `DELETE /internal/reservations/{id}`.
- **Mensageria:** consome `ReleaseStock` (marca todas as reservas da OS como `RELEASED` e devolve estoque, idempotente). Não publica.
- **Config:** `PORT`, `POSTGRES_*`, `JWT_SECRET`, `INTERNAL_TOKEN`, `AWS_ENDPOINT_URL`, `CATALOG_QUEUE_URL`.
- **Dependências de saída:** nenhuma.

### F2 — Customer & Identity (`auto-repair-shop-customer-service`, :8081, PG `customer`) — ficha: `docs/fase4/tasks/F2-customer.md`
- **Origem:** `internal/auth`, `internal/customer`, `internal/vehicle`, seed do admin, `pkg/auth` (JWT + roles).
- **Tabelas:** `user`, `role` (seed ADMIN/MECHANIC/ATTENDANT/CUSTOMER), `user_role`, `customer`, `vehicle` — **idênticas às atuais** (a lambda continua válida); migração do admin (`000002`).
- **Endpoints públicos:** `/v1/auth/register|login`, `PATCH /v1/users/:id/role`, `/v1/customers*`, `/v1/vehicles*`.
- **Internos:** `GET /internal/customers/{id}`, `GET /internal/vehicles/{id}`.
- **Regras preservadas:** criar customer cria user `CUSTOMER` (senha inicial = documento, bcrypt ≥ 12); excluir customer exclui o user.
- **Extra:** no init do Postgres local/Lab, usuário **`lambda_auth_ro`** com `SELECT` em `customer`, `user_role`, `role` (usado pela lambda).
- **Mensageria:** nenhuma. **Dependências de saída:** nenhuma.
- **Config:** `PORT`, `POSTGRES_*`, `JWT_SECRET`, `JWT_EXPIRES_IN`, `BCRYPT_COST`, `INTERNAL_TOKEN`.

### F3 — Billing (`auto-repair-shop-billing-service`, :8083, PG `billing`) — serviço novo — ficha: `docs/fase4/tasks/F3-billing.md`
- **Tabelas:** `budget(id, service_order_id, status GENERATED|CANCELLED, total_amount, created_at)`, `budget_item(id, budget_id, kind WORK|SUPPLY, ref_id, name, quantity, unit_price)`, `payment(id, budget_id, service_order_id, saga_id, amount, status PENDING|APPROVED|REJECTED|REFUNDED|CANCELLED, provider, created_at, updated_at)`, `outbox`, `processed_message`.
- **Regras:** `total = Σ work.unit_price + Σ supply.unit_price × quantity` (mesma fórmula do monólito), calculado dos snapshots recebidos em `GenerateBudget`.
- **`PaymentProvider { Charge, Refund }`**, adaptador `internal` guiado por `PAYMENT_MODE`: `manual` (fica `PENDING` até o atendente agir), `auto-approve`, `auto-reject` (para demonstrar a compensação). O adaptador Mercado Pago será outra implementação futura, sem mexer na saga.
- **Endpoints:** `GET /v1/budgets/{serviceOrderId}`, `GET /v1/payments/{serviceOrderId}`, `POST /v1/payments/{id}/confirm`, `POST /v1/payments/{id}/reject` (atendente).
- **Mensageria:** consome `GenerateBudget`, `CancelBudget`, `RequestPayment`, `RefundPayment`; publica `BudgetGenerated`, `BudgetCancelled`, `PaymentApproved`, `PaymentFailed`, `PaymentRefunded` em `billing-events` via outbox.
- **Entrega:** aqui nasce `internal/pkg/messaging` (S5) — fica como referência para os próximos.
- **Config:** `PORT`, `POSTGRES_*`, `JWT_SECRET`, `AWS_ENDPOINT_URL`, `BILLING_QUEUE_URL`, `BILLING_EVENTS_TOPIC_ARN`, `PAYMENT_MODE`.

### F4 — Execution (`auto-repair-shop-execution-service`, :8084, PG `execution`) — serviço novo — ficha: `docs/fase4/tasks/F4-execution.md`
- **Tabelas:** `execution_task` (guarda o `saga_id` do `EnqueueExecution`), `execution_work`, `execution_work_history`, `outbox`, `processed_message` (DDL em `contracts/ddl.md`).
- **Regras:** ciclo do work igual ao monólito (`AWAITING_START → IN_PROGRESS → COMPLETED`, `CANCELLED` terminal); `finish` exige todos os works `COMPLETED`/`CANCELLED` (regra nova): todos `CANCELLED` → `ExecutionFailed{ALL_WORKS_CANCELLED}`, senão `ExecutionCompleted`.
- **Endpoints (mecânico):** `GET /v1/execution/queue`, `PUT /v1/service-order/{id}/work/{workId}/next|cancel`, `PUT /v1/service-order/{id}/finish` (paths preservados; o gateway os roteia para cá).
- **Mensageria:** consome `EnqueueExecution`, `CancelExecution`; publica `ExecutionQueued`, `ExecutionWorkStatusChanged` (com `work_name`; `previous` nulo na criação), `ExecutionCompleted`, `ExecutionFailed` em `execution-events`, **sempre com o `saga_id` da tarefa**, também nos eventos disparados por REST.
- **Observação:** o diagnóstico continua sendo transição de status da OS (Service Order); Execution cuida da fila e do reparo.
- **Config:** `PORT`, `POSTGRES_*`, `JWT_SECRET`, `AWS_REGION`, `AWS_ENDPOINT_URL`, `EXECUTION_QUEUE_URL`, `EXECUTION_EVENTS_TOPIC_ARN`.

### F5 — History & Reporting (`auto-repair-shop-history-service`, :8085, DynamoDB) — serviço novo — ficha: `docs/fase4/tasks/F5-history.md`
- **Modelo de itens** (`contracts/ddl.md`, seção `history`): eventos de status da OS (`S#…`), transições de work (`W#…`), marcador de início do work (`R#…`), **agregado do relatório por work** (`REPORT#WORK/W#…`: `count`, `total_seconds`) e dedupe (`MSG#…`). Toda escrita é um `TransactWriteItems` (dedupe + item + agregado).
- **Endpoints (formatos idênticos ao monólito):** `GET /v1/service-order/{id}/history` (transições da OS; as transições de work aninhadas só na entrada `IN_PROGRESS`, ADR 0001 do monólito) e `GET /v1/reports/average-execution-time?work_id=…` (**por work**: `[{workId, workName, averageExecutionTimeHours}]`, média entre `IN_PROGRESS` e `COMPLETED` do work).
- **Mensageria:** consome `history-queue` mas **trata só** `ServiceOrderStatusChanged` e `ExecutionWorkStatusChanged`; demais tipos são descartados com log. Não publica. **Sem Postgres.**
- **Config:** `PORT`, `JWT_SECRET`, `AWS_REGION`, `AWS_ENDPOINT_URL`, `HISTORY_QUEUE_URL`, `DYNAMODB_TABLE`.

### F6 — Service Order + orquestrador (`auto-repair-shop-service-order-service`, :8086, PG `service_order`) — ficha: `docs/fase4/tasks/F6-service-order.md`
- **Origem:** `internal/service_order` (sem as partes de histórico, works/execução e relatório, que migraram para Execution/History).
- **Tabelas** (`contracts/ddl.md`): `service_order` com **snapshot de cliente e veículo** (as respostas de listagem/detalhe e a notificação os usam), `service_order_work`, `service_order_supply` (com `reservation_id`), `saga_instance` (flags do que já foi comandado), `outbox`, `processed_message`.
- **Endpoints:** os mesmos de hoje para criação, works/supplies, `received`, `start-diagnosis`, `send` (inicia a saga), `accept` (204, **`IN_PROGRESS` imediato** + `RequestPayment`), `reject`, `cancel`, `deliver` e as leituras (`GET …`, `/{id}`, `/works`, `/supplies`, `/status`). Mais `GET /internal/service-orders/exists` (para o Customer restaurar a checagem de exclusão na onda 4).
- **Saga:** tabela de gatilhos/guardas/efeitos no §3.2 e na ficha (fire-and-forget nas compensações; job de prazos: `BUDGET` falha sem mudar o status, `PAYMENT`/`EXECUTION_QUEUE` compensam e cancelam).
- **REST de saída:** Customer (cliente/veículo na criação), Catalog (work; reserva/liberação de supply); falha de rede → `502`.
- **Mensageria:** publica comandos em `saga-commands` e `ServiceOrderStatusChanged` em `so-events` (outbox); consome `so-saga-replies`.
- **Config:** `PORT`, `POSTGRES_*`, `JWT_SECRET`, `INTERNAL_TOKEN`, `CUSTOMER_URL`, `CATALOG_URL`, `AWS_REGION`, `AWS_ENDPOINT_URL`, `SO_SAGA_REPLIES_QUEUE_URL`, `SAGA_COMMANDS_TOPIC_ARN`, `SO_EVENTS_TOPIC_ARN`, `SAGA_STEP_TIMEOUT`, `SAGA_PAYMENT_TIMEOUT`.
- **Compose isolado:** stubs de Customer e Catalog; Billing/Execution simulados pelos testes (que leem os comandos em filas de teste e publicam as respostas) até a integração real na onda 4.
- **Pendência herdada do F2:** `DELETE /v1/customers/:id` deixou de checar OS; o Customer passa a chamar o endpoint interno acima (ajuste na onda 4).

### F7 — `auto-repair-shop-infra-db` (tudo de dados) — ficha: `docs/fase4/tasks/F7-infra-db.md`
- **Mantém** a instância RDS única, o master `postgres` e o secret legado `${name}/app`; o plano **só adiciona** recursos.
- **Secrets por serviço** no padrão atual `auto-repair-shop-<env>/<nome>` (chaves do JSON = nomes das variáveis de ambiente): `customer`, `catalog`, `billing`, `execution`, `service-order` (`POSTGRES_USER/PASSWORD/DB`, `JWT_SECRET`, `INTERNAL_TOKEN` quando aplicável), `history` (`JWT_SECRET`) e `lambda-auth` (`POSTGRES_PASSWORD` do `lambda_auth_ro` + `JWT_SECRET` — **mesmas chaves que a lambda lê hoje**). O `JWT_SECRET` reaproveita o `random_password.jwt` existente (tokens atuais continuam válidos).
- **Databases e usuários** (`customer`, `catalog`, `billing`, `execution`, `service_order`; usuários `<db>_app`; `lambda_auth_ro` somente leitura no `customer` via `ALTER DEFAULT PRIVILEGES`) criados por **Job de bootstrap no cluster** (`postgres:15-alpine` + `bootstrap.sh` idempotente), disparado pelo workflow depois do `apply` — o RDS está em subnet privada e o runner não alcança.
- **DynamoDB:** state **próprio** `terraform/dynamodb/` (tabela `history`, `PAY_PER_REQUEST`, sem dependência de remote state — por isso é testável no MiniStack).
- CI: `fmt`/`validate` dos dois roots + job `dynamodb-ministack` (apply contra MiniStack e `describe-table`). Docs: README, `database-erd.md` (um ERD por database), ADRs 0005 (um RDS, database por serviço) e 0006 (DynamoDB para History). `lambda_access.tf` permanece.

### F8 — `auto-repair-shop-infra-k8s` (mensageria, registry, nós) — ficha: `docs/fase4/tasks/F8-infra-k8s.md`
- **State `terraform/messaging/`** (novo, sem remote state): 4 tópicos SNS, 5 filas SQS + 5 DLQs, *queue policies*, 8 assinaturas (raw delivery; `filter_policy` por atributo `target` em `saga-commands`). Declara a variável `manage_iam` (sem uso) porque o workflow a repassa a todo layer. Workflow ganha `layer=messaging` e o job `messaging-ministack` (confere 4 tópicos, 10 filas, 8 assinaturas).
- **`shared`:** 6 repositórios ECR novos via `for_each`; o ECR atual (gateway) não muda. No Lab (credenciais estáticas, sem OIDC) não há ajuste de IAM; em conta com IAM livre é preciso ampliar o trust das roles.
- **`aws`:** `aws_launch_template` com IMDS `hop_limit = 2` no node group do Lab (e `metadata_options` no ramo do módulo) para os pods lerem as credenciais da `LabRole`. ⚠️ Se o plano indicar *replace* do node group, **parar e avisar** (derruba os pods do STG).
- **`gateway` (API Gateway AWS):** sem mudança funcional (`/auth/customer-login` → lambda; resto → host do gateway Go via SSM, ADR 0003/0004).
- **Namespace único** `auto-repair-shop`; serviços internos `ClusterIP`; só o gateway Go é `LoadBalancer`. Dashboards/ServiceMonitors da saga ficam para a onda 4. Cada serviço deriva URLs/ARNs de SNS/SQS **por convenção de nome** no próprio `docker.yml` (sem remote state).

### F9 — `auto-repair-shop-lambda-auth`
- **Sem mudança de código Go nem de contrato** (`POST /auth/customer-login`, JWT, throttle, Function URL). As queries (`customer`, `user_role`, `role`) continuam válidas porque o database `customer` as contém.
- **Só `terraform/lambda.tf`** (hoje fixa `DB_USER="postgres"`, `DB_NAME="autorepairshop"`, `APP_SECRET_ID = output app_secret_arn`): passa a `DB_USER="lambda_auth_ro"`, `DB_NAME="customer"`, `APP_SECRET_ID = output lambda_secret_arn` (infra-db). `DB_HOST`/`DB_PORT`/`JWT_EXPIRES_IN` iguais; CI/Deploy inalterados.
- Validar no e2e (F10): login por CPF → JWT aceito por Service Order e Customer.

### F10 — `auto-repair-shop` (monólito → gateway funcional)
- **O que é:** reverse proxy Go (Gin + `httputil.ReverseProxy`), reaproveitando `middleware` (logger/request_id/recovery/metrics), `pkg/env`, `pkg/logger`, `pkg/metrics`. **Não** valida JWT nem acessa banco (cada serviço valida o seu).
- **Tabela de rotas** (`config/routes.yaml`, mais específica primeiro, por método+path; upstreams por env):

| Rota | Upstream |
|---|---|
| `/v1/auth/*`, `/v1/users/*`, `/v1/customers*`, `/v1/vehicles*` | `customer-service:8081` |
| `/v1/works*`, `/v1/supplies*` | `catalog-service:8082` |
| `/v1/budgets*`, `/v1/payments*` | `billing-service:8083` |
| `/v1/execution/*`, `PUT /v1/service-order/:id/work/:workId/{next,cancel}`, `PUT /v1/service-order/:id/finish` | `execution-service:8084` |
| `GET /v1/service-order/:id/history`, `/v1/reports/*` | `history-service:8085` |
| `/v1/service-order*` (demais) | `service-order-service:8086` |
| `/ping`, `/metrics`, `/swagger*` | próprio gateway |
| `/internal/*` | **404** (nunca exposto) |

- **Comportamento:** propaga/gera `X-Request-Id`; timeout por upstream; `502/504` padronizados no formato de erro atual; `/ping` próprio e `/ready` opcional checando os upstreams; Swagger UI único listando o `/swagger.yaml` de cada serviço (via proxy `/docs/<svc>/swagger.yaml`).
- **Remover:** `internal/{auth,customer,vehicle,work,supply,service_order,service_order_history}`, `migrations/`, `internal/pkg/{db,uow,auth}`, Postgres/`migrate`/`JWT_SECRET` do compose e do `k8s`, overlay `local`/Kind/toolbox. Manifests: Deployment + Service `LoadBalancer` + ConfigMap (URLs dos upstreams) + HPA + `ServiceMonitor`; o `docker.yml` continua publicando o host no SSM.
- **Ecossistema local (este repo):** `scripts/clone-services.sh` clona os 6 repos em `services/` (gitignored); `docker-compose.yml` sobe **um** MiniStack (`local/init-aws.sh` cria *todos* os tópicos/filas/tabela com os nomes do Terraform), **um** Postgres (`local/postgres-init/` cria os 5 databases + usuários + `lambda_auth_ro`), os 6 serviços (`build: services/<svc>`), a `lambda-auth` (Lambda Runtime Interface Emulator, rota `/auth/customer-login`) e o gateway em `:8080`. Alvos: `make clone`, `make up-all`, `make down-all`, `make seed`, `make e2e`.
- **e2e/BDD (godog)** rodando contra o ecossistema (job de CI do repo): (1) caminho feliz até `COMPLETED`; (2) pagamento recusado → compensação → `REJECTED` (`reason=PAYMENT_REJECTED`); (3) execução falha → estorno + liberação de estoque → `CANCELLED`; (4) login de cliente por CPF via lambda → `accept` com esse JWT; (5) mensagem duplicada não duplica efeito; (6) cliente rejeita o orçamento → `REJECTED` (`reason=BUDGET_REJECTED`) e estoque liberado.
- **CI:** baseline do §4 (sem integração com Postgres) + job `e2e`; cobertura ≥ 80% do gateway.

### FA — Ajustes e validação local (última onda de implementação)
Objetivo: fechar pendências acumuladas e **provar que o sistema inteiro funciona** antes de gravar o vídeo e montar o PDF (F11).

| # | Tarefa | Detalhe / aceite |
|---|---|---|
| A1 | **Padronizar e endurecer as respostas de erro** | (a) **5xx:** `internal/pkg/web/errors.go` (copiado do monólito) devolve a mensagem da causa raiz em 500/502 e em erros de infra não mapeados (pode vazar texto do driver Postgres). Trocar por mensagem genérica (`internal server error` / `upstream unavailable`) e registrar o erro completo no log com `request_id`. (b) **401/403:** o middleware de autenticação (herdado do monólito) responde `{"error":"Unauthorized"}`, diferente do padrão `{"code":<http>,"errors":["..."]}` usado nos demais erros; padronizar **todos** os 401 e 403 para o formato `{"code","errors"}` (`unauthorized`/`forbidden`). Aplicar nos **6 serviços** e no **gateway**; testes unitários por serviço (corpo 5xx sem a causa; corpo 401/403 no formato padrão) e ajustar o e2e para voltar a asserir `code` no 401 (`e2e/features/security.feature`). (c) **Validação sem mapeamento:** erros de domínio de entrada inválida que hoje caem em 500 — ex.: `ErrInvalidSupplyName/Description/UnitPrice/StockQuantity` no Catalog (já era assim no monólito) — passam a responder 400/422; auditar os `web/errors.go` dos 6 serviços em busca de erros de validação sem mapeamento. (d) **Formato do 409/400 do Catalog** (`{"code":409,"error":"..."}`) usa `error` em vez de `errors`: alinhar ao padrão. Os mapeamentos 4xx de domínio já existentes permanecem. |
| A2 | **Validação local ponta a ponta** | `scripts/validate-local.sh` + `make validate` no repo `auto-repair-shop`: sobe o ecossistema (`make up-all`), espera todos os `/ping`, roda a suíte e2e/BDD e imprime um checklist PASS/FAIL: (1) login de staff e de cliente via lambda por CPF; (2) OS completa até `DELIVERED` com `PAYMENT_MODE=auto-approve`; (3) pagamento recusado (`auto-reject`) ⇒ `REJECTED` e estoque devolvido; (4) falha de execução ⇒ `CANCELLED` e estorno; (5) `GET /history` e relatório de tempo médio; (6) mensagem duplicada não duplica efeito; (7) `/metrics` de cada serviço; (8) rotas `/internal/*` bloqueadas pelo gateway. **Exige daemon Docker** (a máquina de desenvolvimento do agente não tem): roda na CI do repo (job `e2e`) e na máquina do usuário. Aceite: `make validate` termina com 0 falhas. |
| A3 | **Checagem de exclusão de cliente/veículo** | Customer chama `GET /internal/service-orders/exists` antes de `DELETE` (`409` se houver OS); remove os `TODO(F6)`; variável `SERVICE_ORDER_URL` no Customer; teste com `httptest`. |
| A4 | **Observabilidade da saga e padronização dos logs** | Métricas `saga_started/completed/compensated_total`, `saga_step_duration_seconds`, `outbox_pending`, `dlq_messages` nos serviços aplicáveis; `ServiceMonitor` em cada serviço; dashboard Grafana "Saga" em `infra-k8s/terraform/addons/dashboards/`. **Formato dos logs:** todos os serviços e o gateway emitem JSON (Zap) com o mesmo conjunto de campos — `request_id`, `service` (nome do serviço) e, quando houver, `saga_id`, `service_order_id` e `message_id` — e nomes de evento no padrão `<entidade>.<ação>`; nunca CPF/CNPJ/JWT/`Authorization`. Aceite: um `saga_id` rastreável no Loki atravessando todos os serviços e métricas visíveis no Grafana (insumo do vídeo). |
| A5 | **Verificações adiadas por falta de Docker/AWS** | `docker compose up` isolado de cada serviço (checklist por serviço); `bootstrap.sh` contra `postgres:15-alpine` real; `terraform plan` de `shared`/`messaging`/`aws` (ler se o node group é *replace*); login do admin; teste "golden" do `GET /history` contra a saída real do monólito (`monolith-v1`). Resultado de cada item registrado no README do repo correspondente. |
| A6 | **Proteção de branch completa** | Adicionar os status checks obrigatórios (`pre-commit`, `build`, `unit-tests`, `integration-tests`, `sonarqube`, `coverage`) a `main` e `develop` de todos os repos de serviço (S7), agora que a CI já rodou; requer secrets/Sonar configurados. |
| A7 | **Limpeza e entregáveis por serviço** | Auditar `.ai/rules/*`, README (evidência de cobertura), `docs/swagger.yaml` (ou collection Postman) e ADRs de cada serviço; remover `internal/pkg/{db,uow}` não usados; `go mod tidy`; remover código morto. |
Ordem sugerida: A1 e A3 e A4 em paralelo (repos diferentes), depois A2 (depende de A1/A3/A4 para validar o estado final), por fim A5/A6/A7.

### F11 — Fechamento e entrega
- Deploy em STG de todos os serviços que ainda não estiverem implantados (ordem: infra-db → infra-k8s `shared`/`messaging`/`aws`/`addons` → customer, catalog, billing, execution, history → service-order → lambda → gateway) e validação do e2e contra STG.
- README do `auto-repair-shop`: arquitetura final, **escolha e justificativa da saga orquestrada**, justificativa da divisão, diagrama, links dos 9 repos, evidências de cobertura.
- PDF da entrega (participantes, links, vídeo, diagrama, saga, divisão e tecnologias) e vídeo ≤ 15 min (fluxo da OS, saga + falha/compensação, deploy com testes, monitoramento seguindo um `saga_id` no Grafana/Loki).
- Atualizar `PROJECTS_INDEX.md`, `CLAUDE.md` e `AGENTS.md` da raiz para refletir os 9 repos.

## 6. Observabilidade (reuso da Fase 3)
Logs Zap JSON com `request_id`, `saga_id`, `service_order_id`, `message_id` → Promtail/Loki; `request_id` nasce no gateway e viaja em REST e no envelope; `/metrics` + `ServiceMonitor` por serviço. O rastreamento distribuído da demo é feito filtrando um `saga_id` no Grafana (sem OpenTelemetry nesta fase).

## 7. Critérios de aceite e pontos a confirmar

**Aceite final**
- [ ] 6 repos de serviço + gateway (`auto-repair-shop`), cada um com CI verde, cobertura ≥ 80% (evidência no README), Sonar, `main` protegida e deploy em K8s
- [ ] Cada serviço sobe isolado com `make up`; o ecossistema sobe com `make up-all` e passa no `make e2e`
- [ ] Nenhum serviço acessa banco de outro (credenciais por database; lambda lê só o `customer`)
- [ ] `infra-db` e `infra-k8s` provisionam no Lab: RDS (1 instância, N databases), DynamoDB, SNS/SQS, ECRs; partes emuláveis testadas no MiniStack
- [ ] `lambda-auth` funcional, sem mudança de código
- [ ] Gateway `auto-repair-shop` roteia todas as rotas da API atual; `/internal/*` bloqueado
- [ ] Saga com caminho feliz e ≥ 2 falhas compensadas; outbox + idempotência + DLQ em todos os serviços
- [ ] Vídeo, PDF e README conforme §5/F11

**Decisões de domínio registradas (sem pendências)**
1. Pagamento recusado reutiliza `REJECTED` (motivo em `reason`).
2. `accept` é síncrono (204) e leva a OS a `IN_PROGRESS` na hora; o pagamento é processado depois. `AWAITING_APPROVAL` só existe enquanto o cliente não decidiu.
3. Diagnóstico permanece no Service Order, como no monólito; Execution cuida da fila/reparo após o pagamento. Plano B, se vier exigência: `ServiceOrderReceived` cria tarefa `QUEUED` no Execution e `start-diagnosis` passa por ele.
4. `send` mantém a OS em `IN_DIAGNOSIS` até `BudgetGenerated`.
