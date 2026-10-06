package service

import (
	"errors"
	"fmt"
	"strings"
)

// Service errors form the application-level error contract consumed by the
// transport layer. Repository-specific sentinel errors are translated to this contract.
var (
	ErrConciergeVersionImmutable = errors.New("saved concierge versions are immutable")
	ErrConciergeVersionNotFound  = errors.New("concierge version not found")
	ErrConciergeNotFound         = errors.New("concierge not found")
	ErrConciergeNameExists       = errors.New("concierge name already exists")
	ErrConciergeETagMismatch     = errors.New("concierge etag does not match current representation")

	ErrAgentNotFound     = errors.New("agent not found")
	ErrAgentNameExists   = errors.New("agent name already exists")
	ErrAgentETagMismatch = errors.New("agent etag does not match current representation")

	ErrInstructionNotFound     = errors.New("instruction not found")
	ErrInstructionNameExists   = errors.New("instruction name already exists")
	ErrInstructionETagMismatch = errors.New("instruction etag does not match current representation")

	ErrToolNotFound     = errors.New("tool not found")
	ErrToolNameExists   = errors.New("tool name already exists")
	ErrToolETagMismatch = errors.New("tool etag does not match current representation")
)

// ErrReferencedToolsNotFound is returned when tools referenced in instruction content do not exist.
type ErrReferencedToolsNotFound struct {
	ToolIDs []string
}

func (e *ErrReferencedToolsNotFound) Error() string {
	return fmt.Sprintf("referenced tools do not exist: %s", strings.Join(e.ToolIDs, ", "))
}
