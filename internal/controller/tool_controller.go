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
		tools.GET("/:id", c.GetToolByID)
		tools.PUT("/:id", c.UpdateTool)
	}
}

// CreateTool handles POST /api/v1/tools
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
		if errors.Is(err, repository.ErrToolNameExists) {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "Conflict",
				"details": []validator.FieldError{
					{
						Field:  "name",
						Reason: "tool name already exists",
						Value:  req.Name,
					},
				},
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"data": toolResponse,
	})
}

// ListTools handles GET /api/v1/tools
func (c *ToolController) ListTools(ctx *gin.Context) {
	tools, err := c.toolService.ListTools(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": tools,
	})
}

// GetToolByID handles GET /api/v1/tools/:id
func (c *ToolController) GetToolByID(ctx *gin.Context) {
	id := ctx.Param("id")

	toolResponse, err := c.toolService.GetToolByID(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrToolNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "Tool not found",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": toolResponse,
	})
}

// UpdateTool handles PUT /api/v1/tools/:id
func (c *ToolController) UpdateTool(ctx *gin.Context) {
	id := ctx.Param("id")

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

	toolResponse, err := c.toolService.UpdateTool(ctx.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, service.ErrToolNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "Tool not found",
			})
			return
		}
		if errors.Is(err, service.ErrToolVersionConflict) {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "Conflict",
				"details": []validator.FieldError{
					{
						Field:  "version",
						Reason: "tool was modified by another operation (version mismatch)",
						Value:  req.Version,
					},
				},
			})
			return
		}
		if errors.Is(err, repository.ErrToolNameExists) {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "Conflict",
				"details": []validator.FieldError{
					{
						Field:  "name",
						Reason: "tool name already exists",
						Value:  req.Name,
					},
				},
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": toolResponse,
	})
}
