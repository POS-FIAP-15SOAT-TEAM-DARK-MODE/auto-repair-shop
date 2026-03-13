package domain

type (
	ValidationError struct {
		Message string `json:"message"`
	}

	ConflictError struct {
		Message string `json:"message"`
	}

	BusinessRuleError struct {
		Message string `json:"message"`
	}
)

func (e ValidationError) Error() string {
	return e.Message
}

func (e ConflictError) Error() string {
	return e.Message
}

func (e BusinessRuleError) Error() string {
	return e.Message
}
