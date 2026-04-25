package customer

const (
	createCustomerQuery = `
INSERT INTO customer (id, user_id, type, cpf, cnpj, company_name, phone)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`

	getCustomerByIDQuery = `
SELECT c.id, c.user_id, c.type,
       COALESCE(c.cpf, ''), COALESCE(c.cnpj, ''), COALESCE(c.company_name, ''), c.phone,
       u.name, u.email
FROM customer c
JOIN "user" u ON u.id = c.user_id
WHERE c.id = $1
`

	getCustomerByDocumentQuery = `
SELECT c.id, c.user_id, c.type,
       COALESCE(c.cpf, ''), COALESCE(c.cnpj, ''), COALESCE(c.company_name, ''), c.phone,
       u.name, u.email
FROM customer c
JOIN "user" u ON u.id = c.user_id
WHERE c.cpf = $1 OR c.cnpj = $1
`

	getCustomerByUserIDQuery = `
SELECT c.id, c.user_id, c.type,
       COALESCE(c.cpf, ''), COALESCE(c.cnpj, ''), COALESCE(c.company_name, ''), c.phone,
       u.name, u.email
FROM customer c
JOIN "user" u ON u.id = c.user_id
WHERE u.id = $1
`

	updateCustomerQuery = `UPDATE customer SET phone = $1 WHERE id = $2`

	deleteCustomerQuery = `DELETE FROM customer WHERE id = $1`

	countServiceOrdersByCustomerQuery = `SELECT COUNT(*) FROM service_order WHERE customer_id = $1`
)
