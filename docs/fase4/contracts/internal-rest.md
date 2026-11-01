# Contrato — REST interno (`/internal/*`)

Só chamado entre serviços, dentro do cluster. **Nunca roteado pelo gateway** (gateway devolve 404). Autenticação: header `X-Internal-Token: <INTERNAL_TOKEN>`; ausente/errado → `401`. Propagar `X-Request-Id`. Erros no formato atual da API: `{"code": <http>, "errors": ["mensagem"]}`.

## Customer (`customer-service:8081`) — chamado pelo Service Order

### `GET /internal/customers/{id}`
`200`
```json
{"id":"7c9e6679-7425-40de-944b-e07fc1f90ae7","user_id":"3b123072-5ad5-4df8-8d3b-3c2770f11b2a","type":"INDIVIDUAL","name":"João Silva","email":"joao.silva@email.com","document":"52998224725","company_name":null,"phone":"11999990001"}
```
`type`: `INDIVIDUAL` | `COMPANY`. `name` = nome do usuário (sempre); `company_name` só para `COMPANY` (senão `null`). `email` e `phone` existem porque a OS os devolve nas respostas de listagem/detalhe e usa o e-mail na notificação. `404` se não existir.

### `GET /internal/vehicles/{id}`
`200`
```json
{"id":"1f0c1d52-8a10-4c45-9d3c-6a6f0e9b3a10","customer_id":"7c9e6679-7425-40de-944b-e07fc1f90ae7","license_plate":"ABC1D23","brand":"Fiat","model":"Uno","year":2018}
```
`404` se não existir.

## Catalog (`catalog-service:8082`) — chamado pelo Service Order

### `GET /internal/works/{id}`
`200`
```json
{"id":"9b2d6a0e-3c0f-4a0b-8d57-2a5f6a1c9e11","name":"Troca de óleo","description":"Troca de óleo do motor com filtro incluído","unit_price":"150.00","active":true}
```
`404` se não existir. O chamador rejeita work com `active=false`.

### `POST /internal/reservations`
Request `{"service_order_id":"01JD8Z4R3Q8M2V7K5N9XWABCDE","supply_id":"…","quantity":2}`
`201` `{"reservation_id":"…","supply_id":"…","name":"Filtro","description":"Filtro de óleo","unit_price":"35.50","version":3}` (`version` = `supply.version` após o decremento; a OS o devolve como no monólito)
`409` estoque insuficiente · `404` supply inexistente · `409` já existe reserva da mesma peça para a OS · `422` `quantity <= 0`.
Efeito: decremento atômico `UPDATE supply SET stock_quantity = stock_quantity - $q … WHERE id=$1 AND stock_quantity >= $q`.

### `DELETE /internal/reservations/{id}`
`204` (devolve o estoque na mesma transação; idempotente: reserva já liberada também responde `204`). `404` se o id nunca existiu.

## Service Order (`service-order-service:8086`) — chamado pelo Customer (integração na onda 4)

### `GET /internal/service-orders/exists?customer_id={id}` (ou `?vehicle_id={id}`)
`200 {"exists": true}` ou `{"exists": false}`; exatamente um dos dois parâmetros (`422` caso contrário). Usado para restaurar a regra "não excluir cliente/veículo com OS" que o monólito aplicava com `SELECT` direto.

## Regras comuns
- Timeout de cliente: 3 s; sem retry automático em `POST` (quem chama decide); `GET` pode repetir 1×.
- Falha de rede/5xx no chamador → erro ao usuário final (`502`), **sem** estado parcial na OS.
- Snapshots: o Service Order copia `name`/`unit_price`/`document` na resposta e **não** volta a consultar.
