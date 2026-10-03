package controller

import (
	"errors"
	"net/http"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/repository"
	"github.com/bangweiz/dubhu-designer/internal/service"
	"github.com/bangweiz/dubhu-designer/internal/validator"
	"github.com/gin-gonic/gin"
)

// InstructionController handles HTTP requests related to instructions.
type InstructionController struct {
	instructionService *service.InstructionService
}

// NewInstructionController creates a new InstructionController.
func NewInstructionController(instructionService *service.InstructionService) *InstructionController {
	return &InstructionController{
		instructionService: instructionService,
	}
}

// RegisterRoutes registers instruction endpoints onto the gin router/group.
func (c *InstructionController) RegisterRoutes(rg *gin.RouterGroup) {
	instructions := rg.Group("/instructions")
	{
		instructions.POST("", c.CreateInstruction)
		instructions.GET("", c.ListInstructions)
		instructions.GET("/:id", c.GetInstructionByID)
		instructions.PUT("/:id", c.UpdateInstruction)
	}
}

// CreateInstruction handles POST /api/v1/instructions
func (c *InstructionController) CreateInstruction(ctx *gin.Context) {
	var req dto.CreateInstructionDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	req.Trim()

	if errs := validator.ValidateCreateInstruction(&req); len(errs) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": errs,
		})
		return
	}

	resp, err := c.instructionService.CreateInstruction(ctx.Request.Context(), req)
	if err != nil {
		if errors.Is(err, repository.ErrInstructionNameExists) {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "Conflict",
				"details": []validator.FieldError{
					{
						Field:  "name",
						Reason: "instruction name already exists",
						Value:  req.Name,
					},
				},
			})
			return
		}
		var refErr *service.ErrReferencedToolsNotFound
		if errors.As(err, &refErr) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": refErr.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"data": resp,
	})
}

// ListInstructions handles GET /api/v1/instructions
func (c *InstructionController) ListInstructions(ctx *gin.Context) {
	instructions, err := c.instructionService.ListInstructions(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": instructions,
	})
}

// GetInstructionByID handles GET /api/v1/instructions/:id
func (c *InstructionController) GetInstructionByID(ctx *gin.Context) {
	id := ctx.Param("id")

	resp, err := c.instructionService.GetInstructionByID(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrInstructionNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "Instruction not found",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": resp,
	})
}

// UpdateInstruction handles PUT /api/v1/instructions/:id
func (c *InstructionController) UpdateInstruction(ctx *gin.Context) {
	id := ctx.Param("id")

	var req dto.UpdateInstructionDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	req.Trim()

	if errs := validator.ValidateUpdateInstruction(&req); len(errs) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": errs,
		})
		return
	}

	resp, err := c.instructionService.UpdateInstruction(ctx.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, service.ErrInstructionNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "Instruction not found",
			})
			return
		}
		if errors.Is(err, service.ErrInstructionVersionConflict) {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "Conflict",
				"details": []validator.FieldError{
					{
						Field:  "version",
						Reason: "instruction was modified by another operation (version mismatch)",
						Value:  req.Version,
					},
				},
			})
			return
		}
		if errors.Is(err, repository.ErrInstructionNameExists) {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "Conflict",
				"details": []validator.FieldError{
					{
						Field:  "name",
						Reason: "instruction name already exists",
						Value:  req.Name,
					},
				},
			})
			return
		}
		var refErr *service.ErrReferencedToolsNotFound
		if errors.As(err, &refErr) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": refErr.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": resp,
	})
}
