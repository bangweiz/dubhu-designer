package mapper

import (
	"time"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func ToInitialVariableEntity(input dto.CreateVariableDTO) *models.Variable {
	now := time.Now().UTC().Truncate(time.Millisecond)
	return &models.Variable{
		Type:        models.VariableType(input.Type),
		ID:          bson.NewObjectID(),
		Name:        input.Name,
		Description: input.Description,
		AuditFields: models.AuditFields{CreatedAt: now, UpdatedAt: now},
	}
}

func ToVariableResponseDTO(e *models.Variable) dto.VariableResponseDTO {
	return dto.VariableResponseDTO{
		Type:        string(e.EffectiveType()),
		ID:          e.ID.Hex(),
		Name:        e.Name,
		Description: e.Description,
		CreatedBy:   auditID(e.CreatedBy),
		UpdatedBy:   auditID(e.UpdatedBy),
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}
