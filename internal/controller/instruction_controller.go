package controller

import (
	"net/http"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/etag"
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
		instructions.GET("/:instructionId", c.GetInstructionByID)
		instructions.GET("/:instructionId/usages", c.ListInstructionUsages)
		instructions.PUT("/:instructionId", c.UpdateInstruction)
	}
}

// ListInstructionUsages handles GET /api/v1/organisations/:organisationId/instructions/:instructionId/usages.
func (c *InstructionController) ListInstructionUsages(ctx *gin.Context) {
	usages, err := c.instructionService.ListInstructionUsages(ctx.Request.Context(), ctx.Param("instructionId"))
	if err != nil {
		writeServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": usages})
}

// CreateInstruction handles POST /api/v1/organisations/:organisationId/instructions
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
		writeServiceError(ctx, err, errorContext{nameValue: req.Name})
		return
	}

	ctx.Header("ETag", etag.Format(resp.UpdatedAt))
	ctx.JSON(http.StatusCreated, gin.H{
		"data": resp,
	})
}

// ListInstructions handles GET /api/v1/organisations/:organisationId/instructions
func (c *InstructionController) ListInstructions(ctx *gin.Context) {
	instructions, err := c.instructionService.ListInstructions(ctx.Request.Context())
	if err != nil {
		writeServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": instructions,
	})
}

// GetInstructionByID handles GET /api/v1/organisations/:organisationId/instructions/:instructionId
func (c *InstructionController) GetInstructionByID(ctx *gin.Context) {
	id := ctx.Param("instructionId")

	resp, err := c.instructionService.GetInstructionByID(ctx.Request.Context(), id)
	if err != nil {
		writeServiceError(ctx, err)
		return
	}

	ctx.Header("ETag", etag.Format(resp.UpdatedAt))
	ctx.JSON(http.StatusOK, gin.H{
		"data": resp,
	})
}

// UpdateInstruction handles PUT /api/v1/organisations/:organisationId/instructions/:instructionId
func (c *InstructionController) UpdateInstruction(ctx *gin.Context) {
	id := ctx.Param("instructionId")
	ifMatch := ctx.GetHeader("If-Match")
	if ifMatch == "" {
		ctx.JSON(http.StatusPreconditionRequired, gin.H{"error": "If-Match header is required"})
		return
	}

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

	resp, err := c.instructionService.UpdateInstruction(ctx.Request.Context(), id, ifMatch, req)
	if err != nil {
		writeServiceError(ctx, err, errorContext{nameValue: req.Name})
		return
	}

	ctx.Header("ETag", etag.Format(resp.UpdatedAt))
	ctx.JSON(http.StatusOK, gin.H{
		"data": resp,
	})
}
