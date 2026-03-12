package domain

type (
	BadRequestError struct {
		Message string `json:"message"`
	}
)

func (e BadRequestError) Error() string {
	return e.Message
}
