package mapper

import (
	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
)

func ToAccountResponseDTO(a *models.Account) dto.AccountResponseDTO {
	return dto.AccountResponseDTO{ID: a.ID.Hex(), OrganisationID: a.OrganisationID.Hex(), Name: a.Name, Email: a.Email, Role: string(a.Role), CreatedBy: auditID(a.CreatedBy), UpdatedBy: auditID(a.UpdatedBy), CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt}
}

func ToOrganisationResponseDTO(o *models.Organisation) dto.OrganisationResponseDTO {
	return dto.OrganisationResponseDTO{ID: o.ID.Hex(), Name: o.Name, Description: o.Description, RootAccountID: o.RootAccountID.Hex(), CreatedBy: auditID(o.CreatedBy), UpdatedBy: auditID(o.UpdatedBy), CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt}
}
