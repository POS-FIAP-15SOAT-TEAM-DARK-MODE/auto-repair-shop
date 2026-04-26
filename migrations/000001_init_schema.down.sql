-- Drop all tables in reverse order of creation
DROP TABLE IF EXISTS work_service_order_status_history CASCADE;

ALTER TABLE service_order_status_history
DROP CONSTRAINT IF EXISTS uq_service_order_status_history_service_order_id_new_status;

DROP TABLE IF EXISTS service_order_status_history CASCADE;

DROP TABLE IF EXISTS service_order_supply CASCADE;

DROP TABLE IF EXISTS service_order_work CASCADE;

DROP TABLE IF EXISTS service_order CASCADE;

DROP TABLE IF EXISTS supply CASCADE;

DROP TABLE IF EXISTS "work" CASCADE;

DROP TABLE IF EXISTS vehicle CASCADE;

DROP TABLE IF EXISTS customer CASCADE;

DROP TABLE IF EXISTS user_role CASCADE;

DROP TABLE IF EXISTS "role" CASCADE;

DROP TABLE IF EXISTS "user" CASCADE;

DROP INDEX IF EXISTS idx_customer_user_id;

DROP INDEX IF EXISTS idx_vehicle_customer_id;

DROP INDEX IF EXISTS idx_service_order_status;

DROP INDEX IF EXISTS idx_service_order_customer_id;

DROP INDEX IF EXISTS idx_service_order_vehicle_id;

DROP INDEX IF EXISTS idx_service_order_work_service_order_id;

DROP INDEX IF EXISTS idx_service_order_supply_service_order_id;

DROP INDEX IF EXISTS idx_status_history_service_order_id;

DROP INDEX IF EXISTS idx_status_history_work_service_order_id;

DROP INDEX IF EXISTS idx_user_role_user_id;

DROP INDEX IF EXISTS idx_user_role_role_id;
