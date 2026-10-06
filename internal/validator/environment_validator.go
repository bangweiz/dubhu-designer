package validator

import "github.com/bangweiz/dubhu-designer/internal/dto"

func ValidateCreateEnvironment(req *dto.CreateEnvironmentDTO) ValidationErrors {
	var errs ValidationErrors
	if err := validateName("name", req.Name); err != nil {
		errs = append(errs, *err)
	}
	if err := validateDescription("description", req.Description); err != nil {
		errs = append(errs, *err)
	}
	return errs
}

func ValidateUpdateEnvironment(req *dto.UpdateEnvironmentDTO) ValidationErrors {
	return ValidateCreateEnvironment(req)
}
