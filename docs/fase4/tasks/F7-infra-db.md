# Ficha F7 — `auto-repair-shop-infra-db` (dados de todos os serviços)

Repo existente; fluxo do ADR 0001: `feature → develop → main` (**base dos PRs = `develop`**, que já existe; o default do GitHub é `main`). Ler `contracts/{resources,ddl}.md` e `terraform/{rds,locals,variables,outputs,versions,lambda_access}.tf`. Regra de ouro: **o plano não pode alterar nem destruir nada existente** (RDS, SG, secret `${name}/app`) — só adicionar.

## T1 — Segredos por serviço (`terraform/secrets.tf`, novo)
- `locals.pg_services = { customer = {db="customer", user="customer_app", internal=true}, catalog = {…, internal=true}, billing = {…, internal=false}, execution = {…, internal=false}, service_order = {db="service_order", user="service_order_app", internal=true} }`.
- `random_password.service_db` (`for_each = local.pg_services`, `length=24`, `special=false`); `random_password.internal_token` (`length=40`, `special=false`). **Reusar** o `random_password.jwt` existente (não criar outro).
- Um `aws_secretsmanager_secret` + `_version` por serviço, nome `${local.name}/<chave com '_'→'-'>` (`customer`, `catalog`, `billing`, `execution`, `service-order`), `secret_string = jsonencode({POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB, JWT_SECRET, INTERNAL_TOKEN (só se internal=true)})`.
- `${local.name}/history` com `{JWT_SECRET}`.
- `${local.name}/lambda-auth` com `{POSTGRES_PASSWORD = random_password.lambda_ro.result, JWT_SECRET = random_password.jwt.result}` (`random_password.lambda_ro` novo). **Mesmas chaves que a lambda lê hoje.**
- Manter intactos `random_password.db`, `random_password.jwt`, `aws_secretsmanager_secret.app` (legado, até o F10).

## T2 — Outputs (`outputs.tf`)
Manter todos os atuais. Adicionar: `service_secret_arns` (map nome→ARN), `lambda_secret_arn`, `pg_services` (map nome→`{db,user}`) para o workflow. `app_secret_arn` **continua** apontando para o segredo legado.

## T3 — Bootstrap dos databases (Job no cluster)
Arquivos novos: `db-bootstrap/bootstrap.sh`, `db-bootstrap/job.yaml`.
- `bootstrap.sh` (POSIX sh, `set -eu`, idempotente), para cada serviço `s ∈ {customer,catalog,billing,execution,service_order}` com `PW_<S>` do ambiente:
  1. `CREATE ROLE <s>_app LOGIN` se não existir + `ALTER ROLE … PASSWORD` (mantém a senha sincronizada com o secret);
  2. `GRANT <s>_app TO CURRENT_USER` (no RDS o master precisa ser membro do role para definir `OWNER`);
  3. `CREATE DATABASE <s> OWNER <s>_app` se não existir; `REVOKE ALL ON DATABASE <s> FROM PUBLIC`.
  Use `psql -v ON_ERROR_STOP=1 -v role=… -v db=… -v pw=…` com `SELECT format('… %I … %L', :'role', :'pw') WHERE NOT EXISTS (…) \gexec` (padrão que evita `DO` e permite `CREATE DATABASE`).
  Depois, para a lambda: `CREATE ROLE lambda_auth_ro LOGIN PASSWORD :'PW_LAMBDA'` (idempotente); conectado em `-d customer`: `GRANT CONNECT ON DATABASE customer TO lambda_auth_ro; GRANT USAGE ON SCHEMA public TO lambda_auth_ro; ALTER DEFAULT PRIVILEGES FOR ROLE customer_app IN SCHEMA public GRANT SELECT ON TABLES TO lambda_auth_ro;` (as tabelas criadas depois pelas migrations do `customer_app` já nascem legíveis).
- `job.yaml`: `Job db-bootstrap` no namespace `auto-repair-shop`, imagem `postgres:15-alpine`, `backoffMillis` default, `ttlSecondsAfterFinished: 600`, script montado por ConfigMap `db-bootstrap-script`, variáveis de um Secret `db-bootstrap` (`PGHOST`, `PGUSER=postgres`, `PGPASSWORD`, `PGSSLMODE=require`, `PW_CUSTOMER`, `PW_CATALOG`, `PW_BILLING`, `PW_EXECUTION`, `PW_SERVICE_ORDER`, `PW_LAMBDA`).
- Verificação local: `shellcheck db-bootstrap/bootstrap.sh` e execução contra um `postgres:15-alpine` em Docker (`docker run`), **duas vezes seguidas** (idempotência), conferindo `\l` e `\du`.

## T4 — Terraform do DynamoDB (state próprio `terraform/dynamodb/`, novo)
- `versions.tf`: backend `s3` com `key = "dynamodb/terraform.tfstate"` (mesmo padrão parcial do `terraform/versions.tf`), provider `aws ~> 6.0`.
- `variables.tf`: `region`, `project` (default `auto-repair-shop`), `aws_endpoint` (string, default `null`), `state_bucket` (opcional/ignorada, para compatibilidade com o workflow) e nada que dependa de remote state.
- `provider "aws"`: quando `aws_endpoint != null` → `endpoints { dynamodb = var.aws_endpoint … }`, `skip_credentials_validation = true`, `skip_metadata_api_check = true`, `skip_requesting_account_id = true`, credenciais dummy via env.
- `main.tf`: `aws_dynamodb_table.history` — nome `${project}-${environment}-history` (`environment` = workspace, `default`→`stg`, igual ao `locals.tf`), `billing_mode = "PAY_PER_REQUEST"`, `hash_key = "PK"`, `range_key = "SK"`, atributos `S`, `point_in_time_recovery` ligado só em `prd`, `deletion_protection_enabled` só em `prd`, tags do projeto. `outputs.tf`: `history_table_name`, `history_table_arn`.

## T5 — Workflow (`.github/workflows/infra.yml`)
- Novo input `layer` (`db` default, `dynamodb`). Terraform roda em `terraform` ou `terraform/dynamodb`; `-var state_bucket=…` só para `db`.
- `validate` (PR): incluir `terraform/dynamodb` (`init -backend=false` + `validate`) e `terraform fmt -check -recursive terraform`.
- **Novo job** `dynamodb-ministack` (PR, sem credenciais reais): `services: ministack: image: ministackorg/ministack, ports: 4566`; passos: aguardar saúde (`curl` em loop), `terraform -chdir=terraform/dynamodb init -backend=false`, `apply -auto-approve -var aws_endpoint=http://localhost:4566`, e conferir `aws --endpoint-url http://localhost:4566 dynamodb describe-table --table-name auto-repair-shop-stg-history` (com `AWS_ACCESS_KEY_ID=test` etc.).
- **Novo passo** no job `run`, após o `apply` de `layer=db`: `bootstrap-db` — `aws eks update-kubeconfig --name "${EKS_CLUSTER_NAME:-auto-repair-shop-$ENV-eks}"`; criar namespace (idempotente); montar o Secret `db-bootstrap` lendo `terraform output -raw db_host`, a senha master de `auto-repair-shop-$ENV/app` e as senhas dos segredos por serviço/lambda no Secrets Manager (`aws secretsmanager get-secret-value … | jq -r .POSTGRES_PASSWORD`); `kubectl apply` do ConfigMap (`--from-file=db-bootstrap/bootstrap.sh`) e do Job; `kubectl wait --for=condition=complete --timeout=300s job/db-bootstrap`; imprimir `kubectl logs`; remover o Secret `db-bootstrap` ao final (sempre, `if: always()`).

## T6 — Documentação
README (nova estrutura: states `db` e `dynamodb`, bootstrap Job, secrets por serviço, como rodar o MiniStack); `docs/database-erd.md` (um ERD Mermaid por database a partir de `ddl.md`; History = modelo de itens DynamoDB); ADRs `0005-one-rds-instance-database-per-service.md` e `0006-dynamodb-for-history.md` (contexto, decisão, consequências, alternativas — seguir o formato dos ADRs existentes).

## T7 — Entrega
Branch `feat/fase4-databases`, commits Conventional Commits, PR (`gh pr create --base develop`); aguardar `validate`, `dynamodb-ministack`; **não** rodar `apply` em AWS (decisão do usuário, ver abaixo).

## Aceite
- [ ] `terraform fmt -check -recursive terraform` e `validate` dos dois roots ok
- [ ] Job `dynamodb-ministack` verde (tabela criada e descrita no MiniStack)
- [ ] `bootstrap.sh` roda 2× contra Postgres local sem erro e com o mesmo resultado
- [ ] `terraform plan` (stg, quando houver credenciais) mostra **somente adições**: 0 changes/0 destroys em `module.rds`, `aws_security_group.*`, `aws_secretsmanager_secret.app`
- [ ] Docs e ADRs atualizados; PR aberto
