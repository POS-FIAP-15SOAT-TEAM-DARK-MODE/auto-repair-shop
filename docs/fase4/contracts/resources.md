# Contrato — Recursos e nomes (fonte única)

> Estes nomes existem em vários lugares (Terraform `infra-k8s`/`infra-db`, `local/init-aws.sh` de cada serviço e do ecossistema, manifests K8s). **Devem ser idênticos.** Mudou aqui → mudar em todos.
> Ambiente real de execução: **AWS Academy Learner Lab** (`manage_iam=false`, credenciais estáticas, sem IRSA/External Secrets/ALB Controller). Tudo abaixo foi desenhado para isso.

## Repositórios, portas e imagens

| Serviço | Repo | Porta | Repositório ECR | `<svc>` (nome K8s) |
|---|---|---|---|---|
| Gateway | `auto-repair-shop` | 8080 | `auto-repair-shop` (atual) | `auto-repair-shop` (atual) |
| Customer & Identity | `auto-repair-shop-customer-service` | 8081 | `auto-repair-shop-customer-service` | `customer-service` |
| Catalog | `auto-repair-shop-catalog-service` | 8082 | `auto-repair-shop-catalog-service` | `catalog-service` |
| Billing | `auto-repair-shop-billing-service` | 8083 | `auto-repair-shop-billing-service` | `billing-service` |
| Execution | `auto-repair-shop-execution-service` | 8084 | `auto-repair-shop-execution-service` | `execution-service` |
| History & Reporting | `auto-repair-shop-history-service` | 8085 | `auto-repair-shop-history-service` | `history-service` |
| Service Order | `auto-repair-shop-service-order-service` | 8086 | `auto-repair-shop-service-order-service` | `service-order-service` |

URL interna entre serviços: `http://<svc>:<porta>` (namespace único `auto-repair-shop`), ex.: `http://customer-service:8081`.

## Objetos Kubernetes por serviço (todos no namespace `auto-repair-shop`)

Como todos os serviços compartilham o namespace, **todo objeto leva o prefixo `<svc>`** (o monólito usa nomes genéricos como `app-config`/`app-secrets`, que colidiriam):

| Objeto | Nome |
|---|---|
| Deployment, Service, HPA, ServiceAccount | `<svc>` |
| ConfigMap / Secret de ambiente | `<svc>-config` / `<svc>-secrets` |
| Job de migração / ConfigMap de migrations | `<svc>-db-migrate` / `<svc>-migrations` |
| Role/RoleBinding (initContainer `wait-for-migrations`) | `<svc>-wait-migrations` |
| Label `app` | `<svc>` |
| Tipo do Service | `ClusterIP` (só o gateway é `LoadBalancer`) |

## Bancos (RDS PostgreSQL — 1 instância, 1 database por serviço)

A instância atual (`auto-repair-shop-<env>-db`, usuário master `postgres`, database inicial `autorepairshop`) **é mantida**; os databases abaixo são criados nela.

| Serviço | Database | Usuário da aplicação |
|---|---|---|
| Customer & Identity | `customer` | `customer_app` |
| Catalog | `catalog` | `catalog_app` |
| Billing | `billing` | `billing_app` |
| Execution | `execution` | `execution_app` |
| Service Order | `service_order` | `service_order_app` |
| Lambda de login | `customer` (só leitura) | `lambda_auth_ro` |

`REVOKE CONNECT ... FROM PUBLIC` em cada database; cada usuário só conecta no seu. `lambda_auth_ro` recebe `SELECT` (via `ALTER DEFAULT PRIVILEGES FOR ROLE customer_app`) nas tabelas do `customer`.

## Secrets Manager (padrão atual `auto-repair-shop-<env>/<nome>`; **chaves do JSON = nomes das variáveis de ambiente**)

| Secret | Chaves |
|---|---|
| `auto-repair-shop-<env>/customer` | `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `JWT_SECRET`, `INTERNAL_TOKEN` |
| `auto-repair-shop-<env>/catalog` | idem |
| `auto-repair-shop-<env>/service-order` | idem |
| `auto-repair-shop-<env>/billing` | `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `JWT_SECRET` |
| `auto-repair-shop-<env>/execution` | idem billing |
| `auto-repair-shop-<env>/history` | `JWT_SECRET` |
| `auto-repair-shop-<env>/lambda-auth` | `POSTGRES_PASSWORD` (de `lambda_auth_ro`), `JWT_SECRET` — **mesmas chaves que a lambda já lê hoje** |
| `auto-repair-shop-<env>/app` (legado) | `POSTGRES_PASSWORD`, `JWT_SECRET` — mantido até o gateway substituir o monólito (F10) |

`JWT_SECRET` é **o mesmo valor** em todos (reaproveita o `random_password.jwt` existente: tokens já emitidos continuam válidos). `INTERNAL_TOKEN` é um `random_password` novo e único por ambiente. `POSTGRES_HOST` **não** vai no secret: vem do ConfigMap, preenchido no deploy com o `RDS_HOST` (como o monólito faz hoje). No Lab o workflow de deploy lê o secret com `aws secretsmanager get-secret-value` e cria o Secret K8s `<svc>-secrets` com `kubectl` (passo "Provision app secret" do `docker.yml` atual).

### Lambda (estado atual → alvo, sem mudar código Go)
Hoje o Terraform da lambda fixa `DB_USER="postgres"`, `DB_NAME="autorepairshop"`, `APP_SECRET_ID=<infra-db output app_secret_arn>`. Alvo (F9, só `terraform/lambda.tf`): `DB_USER="lambda_auth_ro"`, `DB_NAME="customer"`, `APP_SECRET_ID=<infra-db output lambda_secret_arn>`; `DB_HOST`/`DB_PORT`/`JWT_EXPIRES_IN` iguais.

## DynamoDB (state próprio `terraform/dynamodb/` no infra-db)

| Tabela (`auto-repair-shop-<env>-history`) | Chaves | Modo | Serviço |
|---|---|---|---|
| `history` (nome lógico; a env `DYNAMODB_TABLE` recebe o nome completo) | `PK` (S), `SK` (S) | `PAY_PER_REQUEST` | History |

## SNS / SQS (state próprio `terraform/messaging/` no infra-k8s)

| Tipo | Nome lógico |
|---|---|
| Tópico | `saga-commands` (atributo `target`: `billing`\|`execution`\|`catalog`), `so-events`, `billing-events`, `execution-events` |
| Fila (+ DLQ `<fila>-dlq`, `maxReceiveCount=5`) | `billing-queue`, `execution-queue`, `catalog-queue`, `so-saga-replies`, `history-queue` |

Nome físico = `auto-repair-shop-<env>-<nome lógico>` (AWS) ou `auto-repair-shop-local-<nome lógico>` (MiniStack). O código lê URLs/ARNs por variável de ambiente, **nunca** o nome. Cada fila tem *queue policy* permitindo `sqs:SendMessage` do(s) tópico(s) que a alimentam. Assinaturas e *filter policies*: ver `messaging.md`.

## Variáveis de ambiente por serviço

Comuns: `PORT`, `LOG_LEVEL`, `JWT_SECRET`, `AWS_REGION`, `AWS_ENDPOINT_URL` (só local → `http://ministack:4566`), `INTERNAL_TOKEN` (serviços com `/internal/*` ou que os chamam). Postgres: `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `POSTGRES_SSLMODE` (o `db.go` atual já lê; RDS usa `require`).

| Serviço | Específicas |
|---|---|
| Customer | `BCRYPT_COST`, `JWT_EXPIRES_IN` |
| Catalog | `CATALOG_QUEUE_URL` |
| Billing | `BILLING_QUEUE_URL`, `BILLING_EVENTS_TOPIC_ARN`, `PAYMENT_MODE` (`manual`\|`auto-approve`\|`auto-reject`) |
| Execution | `EXECUTION_QUEUE_URL`, `EXECUTION_EVENTS_TOPIC_ARN` |
| History | `HISTORY_QUEUE_URL`, `DYNAMODB_TABLE` |
| Service Order | `CUSTOMER_URL`, `CATALOG_URL`, `SO_SAGA_REPLIES_QUEUE_URL`, `SAGA_COMMANDS_TOPIC_ARN`, `SO_EVENTS_TOPIC_ARN`, `SAGA_STEP_TIMEOUT`, `SAGA_PAYMENT_TIMEOUT` |
| Gateway | `CUSTOMER_UPSTREAM`, `CATALOG_UPSTREAM`, `BILLING_UPSTREAM`, `EXECUTION_UPSTREAM`, `HISTORY_UPSTREAM`, `SERVICE_ORDER_UPSTREAM`, `UPSTREAM_TIMEOUT` |

## IDs fixos de seed (determinísticos; e2e e fixtures dependem deles)

| Entidade | ID |
|---|---|
| works (5) | `11111111-1111-4111-8111-00000000000{1..5}` |
| supplies | `22222222-2222-4222-8222-0000000000{01..NN}` (NN = quantidade do seed atual) |
| usuário attendant / mechanic | `33333333-3333-4333-8333-000000000001` / `…002` |
| customer A (João) — user / customer / vehicle | `44444444-4444-4444-8444-000000000001` / `…002` / `…003` |
| customer B (Transportes Silva) — user / customer / vehicle | `44444444-4444-4444-8444-000000000011` / `…012` / `…013` |
| admin | `3b123072-5ad5-4df8-8d3b-3c2770f11b2a` (já fixo na migration `000002`) |

`user_role.id` do seed também usa UUIDs fixos derivados (`55555555-…`). Os seeds rodam **só** se `SEED_ENABLED=true` (default `false` em `prd`).

## SSM
O gateway mantém `/auto-repair-shop/<env>/app-backend-host` (ADR 0004). Os demais serviços **não** publicam host (são `ClusterIP`).
