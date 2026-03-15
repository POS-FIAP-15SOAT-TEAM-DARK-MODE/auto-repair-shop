package work

import (
	"errors"
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type createWorkReqDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Status      string `json:"status"`
}

func mapBodyToRequestDTO(body *gin.Context) (*createWorkReqDTO, error) {
	req := new(createWorkReqDTO)

	// TODO: replace by a performant alternative, build a custom serializer
	if err := body.ShouldBindJSON(req); err != nil {
		return nil, err
	}

	return req, nil
}

func (c *createWorkReqDTO) Validate() error {
	var errs []error

	if err := domain.ValidWorkStatusStringValue(c.Status); err != nil {
		errs = append(errs, err)
	}

	// Acceptable: only minus, digits, dots, and commas as characters
	for _, r := range c.Price {
		if !(r == '-' || r == '.' || r == ',' || (r >= '0' && r <= '9')) {
			errs = append(errs, domain.ErrInvalidWorkPriceValue)
			break
		}
	}

	// Remove thousands separators if present (like "1,000.23" -> "1000.23")
	// Only allow dot as decimal separator
	c.Price = strings.TrimSpace(strings.ReplaceAll(c.Price, ",", ""))

	// Reject non-parseable forms (e.g. "-", ".", "-.")
	if _, err := decimal.NewFromString(c.Price); err != nil {
		errs = append(errs, domain.ErrInvalidWorkPriceValue)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (c *createWorkReqDTO) MapToDomain() (*domain.Work, error) {
	return domain.NewWork(c.Name, c.Description, c.Price, domain.StringToWorkStatus(c.Status))
}
