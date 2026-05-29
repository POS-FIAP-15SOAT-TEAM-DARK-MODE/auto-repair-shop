package customer

import (
	"database/sql"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/customer/interfaces"
)

func NewHTTPController(db *sql.DB) interfaces.CustomerHTTPController {
	return NewController(nil)
}
