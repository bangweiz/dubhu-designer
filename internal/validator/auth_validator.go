package validator

import (
	"net/mail"
	"unicode/utf8"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
)

func validateEmail(field, email string) *FieldError {
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || len(email) > 254 {
		return &FieldError{Field: field, Reason: "must be a valid email address", Value: email}
	}

	return nil
}

func validatePassword(field, password string) *FieldError {
	if utf8.RuneCountInString(password) < 15 || len(password) > 72 {
		return &FieldError{
			Field:  field,
			Reason: "must contain at least 15 characters and at most 72 bytes",
			Value:  "[redacted]",
		}
	}

	return nil
}

func ValidateCreateOrganisation(req *dto.CreateOrganisationDTO) ValidationErrors {
	var errs ValidationErrors
	for _, err := range []*FieldError{
		validateName("name", req.Name),
		validateDescription("description", req.Description),
		validateName("rootAccount.name", req.RootAccount.Name),
		validateEmail("rootAccount.email", req.RootAccount.Email),
		validatePassword("rootAccount.password", req.RootAccount.Password),
	} {
		if err != nil {
			errs = append(errs, *err)
		}
	}

	return errs
}

func ValidateCreateAccount(req *dto.CreateAccountDTO) ValidationErrors {
	var errs ValidationErrors
	for _, err := range []*FieldError{
		validateName("name", req.Name),
		validateEmail("email", req.Email),
		validatePassword("password", req.Password),
	} {
		if err != nil {
			errs = append(errs, *err)
		}
	}
	if req.Role != models.RoleAdmin && req.Role != models.RoleUser {
		errs = append(errs, FieldError{Field: "role", Reason: "must be admin or user", Value: req.Role})
	}

	return errs
}

func ValidateLogin(req *dto.LoginDTO) ValidationErrors {
	var errs ValidationErrors
	if err := validateEmail("email", req.Email); err != nil {
		errs = append(errs, *err)
	}
	if len(req.Password) > 72 {
		errs = append(errs, FieldError{Field: "password", Reason: "must not exceed 72 bytes", Value: "[redacted]"})
	}

	return errs
}
