package validator

import (
	"github.com/bangweiz/dubhu-designer/internal/dto"
)

// ValidateCreateInstruction validates all business constraints for CreateInstructionDTO.
func ValidateCreateInstruction(req *dto.CreateInstructionDTO) ValidationErrors {
	var errs ValidationErrors

	if err := validateName("name", req.Name); err != nil {
		errs = append(errs, *err)
	}

	if err := validateContent("content", req.Content); err != nil {
		errs = append(errs, *err)
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

// ValidateUpdateInstruction validates all business constraints for UpdateInstructionDTO.
func ValidateUpdateInstruction(req *dto.UpdateInstructionDTO) ValidationErrors {
	var errs ValidationErrors

	if err := validateName("name", req.Name); err != nil {
		errs = append(errs, *err)
	}

	if err := validateContent("content", req.Content); err != nil {
		errs = append(errs, *err)
	}

	if req.Version <= 0 {
		errs = append(errs, FieldError{
			Field:  "version",
			Reason: "version must be a positive number greater than 0",
			Value:  req.Version,
		})
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}
