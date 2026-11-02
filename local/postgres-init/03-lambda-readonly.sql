-- lambda_auth_ro reads the customer database only (SELECT on the tables created by customer_app).
-- The tables do not exist yet: the default privilege applies to the ones created later by the
-- customer migrations (run as customer_app).
\connect customer
GRANT CONNECT ON DATABASE customer TO lambda_auth_ro;
GRANT USAGE ON SCHEMA public TO lambda_auth_ro;
ALTER DEFAULT PRIVILEGES FOR ROLE customer_app IN SCHEMA public GRANT SELECT ON TABLES TO lambda_auth_ro;
