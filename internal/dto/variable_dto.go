package dto

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type CreateVariableDTO struct {
	Type        string `json:"type" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

func (d *CreateVariableDTO) Trim() {
	d.Name = strings.TrimSpace(d.Name)
	d.Description = strings.TrimSpace(d.Description)
}

// UpdateVariableDTO permits changes only to variable metadata.
type UpdateVariableDTO struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

func (d *UpdateVariableDTO) Trim() {
	d.Name = strings.TrimSpace(d.Name)
	d.Description = strings.TrimSpace(d.Description)
}

// Reject type explicitly, including null, rather than silently ignoring it.
func (d *UpdateVariableDTO) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	for field := range fields {
		if strings.EqualFold(field, "type") {
			return fmt.Errorf("variable type is immutable; omit type from update requests")
		}
	}

	type payload UpdateVariableDTO
	return json.Unmarshal(data, (*payload)(d))
}

type VariableResponseDTO struct {
	Type        string    `json:"type"`
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"createdBy"`
	UpdatedBy   string    `json:"updatedBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
