package dto

import (
	"strings"
	"time"

	"github.com/bangweiz/dubhu-designer/internal/models"
)

type RootAccountDTO struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (d *RootAccountDTO) Trim() {
	d.Name = strings.TrimSpace(d.Name)
	d.Email = strings.ToLower(strings.TrimSpace(d.Email))
}

type CreateOrganisationDTO struct {
	Name        string         `json:"name" binding:"required"`
	Description string         `json:"description" binding:"required"`
	RootAccount RootAccountDTO `json:"rootAccount" binding:"required"`
}

func (d *CreateOrganisationDTO) Trim() {
	d.Name = strings.TrimSpace(d.Name)
	d.Description = strings.TrimSpace(d.Description)
	d.RootAccount.Trim()
}

type CreateAccountDTO struct {
	Name     string             `json:"name" binding:"required"`
	Email    string             `json:"email" binding:"required"`
	Password string             `json:"password" binding:"required"`
	Role     models.AccountRole `json:"role" binding:"required"`
}

func (d *CreateAccountDTO) Trim() {
	d.Name = strings.TrimSpace(d.Name)
	d.Email = strings.ToLower(strings.TrimSpace(d.Email))
}

type LoginDTO struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (d *LoginDTO) Trim() { d.Email = strings.ToLower(strings.TrimSpace(d.Email)) }

type AccountResponseDTO struct {
	ID             string    `json:"id"`
	OrganisationID string    `json:"organisationId"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Role           string    `json:"role"`
	CreatedBy      string    `json:"createdBy"`
	UpdatedBy      string    `json:"updatedBy"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type OrganisationResponseDTO struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	RootAccountID string    `json:"rootAccountId"`
	CreatedBy     string    `json:"createdBy"`
	UpdatedBy     string    `json:"updatedBy"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type CreateOrganisationResponseDTO struct {
	Organisation OrganisationResponseDTO `json:"organisation"`
	RootAccount  AccountResponseDTO      `json:"rootAccount"`
}

type LoginResponseDTO struct {
	AccessToken string             `json:"accessToken"`
	TokenType   string             `json:"tokenType"`
	ExpiresAt   time.Time          `json:"expiresAt"`
	Account     AccountResponseDTO `json:"account"`
}
