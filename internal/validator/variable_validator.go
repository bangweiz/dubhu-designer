package validator

import (
	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
)

func ValidateCreateVariable(req *dto.CreateVariableDTO) ValidationErrors {
	errs := validateVariableMetadata(req.Name, req.Description)
	if !models.VariableType(req.Type).IsValid() {
		errs = append(errs, FieldError{Field: "type", Reason: "must be string, number, or bool", Value: req.Type})
	}
	return errs
}

func ValidateUpdateVariable(req *dto.UpdateVariableDTO) ValidationErrors {
	return validateVariableMetadata(req.Name, req.Description)
}

func validateVariableMetadata(name, description string) ValidationErrors {
	var errs ValidationErrors
	if err := validateName("name", name); err != nil {
		errs = append(errs, *err)
	}
	if err := validateDescription("description", description); err != nil {
		errs = append(errs, *err)
	}
	return errs
}
