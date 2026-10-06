package validator

import "github.com/bangweiz/dubhu-designer/internal/dto"

func ValidateCreateVariable(req *dto.CreateVariableDTO) ValidationErrors {
	var errs ValidationErrors
	if err := validateName("name", req.Name); err != nil {
		errs = append(errs, *err)
	}
	if err := validateDescription("description", req.Description); err != nil {
		errs = append(errs, *err)
	}
	return errs
}

func ValidateUpdateVariable(req *dto.UpdateVariableDTO) ValidationErrors {
	return ValidateCreateVariable(req)
}
