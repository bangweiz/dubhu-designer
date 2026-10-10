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

	if err := validateGoal("goal", req.Goal); err != nil {
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

// ValidateUpdateAgent validates all business constraints for UpdateAgentDTO.
func ValidateUpdateAgent(req *dto.UpdateAgentDTO) ValidationErrors {
	var errs ValidationErrors

	if err := validateName("name", req.Name); err != nil {
		errs = append(errs, *err)
	}

	if err := validateDescription("description", req.Description); err != nil {
		errs = append(errs, *err)
	}

	if err := validateGoal("goal", req.Goal); err != nil {
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
