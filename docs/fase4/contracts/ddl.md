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

## `execution` — migration `000001_init_schema`

```sql
CREATE TABLE execution_task (
    id               VARCHAR(36) PRIMARY KEY,
    service_order_id VARCHAR(26) NOT NULL UNIQUE,
    saga_id          VARCHAR(36) NOT NULL,     -- do envelope do EnqueueExecution; vai em TODOS os eventos da tarefa
    status           VARCHAR(20) NOT NULL DEFAULT 'QUEUED'
                     CHECK (status IN ('QUEUED', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED')),
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    started_at       TIMESTAMP WITH TIME ZONE,
    completed_at     TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_execution_task_status_created ON execution_task (status, created_at);

CREATE TABLE execution_work (
    id        VARCHAR(36) PRIMARY KEY,
    task_id   VARCHAR(36) NOT NULL REFERENCES execution_task (id) ON DELETE CASCADE,
    work_id   VARCHAR(36) NOT NULL,            -- só referência ao catálogo
    name      VARCHAR(255) NOT NULL,
    status    VARCHAR(20) NOT NULL DEFAULT 'AWAITING_START'
              CHECK (status IN ('AWAITING_START', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED')),
    UNIQUE (task_id, work_id)
);

CREATE TABLE execution_work_history (
    id              VARCHAR(26) PRIMARY KEY,   -- ULID
    task_id         VARCHAR(36) NOT NULL REFERENCES execution_task (id) ON DELETE CASCADE,
    work_id         VARCHAR(36) NOT NULL,
    previous_status VARCHAR(20),               -- NULL na criação
    new_status      VARCHAR(20) NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_execution_work_history_task_work ON execution_work_history (task_id, work_id);
-- + processed_message e outbox (comuns)
```
Ciclo do work (igual ao monólito): `AWAITING_START → IN_PROGRESS → COMPLETED`; `CANCELLED` a partir de qualquer não terminal.

## `service_order` — migration `000001_init_schema`

```sql
CREATE TABLE service_order (
    id                    VARCHAR(26) PRIMARY KEY,           -- ULID
    status                VARCHAR(50) NOT NULL DEFAULT 'NEW',
    total_amount          NUMERIC(10, 2) NOT NULL DEFAULT 0,
    -- snapshot do cliente (as respostas de listagem/detalhe e a notificação usam estes campos)
    customer_id           VARCHAR(36) NOT NULL,
    customer_user_id      VARCHAR(36) NOT NULL,               -- escopo do accept/reject (JWT user_id)
    customer_name         VARCHAR(255) NOT NULL,
    customer_email        VARCHAR(255) NOT NULL,
    customer_type         VARCHAR(20) NOT NULL,
    customer_document     VARCHAR(18) NOT NULL,
    customer_company_name VARCHAR(255),
    customer_phone        VARCHAR(25) NOT NULL,
    -- snapshot do veículo
    vehicle_id            VARCHAR(36) NOT NULL,
    vehicle_license_plate VARCHAR(20) NOT NULL,
    vehicle_brand         VARCHAR(100) NOT NULL,
    vehicle_model         VARCHAR(100) NOT NULL,
    vehicle_year          INT NOT NULL,
    quote_sent_at         TIMESTAMP WITH TIME ZONE,
    completed_at          TIMESTAMP WITH TIME ZONE,
    delivered_at          TIMESTAMP WITH TIME ZONE,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_service_order_status ON service_order (status);
CREATE INDEX idx_service_order_customer_id ON service_order (customer_id);
CREATE INDEX idx_service_order_vehicle_id ON service_order (vehicle_id);

CREATE TABLE service_order_work (
    id               VARCHAR(36) PRIMARY KEY,
    service_order_id VARCHAR(26) NOT NULL REFERENCES service_order (id) ON DELETE CASCADE,
    work_id          VARCHAR(36) NOT NULL,
    name             VARCHAR(255) NOT NULL,
    description      TEXT,
    unit_price       NUMERIC(10, 2) NOT NULL,
    UNIQUE (service_order_id, work_id)
);

CREATE TABLE service_order_supply (
    id               VARCHAR(36) PRIMARY KEY,
    service_order_id VARCHAR(26) NOT NULL REFERENCES service_order (id) ON DELETE CASCADE,
    supply_id        VARCHAR(36) NOT NULL,
    reservation_id   VARCHAR(36) NOT NULL,                     -- do Catalog (DELETE /internal/reservations/{id})
    name             VARCHAR(255) NOT NULL,
    description      TEXT,
    version          INT NOT NULL,
    quantity         INT NOT NULL CHECK (quantity > 0),
    unit_price       NUMERIC(10, 2) NOT NULL,
    UNIQUE (service_order_id, supply_id)
);
CREATE INDEX idx_so_work_so_id ON service_order_work (service_order_id);
CREATE INDEX idx_so_supply_so_id ON service_order_supply (service_order_id);

CREATE TABLE saga_instance (
    id                 VARCHAR(36) PRIMARY KEY,                -- = saga_id das mensagens
    service_order_id   VARCHAR(26) NOT NULL REFERENCES service_order (id) ON DELETE CASCADE,
    step               VARCHAR(20) NOT NULL
                       CHECK (step IN ('BUDGET', 'AWAITING_CUSTOMER', 'PAYMENT', 'EXECUTION_QUEUE', 'EXECUTION', 'DONE')),
    status             VARCHAR(20) NOT NULL DEFAULT 'RUNNING'
                       CHECK (status IN ('RUNNING', 'COMPLETED', 'COMPENSATED', 'FAILED')),
    -- o que já foi comandado/concluído (decide quais compensações enviar)
    budget_requested   BOOLEAN NOT NULL DEFAULT FALSE,
    payment_requested  BOOLEAN NOT NULL DEFAULT FALSE,
    payment_approved   BOOLEAN NOT NULL DEFAULT FALSE,
    execution_enqueued BOOLEAN NOT NULL DEFAULT FALSE,
    deadline           TIMESTAMP WITH TIME ZONE,               -- NULL = sem prazo (AWAITING_CUSTOMER, EXECUTION)
    created_at         TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX uq_saga_active_per_so ON saga_instance (service_order_id) WHERE status = 'RUNNING';
CREATE INDEX idx_saga_deadline ON saga_instance (deadline) WHERE status = 'RUNNING' AND deadline IS NOT NULL;
-- + processed_message e outbox (comuns). SEM tabelas de histórico (agora é o serviço History).
```
`SERVICE_ORDER_STATUS` (valores válidos): `NEW`, `RECEIVED`, `IN_DIAGNOSIS`, `AWAITING_APPROVAL`, `IN_PROGRESS`, `COMPLETED`, `DELIVERED`, `REJECTED`, `CANCELLED`.

## `history` (DynamoDB, tabela `auto-repair-shop-<env>-history`, chaves `PK` e `SK` do tipo S)

| Item | `PK` | `SK` | Atributos |
|---|---|---|---|
| Transição de status da OS | `SO#<service_order_id>` | `S#<occurred_at>#<message_id>` | `previous_status` (ausente na criação), `new_status`, `reason?`, `created_at` |
| Transição de work | `SO#<service_order_id>` | `W#<work_id>#<occurred_at>#<message_id>` | `work_id`, `work_name`, `previous_status?`, `new_status`, `created_at` |
| Marcador de execução do work | `SO#<service_order_id>` | `R#<work_id>` | `started_at` (gravado quando o work vai a `IN_PROGRESS`) |
| Agregado do relatório por work | `REPORT#WORK` | `W#<work_id>` | `work_name`, `count` (N), `total_seconds` (N) |
| Dedupe de mensagem | `MSG#<message_id>` | `DEDUPE` | `processed_at` |

Regras: toda escrita de uma mensagem usa **`TransactWriteItems`** (dedupe com `attribute_not_exists(PK)` + item do evento + marcador/agregado quando aplicável) — se a transação falhar por dedupe já existente, a mensagem é duplicada: descartar e apagar da fila. Em `ExecutionWorkStatusChanged` com `new=IN_PROGRESS`: gravar o marcador `R#`. Com `previous=IN_PROGRESS` e `new=COMPLETED`: ler o marcador (leitura consistente); se existir, `duração = occurred_at(envelope) − started_at` e incrementar o agregado (`ADD count 1, total_seconds :d`, `SET work_name`) na **mesma** transação; se não existir (mensagens fora de ordem), gravar só o evento e logar. Datas sempre do envelope (`occurred_at`), nunca do relógio de processamento.

**Formato de `<occurred_at>` nas chaves:** UTC com **largura fixa** — `2006-01-02T15:04:05.000000000Z` (sempre 9 casas de nanossegundo). O RFC 3339 comum omite casas decimais e quebra a ordenação por string (`…00Z` ficaria depois de `…00.1Z`). `processed_at` do item de dedupe usa o `occurred_at` do envelope; em empate de instante no mesmo work, a ordem desempata pelo `message_id`.
