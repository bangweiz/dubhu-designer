package controller

import (
	"net/http"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/etag"
	"github.com/bangweiz/dubhu-designer/internal/service"
	"github.com/bangweiz/dubhu-designer/internal/validator"
	"github.com/gin-gonic/gin"
)

// ToolController handles HTTP requests related to tools.
type ToolController struct {
	toolService *service.ToolService
}

// NewToolController creates a new ToolController.
func NewToolController(toolService *service.ToolService) *ToolController {
	return &ToolController{
		toolService: toolService,
	}
}

// RegisterRoutes registers the tool endpoints onto the gin router/group.
func (c *ToolController) RegisterRoutes(rg *gin.RouterGroup) {
	tools := rg.Group("/tools")
	{
		tools.POST("", c.CreateTool)
		tools.GET("", c.ListTools)
		tools.GET("/:toolId", c.GetToolByID)
		tools.GET("/:toolId/usages", c.ListToolUsages)
		tools.PUT("/:toolId", c.UpdateTool)
	}
}

// ListToolUsages handles GET /api/v1/organisations/:organisationId/tools/:toolId/usages.
func (c *ToolController) ListToolUsages(ctx *gin.Context) {
	usages, err := c.toolService.ListToolUsages(ctx.Request.Context(), ctx.Param("toolId"))
	if err != nil {
		writeServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": usages})
}

// CreateTool handles POST /api/v1/organisations/:organisationId/tools
func (c *ToolController) CreateTool(ctx *gin.Context) {
	var req dto.CreateToolDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	req.Trim()

	if errs := validator.ValidateCreateTool(&req); len(errs) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": errs,
		})
		return
	}

	toolResponse, err := c.toolService.CreateTool(ctx.Request.Context(), req)
	if err != nil {
		writeServiceError(ctx, err, errorContext{nameValue: req.Name})
		return
	}

	ctx.Header("ETag", etag.Format(toolResponse.UpdatedAt))
	ctx.JSON(http.StatusCreated, gin.H{
		"data": toolResponse,
	})
}

// ListTools handles GET /api/v1/organisations/:organisationId/tools
func (c *ToolController) ListTools(ctx *gin.Context) {
	tools, err := c.toolService.ListTools(ctx.Request.Context())
	if err != nil {
		writeServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": tools,
	})
}

// GetToolByID handles GET /api/v1/organisations/:organisationId/tools/:toolId
func (c *ToolController) GetToolByID(ctx *gin.Context) {
	id := ctx.Param("toolId")

	toolResponse, err := c.toolService.GetToolByID(ctx.Request.Context(), id)
	if err != nil {
		writeServiceError(ctx, err)
		return
	}

	ctx.Header("ETag", etag.Format(toolResponse.UpdatedAt))
	ctx.JSON(http.StatusOK, gin.H{
		"data": toolResponse,
	})
}

// UpdateTool handles PUT /api/v1/organisations/:organisationId/tools/:toolId
func (c *ToolController) UpdateTool(ctx *gin.Context) {
	id := ctx.Param("toolId")
	ifMatch := ctx.GetHeader("If-Match")
	if ifMatch == "" {
		ctx.JSON(http.StatusPreconditionRequired, gin.H{"error": "If-Match header is required"})
		return
	}

	var req dto.UpdateToolDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	req.Trim()

	if errs := validator.ValidateUpdateTool(&req); len(errs) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": errs,
		})
		return
	}

	toolResponse, err := c.toolService.UpdateTool(ctx.Request.Context(), id, ifMatch, req)
	if err != nil {
		writeServiceError(ctx, err, errorContext{nameValue: req.Name})
		return
	}

	ctx.Header("ETag", etag.Format(toolResponse.UpdatedAt))
	ctx.JSON(http.StatusOK, gin.H{
		"data": toolResponse,
	})
}
