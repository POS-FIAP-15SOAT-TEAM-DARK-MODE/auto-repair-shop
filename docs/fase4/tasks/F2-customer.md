# Ficha F2 — Customer & Identity (`auto-repair-shop-customer-service`, :8081, PG `customer`)

Pré-requisitos: F0 concluída; ler `_baseline.md`, `contracts/{resources,internal-rest,ddl}.md`. `<svc>=customer-service`, `<db>=customer`, secret `auto-repair-shop-<env>/customer`. **Sem mensageria** (sem `ministack`/`aws-init`/`messaging`).

## T1 — Baseline
B1–B3 e B6–B7 de `_baseline.md` (compose só com `app`, `db`, `migrate`).

## T2 — Código de domínio (mover do monólito `$SRC`)
- `cp -r $SRC/internal/auth $SRC/internal/customer $SRC/internal/vehicle` para `internal/`.
- `internal/app/middleware/auth.go` **permanece** importando `internal/auth/domain` (o tipo `Role` mora aqui, como hoje).
- `internal/pkg/auth/jwt.go` literal (emissão e validação do JWT).
- `bootstrap`/`container`/`routing`: `HandlersWrapper{UserHandler, CustomerHandler, VehicleHandler}`; rotas **idênticas ao original** para `/v1/auth/register` (ADMIN), `/v1/auth/login`, `PATCH /v1/users/:id/role` (ADMIN), `/v1/customers*` (AttendantRoles), `/v1/vehicles*` (GET atendente+mecânico; escrita atendente).
- Verificar: `grep -rn 'internal/work\|internal/supply\|internal/service_order' --include=*.go .` **vazio**.
- O comportamento "criar customer cria user `CUSTOMER` (senha inicial = CPF/CNPJ, bcrypt ≥ 12) e excluir customer exclui o user" já é transacional via `uow` e **continua igual** (agora no mesmo database).

## T3 — Endpoints internos (código novo)
- Copiar `InternalAuth()` (mesma implementação da ficha F1/T3; duplicar sem cerimônia).
- `GET /internal/customers/:id` → `{id, user_id, type, name, document}`; `name` = `company_name` quando `COMPANY`, senão `user.name`; `document` = `cpf` ou `cnpj` **sem formatação**; `404` se inexistente. Reusar o repositório existente de customer (buscar por id com o usuário).
- `GET /internal/vehicles/:id` → `{id, customer_id, license_plate, brand, model, year}`; `404` se inexistente.
- Respostas e erros exatamente como `internal-rest.md`.

## T4 — Migrations, seed e compatibilidade da lambda
- `migrations/000001_init_schema` e `000002_create_user_admin` conforme `contracts/ddl.md` (seção customer).
- `internal/app/seed.go`: copiar do original **apenas** staff (attendant/mechanic), customers A/B e veículos; trocar `uuid.NewString()` pelos **IDs fixos** de `resources.md` (inclusive `user_role.id`); remover `insertWorks`/`insertSupplies`; `isAlreadySeeded` passa a checar `COUNT(*) FROM customer`; só roda com `SEED_ENABLED=true`.
- `local/postgres-init/01-lambda-ro.sql` (montado em `/docker-entrypoint-initdb.d` do `db`): cria role `lambda_auth_ro` (`LOGIN PASSWORD 'lambda'`) e `ALTER DEFAULT PRIVILEGES … GRANT SELECT ON TABLES TO lambda_auth_ro` para o usuário dono; após as migrations, `GRANT SELECT ON ALL TABLES IN SCHEMA public TO lambda_auth_ro` (documentar o passo no README; no Lab quem faz é o Job de bootstrap do F7).

## T5 — Testes e docs
- Mover testes unitários de `auth`, `customer`, `vehicle`; mover de `$SRC/internal/integration` **apenas** `setup_test.go` (reduzido a auth/customer/vehicle), `helpers_test.go`, `auth_test.go`, `customer_test.go`, `vehicle_test.go`, ajustando imports e removendo referências a work/supply/service_order.
- Novo teste de integração `lambda_compat_test.go`: com o Postgres real, conectar como `lambda_auth_ro` e executar **as duas queries literais da lambda** (`SELECT c.user_id FROM customer c WHERE c.cpf = $1` e `SELECT r.name FROM user_role ur JOIN role r ON r.id = ur.role_id WHERE ur.user_id = $1`) para o customer A semeado (CPF `529.982.247-25`; **verificar como o CPF é gravado** — formatado ou só dígitos — e registrar no README qual formato a lambda deve receber); garantir que um `INSERT` como `lambda_auth_ro` falha.
- Testes das rotas internas (200/404/401). `docs/swagger.yaml`: paths de auth/users/customers/vehicles do original. README + ADR `0002-customer-and-identity-in-one-service.md` (por que juntos: acoplamento transacional e lambda).

## T6 — Entrega
B8 de `_baseline.md`.

## Aceite
- [ ] `go build ./... && go vet ./...`; `go test ./... --race`; **cobertura ≥ 80%**
- [ ] `docker compose up -d --build`: `GET :8081/ping` = 200; `POST /v1/auth/login` com o admin (`admin@email.com`, senha original da migration) devolve token; criar customer cria user (verificar no banco)
- [ ] Teste `lambda_compat` verde (queries da lambda funcionam como `lambda_auth_ro`; escrita negada)
- [ ] `GET /internal/customers/<id fixo>` e `/internal/vehicles/<id fixo>` conforme contrato, com e sem `X-Internal-Token`
- [ ] 3 overlays do kustomize ok; PR aberto para `develop` com jobs verdes (`sonarqube` conforme secret)
