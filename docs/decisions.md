# Registro de decisões — Fase 4 (microsserviços + saga)

Este repositório (`auto-repair-shop`) é o ponto de entrada do sistema: o gateway, o ecossistema local e a documentação da arquitetura. Aqui ficam o **spec**, os **contratos** e as **decisões** que valem para os 9 repositórios.

- Spec completo: [`fase4/SPEC.md`](fase4/SPEC.md)
- Contratos (nomes, mensagens, REST interno, DDL): [`fase4/contracts/`](fase4/contracts)
- Fichas de tarefa por fase: [`fase4/tasks/`](fase4/tasks)
- ADRs deste repositório: [`adr/`](adr) (a decisão de arquitetura geral é a [ADR 0006](adr/0006-microservices-and-saga-architecture.md))

## Mapa dos repositórios

| Repositório | Papel |
|---|---|
| `auto-repair-shop` | gateway (reverse proxy), ecossistema local (`make up-all`), e2e BDD, documentação |
| `auto-repair-shop-customer-service` | usuários, roles, clientes, veículos, login de staff (PostgreSQL `customer`) |
| `auto-repair-shop-catalog-service` | serviços (`work`), peças e estoque com reservas (PostgreSQL `catalog`) |
| `auto-repair-shop-billing-service` | orçamento e pagamento, sem gateway de pagamento real (PostgreSQL `billing`) |
| `auto-repair-shop-execution-service` | fila de execução e progresso dos works (PostgreSQL `execution`) |
| `auto-repair-shop-history-service` | linha do tempo da OS e relatório de tempo médio (DynamoDB) |
| `auto-repair-shop-service-order-service` | OS e orquestrador da saga (PostgreSQL `service_order`) |
| `auto-repair-shop-lambda-auth` | login do cliente por CPF (Lambda), mantida |
| `auto-repair-shop-infra-db` / `-infra-k8s` | RDS, DynamoDB, segredos / EKS, SNS/SQS, ECR, API Gateway |

## Decisões de arquitetura

| # | Decisão | Motivo | Onde está detalhada |
|---|---|---|---|
| D1 | **6 microsserviços**, cada um com repositório, banco e CI próprios (mínimo exigido: 3) | dono do dado = dono da regra; ciclo de vida da OS (atendimento → orçamento/pagamento → execução) | SPEC §2–3 |
| D2 | **Customer e Identity num serviço só** | `customer` já depende de `auth` no código; a lambda lê `customer`, `user_role` e `role` e continua **sem mudança de contrato** | SPEC §2 |
| D3 | **Saga orquestrada**, orquestrador no Service Order | fluxo linear de ordem fixa, compensação em ordem inversa, estado visível em `saga_instance`; reaproveita a máquina de estados da OS | SPEC §3.2; ADR do service-order |
| D4 | **SNS + SQS** (filas Standard, DLQ por fila, raw delivery, *filter policy* por atributo `target`) | exigência do time; MiniStack permite rodar igual localmente | `contracts/messaging.md`; ADR 0005 do infra-k8s |
| D5 | **Outbox transacional + consumidor idempotente** (`processed_message`) em todo serviço | sem dual-write; entrega at-least-once com efeito único | `contracts/messaging.md` |
| D6 | **REST interno só onde a resposta é imediata** (`/internal/*` com `X-Internal-Token`, nunca roteado pelo gateway) | reserva de estoque, snapshots de cliente/work, exclusão de cliente | `contracts/internal-rest.md` |
| D7 | **Snapshots**: a OS guarda cópia de cliente, veículo, works e supplies | nenhum acesso a banco de outro serviço; respostas de listagem/detalhe sem chamadas cruzadas | `contracts/ddl.md` |
| D8 | **PostgreSQL**: uma instância RDS, um database e um usuário por serviço. **DynamoDB** só no History | isolamento sem custo de várias instâncias; NoSQL onde o acesso é por chave e append-only | `contracts/resources.md`; ADRs 0005/0006 do infra-db |
| D9 | **O monólito vira o gateway** (reverse proxy), com o API Gateway da AWS ainda à frente | um único ponto público; serviços `ClusterIP` | ADR 0005 deste repositório |
| D10 | **Lambda mantida** com o mesmo contrato; só lê o database `customer` como `lambda_auth_ro`. `DB_SSLMODE` opcional (default `require`) existe só para o ecossistema local | requisito do projeto; sem acesso a banco de outro serviço | `fase4/tasks/F9-lambda-auth.md` |
| D11 | **Billing sem Mercado Pago**: interface `PaymentProvider` e adaptador interno (`manual`, `auto-approve`, `auto-reject`) | começar com o mínimo e plugar o Mercado Pago depois sem mudar a saga | SPEC F3 |
| D12 | **Banco de dados e mensageria rodam no Lab sem IAM próprio** (`manage_iam=false`, role do node, sem IRSA) | restrição do AWS Academy Learner Lab | `contracts/resources.md` |
| D13 | **Repositórios públicos** (org no plano Free) | proteção de branch e SonarCloud gratuitos exigem repos públicos | SPEC §4 |

## Decisões de domínio (comportamento da OS)

| # | Decisão |
|---|---|
| R1 | Pagamento recusado **reutiliza o status `REJECTED`**; o motivo (`BUDGET_REJECTED` ou `PAYMENT_REJECTED`) vai em `reason`. Não há status novo. |
| R2 | **`AWAITING_APPROVAL` só existe enquanto o cliente não decidiu.** O `accept` leva a OS a `IN_PROGRESS` imediatamente (204); o pagamento e a fila de execução são processados depois, sem mudar o status. |
| R3 | `send` mantém a OS em `IN_DIAGNOSIS` até o orçamento existir (`BudgetGenerated`). |
| R4 | O diagnóstico continua sendo transição de status da OS (Service Order); o Execution cuida da fila e do reparo depois do pagamento. |
| R5 | **`finish` exige todos os works `COMPLETED` ou `CANCELLED`** (regra nova, não existia no monólito); todos `CANCELLED` ⇒ `ExecutionFailed` e a OS vai a `CANCELLED`. |
| R6 | **`cancel` sempre libera o estoque reservado** (no monólito só o `reject` devolvia). Reject, falha de pagamento e falha de execução também compensam. |
| R7 | Compensações são **fire-and-forget** e idempotentes; a saga é marcada `COMPENSATED` ao publicar os comandos. |
| R8 | Documentos (CPF/CNPJ) são armazenados **sem formatação**; o seed usa os mesmos IDs fixos em todos os ambientes. |
| R9 | Work inativo é tratado como não encontrado pela OS. `RemoveSupply` libera a reserva **depois** de gravar a remoção do vínculo (falha conservadora: estoque fica reservado). |
| R10 | **O cliente não vê o pagamento da própria OS.** `GET /v1/payments/:serviceOrderId` é restrito a `Attendant` e `Admin` (o pagamento é uma operação do balcão); o cliente vê apenas o orçamento (`GET /v1/budgets/:serviceOrderId`, com checagem de dono) e o status da OS. |

## Decisões de engenharia

- Código repetido entre serviços é **duplicado** (nada de módulo compartilhado); o pacote `messaging` nasceu no Billing e os demais o copiaram.
- Todo serviço nasce com a mesma CI: `pre-commit`, `build`, `unit-tests`, `integration-tests`, `sonarqube`, `coverage` (≥ 80%, **sem contar mocks gerados**; ≥ 90% nos pacotes críticos).
- Branches: `feature → develop → main`; commits em Conventional Commits.
- Valores monetários nas mensagens são strings decimais (`"150.00"`), nunca float.

## Decisões pendentes

| # | Pendência | Recomendação |
|---|---|---|
| P1 | `internal/pkg/web/errors.go`, herdado do monólito, devolve a mensagem da causa raiz em respostas 5xx | trocar por mensagem genérica e manter o erro completo no log (tarefa A1 da onda 5) |
| P1b | Os 401/403 dos serviços respondem `{"error":"Unauthorized"}`, diferente do padrão `{"code","errors"}` dos demais erros | padronizar o corpo em todos os serviços e no gateway (tarefa A1) |
| P1c | Formato dos logs não é uniforme entre os serviços | mesmo conjunto de campos JSON (`request_id`, `service`, `saga_id`, `service_order_id`, `message_id`) (tarefa A4) |
| P2 | Exclusão de cliente/veículo não verifica OS vinculadas desde a separação | Customer chamar `GET /internal/service-orders/exists` (tarefa A3) |
| P3 | Checks obrigatórios nas branches protegidas | adicionar depois que a CI tiver rodado em `develop` (tarefa A6) |
