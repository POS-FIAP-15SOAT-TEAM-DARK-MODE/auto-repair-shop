-- Local development only: these passwords are fixed, public and NOT secrets.
-- One login role per service database, plus the read-only role used by the login lambda.
CREATE ROLE customer_app        LOGIN PASSWORD 'local-dev-only';
CREATE ROLE catalog_app         LOGIN PASSWORD 'local-dev-only';
CREATE ROLE billing_app         LOGIN PASSWORD 'local-dev-only';
CREATE ROLE execution_app       LOGIN PASSWORD 'local-dev-only';
CREATE ROLE service_order_app   LOGIN PASSWORD 'local-dev-only';
CREATE ROLE lambda_auth_ro      LOGIN PASSWORD 'local-dev-only';
