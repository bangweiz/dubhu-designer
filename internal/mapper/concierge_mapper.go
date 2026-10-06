package mapper

import (
	"time"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ToInitialConciergeEntity converts a CreateConciergeDTO into a new Concierge entity
// with version 1, a new ObjectID, and UTC timestamps.
func ToInitialConciergeEntity(input dto.CreateConciergeDTO) *models.Concierge {
	now := time.Now().UTC()
	return &models.Concierge{
		ID:          bson.NewObjectID(),
		Name:        input.Name,
		Description: input.Description,
		Agents:      []models.Agent{},
		ETagVersion: 1,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// ToConciergeResponseDTO converts a domain Concierge model to ConciergeResponseDTO,
// ensuring Agents is always an initialized array of AgentSummaryResponseDTO ([] if empty).
func ToConciergeResponseDTO(c *models.Concierge) dto.ConciergeResponseDTO {
	agents := make([]models.Agent, len(c.Agents))
	for i := range c.Agents {
		agents[i] = c.Agents[i]
		agents[i].ConciergeID = c.ID
	}

	return dto.ConciergeResponseDTO{
		ID:          c.ID.Hex(),
		Name:        c.Name,
		Description: c.Description,
		Agents:      ToAgentSummaryResponseDTOList(agents),
		Version:     c.Version,
		ETagVersion: c.ETagVersion,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// ToConciergeResponseDTOList converts a slice of domain Concierge models to a slice of ConciergeResponseDTO.
func ToConciergeResponseDTOList(concierges []models.Concierge) []dto.ConciergeResponseDTO {
	res := make([]dto.ConciergeResponseDTO, 0, len(concierges))
	for i := range concierges {
		res = append(res, ToConciergeResponseDTO(&concierges[i]))
	}
	return res
}

// ToConciergeSummaryResponseDTO converts a domain Concierge model to ConciergeSummaryResponseDTO.
func ToConciergeSummaryResponseDTO(c *models.Concierge) dto.ConciergeSummaryResponseDTO {
	return dto.ConciergeSummaryResponseDTO{
		ID:          c.ID.Hex(),
		Name:        c.Name,
		Description: c.Description,
		Version:     c.Version,
		ETagVersion: c.ETagVersion,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// ToConciergeSummaryResponseDTOList converts a slice of domain Concierge models to a slice of ConciergeSummaryResponseDTO.
func ToConciergeSummaryResponseDTOList(concierges []models.Concierge) []dto.ConciergeSummaryResponseDTO {
	res := make([]dto.ConciergeSummaryResponseDTO, 0, len(concierges))
	for i := range concierges {
		res = append(res, ToConciergeSummaryResponseDTO(&concierges[i]))
	}
	return res
}
