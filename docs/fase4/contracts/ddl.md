# Contrato — DDL por serviço (migrations `golang-migrate`)

Convenções: ids `VARCHAR(36)` (UUID) salvo OS (`VARCHAR(26)`, ULID); `TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP`; dinheiro `NUMERIC(10,2)`; **sem FK entre databases**. Cada migration tem `.up.sql` e `.down.sql` (down = `DROP TABLE … CASCADE` na ordem inversa).

## Tabelas comuns (todo serviço Postgres que consome/publica mensagens)

```sql
-- dedupe de consumidor
CREATE TABLE processed_message (
    message_id   UUID PRIMARY KEY,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- outbox (só quem publica). `body` é o envelope JSON completo (messaging.md)
CREATE TABLE outbox (
    id           BIGSERIAL PRIMARY KEY,
    message_id   UUID NOT NULL UNIQUE,
    topic        VARCHAR(50) NOT NULL,          -- nome lógico: saga-commands | so-events | billing-events | execution-events
    target       VARCHAR(20),                   -- só para saga-commands
    type         VARCHAR(50) NOT NULL,
    body         JSONB NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_outbox_unpublished ON outbox (id) WHERE published_at IS NULL;
```

## `customer` — migrations `000001_init_schema` e `000002_create_user_admin`

`000001_init_schema.up.sql`: as tabelas `"user"`, `"role"`, `user_role`, `customer`, `vehicle` e os índices `idx_customer_user_id`, `idx_vehicle_customer_id`, `idx_user_role_user_id`, `idx_user_role_role_id` **exatamente como no monólito** (`auto-repair-shop/migrations/000001_init_schema.up.sql`), mais os `INSERT INTO "role"` (ids `'1'..'4'` = ADMIN, MECHANIC, ATTENDANT, CUSTOMER). **Omitir** `work`, `supply`, `service_order*` e os índices deles.
`000002_create_user_admin.up.sql` / `.down.sql`: cópia literal da `000002` do monólito.
Sem `processed_message`/`outbox` (o serviço não usa mensageria).

## `catalog` — migration `000001_init_schema`

```sql
CREATE TABLE "work" (            -- cópia do monólito
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    unit_price NUMERIC(10, 2) NOT NULL,
    status BOOLEAN NOT NULL DEFAULT TRUE,   -- true = ativo (a API expõe como `active`)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE supply (            -- cópia do monólito
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    unit_price NUMERIC(10, 2) NOT NULL,
    stock_quantity INT NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE stock_reservation (
    id               VARCHAR(36) PRIMARY KEY,
    service_order_id VARCHAR(26) NOT NULL,
    supply_id        VARCHAR(36) NOT NULL REFERENCES supply (id),
    quantity         INT NOT NULL CHECK (quantity > 0),
    status           VARCHAR(20) NOT NULL DEFAULT 'RESERVED' CHECK (status IN ('RESERVED', 'RELEASED')),
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    released_at      TIMESTAMP WITH TIME ZONE,
    UNIQUE (service_order_id, supply_id)
);
CREATE INDEX idx_stock_reservation_service_order_id ON stock_reservation (service_order_id);
-- + processed_message (comum). Sem outbox (Catalog não publica).
```
Regra de reserva: `INSERT … ON CONFLICT (service_order_id, supply_id) DO UPDATE SET status='RESERVED', quantity=EXCLUDED.quantity, released_at=NULL WHERE stock_reservation.status='RELEASED'`; se já `RESERVED` → `409`. Decremento e devolução de estoque na **mesma transação** da linha de reserva (`UPDATE supply SET stock_quantity = stock_quantity ∓ $q, version = version + 1 …`).

## `billing` — migration `000001_init_schema`

```sql
CREATE TABLE budget (
    id               VARCHAR(36) PRIMARY KEY,
    service_order_id VARCHAR(26) NOT NULL,
    customer_user_id VARCHAR(36) NOT NULL,     -- escopo: o CUSTOMER só lê o orçamento da própria OS (JWT user_id)
    status           VARCHAR(20) NOT NULL DEFAULT 'GENERATED' CHECK (status IN ('GENERATED', 'CANCELLED')),
    total_amount     NUMERIC(10, 2) NOT NULL,
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX uq_budget_active_per_so ON budget (service_order_id) WHERE status = 'GENERATED';

CREATE TABLE budget_item (
    id         VARCHAR(36) PRIMARY KEY,
    budget_id  VARCHAR(36) NOT NULL REFERENCES budget (id) ON DELETE CASCADE,
    kind       VARCHAR(10) NOT NULL CHECK (kind IN ('WORK', 'SUPPLY')),
    ref_id     VARCHAR(36) NOT NULL,           -- work_id ou supply_id (só referência)
    name       VARCHAR(255) NOT NULL,
    quantity   INT NOT NULL DEFAULT 1 CHECK (quantity > 0),
    unit_price NUMERIC(10, 2) NOT NULL
);
CREATE INDEX idx_budget_item_budget_id ON budget_item (budget_id);

CREATE TABLE payment (
    id               VARCHAR(36) PRIMARY KEY,
    budget_id        VARCHAR(36) NOT NULL REFERENCES budget (id),
    service_order_id VARCHAR(26) NOT NULL,
    saga_id          VARCHAR(36) NOT NULL,     -- do envelope do RequestPayment; reutilizado nos eventos de confirm/reject manuais
    amount           NUMERIC(10, 2) NOT NULL,
    status           VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                     CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED', 'REFUNDED', 'CANCELLED')),
    provider         VARCHAR(30) NOT NULL,     -- 'internal' (futuro: 'mercadopago')
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX uq_payment_open_per_so ON payment (service_order_id) WHERE status IN ('PENDING', 'APPROVED');
CREATE INDEX idx_payment_service_order_id ON payment (service_order_id);
-- + processed_message e outbox (comuns)
```
Transições de pagamento: `PENDING → APPROVED | REJECTED | CANCELLED`; `APPROVED → REFUNDED`. `RefundPayment` em `PENDING` → `CANCELLED`; em `APPROVED` → `REFUNDED`; em estado já final → ignora (idempotente) e publica o evento correspondente uma única vez.
