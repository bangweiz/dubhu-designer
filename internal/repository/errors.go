package repository

import "errors"

// Repository errors describe persistence-layer outcomes. Services translate
// these errors before returning them to callers outside this package.
var (
	ErrAgentLimitReached      = errors.New("a concierge can have at most 10 agents")
	ErrOrganisationNameExists = errors.New("organisation name already exists")
	ErrAccountConflict        = errors.New("account already exists")
	ErrVariableNameExists     = errors.New("variable name already exists")
	ErrEnvironmentNameExists  = errors.New("environment name already exists")
	ErrConciergeNameExists    = errors.New("concierge name already exists")
	ErrConciergeNotFound      = errors.New("concierge not found")

	ErrAgentNameExists     = errors.New("agent name already exists")
	ErrAgentNotFound       = errors.New("agent not found")
	ErrAgentUpdateConflict = errors.New("agent update conflict")

	ErrInstructionNameExists = errors.New("instruction name already exists")
	ErrInstructionNotFound   = errors.New("instruction not found")
	ErrInstructionConflict   = errors.New("instruction update conflict")

	ErrToolNameExists = errors.New("tool name already exists")
	ErrUpdateConflict = errors.New("update conflict")
)
