# Contrato — REST interno (`/internal/*`)

Só chamado entre serviços, dentro do cluster. **Nunca roteado pelo gateway** (gateway devolve 404). Autenticação: header `X-Internal-Token: <INTERNAL_TOKEN>`; ausente/errado → `401`. Propagar `X-Request-Id`. Erros no formato atual da API: `{"code": <http>, "errors": ["mensagem"]}`.

## Customer (`customer-service:8081`) — chamado pelo Service Order

### `GET /internal/customers/{id}`
`200`
```json
{"id":"7c9e6679-7425-40de-944b-e07fc1f90ae7","user_id":"3b123072-5ad5-4df8-8d3b-3c2770f11b2a","type":"INDIVIDUAL","name":"João Silva","document":"52998224725"}
```
`type`: `INDIVIDUAL` | `COMPANY` (para `COMPANY`, `name` = razão social). `404` se não existir.

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
{"id":"9b2d6a0e-3c0f-4a0b-8d57-2a5f6a1c9e11","name":"Troca de óleo","unit_price":"150.00","active":true}
```
`404` se não existir. O chamador rejeita work com `active=false`.

### `POST /internal/reservations`
Request `{"service_order_id":"01JD8Z4R3Q8M2V7K5N9XWABCDE","supply_id":"…","quantity":2}`
`201` `{"reservation_id":"…","supply_id":"…","name":"Filtro","unit_price":"35.50"}`
`409` estoque insuficiente · `404` supply inexistente · `409` já existe reserva da mesma peça para a OS · `422` `quantity <= 0`.
Efeito: decremento atômico `UPDATE supply SET stock_quantity = stock_quantity - $q … WHERE id=$1 AND stock_quantity >= $q`.

### `DELETE /internal/reservations/{id}`
`204` (devolve o estoque na mesma transação; idempotente: reserva já liberada também responde `204`). `404` se o id nunca existiu.

## Regras comuns
- Timeout de cliente: 3 s; sem retry automático em `POST` (quem chama decide); `GET` pode repetir 1×.
- Falha de rede/5xx no chamador → erro ao usuário final (`502`), **sem** estado parcial na OS.
- Snapshots: o Service Order copia `name`/`unit_price`/`document` na resposta e **não** volta a consultar.
