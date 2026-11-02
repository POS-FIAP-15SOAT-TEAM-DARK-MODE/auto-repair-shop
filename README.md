# Auto Repair Shop: gateway e ambiente de execução

Este repositório tem o **gateway** (ponto de entrada público do sistema, um proxy reverso em Go) e o **ambiente local** que sobe o sistema inteiro com Docker: o gateway, os seis microsserviços, a Lambda de login do cliente, um PostgreSQL e um MiniStack (emulador de SNS, SQS, DynamoDB e Secrets Manager). Também contém o e2e BDD que percorre uma ordem de serviço do começo ao fim.

Este guia é para quem nunca viu o projeto: siga a ordem das seções. Decisões de arquitetura estão em [`docs/adr/`](docs/adr) e [`docs/decisions.md`](docs/decisions.md); a especificação da Fase 4 está em [`docs/fase4/SPEC.md`](docs/fase4/SPEC.md).

## 1. Pré-requisitos

| Ferramenta | Para quê | Como checar |
|---|---|---|
| Docker com Compose v2 | subir o ambiente | `docker compose version` |
| Docker em execução | o daemon precisa estar de pé | `docker info` |
| Memória do Docker de pelo menos 4 GB | os 14 contêineres e as builds | Docker Desktop: *Settings > Resources* |
| `git`, `make`, `curl` | clonar, comandos, checagens | `git --version`, `make --version`, `curl --version` |
| Go 1.27 (opcional) | `make test`, `make run` e `make e2e` no host | `go version` |

Portas usadas no host: `8080` (gateway), `8081`-`8086` (serviços), `9001` (adaptador da Lambda), `4566` (MiniStack) e `5432` (PostgreSQL). Se alguma estiver ocupada, pare o processo que a usa antes de subir o ambiente.

O `make start` checa tudo isso e explica o que falta. Ele não inicia o Docker sozinho: se o daemon estiver parado, abra o Docker Desktop e rode de novo.

## 2. Subir o ambiente

```bash
git clone https://github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop.git
cd auto-repair-shop
make start
```

O `make start` (ou `sh scripts/start.sh`) faz, nesta ordem:

1. checa as ferramentas, o daemon do Docker e a memória (aviso abaixo de 4 GB, sem bloquear);
2. clona os repositórios de serviço que faltam em `services/` (os existentes não são alterados);
3. faz `docker pull` das imagens base, para não usar uma imagem antiga do cache local;
4. constrói as imagens **uma por vez**;
5. sobe os contêineres (`docker compose up -d`);
6. espera o `/ping` de cada serviço e o `/ready` do gateway, por até 3 minutos;
7. imprime o resumo com URLs, usuários de seed e comandos úteis.

É idempotente: rodar de novo não quebra nada e reaproveita o que já existe.

Para usar clones locais (por exemplo, branches que ainda não foram para `develop`), aponte `SERVICES_DIR` para o diretório que os contém:

```bash
SERVICES_DIR=/caminho/para/os/clones make start
```

Os clones que faltarem são criados nesse diretório. Caminhos relativos são relativos à raiz deste repositório.

### Os comandos do dia a dia

| Comando | O que faz |
|---|---|
| `make start` | sobe o ambiente (ver acima) |
| `make status` | estado de cada contêiner, `/ping` de cada serviço e `/ready` do gateway; sai com código diferente de zero se algo falhar |
| `make smoke` | um fluxo de ordem de serviço em cerca de um minuto, sem o godog (seção 4) |
| `make e2e` | a suíte BDD completa contra o gateway (seção 6) |
| `make logs` | segue os logs de todos os contêineres |
| `make down-all` | para o ambiente e remove os volumes (os dados somem) |
| `make reset` | remove os volumes e sobe de novo do zero |

## 3. Endpoints e mapa do ambiente

### Portas

| Serviço | Porta no host | Papel |
|---|---|---|
| gateway | `8080` | único ponto público; roteia para os serviços |
| customer-service | `8081` | usuários, papéis, clientes, veículos, login de atendente e mecânico |
| catalog-service | `8082` | serviços (works), peças e estoque |
| billing-service | `8083` | orçamento e pagamento (`PAYMENT_MODE=manual`) |
| execution-service | `8084` | fila de execução e progresso dos works |
| history-service | `8085` | linha do tempo da OS e relatório de tempo médio (DynamoDB) |
| service-order-service | `8086` | ordem de serviço e orquestrador da saga |
| lambda-adapter | `9001` | adaptador HTTP da Lambda de login (interno ao gateway) |
| PostgreSQL | `5432` | um banco por serviço: `customer`, `catalog`, `billing`, `execution`, `service_order` |
| MiniStack | `4566` | SNS, SQS, DynamoDB e Secrets Manager |

Os serviços só são acessados diretamente para depuração. Os clientes falam **sempre com o gateway**.

### Endpoints do gateway

| Endpoint | Resultado |
|---|---|
| `GET http://localhost:8080/ping` | `200 {"message":"pong"}` |
| `GET http://localhost:8080/ready` | `200 {"message":"ready"}`, ou `503` com os nomes dos upstreams que falham |
| `GET http://localhost:8080/swagger/index.html` | página com a documentação de cada serviço |
| `GET http://localhost:8080/docs/<serviço>/swagger.yaml` | OpenAPI de um serviço (`customer`, `catalog`, `billing`, `execution`, `history`, `service_order`) |
| `GET http://localhost:8080/metrics` | métricas Prometheus |

Rotas que o gateway não conhece, e as de `/internal/*`, respondem `404 {"code":404,"error":"route not found"}`. Os serviços não são alcançados por esse caminho.

Erros que o próprio gateway produz seguem o formato `{"code":N,"error":"mensagem"}`. Quando um upstream não responde, a mensagem é fixa (`upstream timeout` com `504`, `upstream unavailable` com `502`) e a causa vai só para o log do servidor. A resposta de um serviço (inclusive 4xx e 5xx) passa sem alteração pelo gateway.

### Tabela de rotas

A tabela está em [`config/routes.yaml`](config/routes.yaml), embutida no binário. As rotas são testadas em ordem, e vence a primeira cujo método e caminho casam.

| Método | Caminho | Upstream |
|---|---|---|
| qualquer | `/internal/*` | bloqueada (404) |
| PUT | `/v1/service-order/:id/work/:workId/next`, `.../cancel`, `/v1/service-order/:id/finish` | `execution` |
| GET | `/v1/execution/*` | `execution` |
| GET | `/v1/service-order/:id/history`, `/v1/reports/*` | `history` |
| qualquer | `/v1/auth/*`, `/v1/users/*`, `/v1/customers*`, `/v1/vehicles*` | `customer` |
| qualquer | `/v1/works*`, `/v1/supplies*` | `catalog` |
| qualquer | `/v1/budgets*`, `/v1/payments*` | `billing` |
| qualquer | `/v1/service-order*` (o restante) | `service_order` |
| POST | `/auth/customer-login` | `lambda` (ativa só com `LAMBDA_UPSTREAM`; o ambiente local a liga) |

### Usuários de seed

Os seeds rodam porque o ambiente local define `SEED_ENABLED=true`. Os IDs são fixos (o e2e depende deles).

| Quem | Como entrar | Papel |
|---|---|---|
| Atendente | e-mail `attendant@autorepairshop.com`, senha `attendant123` (`POST /v1/auth/login`) | recebe a OS, aprova pagamentos |
| Mecânico | e-mail `mechanic@autorepairshop.com`, senha `mechanic123` (`POST /v1/auth/login`) | diagnóstico, orçamento, execução |
| Cliente A (João Silva) | `POST /auth/customer-login` com o CPF do seed | aceita ou recusa o orçamento |

Estas senhas e o CPF são **valores fixos de desenvolvimento**, não segredos. Não os reutilize fora deste ambiente. O CPF do cliente A está em `e2e/steps_test.go` (`customerACPF`).

IDs fixos úteis (`docs/fase4/contracts/resources.md`):

| Entidade | ID |
|---|---|
| cliente A | `44444444-4444-4444-8444-000000000002` |
| veículo do cliente A | `44444444-4444-4444-8444-000000000003` |
| work "Troca de Óleo" | `11111111-1111-4111-8111-000000000001` |
| supply "Filtro de Óleo" (estoque 50) | `22222222-2222-4222-8222-000000000001` |

## 4. Fluxo de exemplo com curl

O `make smoke` executa este mesmo fluxo. Para fazer à mão, use o gateway em `http://localhost:8080`. Os exemplos usam `jq` para ler os campos da resposta (sem ele, leia o JSON na saída).

```bash
G=http://localhost:8080

# 1. Atendente e mecânico entram e recebem um token (campo "token").
ATT=$(curl -s -X POST $G/v1/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"attendant@autorepairshop.com","password":"attendant123"}' | jq -r .token)
MEC=$(curl -s -X POST $G/v1/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"mechanic@autorepairshop.com","password":"mechanic123"}' | jq -r .token)

# 2. O atendente abre a OS para o cliente A, o veículo dele e o work "Troca de Óleo" (201, campo "id").
SO=$(curl -s -X POST $G/v1/service-order -H "Authorization: Bearer $ATT" -H 'Content-Type: application/json' \
  -d '{"client":"44444444-4444-4444-8444-000000000002","vehicle":"44444444-4444-4444-8444-000000000003","works":["11111111-1111-4111-8111-000000000001"]}' | jq -r .id)

# 3. Recebimento e diagnóstico (204).
curl -s -o /dev/null -w '%{http_code}\n' -X PUT $G/v1/service-order/$SO/received -H "Authorization: Bearer $ATT"
curl -s -o /dev/null -w '%{http_code}\n' -X PUT $G/v1/service-order/$SO/start-diagnosis -H "Authorization: Bearer $MEC"

# 4. O mecânico envia o orçamento (204). A saga gera o orçamento em segundo plano.
curl -s -o /dev/null -w '%{http_code}\n' -X PUT $G/v1/service-order/$SO/send -H "Authorization: Bearer $MEC"

# 5. Consulta pública do status: quando chega a AWAITING_APPROVAL, o cliente decide.
curl -s $G/v1/service-order/$SO/status          # {"status":"AWAITING_APPROVAL",...} quando estiver pronto

# 6. O cliente entra pelo CPF do seed (veja e2e/steps_test.go) e aceita o orçamento (204).
CPF=...                                          # o CPF do cliente A, ver e2e/steps_test.go
CUS=$(curl -s -X POST $G/auth/customer-login -H 'Content-Type: application/json' -d "{\"cpf\":\"$CPF\"}" | jq -r .token)
curl -s -o /dev/null -w '%{http_code}\n' -X PUT $G/v1/service-order/$SO/accept -H "Authorization: Bearer $CUS"

# 7. O atendente confirma o pagamento pendente (204). Primeiro descubra o id do pagamento.
PAY=$(curl -s $G/v1/payments/$SO -H "Authorization: Bearer $ATT" | jq -r .id)
curl -s -o /dev/null -w '%{http_code}\n' -X POST $G/v1/payments/$PAY/confirm -H "Authorization: Bearer $ATT"

# 8. A execução entra na fila do mecânico (a saga é assíncrona: pode levar alguns segundos).
curl -s $G/v1/execution/queue -H "Authorization: Bearer $MEC" | jq ".items[] | select(.serviceOrderId==\"$SO\")"
```

Sem token, uma rota protegida responde `401` do serviço (não do gateway). O formato alvo é `{"code":401,"error":"Unauthorized"}`; enquanto a mudança de erros não chega aos repositórios dos serviços, o corpo ainda é `{"error":"Unauthorized"}` (seção 7). O histórico da OS está em `GET /v1/service-order/$SO/history`, com o mesmo token do atendente.

## 5. Parar, limpar e reiniciar

- `make down-all`: para os contêineres e remove os volumes. O próximo `make start` recria o banco e o seed.
- `make reset`: a mesma limpeza, seguida de um `make start`.
- Reconstruir um serviço depois de mudar o código: `docker compose build <serviço>` e `docker compose up -d <serviço>` (os nomes estão em `docker compose config --services`).
- Parar sem apagar dados: `docker compose stop`; retomar com `docker compose start`.

### Inspecionar mensageria e banco

- Filas e DLQs no MiniStack (credenciais de desenvolvimento `test`/`test`):

  ```bash
  AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_REGION=us-east-1 \
    aws --endpoint-url http://localhost:4566 sqs list-queues
  ```

  As filas têm o prefixo `auto-repair-shop-local-`; cada uma tem uma DLQ com sufixo `-dlq`.
- Banco de um serviço (o outbox fica nas tabelas do próprio banco):

  ```bash
  docker compose exec postgres psql -U postgres -d billing -c '\dt'
  ```
- Logs de uma saga: `docker compose logs <serviço> | grep <saga_id>`, nos serviços que registram o `saga_id`.

## 6. Testes end-to-end

`e2e/` é um módulo Go separado (godog, o Cucumber de Go). Os arquivos `e2e/features/*.feature` descrevem os cenários em português ou inglês, e os passos falam **só com o gateway**, pelas rotas de `config/routes.yaml`.

```bash
make e2e           # contra http://localhost:8080 (E2E_BASE_URL)
```

| Feature | O que prova |
|---|---|
| `happy_path` | OS recebida, diagnosticada, orçada, aceita, paga, executada, finalizada e entregue; histórico e relatório de tempo médio |
| `payment_rejected` | o atendente recusa o pagamento: OS `REJECTED` e estoque devolvido |
| `execution_failed` | todos os works cancelados: OS `CANCELLED`, pagamento `REFUNDED` e estoque devolvido |
| `budget_rejected` | o cliente recusa o orçamento: OS `REJECTED` e estoque devolvido |
| `customer_login` | CPF conhecido recebe token aceito por um serviço; CPF desconhecido recebe `404` |
| `security` | `/internal/*` responde `404` pelo gateway; rota protegida sem token responde `401` com `code` |
| `idempotency` | `accept` e confirmação de pagamento repetidos respondem `409` e não duplicam a tarefa de execução |

**Como ler o resultado.** O `go test -v` imprime cada cenário com os passos. A linha final `ok` significa que todos passaram; `FAIL` indica o passo que falhou, com o status HTTP esperado e o recebido. Esperas assíncronas (saga) têm limite de 20 s por passo, então uma falha do tipo "status X, want Y" costuma ser um serviço parado: rode `make status` e `docker compose logs <serviço>`.

Para o e2e, o ambiente precisa estar pronto: `make start` já espera por isso. Em CI, o job `e2e` roda a mesma sequência.

## 7. Solução de problemas

- **`the Docker daemon is not running`**: abra o Docker Desktop (ou o serviço `docker`) e rode `make start` de novo. O script não inicia o Docker.
- **Aviso de memória, ou `ResourceExhausted` durante o build**: a VM do Docker ficou sem memória. Aumente a memória em *Settings > Resources* (4 GB ou mais). O `make start` já constrói uma imagem por vez; se o problema persistir, aumente a memória mais.
- **`go.mod requires go >= 1.27.3 (running go 1.27.x)`** durante o build: o cache local tem uma imagem `golang` antiga. O `make start` faz `docker pull` das imagens base; se a falha continuar, rode `docker pull golang:1.27-alpine` e depois `docker compose build --no-cache <serviço>`.
- **Um serviço fica `unhealthy` ou `starting` para sempre**: o healthcheck do MiniStack usa `python3` dentro do contêiner (a imagem não tem `wget` nem `curl`). Veja `docker compose logs ministack` e `make status`. Se o MiniStack não ficar saudável, `make reset` costuma resolver.
- **`bind: address already in use` na porta 5432 (ou outra da lista)**: há um PostgreSQL local rodando. Pare-o ou troque o mapeamento de porta no `docker-compose.yml`.
- **Um serviço não sobe ou sai com erro**: `docker compose logs <serviço>`. As migrações rodam antes dos serviços; se `migrate-<serviço>` falhou, o log dele mostra o motivo.
- **Serviço clonado no branch errado**: `SERVICES_BRANCH=<branch> sh scripts/clone-services.sh` atualiza os clones existentes (`git pull --ff-only`). Para usar clones locais, veja `SERVICES_DIR` na seção 2.
- **`body code is 401` falha no e2e**: o 401 vem do serviço, não do gateway. Enquanto os serviços clonados responderem `{"error":"Unauthorized"}` (sem `code`), o cenário `security` falha nesse único passo. A correção é a tarefa de erros nos repositórios dos serviços; o gateway já repassa o corpo sem alterar.
- **Clone falha**: a clonagem é por HTTPS sem credenciais; os repositórios de serviço precisam ser públicos.

## 8. Referência do gateway

### Configuração

| Variável | Padrão | Significado |
|---|---|---|
| `PORT` | `8080` | porta de escuta |
| `LOG_LEVEL` | `INFO` | `DEBUG`, `INFO` ou `PANIC` |
| `UPSTREAM_TIMEOUT` | `30s` | espera pelos cabeçalhos de resposta do upstream |
| `CUSTOMER_UPSTREAM` … `SERVICE_ORDER_UPSTREAM` | `http://<serviço>:<porta>` da rede do ambiente | endereços base dos serviços |
| `LAMBDA_UPSTREAM` | vazio | endereço do adaptador da Lambda; vazio desliga `POST /auth/customer-login` |

Valor inválido impede o gateway de subir, e a mensagem de erro nunca mostra a URL configurada.

### Comportamento do proxy

- Método, caminho, query, corpo e cabeçalhos são repassados, inclusive `Authorization`. São acrescentados `X-Forwarded-For`, `-Host` e `-Proto`.
- `X-Request-Id` válido (até 128 caracteres de `A-Z a-z 0-9 - _ . :`) é mantido; senão é gerado. Ele volta na resposta, vai ao upstream e aparece em toda linha de log.
- Os logs são JSON, uma linha por requisição, com `request_id`, `method`, `path`, `upstream`, `status` e `duration_ms`. A query string não é registrada.

### Verificações de qualidade

- `make test`: testes unitários com o detector de corrida; `make coverage`: cobertura e relatório HTML; `make check`: formatação, `go vet` e testes.
- A cobertura total é de pelo menos 80% (mocks gerados excluídos) e o matcher de rotas de pelo menos 90%. Os dois limites são verificados pelo job `coverage` do CI.

## 9. Deploy

- `k8s/manifests/base`: Deployment, Service `ClusterIP`, HPA e o ConfigMap com os endereços dos upstreams. O gateway não tem segredos, banco nem migrações.
- `k8s/manifests/overlays/lab`: Service `LoadBalancer` (AWS Academy Learner Lab, sem controlador de Ingress).
- `k8s/manifests/overlays/stg` e `prd`: Ingress (ALB), limites de HPA e nível de log; `components/observability` adiciona o `ServiceMonitor`.
- `.github/workflows/docker.yml` constrói e publica a imagem, aplica o overlay e espera o rollout. No lab, publica o host do load balancer no SSM (`/auto-repair-shop/<env>/app-backend-host`), que o API Gateway da AWS lê.

## 10. Estrutura

```
cmd/service                  ponto de entrada
config                       tabela de rotas embutida (routes.yaml)
internal/app                 servidor HTTP, bootstrap, middlewares do Gin (log, recuperação de pânico, métricas)
internal/gateway/config      carregamento e validação de configuração e rotas
internal/gateway/match       casamento de método e caminho (só biblioteca padrão)
internal/gateway/proxy       httputil.ReverseProxy por upstream, timeouts, mapeamento 502/504
internal/gateway/handlers    endpoints do gateway, proxy do NoRoute, roteador
internal/pkg                 env, logger, métricas, ciclo de vida, web (envelope de erro)
docs                         swagger.yaml do gateway e página de índice (embutidos)
k8s/manifests                base e overlays Kustomize
e2e                          suíte BDD (godog), módulo Go próprio
local                        ambiente: postgres-init, init-aws.sh, imagem e adaptador da Lambda
scripts                      start.sh, clone-services.sh, wait-ready.sh
docker-compose.yml           ambiente local
```

As regras que os agentes de IA seguem estão em [`.ai/`](.ai/README.md); este README não as repete.
