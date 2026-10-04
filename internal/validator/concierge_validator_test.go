package validator

import (
	"strings"
	"testing"

	"github.com/bangweiz/dubhu-designer/internal/dto"
)

func TestValidateCreateConcierge(t *testing.T) {
	tests := []struct {
		name        string
		req         dto.CreateConciergeDTO
		expectError bool
		errorFields []string
	}{
		{
			name: "valid concierge",
			req: dto.CreateConciergeDTO{
				Name:        "front_desk_assistant",
				Description: "Handles incoming customer requests and routes inquiries.",
			},
			expectError: false,
		},
		{
			name: "empty name and description",
			req: dto.CreateConciergeDTO{
				Name:        "",
				Description: "",
			},
			expectError: true,
			errorFields: []string{"name", "description"},
		},
		{
			name: "name exceeds max length",
			req: dto.CreateConciergeDTO{
				Name:        strings.Repeat("a", 201),
				Description: "Valid description",
			},
			expectError: true,
			errorFields: []string{"name"},
		},
		{
			name: "description exceeds max length",
			req: dto.CreateConciergeDTO{
				Name:        "Valid Name",
				Description: strings.Repeat("b", 2001),
			},
			expectError: true,
			errorFields: []string{"description"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCreateConcierge(&tt.req)
			if tt.expectError {
				if len(errs) == 0 {
					t.Errorf("expected validation errors, got none")
				}
				for _, expectedField := range tt.errorFields {
					found := false
					for _, err := range errs {
						if err.Field == expectedField {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error on field %q, but not found", expectedField)
					}
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("expected no validation errors, got %v", errs)
				}
			}
		})
	}
}
