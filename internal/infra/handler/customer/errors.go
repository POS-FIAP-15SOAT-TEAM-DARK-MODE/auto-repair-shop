package customer

import "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"

var (
	ErrDocumentRequired      = domain.ValidationError{Message: "document is required"}
	ErrIDRequired            = domain.ValidationError{Message: "id is required"}
	ErrDocumentQueryRequired = domain.ValidationError{Message: "document query param is required"}
)
