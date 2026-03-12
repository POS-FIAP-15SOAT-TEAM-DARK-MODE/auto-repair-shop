package dto

type (
	CreateCustomerResponse struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Email       string `json:"email"`
		Type        string `json:"type"`
		Document    string `json:"document"`
		CompanyName string `json:"company_name,omitempty"`
		Phone       string `json:"phone"`
	}
)
