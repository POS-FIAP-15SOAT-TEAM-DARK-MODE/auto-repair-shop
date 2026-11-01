# Ficha F10 — `auto-repair-shop`: monólito → gateway + ecossistema local

Repo existente (hoje o monólito). Fluxo: **base `develop`**; branch **local** `feat/gateway` criada de `origin/develop` — **sem push e sem PR por ora**. Ler `_baseline.md` (B9), `contracts/{resources,messaging,internal-rest}.md` e `SPEC.md` (§2, F10). Baseline do monólito congelada em `$SRC` (tag `monolith-v1`).
Dividida em **F10a** (código do gateway, testes, k8s, CI, remoção do legado) e **F10b** (ecossistema local, scripts e e2e BDD) — F10b só começa depois de F10a. Nada de `docker compose up` neste ambiente (sem daemon): o ecossistema e o e2e são **escritos agora e validados na CI/máquina do usuário depois dos merges**.

---
## F10a — Gateway (reverse proxy Go)

### T1 — O que o gateway é
Gin + `net/http/httputil.ReverseProxy`. **Não** valida JWT nem acessa banco (cada serviço valida o seu). Reaproveita do monólito: `internal/app/{server.go}` (servidor HTTP com shutdown), `internal/app/middleware/{logger,recovery,metrics}.go` (+ testes), `internal/pkg/{app,env,logger,metrics(só HTTPRequestDuration),web(formato de erro)}`. Remove o resto (T7).

### T2 — Tabela de rotas (`config/routes.yaml`, embutida com `go:embed`; ordem = prioridade; casa **método + path**)
```yaml
routes:
  - {methods: ["*"],   path: "/internal/*",                                  block: true}      # 404, nunca exposto
  - {methods: ["PUT"], path: "/v1/service-order/:id/work/:workId/next",      upstream: execution}
  - {methods: ["PUT"], path: "/v1/service-order/:id/work/:workId/cancel",    upstream: execution}
  - {methods: ["PUT"], path: "/v1/service-order/:id/finish",                 upstream: execution}
  - {methods: ["GET"], path: "/v1/execution/*",                              upstream: execution}
  - {methods: ["GET"], path: "/v1/service-order/:id/history",                upstream: history}
  - {methods: ["GET"], path: "/v1/reports/*",                                upstream: history}
  - {methods: ["*"],   path: "/v1/auth/*",                                   upstream: customer}
  - {methods: ["*"],   path: "/v1/users/*",                                  upstream: customer}
  - {methods: ["*"],   path: "/v1/customers*",                               upstream: customer}
  - {methods: ["*"],   path: "/v1/vehicles*",                                upstream: customer}
  - {methods: ["*"],   path: "/v1/works*",                                   upstream: catalog}
  - {methods: ["*"],   path: "/v1/supplies*",                                upstream: catalog}
  - {methods: ["*"],   path: "/v1/budgets*",                                 upstream: billing}
  - {methods: ["*"],   path: "/v1/payments*",                                upstream: billing}
  - {methods: ["*"],   path: "/v1/service-order*",                           upstream: service_order}   # o restante
  - {methods: ["POST"],path: "/auth/customer-login",                         upstream: lambda, optional: true}
```
`*` no fim de segmento casa qualquer sufixo; `:param` casa um segmento. Upstreams por env (default = DNS do cluster, `resources.md`): `CUSTOMER_UPSTREAM=http://customer-service:8081`, `CATALOG_UPSTREAM=http://catalog-service:8082`, `BILLING_UPSTREAM=http://billing-service:8083`, `EXECUTION_UPSTREAM=http://execution-service:8084`, `HISTORY_UPSTREAM=http://history-service:8085`, `SERVICE_ORDER_UPSTREAM=http://service-order-service:8086`, `LAMBDA_UPSTREAM` (vazio = rota `optional` desligada; só o ecossistema local a liga), `UPSTREAM_TIMEOUT` (default `30s`). Rotas não casadas → `404` no formato padrão `{"code":404,"errors":["route not found"]}`.
**Não use `router.Any` do Gin para estas rotas** (conflitos entre `:id`/`*`): registre `/ping`, `/ready`, `/metrics`, `/swagger*`, `/docs/*` explicitamente e trate o resto em `router.NoRoute` com um **matcher próprio** (pacote `internal/gateway/match`, sem dependências).

### T3 — Proxy e comportamento
- Preserva método, path, query, corpo e cabeçalhos (inclusive `Authorization`); acrescenta `X-Forwarded-For/Host/Proto`.
- `X-Request-Id`: usa o recebido ou gera um UUID; devolve na resposta e propaga ao upstream; entra no log.
- Timeout por upstream (`UPSTREAM_TIMEOUT`, `ResponseHeaderTimeout`): estouro → `504`; upstream indisponível/erro de conexão → `502`. Corpo `{"code":502|504,"errors":["upstream unavailable"|"upstream timeout"]}` — **nunca** a causa original (log completo no servidor). Resposta do upstream (inclusive 4xx/5xx) é repassada sem alteração.
- `GET /ping` → `{"message":"pong"}`; `GET /ready` → `200` se todos os upstreams configurados respondem `/ping` em ≤ 2 s, senão `503` com a lista (nomes, sem URLs).
- `GET /metrics` (Prometheus, `http_request_duration_seconds` com labels `route` e `upstream`), logs JSON com `request_id`, `upstream`, `status`, `duration_ms`.
- `GET /docs/<svc>/swagger.yaml` faz proxy para `<upstream>/swagger.yaml`; `GET /swagger/index.html` = página HTML estática (embutida) com links para o swagger de cada serviço.

### T4 — Estrutura
`cmd/service/main.go` (ligando o bootstrap), `internal/gateway/{config,match,proxy,handlers}`, `internal/app/{server.go,middleware/*}`, `internal/pkg/{app,env,logger,metrics,web}`, `config/routes.yaml`, `docs/` (embed do swagger/HTML), testes ao lado.

### T5 — Testes (cobertura total ≥ 80%; **≥ 90%** em `internal/gateway/match`)
Unitários com `httptest`: matcher (precedência, `:param`, `*`, método, `/internal/*` bloqueado, rota inexistente), proxy (repasse de método/corpo/query/headers, `X-Request-Id` gerado e propagado, upstream 4xx/5xx repassado, upstream fora → 502, lento → 504 sem vazar a causa, rota `optional` desligada), `/ready`, `/ping`, `/metrics`, tabela completa de rotas (um teste por linha de `routes.yaml` confirmando o upstream escolhido).

### T6 — Kubernetes e CI/CD
- `k8s/manifests`: manter `base` (Deployment/Service/HPA), `overlays/{lab,stg,prd}` e `components/observability`; **remover** `migrate-job.yaml`, `rbac.yaml`, o initContainer `wait-for-migrations`, `overlays/local`, `k8s/kind`, `k8s/toolbox`, secrets/ExternalSecrets de banco e o `JWT_SECRET`. ConfigMap `app-config` com os `*_UPSTREAM`. Nomes e labels permanecem `auto-repair-shop` (é o gateway; já é o que o API Gateway AWS e o SSM `/auto-repair-shop/<env>/app-backend-host` esperam). Service `LoadBalancer` no `lab` (como hoje).
- `.github/workflows/docker.yml`: remover "Sync migrations ConfigMap", "Provision app secret" e o patch de `POSTGRES_HOST`/RDS; **manter** build/push, deploy, rollout e "Publish app URL" (SSM). `.github/workflows/ci.yml`: manter `pre-commit`, `build`, `unit-tests`, `sonarqube`, `coverage` (gate sem mocks, ≥ 80%); **remover** `integration-tests` (sem banco) — o job `e2e` entra no F10b.
- `Makefile`, `.env.example`, `sonar-project.properties` (mantém a chave atual), `.ai/**`, `CLAUDE.md`, `AGENTS.md`, `README.md`, `docs/swagger.yaml` (só `/ping`, `/ready`, `/metrics`) reescritos para o gateway. ADR novo `docs/adr/0005-monolith-becomes-gateway.md` (por que reverse proxy; o API Gateway AWS continua à frente).

### T7 — Remoção do legado (commit separado `refactor!: ...`)
Apagar `internal/{auth,customer,vehicle,work,supply,service_order,service_order_history,integration}`, `internal/app/{db.go,seed.go,container,routing}` (reescrever o necessário), `internal/pkg/{db,uow,auth,json,id}` (se sem uso), `migrations/`, `docker-compose.yml` atual e `Dockerfile` ajustado (sem migrations). `go mod tidy`. **Não** apagar `docs/fase4/` (se existir untracked/tracked) nem `.git`.

### Aceite F10a
- [ ] `go build ./...`, `go vet ./...`, `gofmt -l .` vazio, `LOG_LEVEL=PANIC go test ./... --race`; cobertura ≥ 80% (≥ 90% no matcher)
- [ ] `grep -rn 'internal/\(auth\|customer\|vehicle\|work\|supply\|service_order\)' --include=*.go .` **vazio**; `kubectl kustomize` dos 3 overlays ok; nenhum `migrate`/`POSTGRES`/`JWT_SECRET` restante no `k8s/` e nos workflows
- [ ] Branch local `feat/gateway`; **nenhum push, nenhum PR**

---
## F10b — Ecossistema local, scripts e e2e BDD (depois de F10a, mesma branch)

### T1 — `scripts/clone-services.sh` e Makefile
`clone-services.sh` clona (ou faz `git pull --ff-only` em) os 6 repos de serviço + `auto-repair-shop-lambda-auth` em `services/<nome>` (HTTPS; `services/` no `.gitignore`); aceita `SERVICES_BRANCH` (default `develop`). Alvos: `make clone`, `make up-all` (clone se faltar + `docker compose up -d --build` + espera dos `/ping`), `make down-all` (`down -v`), `make logs`, `make e2e`, `make test`.

### T2 — `docker-compose.yml` do ecossistema
Serviços: **um** `ministack` (porta 4566, healthcheck) e `aws-init` (`local/init-aws.sh`: **todos** os tópicos/filas/DLQs/assinaturas do `messaging.md`/`resources.md` com prefixo `auto-repair-shop-local-`, a tabela DynamoDB `auto-repair-shop-local-history`); **um** `postgres:15-alpine` com `local/postgres-init/*.sql` (cria os 5 databases/usuários de `resources.md` com senhas locais fixas e o `lambda_auth_ro` com `ALTER DEFAULT PRIVILEGES`); um `migrate` por serviço relacional (imagem `migrate/migrate:v4.17.1`, volume `services/<svc>/migrations`); os 6 serviços (`build: services/<svc>`, portas 8081–8086, variáveis do `resources.md`, `SEED_ENABLED=true`, **`PAYMENT_MODE=manual`**, `AWS_ENDPOINT_URL=http://ministack:4566`, mesmo `JWT_SECRET` e `INTERNAL_TOKEN` de desenvolvimento); a lambda (T3); o gateway (`build: .`, porta 8080, `*_UPSTREAM` para os serviços e `LAMBDA_UPSTREAM`). Rede `auto-repair-shop-net`; `depends_on` com `condition: service_healthy/completed_successfully`.

### T3 — Lambda local (`local/lambda/`)
`Dockerfile` multi-stage: compila `services/auto-repair-shop-lambda-auth/cmd/lambda` para `bootstrap` e usa `public.ecr.aws/lambda/provided:al2023` (Lambda Runtime Interface Emulator já embutido, porta 8080 interna). Variáveis: `DB_HOST=postgres`, `DB_USER=lambda_auth_ro`, `DB_NAME=customer`, `APP_SECRET_ID` — como a lambda lê o Secrets Manager, o `aws-init` cria o segredo `…/lambda-auth` no MiniStack com `{"POSTGRES_PASSWORD":"<senha local do lambda_auth_ro>","JWT_SECRET":"<o mesmo dos serviços>"}` e a lambda recebe `AWS_ENDPOINT_URL` + credenciais dummy. **Adaptador HTTP→evento** (`local/lambda/adapter/main.go`, ~60 linhas, container próprio ou no mesmo): serve `POST /auth/customer-login`, monta um `APIGatewayV2HTTPRequest` (`rawPath`, `headers`, `body`, `requestContext.http.method`) e faz `POST` em `http://lambda:8080/2015-03-31/functions/function/invocations`, devolvendo status/headers/corpo do `APIGatewayV2HTTPResponse`. O gateway aponta `LAMBDA_UPSTREAM` para o adaptador. **O adaptador também serve `GET /ping` → `200 {"message":"pong"}`**: com `LAMBDA_UPSTREAM` configurado o `/ready` do gateway chama `<lambda>/ping`; sem isso ele ficaria `503`. (O gateway já implementa as rotas de docs como `/docs/<nome>/swagger.yaml` com o nome do upstream em snake_case, ex.: `service_order`.)

### T4 — e2e BDD (`e2e/`, módulo Go próprio com `godog`)
`features/*.feature` em português ou inglês, steps em Go falando **só com o gateway** (`http://localhost:8080`). Token de staff via `POST /v1/auth/login` com os usuários do seed (`attendant@autorepairshop.com`/`attendant123`, `mechanic@autorepairshop.com`/`mechanic123`); token do cliente via `POST /auth/customer-login {"cpf":"52998224725"}`. IDs fixos do seed (`resources.md`). Pagamento sempre em modo `manual`: aprovar com `POST /v1/payments/{id}/confirm`, recusar com `/reject` (atendente). Cenários (espera com polling de até 20 s por status):
 1. **Caminho feliz** — criar OS → `received` → `start-diagnosis` (mecânico) → adicionar supply → `send` → orçamento disponível (`GET /v1/budgets/{so}`) e OS `AWAITING_APPROVAL` → `accept` do cliente ⇒ OS `IN_PROGRESS` imediato → `confirm` do pagamento → fila de execução tem a tarefa → mecânico `next`/`next` em cada work → `finish` ⇒ `COMPLETED` → `deliver` ⇒ `DELIVERED`; `GET /history` lista as transições com `workTransitions`; relatório `average-execution-time` devolve os works.
 2. **Pagamento recusado** — após `accept`, atendente `reject` do pagamento ⇒ OS `REJECTED` e o estoque da peça volta ao valor inicial (`GET /v1/supplies`).
 3. **Execução falha** — pagamento confirmado, mecânico cancela todos os works, `finish` ⇒ OS `CANCELLED`, pagamento `REFUNDED`, estoque devolvido.
 4. **Cliente rejeita o orçamento** — `reject` ⇒ `REJECTED` e estoque devolvido.
 5. **Login do cliente pela lambda** — `POST /auth/customer-login` ⇒ 200 com `token`; o token aceito em `GET /v1/service-order/{id}/status` e em `accept`; CPF inexistente ⇒ 404.
 6. **Segurança de rota** — `GET /internal/customers/<id>` ⇒ 404 pelo gateway; rota sem token ⇒ 401 vindo do serviço.
 7. **Idempotência** — (se viável via gateway) reenviar `accept`/`confirm` não duplica efeitos; caso contrário documentar que está coberto pelos testes de integração de cada serviço.
`make e2e` roda `go test ./e2e/...` (`godog`) apontando para `E2E_BASE_URL` (default `http://localhost:8080`).

### T5 — CI: job `e2e` (em `.github/workflows/ci.yml`)
Em PR: `make clone` (branch `develop` dos serviços), `docker compose up -d --build`, esperar `/ready` do gateway (até 3 min), `make e2e`, em falha `docker compose logs` e `docker compose down -v` (sempre). Nome do job: `e2e`. (Só passa depois que os PRs das ondas 2 e 3 estiverem em `develop`.)

### T6 — Docs
README do gateway/ecossistema: diagrama, como subir (`make up-all`), tabela de rotas, e2e, limites (token do Lab, etc.). Atualizar `PROJECTS_INDEX.md`, `CLAUDE.md` e `AGENTS.md` **da raiz de `fases-pos`** só se o usuário pedir (fora deste repo — **não tocar**).

### Aceite F10b
- [ ] `docker compose config -q` ok (com os services clonados ou com `COMPOSE_...` ajustado — documentar); `go vet ./e2e/...` e `gofmt -l .` ok; `shellcheck`-equivalente (`sh -n`) nos scripts
- [ ] `local/init-aws.sh` cria **exatamente** os recursos de `resources.md` (contar: 4 tópicos, 10 filas incluindo DLQs, 8 assinaturas, 1 tabela, 1 segredo da lambda)
- [ ] Branch local `feat/gateway` com commits temáticos; **nenhum push, nenhum PR**; a execução real do e2e fica registrada como pendente de Docker no README
