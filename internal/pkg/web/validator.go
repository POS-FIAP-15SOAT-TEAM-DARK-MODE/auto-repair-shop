package web

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = initValidator()

func initValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})
	return v
}
