package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
)

var (
	ErrWrongPayloadFormat   = fmt.Errorf("invalid payload")
	ErrJSONSyntax           = fmt.Errorf("invalid JSON syntax")
	ErrJSONType             = fmt.Errorf("JSON type mismatch")
	ErrJSONInvalidUnmarshal = fmt.Errorf("invalid unmarshal target")
	ErrJSONUnsupportedType  = fmt.Errorf("unsupported JSON type")
	ErrJSONUnsupportedValue = fmt.Errorf("unsupported JSON value")
	ErrJSONUnexpectedEOF    = fmt.Errorf("unexpected end of JSON input")
	ErrJSONEmptyBody        = fmt.Errorf("JSON body is empty")
	ErrJSONUnknown          = fmt.Errorf("unknown JSON error")
)

func mapJSONErrorType(err error) error {
	switch err.(type) {
	case *json.SyntaxError:
		return ErrJSONSyntax
	case *json.UnmarshalTypeError:
		return ErrJSONType
	case *json.InvalidUnmarshalError:
		return ErrJSONInvalidUnmarshal
	case *json.UnsupportedTypeError:
		return ErrJSONUnsupportedType
	case *json.UnsupportedValueError:
		return ErrJSONUnsupportedValue
	}
	return nil
}

func CheckJsonError(err error) error {
	if err == nil {
		return nil
	}

	if ginErr, ok := err.(*gin.Error); ok && ginErr.Err != nil {
		if mapped := mapJSONErrorType(ginErr.Err); mapped != nil {
			return mapped
		}

		err = ginErr.Err
	}

	if mapped := mapJSONErrorType(err); mapped != nil {
		return mapped
	}

	if errors.Is(err, io.EOF) {
		return ErrJSONEmptyBody
	}

	if errors.Is(err, io.ErrUnexpectedEOF) {
		return ErrJSONUnexpectedEOF
	}

	switch err.Error() {
	case "unexpected EOF", "unexpected end of JSON input":
		return ErrJSONUnexpectedEOF
	case "EOF":
		return ErrJSONEmptyBody
	}

	return ErrWrongPayloadFormat
}
