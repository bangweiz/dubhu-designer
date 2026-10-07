package mapper

import (
	"time"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func ToInitialConciergeEntity(input dto.CreateConciergeDTO) *models.Concierge {
	now := time.Now().UTC()
	return &models.Concierge{ID: bson.NewObjectID(), Name: input.Name, Description: input.Description, Version: 1, CreatedAt: now, UpdatedAt: now}
}
func ToConciergeResponseDTO(c *models.Concierge) dto.ConciergeResponseDTO {
	refs := make([]dto.ConciergeVersionDescriptorDTO, 0, len(c.ConciergeVersions))
	for _, v := range c.ConciergeVersions {
		refs = append(refs, dto.ConciergeVersionDescriptorDTO{ConciergeVersionID: v.ConciergeVersionID.Hex(), Version: v.Version})
	}
	return dto.ConciergeResponseDTO{ID: c.ID.Hex(), Name: c.Name, Description: c.Description, Version: c.Version, CreatedBy: auditID(c.CreatedBy), UpdatedBy: auditID(c.UpdatedBy), CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt, ConciergeVersions: refs}
}
