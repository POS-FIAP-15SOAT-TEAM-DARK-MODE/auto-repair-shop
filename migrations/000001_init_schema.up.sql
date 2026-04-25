-- Create USER table
CREATE TABLE "user" (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create ROLE table
CREATE TABLE "role" (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create USER_ROLE join table
CREATE TABLE user_role (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES "user" (id) ON DELETE CASCADE,
    role_id VARCHAR(36) NOT NULL REFERENCES "role" (id) ON DELETE CASCADE,
    UNIQUE (user_id, role_id)
);

-- Create CUSTOMER table
CREATE TABLE customer (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL UNIQUE REFERENCES "user" (id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,
    cpf VARCHAR(14) UNIQUE,
    cnpj VARCHAR(18) UNIQUE,
    company_name VARCHAR(255),
    phone VARCHAR(25) NOT NULL,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        CHECK (
            (
                type = 'INDIVIDUAL'
                AND cpf IS NOT NULL
                AND cnpj IS NULL
                AND company_name IS NULL
            )
            OR (
                type = 'COMPANY'
                AND cpf IS NULL
                AND cnpj IS NOT NULL
                AND company_name IS NOT NULL
            )
        )
);

-- Create VEHICLE table
CREATE TABLE vehicle (
    id VARCHAR(36) PRIMARY KEY,
    license_plate VARCHAR(20) UNIQUE NOT NULL,
    brand VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    year INT NOT NULL,
    customer_id VARCHAR(36) NOT NULL REFERENCES customer (id) ON DELETE CASCADE,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create WORK table
CREATE TABLE "work" (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    unit_price NUMERIC(10, 2) NOT NULL,
    status BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create SUPPLY table
CREATE TABLE supply (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    unit_price NUMERIC(10, 2) NOT NULL,
    stock_quantity INT NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create SERVICE_ORDER table
CREATE TABLE service_order (
    id VARCHAR(26) PRIMARY KEY,
    customer_id VARCHAR(36) NOT NULL REFERENCES customer (id) ON DELETE CASCADE,
    vehicle_id VARCHAR(36) NOT NULL REFERENCES vehicle (id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'RECEIVED',
    total_amount NUMERIC(10, 2) NOT NULL DEFAULT 0,
    quote_sent_at TIMESTAMP
    WITH
        TIME ZONE,
        completed_at TIMESTAMP
    WITH
        TIME ZONE,
        delivered_at TIMESTAMP
    WITH
        TIME ZONE,
        created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create SERVICE_ORDER_WORK join table
CREATE TABLE service_order_work (
    id VARCHAR(36) PRIMARY KEY,
    service_order_id VARCHAR(36) NOT NULL REFERENCES service_order (id) ON DELETE CASCADE,
    work_id VARCHAR(36) NOT NULL REFERENCES work (id) ON DELETE CASCADE,
    unit_price NUMERIC(10, 2) NOT NULL,
    UNIQUE (service_order_id, work_id)
);

-- Create SERVICE_ORDER_SUPPLY join table
CREATE TABLE service_order_supply (
    id VARCHAR(36) PRIMARY KEY,
    service_order_id VARCHAR(36) NOT NULL REFERENCES service_order (id) ON DELETE CASCADE,
    supply_id VARCHAR(36) NOT NULL REFERENCES supply (id) ON DELETE CASCADE,
    quantity INT NOT NULL,
    unit_price NUMERIC(10, 2) NOT NULL,
    UNIQUE (service_order_id, supply_id)
);

-- Create SERVICE_ORDER_STATUS_HISTORY table
CREATE TABLE service_order_status_history (
    id VARCHAR(26) PRIMARY KEY,
    service_order_id VARCHAR(36) NOT NULL REFERENCES service_order (id) ON DELETE CASCADE,
    previous_status VARCHAR(50),
    new_status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP
    WITH
        TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for foreign keys and search columns
CREATE INDEX idx_customer_user_id ON customer (user_id);

CREATE INDEX idx_vehicle_customer_id ON vehicle (customer_id);

CREATE INDEX idx_service_order_status ON service_order (status);

CREATE INDEX idx_service_order_customer_id ON service_order (customer_id);

CREATE INDEX idx_service_order_vehicle_id ON service_order (vehicle_id);

CREATE INDEX idx_service_order_work_service_order_id ON service_order_work (service_order_id);

CREATE INDEX idx_service_order_supply_service_order_id ON service_order_supply (service_order_id);

CREATE INDEX idx_status_history_service_order_id ON service_order_status_history (service_order_id);

CREATE INDEX idx_user_role_user_id ON user_role (user_id);

CREATE INDEX idx_user_role_role_id ON user_role (role_id);

-- Insert default roles
INSERT INTO "role" (id, name) VALUES ('1', 'ADMIN');

INSERT INTO "role" (id, name) VALUES ('2', 'MECHANIC');

INSERT INTO "role" (id, name) VALUES ('3', 'ATTENDANT');

INSERT INTO "role" (id, name) VALUES ('4', 'CUSTOMER');

-- Constraints
ALTER TABLE service_order_status_history
ADD CONSTRAINT uq_service_order_status_history_service_order_id_new_status
UNIQUE (service_order_id, new_status);
