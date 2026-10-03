package validator

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	maxNameLength        = 200
	maxDescriptionLength = 2000
)

// FieldError represents a single field validation failure.
type FieldError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
	Value  any    `json:"value"`
}

// ValidationErrors collects multiple FieldError instances.
type ValidationErrors []FieldError

func (ve ValidationErrors) Error() string {
	msgs := make([]string, 0, len(ve))
	for _, fe := range ve {
		msgs = append(msgs, fmt.Sprintf("%s: %s (value: %v)", fe.Field, fe.Reason, fe.Value))
	}
	return strings.Join(msgs, "; ")
}

// validateName validates that a pre-trimmed name is non-empty and does not exceed maxNameLength.
func validateName(field, value string) *FieldError {
	if value == "" {
		return &FieldError{
			Field:  field,
			Reason: fmt.Sprintf("%s is required and cannot be blank", field),
			Value:  value,
		}
	}
	if utf8.RuneCountInString(value) > maxNameLength {
		return &FieldError{
			Field:  field,
			Reason: fmt.Sprintf("%s cannot exceed %d characters", field, maxNameLength),
			Value:  value,
		}
	}
	return nil
}

// validateDescription validates that a pre-trimmed description is non-empty and does not exceed maxDescriptionLength.
func validateDescription(field, value string) *FieldError {
	if value == "" {
		return &FieldError{
			Field:  field,
			Reason: fmt.Sprintf("%s is required and cannot be blank", field),
			Value:  value,
		}
	}
	if utf8.RuneCountInString(value) > maxDescriptionLength {
		return &FieldError{
			Field:  field,
			Reason: fmt.Sprintf("%s cannot exceed %d characters", field, maxDescriptionLength),
			Value:  value,
		}
	}
	return nil
}
