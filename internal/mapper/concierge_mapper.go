package mapper

import (
	"time"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func ToInitialConciergeEntity(input dto.CreateConciergeDTO) *models.Concierge {
	now := time.Now().UTC().Truncate(time.Millisecond)
	return &models.Concierge{ID: bson.NewObjectID(), Name: input.Name, Description: input.Description, NextVersion: 1, Agents: []models.Agent{}, ConciergeVersions: []models.ConciergeVersionReference{}, AuditFields: models.AuditFields{CreatedAt: now, UpdatedAt: now}}
}
func ToConciergeResponseDTO(c *models.Concierge) dto.ConciergeResponseDTO {
	refs := make([]dto.ConciergeVersionDescriptorDTO, 0, len(c.ConciergeVersions))
	for _, v := range c.ConciergeVersions {
		refs = append(refs, dto.ConciergeVersionDescriptorDTO{ConciergeVersionID: v.ConciergeVersionID.Hex(), Version: v.Version})
	}
	return dto.ConciergeResponseDTO{Agents: ToAgentSummaryResponseDTOList(c.Agents), NextVersion: c.NextVersion, ID: c.ID.Hex(), Name: c.Name, Description: c.Description, CreatedBy: auditID(c.CreatedBy), UpdatedBy: auditID(c.UpdatedBy), CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt, ConciergeVersions: refs}
}

// ToSavedConciergeVersionResponseDTO maps the complete snapshot without exposing models.
func ToSavedConciergeVersionResponseDTO(saved *models.SavedConciergeVersion) dto.SavedConciergeVersionResponseDTO {
	agents := make([]dto.SavedAgentResponseDTO, 0, len(saved.Agents))
	for _, agent := range saved.Agents {
		agents = append(agents, dto.SavedAgentResponseDTO{
			ID: agent.ID.Hex(), ConciergeID: auditID(agent.ConciergeID),
			Name: agent.Name, Description: agent.Description, Goal: agent.Goal, Model: string(agent.Model),
			Instructions: objectIDStrings(agent.Instructions), Tools: objectIDStrings(agent.Tools),
			CreatedBy: auditID(agent.CreatedBy), UpdatedBy: auditID(agent.UpdatedBy),
			CreatedAt: agent.CreatedAt, UpdatedAt: agent.UpdatedAt,
		})
	}
	instructions := make([]dto.SavedInstructionResponseDTO, 0, len(saved.Instructions))
	for _, instruction := range saved.Instructions {
		instructions = append(instructions, dto.SavedInstructionResponseDTO{
			ID: instruction.ID.Hex(), Name: instruction.Name, Content: instruction.Content,
			Tools: objectIDStrings(instruction.Tools), CreatedBy: auditID(instruction.CreatedBy),
			UpdatedBy: auditID(instruction.UpdatedBy), CreatedAt: instruction.CreatedAt, UpdatedAt: instruction.UpdatedAt,
		})
	}
	return dto.SavedConciergeVersionResponseDTO{
		ID: saved.ID.Hex(), ConciergeID: saved.ConciergeID.Hex(), Name: saved.Name, Description: saved.Description,
		Version: saved.Version, Agents: agents, Instructions: instructions, Tools: ToToolResponseDTOList(saved.Tools),
		CreatedBy: auditID(saved.CreatedBy), UpdatedBy: auditID(saved.UpdatedBy), CreatedAt: saved.CreatedAt, UpdatedAt: saved.UpdatedAt,
	}
}

func objectIDStrings(ids []bson.ObjectID) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		result = append(result, id.Hex())
	}
	return result
}
