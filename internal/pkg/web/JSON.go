package web

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-json"
)

func ValidateAndDecodeJSON[T any](r *http.Request, ptr *T) error {

	if err := json.NewDecoder(r.Body).Decode(ptr); err != nil {
		return err //TODO: add custom error handling
	}

	if err := validate.Struct(*ptr); err != nil {
		return errors.New(formatValidationError(err)) //TODO: add custom error handling
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
