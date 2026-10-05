package dto

import (
	"strings"
)

// --- Agent Request DTOs ---

// CreateAgentDTO represents the request payload for creating an agent.
type CreateAgentDTO struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Goal        string `json:"goal" binding:"required"`
	Model       string `json:"model" binding:"required"`
}

// Trim trims leading and trailing whitespace from string fields in CreateAgentDTO.
func (d *CreateAgentDTO) Trim() {
	d.Name = strings.TrimSpace(d.Name)
	d.Description = strings.TrimSpace(d.Description)
	d.Goal = strings.TrimSpace(d.Goal)
	d.Model = strings.TrimSpace(d.Model)
}

// UpdateAgentDTO represents the request payload for updating an agent.
type UpdateAgentDTO struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Goal        string `json:"goal" binding:"required"`
	Model       string `json:"model" binding:"required"`
}

// Trim trims leading and trailing whitespace from string fields in UpdateAgentDTO.
func (d *UpdateAgentDTO) Trim() {
	d.Name = strings.TrimSpace(d.Name)
	d.Description = strings.TrimSpace(d.Description)
	d.Goal = strings.TrimSpace(d.Goal)
	d.Model = strings.TrimSpace(d.Model)
}

// --- Agent Response DTOs ---

// AgentSummaryResponseDTO represents an agent without assigned instructions.
// Used for listing agents.
type AgentSummaryResponseDTO struct {
	ID          string `json:"id"`
	ConciergeID string `json:"conciergeId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Goal        string `json:"goal"`
	Model       string `json:"model"`
}

// AgentResponseDTO represents an agent returned in HTTP responses.
type AgentResponseDTO struct {
	ID           string                          `json:"id"`
	ConciergeID  string                          `json:"conciergeId"`
	Name         string                          `json:"name"`
	Description  string                          `json:"description"`
	Goal         string                          `json:"goal"`
	Model        string                          `json:"model"`
	Instructions []InstructionSummaryResponseDTO `json:"instructions"`
	Tools        []ToolResponseDTO               `json:"tools"`
	// Version is internal state used to generate the HTTP ETag.
	Version int `json:"-"`
}
