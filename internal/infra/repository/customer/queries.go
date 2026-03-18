package customer

const (
	createCustomerQuery = `
INSERT INTO customer (id, user_id, type, cpf, cnpj, company_name, phone)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`
)
