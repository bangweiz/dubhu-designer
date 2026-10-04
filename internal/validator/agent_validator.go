package validator

import (
	"github.com/bangweiz/dubhu-designer/internal/dto"
)

// ValidateCreateAgent validates all business constraints for CreateAgentDTO.
func ValidateCreateAgent(req *dto.CreateAgentDTO) ValidationErrors {
	var errs ValidationErrors

	if err := validateName("name", req.Name); err != nil {
		errs = append(errs, *err)
	}

	if err := validateDescription("description", req.Description); err != nil {
		errs = append(errs, *err)
	}

	if err := validateModel("model", req.Model); err != nil {
		errs = append(errs, *err)
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}
