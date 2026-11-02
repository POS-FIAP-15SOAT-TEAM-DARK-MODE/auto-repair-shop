# Ficha F9 — `auto-repair-shop-lambda-auth` (apontar para o database `customer`)

Repo existente. Fluxo do ADR 0001: `feature → develop → main`; **base = `develop`**. **Sem mudança de código Go nem de contrato** (`POST /auth/customer-login`, JWT, throttle, Function URL). Branch **local** `feat/lambda-customer-database` criada de `origin/develop` — **não fazer push, não abrir PR** (os PRs são abertos depois, quando o usuário mergear as ondas 2 e 3).

## T1 — `terraform/lambda.tf` (hoje fixa `postgres`/`autorepairshop` e lê `app_secret_arn`)
- Linha da política de IAM (ramo `manage_iam=true`): `resources = [data.terraform_remote_state.db.outputs.lambda_secret_arn]` (era `app_secret_arn`).
- `environment.variables`: `DB_USER = "lambda_auth_ro"`, `DB_NAME = "customer"`, `APP_SECRET_ID = data.terraform_remote_state.db.outputs.lambda_secret_arn`. `DB_HOST`, `DB_PORT`, `JWT_EXPIRES_IN` iguais.
- Os dois outputs (`db_host`, `db_port`) e o `lambda_access_security_group_id` continuam vindo do state do `infra-db`; **`lambda_secret_arn` é um output novo** criado no PR #9 do `infra-db` (F7). O formato do segredo (`POSTGRES_PASSWORD`, `JWT_SECRET`) é o que a lambda já lê — nada muda no Go.

## T2 — Documentação e verificação
- `README.md` (linha que lista os outputs do `infra-db`, ~82): citar `lambda_secret_arn` no lugar de `app_secret_arn`; acrescentar uma seção curta "Banco de dados": a lambda lê o database `customer` como `lambda_auth_ro` (somente leitura em `customer`, `user_role`, `role`), CPF armazenado só com dígitos.
- Verificar: `terraform fmt -check -recursive terraform`; `mkdir -p build && touch build/lambda.zip && terraform -chdir=terraform init -backend=false -input=false && terraform -chdir=terraform validate`; `gofmt -l .` vazio; `go vet ./... && go test ./... --race` (nada de Go deve mudar).
- Commits Conventional Commits (`fix(terraform): ...`, `docs: ...`) com o trailer `Co-Authored-By`.

## Aceite
- [ ] `git diff origin/develop --stat` toca **apenas** `terraform/lambda.tf` e `README.md`
- [ ] `terraform fmt`/`validate` ok; Go inalterado e testes passando
- [ ] Branch local pronta; **nenhum push, nenhum PR**
- [ ] (Depois do merge do `infra-db` #9 e do apply) `lambda_secret_arn` existe no state do `db`; sem isso o `plan` da lambda falha ao ler o output — registrar no README como pré-requisito de deploy
