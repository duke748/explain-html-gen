package schema

import "fmt"

// ValidationError represents a validation problem in the input.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func NewValidationError(format string, args ...interface{}) *ValidationError {
	return &ValidationError{
		Message: fmt.Sprintf(format, args...),
	}
}
