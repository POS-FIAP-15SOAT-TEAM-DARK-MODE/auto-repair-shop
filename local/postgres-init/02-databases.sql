-- One database per service (contracts/resources.md). The owner is the service role, so the
-- migrations (run as that role) can create the tables in schema public (PostgreSQL 15 gives
-- CREATE on public to the database owner only). No other role may connect.
CREATE DATABASE customer      OWNER customer_app;
CREATE DATABASE catalog       OWNER catalog_app;
CREATE DATABASE billing       OWNER billing_app;
CREATE DATABASE execution     OWNER execution_app;
CREATE DATABASE service_order OWNER service_order_app;

REVOKE CONNECT ON DATABASE customer      FROM PUBLIC;
REVOKE CONNECT ON DATABASE catalog       FROM PUBLIC;
REVOKE CONNECT ON DATABASE billing       FROM PUBLIC;
REVOKE CONNECT ON DATABASE execution     FROM PUBLIC;
REVOKE CONNECT ON DATABASE service_order FROM PUBLIC;
