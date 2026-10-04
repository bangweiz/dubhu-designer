package mapper

import (
	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ToInitialAgentEntity converts a CreateAgentDTO and concierge ObjectID into a new Agent entity with version 1.
func ToInitialAgentEntity(conciergeID bson.ObjectID, input dto.CreateAgentDTO) *models.Agent {
	return &models.Agent{
		ID:          bson.NewObjectID(),
		ConciergeID: conciergeID,
		Name:        input.Name,
		Description: input.Description,
		Model:       models.Model(input.Model),
		Version:     1,
	}
}

// ToAgentResponseDTO converts a domain Agent model to AgentResponseDTO.
func ToAgentResponseDTO(a *models.Agent) dto.AgentResponseDTO {
	return dto.AgentResponseDTO{
		ID:          a.ID.Hex(),
		ConciergeID: a.ConciergeID.Hex(),
		Name:        a.Name,
		Description: a.Description,
		Model:       string(a.Model),
		Version:     a.Version,
	}
}

// ToAgentResponseDTOList converts a slice of domain Agent models to a slice of AgentResponseDTO.
func ToAgentResponseDTOList(agents []models.Agent) []dto.AgentResponseDTO {
	res := make([]dto.AgentResponseDTO, 0, len(agents))
	for i := range agents {
		res = append(res, ToAgentResponseDTO(&agents[i]))
	}
	return res
}
