package validator

import (
	"strings"
	"testing"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
)

func TestValidateCreateAgent(t *testing.T) {
	tests := []struct {
		name        string
		req         dto.CreateAgentDTO
		expectError bool
		errorFields []string
	}{
		{
			name: "valid agent with gemini-3.5-flash",
			req: dto.CreateAgentDTO{
				Name:        "research_assistant",
				Description: "Conducts deep literature research and analysis.",
				Model:       string(models.ModelGemini35Flash),
			},
			expectError: false,
		},
		{
			name: "valid agent with gemini-3.5-flash-lite",
			req: dto.CreateAgentDTO{
				Name:        "quick_assistant",
				Description: "Fast lightweight assistant.",
				Model:       string(models.ModelGemini35FlashLite),
			},
			expectError: false,
		},
		{
			name: "empty name, description, and model",
			req: dto.CreateAgentDTO{
				Name:        "",
				Description: "",
				Model:       "",
			},
			expectError: true,
			errorFields: []string{"name", "description", "model"},
		},
		{
			name: "name exceeds max length",
			req: dto.CreateAgentDTO{
				Name:        strings.Repeat("a", 201),
				Description: "Valid description",
				Model:       string(models.ModelGemini35Flash),
			},
			expectError: true,
			errorFields: []string{"name"},
		},
		{
			name: "description exceeds max length",
			req: dto.CreateAgentDTO{
				Name:        "Valid Name",
				Description: strings.Repeat("b", 2001),
				Model:       string(models.ModelGemini35Flash),
			},
			expectError: true,
			errorFields: []string{"description"},
		},
		{
			name: "unsupported model value",
			req: dto.CreateAgentDTO{
				Name:        "Valid Name",
				Description: "Valid description",
				Model:       "gemini-1.5-pro",
			},
			expectError: true,
			errorFields: []string{"model"},
		},
		{
			name: "arbitrary invalid model string",
			req: dto.CreateAgentDTO{
				Name:        "Valid Name",
				Description: "Valid description",
				Model:       "gpt-4",
			},
			expectError: true,
			errorFields: []string{"model"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCreateAgent(&tt.req)
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
