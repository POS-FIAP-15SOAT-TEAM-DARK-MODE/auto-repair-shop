# Fichas F1b e F2b — ajustes de contrato nos PRs da onda 2 (necessários ao F6)

Mudanças **aditivas** nos contratos internos (`contracts/internal-rest.md`), no mesmo PR de cada serviço (sem merge, sem force-push, commits Conventional Commits com o trailer, sem apply na AWS).

## F1b — Catalog (`auto-repair-shop-catalog-service`, PR #1)
- `GET /internal/works/{id}`: acrescentar `"description"` (texto; vazio se não houver) ao JSON.
- `POST /internal/reservations` (`201`): acrescentar `"description"` e `"version"` (`supply.version` **após** o decremento) ao JSON de resposta.
- Atualizar testes unitários/integração, o README (contrato interno) e o swagger se houver menção.
- Aceite: `go test ./... --race` ok, cobertura ≥ 80% (gate sem mocks), CI verde (Sonar conforme token), exemplo de resposta nos testes igual ao de `internal-rest.md`.

## F2b — Customer & Identity (`auto-repair-shop-customer-service`, PR #1)
- `GET /internal/customers/{id}`: acrescentar `"email"` (do usuário), `"phone"` e `"company_name"` (string para `COMPANY`; `null` para `INDIVIDUAL`) ao JSON. `name` continua sendo o nome do usuário.
- Atualizar testes (`internal_test.go`, `routing_test.go`), README e o exemplo de resposta.
- Aceite: igual ao F1b.
