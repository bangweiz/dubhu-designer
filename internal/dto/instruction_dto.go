package dto

import (
	"strings"
	"time"
)

type InstructionUsageResponseDTO struct {
	ConciergeID        string `json:"conciergeId"`
	ConciergeName      string `json:"conciergeName"`
	ConciergeVersionID string `json:"conciergeVersionId"`
	Version            int    `json:"version"`
	AgentID            string `json:"agentId"`
	AgentName          string `json:"agentName"`
}

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
type UpdateInstructionDTO struct {
	Name    string `json:"name" binding:"required"`
	Content string `json:"content" binding:"required"`
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
	Version   int       `json:"-"`
	CreatedBy string    `json:"createdBy"`
	UpdatedBy string    `json:"updatedBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// InstructionResponseDTO represents a full instruction returned in HTTP responses.
// Tools contains the fully resolved ToolResponseDTO objects ([] if empty).
type InstructionResponseDTO struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Content   string            `json:"content"`
	Tools     []ToolResponseDTO `json:"tools"`
	Version   int               `json:"-"`
	CreatedBy string            `json:"createdBy"`
	UpdatedBy string            `json:"updatedBy"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
}
