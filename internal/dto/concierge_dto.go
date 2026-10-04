package dto

import (
	"strings"
	"time"
)

// --- Concierge Request DTOs ---

// CreateConciergeDTO represents the request payload for creating a concierge.
type CreateConciergeDTO struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

// Trim trims leading and trailing whitespace from string fields in CreateConciergeDTO.
func (d *CreateConciergeDTO) Trim() {
	d.Name = strings.TrimSpace(d.Name)
	d.Description = strings.TrimSpace(d.Description)
}

// --- Concierge Response DTOs ---

// ConciergeSummaryResponseDTO represents a summary of a concierge without agent IDs.
// Used for listing concierges.
type ConciergeSummaryResponseDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ConciergeResponseDTO represents a concierge returned in HTTP responses.
type ConciergeResponseDTO struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Agents      []AgentResponseDTO `json:"agents"`
	Version     int                `json:"version"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
}
