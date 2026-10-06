package dto

import (
	"strings"
	"time"
)

type UpdateConciergeDTO = CreateConciergeDTO

type CreateConciergeDTO struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

func (d *CreateConciergeDTO) Trim() {
	d.Name = strings.TrimSpace(d.Name)
	d.Description = strings.TrimSpace(d.Description)
}

type ConciergeVersionDescriptorDTO struct {
	ConciergeVersionID string `json:"conciergeVersionId"`
	Version            int    `json:"version"`
}
type ConciergeResponseDTO struct {
	ID                string                          `json:"id"`
	Name              string                          `json:"name"`
	Description       string                          `json:"description"`
	ConciergeVersions []ConciergeVersionDescriptorDTO `json:"conciergeVersions"`
	Version           int                             `json:"-"`
	CreatedAt         time.Time                       `json:"createdAt"`
	UpdatedAt         time.Time                       `json:"updatedAt"`
}
type ConciergeSummaryResponseDTO = ConciergeResponseDTO
