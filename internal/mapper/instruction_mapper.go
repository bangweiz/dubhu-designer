package mapper

import (
	"time"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func ToInstructionUsageResponseDTOList(usages []models.InstructionUsage) []dto.InstructionUsageResponseDTO {
	result := make([]dto.InstructionUsageResponseDTO, 0, len(usages))
	for _, usage := range usages {
		result = append(result, dto.InstructionUsageResponseDTO{
			ConciergeID: usage.ConciergeID.Hex(), ConciergeName: usage.ConciergeName,
			ConciergeVersionID: usage.ConciergeVersionID.Hex(), Version: usage.Version,
			AgentID: usage.AgentID.Hex(), AgentName: usage.AgentName,
		})
	}
	return result
}

// ToInitialInstructionEntity converts a CreateInstructionDTO into a new Instruction entity
// with version 1, a new ObjectID, and UTC timestamps.
func ToInitialInstructionEntity(input dto.CreateInstructionDTO, toolIDs []bson.ObjectID) *models.Instruction {
	now := time.Now().UTC()
	if toolIDs == nil {
		toolIDs = []bson.ObjectID{}
	}

	return &models.Instruction{
		ID:        bson.NewObjectID(),
		Name:      input.Name,
		Content:   input.Content,
		Tools:     toolIDs,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ToInstructionSummaryResponseDTO converts a domain Instruction model to InstructionSummaryResponseDTO.
func ToInstructionSummaryResponseDTO(inst *models.Instruction) dto.InstructionSummaryResponseDTO {
	return dto.InstructionSummaryResponseDTO{
		ID:        inst.ID.Hex(),
		Name:      inst.Name,
		Content:   inst.Content,
		Version:   inst.Version,
		CreatedAt: inst.CreatedAt,
		UpdatedAt: inst.UpdatedAt,
	}
}

// ToInstructionSummaryResponseDTOList converts a slice of domain Instruction models to a slice of InstructionSummaryResponseDTO.
func ToInstructionSummaryResponseDTOList(instructions []models.Instruction) []dto.InstructionSummaryResponseDTO {
	res := make([]dto.InstructionSummaryResponseDTO, 0, len(instructions))
	for i := range instructions {
		res = append(res, ToInstructionSummaryResponseDTO(&instructions[i]))
	}
	return res
}

// ToInstructionResponseDTO converts a domain Instruction model and its resolved tools to InstructionResponseDTO,
// ensuring Tools is always an initialized array of ToolResponseDTO ([] if empty) ordered by instruction.Tools.
func ToInstructionResponseDTO(inst *models.Instruction, tools []models.Tool) dto.InstructionResponseDTO {
	toolMap := make(map[bson.ObjectID]dto.ToolResponseDTO, len(tools))
	for i := range tools {
		toolMap[tools[i].ID] = ToToolResponseDTO(&tools[i])
	}

	toolDTOs := make([]dto.ToolResponseDTO, 0, len(inst.Tools))
	for _, id := range inst.Tools {
		if t, ok := toolMap[id]; ok {
			toolDTOs = append(toolDTOs, t)
		}
	}

	return dto.InstructionResponseDTO{
		ID:        inst.ID.Hex(),
		Name:      inst.Name,
		Content:   inst.Content,
		Tools:     toolDTOs,
		Version:   inst.Version,
		CreatedAt: inst.CreatedAt,
		UpdatedAt: inst.UpdatedAt,
	}
}

// ToPopulatedInstructionResponseDTO converts a domain PopulatedInstruction model to InstructionResponseDTO,
// ensuring Tools is ordered by the instruction's Tools slice.
func ToPopulatedInstructionResponseDTO(populated *models.PopulatedInstruction) dto.InstructionResponseDTO {
	return ToInstructionResponseDTO(&populated.Instruction, populated.ResolvedTools)
}
