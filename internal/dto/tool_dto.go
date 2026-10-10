package dto

import (
	"strings"
	"time"
)

// --- Request DTOs ---

type ToolUsageResponseDTO = InstructionUsageResponseDTO

// ToolInputDTO represents an input parameter in create or update requests.
type ToolInputDTO struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Required    bool   `json:"required"`
}

// ToolOutputDTO represents an output parameter in create or update requests.
type ToolOutputDTO struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

// CreateToolDTO represents the request payload for creating a new tool.
type CreateToolDTO struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description" binding:"required"`
	Inputs      []ToolInputDTO  `json:"inputs"`
	Outputs     []ToolOutputDTO `json:"outputs"`
}

// Trim trims leading and trailing whitespace from all string fields in CreateToolDTO.
func (d *CreateToolDTO) Trim() {
	d.Name = strings.TrimSpace(d.Name)
	d.Description = strings.TrimSpace(d.Description)
	for i := range d.Inputs {
		d.Inputs[i].Name = strings.TrimSpace(d.Inputs[i].Name)
		d.Inputs[i].Description = strings.TrimSpace(d.Inputs[i].Description)
	}

	for i := range d.Outputs {
		d.Outputs[i].Name = strings.TrimSpace(d.Outputs[i].Name)
		d.Outputs[i].Description = strings.TrimSpace(d.Outputs[i].Description)
	}
}

// UpdateToolDTO represents the request payload for updating an existing tool.
type UpdateToolDTO struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description" binding:"required"`
	Inputs      []ToolInputDTO  `json:"inputs"`
	Outputs     []ToolOutputDTO `json:"outputs"`
}

// Trim trims leading and trailing whitespace from all string fields in UpdateToolDTO.
func (d *UpdateToolDTO) Trim() {
	d.Name = strings.TrimSpace(d.Name)
	d.Description = strings.TrimSpace(d.Description)
	for i := range d.Inputs {
		d.Inputs[i].Name = strings.TrimSpace(d.Inputs[i].Name)
		d.Inputs[i].Description = strings.TrimSpace(d.Inputs[i].Description)
	}

	for i := range d.Outputs {
		d.Outputs[i].Name = strings.TrimSpace(d.Outputs[i].Name)
		d.Outputs[i].Description = strings.TrimSpace(d.Outputs[i].Description)
	}
}

// --- Response DTOs ---

// ToolInputResponseDTO represents an input in the response.
type ToolInputResponseDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// ToolOutputResponseDTO represents an output in the response.
type ToolOutputResponseDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ToolResponseDTO represents a tool returned in HTTP responses.
// Guarantees non-nil slices (always serialized as [] instead of null in JSON).
type ToolResponseDTO struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Inputs      []ToolInputResponseDTO  `json:"inputs"`
	Outputs     []ToolOutputResponseDTO `json:"outputs"`
	CreatedBy   string                  `json:"createdBy"`
	UpdatedBy   string                  `json:"updatedBy"`
	CreatedAt   time.Time               `json:"createdAt"`
	UpdatedAt   time.Time               `json:"updatedAt"`
}
