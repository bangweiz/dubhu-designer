package mapper

import (
	"time"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func ToInitialEnvironmentEntity(input dto.CreateEnvironmentDTO) *models.Environment {
	now := time.Now().UTC().Truncate(time.Millisecond).Truncate(time.Millisecond)
	return &models.Environment{
		ID:          bson.NewObjectID(),
		Name:        input.Name,
		Description: input.Description,
		AuditFields: models.AuditFields{CreatedAt: now, UpdatedAt: now},
	}
}

func ToEnvironmentResponseDTO(e *models.Environment) dto.EnvironmentResponseDTO {
	return dto.EnvironmentResponseDTO{
		ID:          e.ID.Hex(),
		Name:        e.Name,
		Description: e.Description,
		CreatedBy:   auditID(e.CreatedBy),
		UpdatedBy:   auditID(e.UpdatedBy),
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}
