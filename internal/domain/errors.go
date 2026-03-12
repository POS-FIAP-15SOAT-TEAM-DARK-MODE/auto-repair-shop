package domain

type (
	BadRequestError struct {
		Message string `json:"message"`
	}

	ConflictError struct {
		Message string `json:"message"`
	}
)

func (e BadRequestError) Error() string {
	return e.Message
}

func (e ConflictError) Error() string {
	return e.Message
}
