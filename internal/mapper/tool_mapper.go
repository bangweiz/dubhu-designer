package mapper

import (
	"time"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ToInitialToolEntity converts a pre-trimmed CreateToolDTO into a newly initialized Tool entity
// with initial version 1, a new ObjectID, and UTC timestamps.
func ToInitialToolEntity(input dto.CreateToolDTO) *models.Tool {
	now := time.Now().UTC()

	return &models.Tool{
		ID:          bson.NewObjectID(),
		Name:        input.Name,
		Description: input.Description,
		Version:     1,
		Inputs:      ToToolInputs(input.Inputs),
		Outputs:     ToToolOutputs(input.Outputs),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// ToToolInputs converts a slice of ToolInputDTO into models.ToolInput.
func ToToolInputs(inputs []dto.ToolInputDTO) []models.ToolInput {
	res := make([]models.ToolInput, 0, len(inputs))
	for _, in := range inputs {
		res = append(res, models.ToolInput{
			Name:        in.Name,
			Description: in.Description,
			Required:    in.Required,
		})
	}
	return res
}

// ToToolOutputs converts a slice of ToolOutputDTO into models.ToolOutput.
func ToToolOutputs(outputs []dto.ToolOutputDTO) []models.ToolOutput {
	res := make([]models.ToolOutput, 0, len(outputs))
	for _, out := range outputs {
		res = append(res, models.ToolOutput{
			Name:        out.Name,
			Description: out.Description,
		})
	}
	return res
}

// ToToolResponseDTO converts a domain Tool model to ToolResponseDTO,
// ensuring slices are initialized to empty arrays instead of nil.
func ToToolResponseDTO(tool *models.Tool) dto.ToolResponseDTO {
	inputs := make([]dto.ToolInputResponseDTO, 0, len(tool.Inputs))
	for _, in := range tool.Inputs {
		inputs = append(inputs, dto.ToolInputResponseDTO{
			Name:        in.Name,
			Description: in.Description,
			Required:    in.Required,
		})
	}

	outputs := make([]dto.ToolOutputResponseDTO, 0, len(tool.Outputs))
	for _, out := range tool.Outputs {
		outputs = append(outputs, dto.ToolOutputResponseDTO{
			Name:        out.Name,
			Description: out.Description,
		})
	}

	return dto.ToolResponseDTO{
		ID:          tool.ID.Hex(),
		Name:        tool.Name,
		Description: tool.Description,
		Inputs:      inputs,
		Outputs:     outputs,
		Version:     tool.Version,
		CreatedAt:   tool.CreatedAt,
		UpdatedAt:   tool.UpdatedAt,
	}
}

// ToToolResponseDTOList converts a slice of domain Tool models to a slice of ToolResponseDTO,
// guaranteeing a non-nil slice (serialized as [] instead of null in JSON).
func ToToolResponseDTOList(tools []models.Tool) []dto.ToolResponseDTO {
	res := make([]dto.ToolResponseDTO, 0, len(tools))
	for i := range tools {
		res = append(res, ToToolResponseDTO(&tools[i]))
	}
	return res
}
