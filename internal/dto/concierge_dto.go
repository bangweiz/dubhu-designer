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
	Agents            []AgentSummaryResponseDTO       `json:"agents"`
	NextVersion       int                             `json:"nextVersion"`
	ID                string                          `json:"id"`
	Name              string                          `json:"name"`
	Description       string                          `json:"description"`
	ConciergeVersions []ConciergeVersionDescriptorDTO `json:"conciergeVersions"`
	CreatedBy         string                          `json:"createdBy"`
	UpdatedBy         string                          `json:"updatedBy"`
	CreatedAt         time.Time                       `json:"createdAt"`
	UpdatedAt         time.Time                       `json:"updatedAt"`
}

type ConciergeSummaryResponseDTO = ConciergeResponseDTO

// ConciergeVersionDeploymentDTO identifies the target environment for a version action.
type ConciergeVersionDeploymentDTO struct {
	EnvironmentID string `json:"environmentId" binding:"required"`
}

// SavedConciergeVersionResponseDTO exposes snapshot content and deployment metadata.
type SavedConciergeVersionResponseDTO struct {
	EnvironmentIDs []string                      `json:"environmentIds"`
	Variables      []VariableResponseDTO         `json:"variables"`
	ID             string                        `json:"id"`
	ConciergeID    string                        `json:"conciergeId"`
	Name           string                        `json:"name"`
	Description    string                        `json:"description"`
	Version        int                           `json:"version"`
	Agents         []SavedAgentResponseDTO       `json:"agents"`
	Instructions   []SavedInstructionResponseDTO `json:"instructions"`
	Tools          []ToolResponseDTO             `json:"tools"`
	CreatedBy      string                        `json:"createdBy"`
	UpdatedBy      string                        `json:"updatedBy"`
	CreatedAt      time.Time                     `json:"createdAt"`
	UpdatedAt      time.Time                     `json:"updatedAt"`
}

// SavedAgentResponseDTO retains references to documents contained in the snapshot.
type SavedAgentResponseDTO struct {
	ID           string    `json:"id"`
	ConciergeID  string    `json:"conciergeId,omitempty"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Goal         string    `json:"goal"`
	Model        string    `json:"model"`
	Instructions []string  `json:"instructions"`
	Tools        []string  `json:"tools"`
	CreatedBy    string    `json:"createdBy"`
	UpdatedBy    string    `json:"updatedBy"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// SavedInstructionResponseDTO retains tool references within the snapshot.
type SavedInstructionResponseDTO struct {
	Variables []string  `json:"variables"`
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	Tools     []string  `json:"tools"`
	CreatedBy string    `json:"createdBy"`
	UpdatedBy string    `json:"updatedBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
