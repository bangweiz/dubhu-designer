package mapper

import (
	"time"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func ToInitialVariableEntity(input dto.CreateVariableDTO) *models.Variable {
	now := time.Now().UTC().Truncate(time.Millisecond)
	return &models.Variable{ID: bson.NewObjectID(), Name: input.Name, Description: input.Description, Version: 1, CreatedAt: now, UpdatedAt: now}
}

func ToVariableResponseDTO(e *models.Variable) dto.VariableResponseDTO {
	return dto.VariableResponseDTO{ID: e.ID.Hex(), Name: e.Name, Description: e.Description, Version: e.Version, CreatedBy: auditID(e.CreatedBy), UpdatedBy: auditID(e.UpdatedBy), CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt}
}
