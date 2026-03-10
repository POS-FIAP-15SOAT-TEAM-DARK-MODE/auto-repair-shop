-- Create the first user admin of the application
INSERT INTO "user" (id, name, email, password_hash)
VALUES ('3b123072-5ad5-4df8-8d3b-3c2770f11b2a', 'admin', 'admin@email.com', '8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918');

-- Insert role to this user
INSERT INTO "user_role" (id, user_id, role_id)
VALUES ('0b4b3baf-ff7b-461f-9b90-81f6b0c5e5b2', '3b123072-5ad5-4df8-8d3b-3c2770f11b2a', '1');