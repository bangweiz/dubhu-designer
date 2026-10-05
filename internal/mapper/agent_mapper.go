package mapper

import (
	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ToInitialAgentEntity converts a CreateAgentDTO and concierge ObjectID into a new Agent entity with version 1.
func ToInitialAgentEntity(conciergeID bson.ObjectID, input dto.CreateAgentDTO) *models.Agent {
	return &models.Agent{
		ID:           bson.NewObjectID(),
		ConciergeID:  conciergeID,
		Name:         input.Name,
		Description:  input.Description,
		Goal:         input.Goal,
		Model:        models.Model(input.Model),
		Instructions: []bson.ObjectID{},
		Tools:        []bson.ObjectID{},
		Version:      1,
	}
}

// ToAgentResponseDTO converts a domain Agent model and its resolved instructions to AgentResponseDTO.
func ToAgentResponseDTO(a *models.Agent, instructions []models.Instruction, tools []models.Tool) dto.AgentResponseDTO {
	instructionMap := make(map[bson.ObjectID]dto.InstructionSummaryResponseDTO, len(instructions))
	for i := range instructions {
		instructionMap[instructions[i].ID] = ToInstructionSummaryResponseDTO(&instructions[i])
	}

	instructionDTOs := make([]dto.InstructionSummaryResponseDTO, 0, len(a.Instructions))
	for _, instructionID := range a.Instructions {
		if instruction, ok := instructionMap[instructionID]; ok {
			instructionDTOs = append(instructionDTOs, instruction)
		}
	}

	toolMap := make(map[bson.ObjectID]dto.ToolResponseDTO, len(tools))
	for i := range tools {
		toolMap[tools[i].ID] = ToToolResponseDTO(&tools[i])
	}

	toolDTOs := make([]dto.ToolResponseDTO, 0, len(a.Tools))
	for _, toolID := range a.Tools {
		if tool, ok := toolMap[toolID]; ok {
			toolDTOs = append(toolDTOs, tool)
		}
	}

	return dto.AgentResponseDTO{
		ID:           a.ID.Hex(),
		ConciergeID:  a.ConciergeID.Hex(),
		Name:         a.Name,
		Description:  a.Description,
		Goal:         a.Goal,
		Model:        string(a.Model),
		Instructions: instructionDTOs,
		Tools:        toolDTOs,
		Version:      a.Version,
	}
}

// ToAgentSummaryResponseDTO converts a domain Agent model to AgentSummaryResponseDTO.
func ToAgentSummaryResponseDTO(a *models.Agent) dto.AgentSummaryResponseDTO {
	return dto.AgentSummaryResponseDTO{
		ID:          a.ID.Hex(),
		ConciergeID: a.ConciergeID.Hex(),
		Name:        a.Name,
		Description: a.Description,
		Goal:        a.Goal,
		Model:       string(a.Model),
	}
}

// ToAgentSummaryResponseDTOList converts Agent models to summaries without assigned instructions.
func ToAgentSummaryResponseDTOList(agents []models.Agent) []dto.AgentSummaryResponseDTO {
	res := make([]dto.AgentSummaryResponseDTO, 0, len(agents))
	for i := range agents {
		res = append(res, ToAgentSummaryResponseDTO(&agents[i]))
	}
	return res
}
