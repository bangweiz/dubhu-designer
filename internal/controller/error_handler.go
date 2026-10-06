package controller

import (
	"errors"
	"log"
	"net/http"

	"github.com/bangweiz/dubhu-designer/internal/service"
	"github.com/bangweiz/dubhu-designer/internal/validator"
	"github.com/gin-gonic/gin"
)

type errorContext struct {
	nameValue string
}

// writeServiceError is the single mapping from the service error contract to
// the HTTP API. Optional context supplies request values for structured errors.
func writeServiceError(ctx *gin.Context, err error, opts ...errorContext) {
	var meta errorContext
	if len(opts) > 0 {
		meta = opts[0]
	}

	var referencedToolsErr *service.ErrReferencedToolsNotFound
	switch {
	case errors.Is(err, service.ErrConciergeETagMismatch):
		ctx.JSON(http.StatusPreconditionFailed, gin.H{"error": "Concierge was modified; fetch the latest representation and retry"})
	case errors.Is(err, service.ErrConciergeVersionImmutable):
		ctx.JSON(http.StatusConflict, gin.H{"error": "Saved concierge versions are immutable"})
	case errors.Is(err, service.ErrConciergeVersionNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Concierge version not found"})
	case errors.Is(err, service.ErrConciergeNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Concierge not found"})
	case errors.Is(err, service.ErrAgentNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
	case errors.Is(err, service.ErrInstructionNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Instruction not found"})
	case errors.Is(err, service.ErrToolNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Tool not found"})
	case errors.Is(err, service.ErrAgentETagMismatch):
		ctx.JSON(http.StatusPreconditionFailed, gin.H{"error": "Agent was modified; fetch the latest representation and retry"})
	case errors.Is(err, service.ErrInstructionETagMismatch):
		ctx.JSON(http.StatusPreconditionFailed, gin.H{"error": "Instruction was modified; fetch the latest representation and retry"})
	case errors.Is(err, service.ErrToolETagMismatch):
		ctx.JSON(http.StatusPreconditionFailed, gin.H{"error": "Tool was modified; fetch the latest representation and retry"})
	case errors.Is(err, service.ErrConciergeNameExists):
		writeNameConflict(ctx, "concierge name already exists", meta.nameValue)
	case errors.Is(err, service.ErrAgentNameExists):
		writeNameConflict(ctx, "agent name already exists for this concierge", meta.nameValue)
	case errors.Is(err, service.ErrInstructionNameExists):
		writeNameConflict(ctx, "instruction name already exists", meta.nameValue)
	case errors.Is(err, service.ErrToolNameExists):
		writeNameConflict(ctx, "tool name already exists", meta.nameValue)
	case errors.As(err, &referencedToolsErr):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": referencedToolsErr.Error()})
	default:
		log.Printf("unhandled service error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}

func writeNameConflict(ctx *gin.Context, reason, value string) {
	ctx.JSON(http.StatusConflict, gin.H{
		"error": "Conflict",
		"details": []validator.FieldError{{
			Field:  "name",
			Reason: reason,
			Value:  value,
		}},
	})
}
