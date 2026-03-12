package web

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-json"
)

const (
	contentTypeHeader       = "Content-Type"
	applicationJSON         = "application/json"
	contentTypeErrorMessage = "Content-Type must be application/json"
	invalidJSONMessage      = "Invalid JSON payload"
)

func ValidateAndDecodeJSON[T any](r *http.Request, ptr *T) error {

	if r.Header.Get(contentTypeHeader) != applicationJSON {
		return domain.BadRequestError{
			Message: contentTypeErrorMessage,
		}
	}

	if err := json.NewDecoder(r.Body).Decode(ptr); err != nil {
		return domain.BadRequestError{
			Message: invalidJSONMessage,
		}
	}

	if err := validate.Struct(*ptr); err != nil {
		return domain.BadRequestError{
			Message: formatValidationError(err),
		}
	}

	return nil
}

func formatValidationError(err error) string {
	var errs validator.ValidationErrors
	if !errors.As(err, &errs) {
		return err.Error()
	}

	messages := make([]string, 0, len(errs))
	for _, e := range errs {
		messages = append(messages, fmt.Sprintf("field validation for '%s' failed on the '%s' tag", e.Field(), e.Tag()))
	}
	return strings.Join(messages, "; ")
}
