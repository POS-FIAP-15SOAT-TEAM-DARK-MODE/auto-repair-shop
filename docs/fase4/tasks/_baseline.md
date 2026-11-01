# Receita comum — baseline de um repo de serviço (referenciada por F1–F6)

> Executar dentro do clone `fases-pos/<repo>`, em branch `feat/bootstrap-<svc>` criada a partir de `develop`. Cada bloco termina com um **comando de verificação**. Se a verificação falhar, parar e reportar (não improvisar).

## B1 — Extrair a baseline congelada (não mexe no monólito)
```bash
SRC=/private/tmp/claude-501/-Users-caetanofreitas-projects-study-fases-pos/384e6c41-4feb-4520-af12-47f6c534e3ee/scratchpad/monolith-v1
mkdir -p "$SRC" && git -C /Users/caetanofreitas/projects/study/fases-pos/auto-repair-shop archive monolith-v1 | tar -x -C "$SRC"
test -f "$SRC/go.mod" && echo OK
```

## B2 — O que copiar de `$SRC` para o repo novo (e o que NÃO copiar)

| Copiar | Ajuste |
|---|---|
| `Dockerfile`, `.gitignore`, `.pre-commit-config.yaml` | só o nome do binário/porta (`EXPOSE ${PORT}`) |
| `Makefile` | remover alvos `k8s-*`, `sonarqube-run` (projectKey) e `migrate-*` apontando para `autorepairshop`; DB = `<db>` do serviço; adicionar `up`/`down` (`docker compose up --build` / `down -v`) |
| `.env.example` | variáveis do serviço (resources.md); sem `SONARQUBE_*` |
| `go.mod`, `go.sum` | trocar o module path (B3); `go mod tidy` no fim |
| `cmd/service/main.go` | só o import do bootstrap |
| `internal/app/{server.go,errors.go,responses.go}` | — |
| `internal/app/db.go` | **só** se o serviço usa Postgres; default `POSTGRES_DB` = `<db>`; lê `POSTGRES_SSLMODE` |
| `internal/app/middleware/*` | `auth.go`: ver cada ficha (origem do tipo `Role`) |
| `internal/app/{bootstrap,container,routing}` | **reescrever** só com os domínios do serviço |
| `internal/pkg/{app,env,id,json,logger,web}` | cópia literal |
| `internal/pkg/{db,uow}` | só se Postgres |
| `internal/pkg/auth` | `jwt.go` literal (+ `roles.go` onde a ficha mandar) |
| `internal/pkg/metrics/metrics.go` | **manter só** `HTTPRequestDuration`; remover contadores/collector de service order |
| `docs/embed.go`, `docs/swagger.yaml` | swagger: manter apenas os paths do serviço |
| `docs/adr/0003-git-branching-strategy.md` | copiar como `docs/adr/0001-git-branching-strategy.md` (ajustar número/título) |
| `.github/workflows/{ci.yml,docker.yml}`, `.github/{dependabot.yml,pull_request_template.md,commit_message_template.txt}` | ver B4/B5 |
| `.ai/**`, `CLAUDE.md`, `AGENTS.md`, `.cursor/**`, `.windsurf/**` | trocar descrições de domínio pelas do serviço; `.ai/rules/domain.md` só com as entidades do serviço; remover regras de outros domínios |
| `k8s/manifests/{base,components/observability,overlays/lab,overlays/stg,overlays/prd}` | ver B6 |

**Não copiar:** `k8s/kind`, `k8s/toolbox`, `k8s/manifests/overlays/local`, `overlays/*/ingress.yaml`, `overlays/lab/patch-service-loadbalancer.yaml`, `migrations/` (escrever novas — ddl.md), `docker-compose.yml` (escrever novo — B7), `sonar-project.properties` original (recriar), qualquer pasta `internal/<outro domínio>`, `internal/integration/*` de outros domínios, `docs/{auth-middleware.md,sonarqube.md}`, `README.md` (escrever novo).

## B3 — Renomear
- Module path: `github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop` → `github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/<repo>` em `go.mod` e em **todos** os `*.go`.
  `grep -rl 'POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop"\|POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/' . | xargs sed -i '' 's#POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop\([/"]\)#POS-FIAP-15SOAT-TEAM-DARK-MODE/<repo>\1#g'`
- Verificação: `grep -rn 'auto-repair-shop/internal' --include=*.go . | head` deve vir **vazio** (módulo antigo); `go build ./...` ok.
- `sonar-project.properties`: `sonar.projectKey=POS-FIAP-15SOAT-TEAM-DARK-MODE_<repo>`, `sonar.projectName=<Nome do serviço>`, demais chaves como no original.

## B4 — `.github/workflows/docker.yml` (deploy)
Partir do original e alterar:
- `ECR repo` / imagem: `<repo>` (o ECR é criado no F8).
- `Resolve cluster and RDS endpoint`: igual (usa `vars.EKS_CLUSTER_NAME`, `vars.RDS_HOST`, detecta Lab).
- `Sync migrations ConfigMap`: `kubectl create configmap <svc>-migrations --from-file=migrations …` e `kubectl -n auto-repair-shop delete job <svc>-db-migrate --ignore-not-found` (pular o passo se o serviço não usa Postgres).
- `Provision … secret (lab)`: ler `auto-repair-shop-${ENV_NAME}/<nome-do-secret>` (resources.md) e criar o Secret `<svc>-secrets` com **todas** as chaves do JSON (loop `jq -r 'to_entries[] | "--from-literal=\(.key)=\(.value)"'`); **remover** o `delete ingress`.
- **Novo passo** `Resolve messaging endpoints` (só serviços com mensageria), antes do Deploy: derivar por convenção, sem remote state —
  `ACCOUNT=$(aws sts get-caller-identity --query Account --output text)`;
  `TOPIC_ARN=arn:aws:sns:$REGION:$ACCOUNT:auto-repair-shop-$ENV_NAME-<topico>`;
  `QUEUE_URL=https://sqs.$REGION.amazonaws.com/$ACCOUNT/auto-repair-shop-$ENV_NAME-<fila>`;
  e aplicar no ConfigMap `<svc>-config` (`kubectl patch configmap … --type merge`) junto com `POSTGRES_HOST`.
- `Deploy`/`set image`/`rollout`: `deploy/<svc>`. **Remover** `Publish app URL` (só o gateway publica host).
Verificação: `yq` ou `python -c 'import yaml…'` parseando o YAML; `grep -n 'auto-repair-shop"' .github/workflows/docker.yml` não deve sobrar referência ao deploy antigo `deploy/auto-repair-shop`.

## B5 — `.github/workflows/ci.yml`
Manter os jobs `pre-commit`, `build`, `unit-tests`, `sonarqube` do original e **acrescentar**:
- `integration-tests`: `services:` `postgres:15-alpine` (env `POSTGRES_USER/PASSWORD/DB` do serviço, `--health-cmd pg_isready`) e, se o serviço usa mensageria, `ministack: ministackorg/ministack` (porta 4566); rodar migrations (`migrate` CLI ou `go run`) antes; `go test -tags=integration ./internal/integration/... -v` com `AWS_ENDPOINT_URL=http://localhost:4566`, `AWS_ACCESS_KEY_ID=test`, `AWS_SECRET_ACCESS_KEY=test`, `AWS_REGION=us-east-1`.
- `coverage`: `go test ./... -coverprofile=coverage.out`, **excluindo mocks gerados** (`grep -v '/mocks/' coverage.out > coverage.filtered.out`, mesma exclusão do `sonar.exclusions`) + gate sobre o arquivo filtrado: `COV=$(go tool cover -func=coverage.filtered.out | awk '/^total:/ {gsub("%","",$3); print $3}')` e `awk -v c="$COV" 'BEGIN{exit !(c>=80)}'` (falha se < 80). Publicar o percentual em `$GITHUB_STEP_SUMMARY`.
Nomes de job **exatos**: `pre-commit`, `build`, `unit-tests`, `integration-tests`, `sonarqube`, `coverage` (usados na proteção de branch no S7).

## B6 — `k8s/manifests`
- Pastas: `base/`, `components/observability/`, `overlays/{lab,stg,prd}/`.
- `base/`: remover `namespace.yaml` do `resources`; renomear objetos pelo padrão `<svc>` (resources.md): Deployment/Service/HPA/ServiceAccount `<svc>`, ConfigMap `<svc>-config`, Secret ref `<svc>-secrets`, Job `<svc>-db-migrate`, ConfigMap de migrations `<svc>-migrations`, Role/RoleBinding `<svc>-wait-migrations`; `containerPort`/probes `/ping` na porta do serviço; imagem `<repo>`; Service `type: ClusterIP`. Serviços sem Postgres: remover `migrate-job.yaml`, `rbac.yaml` e o initContainer `wait-for-migrations`.
- `overlays/lab`: manter patch do ConfigMap (`POSTGRES_HOST`, sslmode `require`), **sem** patch de LoadBalancer.
- `overlays/stg` e `prd`: manter patch do ConfigMap, HPA e `secrets.yaml` (ExternalSecret, só vale com IRSA — manter por paridade) e **sem** `ingress.yaml`.
- `components/observability/servicemonitor.yaml`: seletor `app: <svc>`, porta `http`.
Verificação: `kubectl kustomize k8s/manifests/overlays/lab >/dev/null && kubectl kustomize k8s/manifests/overlays/stg >/dev/null && kubectl kustomize k8s/manifests/overlays/prd >/dev/null` (se não houver `kubectl`, `kustomize build`), e `grep -rn "app-config\|app-secrets\|name: auto-repair-shop$" k8s/` **sem** resultados.

## B7 — `docker-compose.yml` (independente)
Serviços (nomes fixos): `app` (build `.`, porta do serviço, `env_file: .env`, `depends_on` `migrate` concluído e, se mensageria, `aws-init` concluído), `db` (`postgres:15-alpine`, database do serviço, healthcheck), `migrate` (`migrate/migrate`, `-path=/migrations`), e — só com mensageria — `ministack` (`ministackorg/ministack`, 4566, healthcheck HTTP) e `aws-init` (`amazon/aws-cli`, `entrypoint: ["/bin/sh","/init/init-aws.sh"]`, volume `./local:/init`, `AWS_ENDPOINT_URL=http://ministack:4566`, credenciais `test`). Rede `backend`. Dependências REST de outros serviços: variáveis `*_URL` com default apontando para um **stub** (`wiremock/wiremock` com mappings em `local/stubs/`) sob `profiles: ["stubs"]`.
`local/init-aws.sh`: idempotente (`|| true` em recursos já existentes), cria **apenas** os tópicos/filas deste serviço com o prefixo `auto-repair-shop-local-`, DLQ + `RedrivePolicy` (`maxReceiveCount=5`), *queue policy* para o tópico, assinatura com `RawMessageDelivery=true` e `FilterPolicy` quando for `saga-commands` (nomes e filtros em `messaging.md`/`resources.md`).
Verificação: `docker compose config -q` e `docker compose up -d --build` seguido de `curl -fsS http://localhost:<porta>/ping`; depois `docker compose down -v`.

## B8 — Entrega
1. `git add -A && git commit` em Conventional Commits (`feat(bootstrap): …`, `test: …`, `docs: …`; commits pequenos por tema).
2. `git push -u origin feat/bootstrap-<svc>` e `gh pr create --base develop --title "feat: bootstrap <svc>" --body "<resumo + evidência de cobertura>"`.
3. Aguardar a CI: `gh pr checks --watch`. Falha de `sonarqube` por falta de `SONAR_TOKEN` **não** é falha de código — reportar.
4. **Não** fazer merge sem a autorização definida pelo usuário (ver F0/decisões). Reportar URL do PR, resultado de cada job e % de cobertura.
