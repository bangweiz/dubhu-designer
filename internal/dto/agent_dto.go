package dto

import (
	"strings"
)

// --- Agent Request DTOs ---

// CreateAgentDTO represents the request payload for creating an agent.
type CreateAgentDTO struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Model       string `json:"model" binding:"required"`
}

// Trim trims leading and trailing whitespace from string fields in CreateAgentDTO.
func (d *CreateAgentDTO) Trim() {
	d.Name = strings.TrimSpace(d.Name)
	d.Description = strings.TrimSpace(d.Description)
	d.Model = strings.TrimSpace(d.Model)
}

// --- Agent Response DTOs ---

// AgentResponseDTO represents an agent returned in HTTP responses.
type AgentResponseDTO struct {
	ID          string `json:"id"`
	ConciergeID string `json:"conciergeId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Model       string `json:"model"`
	Version     int    `json:"version"`
}
