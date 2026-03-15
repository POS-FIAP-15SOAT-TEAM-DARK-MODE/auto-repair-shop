package web

// Validatable is implemented by request structs that perform DTO-level validation
// (fields that have no direct equivalent in the domain layer).
type Validatable interface {
	Validate() error
}
