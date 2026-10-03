package dto

import (
	"strings"
	"time"
)

// --- Instruction Request DTOs ---

// CreateInstructionDTO represents the request payload for creating an instruction.
type CreateInstructionDTO struct {
	Name    string `json:"name" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// Trim trims leading and trailing whitespace from string fields in CreateInstructionDTO.
func (d *CreateInstructionDTO) Trim() {
	d.Name = strings.TrimSpace(d.Name)
	d.Content = strings.TrimSpace(d.Content)
}

// UpdateInstructionDTO represents the request payload for updating an instruction.
// Version is required for optimistic concurrency control.
type UpdateInstructionDTO struct {
	Name    string `json:"name" binding:"required"`
	Content string `json:"content" binding:"required"`
	Version int    `json:"version" binding:"required"`
}

// Trim trims leading and trailing whitespace from string fields in UpdateInstructionDTO.
func (d *UpdateInstructionDTO) Trim() {
	d.Name = strings.TrimSpace(d.Name)
	d.Content = strings.TrimSpace(d.Content)
}

// --- Instruction Response DTOs ---

// InstructionSummaryResponseDTO represents a summary of an instruction without embedded tools.
// Used for listing instructions.
type InstructionSummaryResponseDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// InstructionResponseDTO represents a full instruction returned in HTTP responses.
// Tools contains the fully resolved ToolResponseDTO objects ([] if empty).
type InstructionResponseDTO struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Content   string            `json:"content"`
	Tools     []ToolResponseDTO `json:"tools"`
	Version   int               `json:"version"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}
