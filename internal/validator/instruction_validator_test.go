package validator

import (
	"strings"
	"testing"

	"github.com/bangweiz/dubhu-designer/internal/dto"
)

func TestValidateCreateInstruction(t *testing.T) {
	tests := []struct {
		name        string
		req         dto.CreateInstructionDTO
		expectError bool
		errorFields []string
	}{
		{
			name: "valid instruction",
			req: dto.CreateInstructionDTO{
				Name:    "search_and_summarize",
				Content: "Use {{tool:651da3b57f8641a27e36c579}} to find articles and summarize.",
			},
			expectError: false,
		},
		{
			name: "empty name and content",
			req: dto.CreateInstructionDTO{
				Name:    "",
				Content: "",
			},
			expectError: true,
			errorFields: []string{"name", "content"},
		},
		{
			name: "name exceeds max length",
			req: dto.CreateInstructionDTO{
				Name:    strings.Repeat("a", 201),
				Content: "Valid content",
			},
			expectError: true,
			errorFields: []string{"name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCreateInstruction(&tt.req)
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

func TestValidateUpdateInstruction(t *testing.T) {
	tests := []struct {
		name        string
		req         dto.UpdateInstructionDTO
		expectError bool
		errorFields []string
	}{
		{
			name: "valid update",
			req: dto.UpdateInstructionDTO{
				Name:    "updated_instruction",
				Content: "Updated content with {{tool:651da3b57f8641a27e36c579}}",
				Version: 1,
			},
			expectError: false,
		},
		{
			name: "invalid version 0",
			req: dto.UpdateInstructionDTO{
				Name:    "updated_instruction",
				Content: "Updated content",
				Version: 0,
			},
			expectError: true,
			errorFields: []string{"version"},
		},
		{
			name: "negative version",
			req: dto.UpdateInstructionDTO{
				Name:    "updated_instruction",
				Content: "Updated content",
				Version: -1,
			},
			expectError: true,
			errorFields: []string{"version"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateUpdateInstruction(&tt.req)
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
