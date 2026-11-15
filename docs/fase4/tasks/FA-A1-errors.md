# Ficha FA-A1 — Padronizar e endurecer as respostas de erro (por serviço)

Aplica-se a cada serviço (`customer`, `catalog`, `billing`, `execution`, `history`, `service-order`), **partindo de `origin/develop`**, em branch **local** `feat/fa-error-format` — **sem push e sem PR**. Leia `SPEC.md` (tarefa A1), `_baseline.md` (B9) e o código atual de `internal/pkg/web/errors.go`, `internal/app/middleware/{auth.go,internal_auth.go}` e `internal/app/routing/routing.go`.

## Formato padrão (não mude o contrato existente, só complete)
`{"code":<http>,"error":"<mensagem>"}`; e `{"code":400,"errors":["...","..."]}` (lista) **somente** para validação de entrada (JSON inválido, campos). Nada de corpo sem `code`.

## T1 — Não vazar a causa raiz (`internal/pkg/web/errors.go`)
Hoje `buildErrorMessages` devolve `root.Error()` (a causa mais profunda) em 409, 422 e 500, o que pode expor texto do driver Postgres (nomes de constraint, valores). Mude assim:
- 500 (e qualquer status sem erro de domínio reconhecido): corpo `{"code":500,"error":"internal server error"}`.
- 409/422 vindos de **infraestrutura** (`app.ErrDataConflict`, `app.ErrInfraConflict`, `app.ErrDataViolation` ou equivalentes): mensagens genéricas `conflict` e `invalid data`.
- 409/422/404/400 vindos de **erros de domínio** conhecidos (os sentinelas listados em `isConflictError`, `isUnprocessableEntityError`, `isNotFoundError`, listas de validação — ex.: `payment.ErrInvalidTransition`, `ErrSupplyVersionConflict`, `ErrServiceOrder*`): **mantêm** a mensagem do próprio erro de domínio (`err.Error()` do sentinela, não a causa raiz de um wrap de infra).
- O erro **completo** (com causa raiz) deve ir para o log: garanta que o caminho que chama `web.Error` (os adaptadores `GinHandler`/`GinOnlyErrorHandler` em `internal/app/routing`) registre `logger.Of(ctx).Error(err)` (ou equivalente) para status ≥ 500 e para 409/422 de infraestrutura, com `request_id` no contexto.
Testes unitários (tabela): para cada status, um erro de infraestrutura com texto sensível (ex.: `pq: duplicate key value violates unique constraint "x" (value)`) **não** pode aparecer no corpo; um erro de domínio mantém a mensagem.

## T2 — 401/403 com `code`
No `auth.go`: troque `gin.H{"error": "..."}` por `gin.H{"code": <status>, "error": "..."}` em **todos** os retornos (sem header/esquema: 401 `Unauthorized`; token inválido: 401 `Unauthorized`; expirado: 401 `Token expired`; papel insuficiente: 403 `Forbidden`). Idem no `InternalAuth` (401). Ajuste os testes do middleware para assertar o JSON (`code` e `error`), não só o status.

## T3 — Validação sem mapeamento
Liste todos os `Err*` exportados nos pacotes de domínio do serviço (`grep -rn "= errors.New" internal/*/domain internal/*/*.go`) cujo nome indique entrada inválida (`ErrInvalid*`, `Err*Empty*`, `Err*Required`, `Err*Shorter*`, `Err*Invalid*Value`, `Err*LessThen*`) e **não** apareçam em nenhuma lista do `web/errors.go`. Inclua-os na lista de validação (400, mesma função que já trata os demais erros de validação do serviço). Teste: um caso por erro adicionado (retorna 400, não 500). Registre no PR/relatório quais foram adicionados (no Catalog espera-se `ErrInvalidSupplyName`, `ErrInvalidSupplyDescription`, `ErrInvalidSupplyUnitPrice`, `ErrInvalidSupplyStockQuantity`, `ErrInvalidSupplyVersion`).

## T4 — Fechamento
`gofmt -l .` vazio, `go vet ./...` (e com `-tags=integration`), `LOG_LEVEL=PANIC go test ./... --race` (suíte COMPLETA, não só o pacote alterado), cobertura total ≥ 80% (sem mocks). Atualize testes de integração/README/swagger que citem o corpo antigo. Commits pequenos em Conventional Commits (`fix(web): ...`, `fix(auth): ...`, `test: ...`) com o trailer `Co-Authored-By: Claude Haiku 5.5 <noreply@anthropic.com>`. **Não faça push, não abra PR, não use `docker`.**

## Aceite
- [ ] Nenhum corpo de erro 5xx/409/422 contém texto de causa raiz de infraestrutura (teste prova)
- [ ] 401/403/InternalAuth respondem `{"code":N,"error":"..."}` (teste prova)
- [ ] Erros de validação listados em T3 respondem 400
- [ ] Suíte completa verde, cobertura ≥ 80%; branch local `feat/fa-error-format` com commits temáticos; sem push

## Ficha FA-A3 (somente `customer-service`, depois da A1) — checagem de exclusão
Em branch `feat/fa-customer-delete-check` criada a partir de `feat/fa-error-format`. `DELETE /v1/customers/:id` e `DELETE /v1/vehicles/:id` (hoje sem a checagem de OS; há `TODO(F6)` no código) devem chamar o Service Order antes de excluir: `GET {SERVICE_ORDER_URL}/internal/service-orders/exists?customer_id=<id>` (ou `?vehicle_id=<id>`) com o header `X-Internal-Token: $INTERNAL_TOKEN`, timeout 3 s, sem retry. `{"exists":true}` ⇒ `409` com mensagem de domínio (reaproveitar o erro 409 existente "customer has service orders"); `{"exists":false}` ⇒ segue; indisponível/erro/5xx ⇒ `502` genérico (`upstream unavailable`) e **não exclui**. Cliente HTTP em `internal/clients/serviceorder` (interface + `net/http`, testes com `httptest`); nova variável `SERVICE_ORDER_URL` (default `http://service-order-service:8086`) em `.env.example`, `docker-compose.yml` e no ConfigMap dos overlays; remova os `TODO(F6)`. Testes: exists true/false, upstream fora, token enviado; cobertura ≥ 80%. Contrato: `contracts/internal-rest.md` (seção Service Order).
