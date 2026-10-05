package validator

import (
	"fmt"

	"github.com/bangweiz/dubhu-designer/internal/dto"
)

// validateToolInput validates an individual tool input parameter,
// ensuring its name and description satisfy length and non-blank rules.
func validateToolInput(prefix string, in dto.ToolInputDTO) ValidationErrors {
	var errs ValidationErrors

	if err := validateName(fmt.Sprintf("%s.name", prefix), in.Name); err != nil {
		errs = append(errs, *err)
	}
	if err := validateDescription(fmt.Sprintf("%s.description", prefix), in.Description); err != nil {
		errs = append(errs, *err)
	}

	return errs
}

// validateToolInputs validates a slice of tool input parameters.
func validateToolInputs(inputs []dto.ToolInputDTO) ValidationErrors {
	var errs ValidationErrors
	for i, in := range inputs {
		errs = append(errs, validateToolInput(fmt.Sprintf("inputs[%d]", i), in)...)
	}
	return errs
}

// validateToolOutput validates an individual tool output parameter,
// ensuring its name and description satisfy length and non-blank rules.
func validateToolOutput(prefix string, out dto.ToolOutputDTO) ValidationErrors {
	var errs ValidationErrors

	if err := validateName(fmt.Sprintf("%s.name", prefix), out.Name); err != nil {
		errs = append(errs, *err)
	}
	if err := validateDescription(fmt.Sprintf("%s.description", prefix), out.Description); err != nil {
		errs = append(errs, *err)
	}

	return errs
}

// validateToolOutputs validates a slice of tool output parameters.
func validateToolOutputs(outputs []dto.ToolOutputDTO) ValidationErrors {
	var errs ValidationErrors
	for i, out := range outputs {
		errs = append(errs, validateToolOutput(fmt.Sprintf("outputs[%d]", i), out)...)
	}
	return errs
}

// ValidateCreateTool validates all business constraints for CreateToolDTO.
func ValidateCreateTool(req *dto.CreateToolDTO) ValidationErrors {
	var errs ValidationErrors

	if err := validateName("name", req.Name); err != nil {
		errs = append(errs, *err)
	}

	if err := validateDescription("description", req.Description); err != nil {
		errs = append(errs, *err)
	}

	errs = append(errs, validateToolInputs(req.Inputs)...)
	errs = append(errs, validateToolOutputs(req.Outputs)...)

	if len(errs) == 0 {
		return nil
	}
	return errs
}

// ValidateUpdateTool validates all business constraints for UpdateToolDTO.
func ValidateUpdateTool(req *dto.UpdateToolDTO) ValidationErrors {
	var errs ValidationErrors

	if err := validateName("name", req.Name); err != nil {
		errs = append(errs, *err)
	}

	if err := validateDescription("description", req.Description); err != nil {
		errs = append(errs, *err)
	}

	errs = append(errs, validateToolInputs(req.Inputs)...)
	errs = append(errs, validateToolOutputs(req.Outputs)...)

	if len(errs) == 0 {
		return nil
	}
	return errs
}
