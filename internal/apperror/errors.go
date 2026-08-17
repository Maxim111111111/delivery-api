package apperror

import (
	"errors"
	"strings"
)

var ErrOrderNotFound = errors.New("order not found")

type ValidationError struct {
	Messages []string
}

func (e *ValidationError) Error() string {
	return strings.Join(e.Messages, "; ")
}
