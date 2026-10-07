package repository

import "errors"

// Repository errors describe persistence-layer outcomes. Services translate
// these errors before returning them to callers outside this package.
var (
	ErrOrganisationNameExists = errors.New("organisation name already exists")
	ErrAccountConflict        = errors.New("account already exists")
	ErrVariableNameExists     = errors.New("variable name already exists")
	ErrEnvironmentNameExists  = errors.New("environment name already exists")
	ErrConciergeNameExists    = errors.New("concierge name already exists")
	ErrConciergeNotFound      = errors.New("concierge not found")

	ErrAgentNameExists      = errors.New("agent name already exists")
	ErrAgentNotFound        = errors.New("agent not found")
	ErrAgentVersionConflict = errors.New("agent version conflict")

	ErrInstructionNameExists = errors.New("instruction name already exists")
	ErrInstructionNotFound   = errors.New("instruction not found")
	ErrInstructionConflict   = errors.New("instruction version conflict")

	ErrToolNameExists  = errors.New("tool name already exists")
	ErrVersionConflict = errors.New("version conflict")
)
