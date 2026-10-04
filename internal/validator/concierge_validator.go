package validator

import (
	"github.com/bangweiz/dubhu-designer/internal/dto"
)

// ValidateCreateConcierge validates all business constraints for CreateConciergeDTO.
func ValidateCreateConcierge(req *dto.CreateConciergeDTO) ValidationErrors {
	var errs ValidationErrors

	if err := validateName("name", req.Name); err != nil {
		errs = append(errs, *err)
	}

	if err := validateDescription("description", req.Description); err != nil {
		errs = append(errs, *err)
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}
