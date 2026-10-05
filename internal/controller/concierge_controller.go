package controller

import (
	"net/http"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/service"
	"github.com/bangweiz/dubhu-designer/internal/validator"
	"github.com/gin-gonic/gin"
)

// ConciergeController handles HTTP requests related to concierges.
type ConciergeController struct {
	conciergeService *service.ConciergeService
}

// NewConciergeController creates a new ConciergeController.
func NewConciergeController(conciergeService *service.ConciergeService) *ConciergeController {
	return &ConciergeController{
		conciergeService: conciergeService,
	}
}

// RegisterRoutes registers the concierge endpoints onto the gin router/group.
func (c *ConciergeController) RegisterRoutes(rg *gin.RouterGroup) {
	concierges := rg.Group("/concierges")
	{
		concierges.POST("", c.CreateConcierge)
		concierges.GET("", c.ListConcierges)
		concierges.GET("/:conciergeId", c.GetConciergeByID)
	}
}

// CreateConcierge handles POST /api/v1/concierges
func (c *ConciergeController) CreateConcierge(ctx *gin.Context) {
	var req dto.CreateConciergeDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	req.Trim()

	if errs := validator.ValidateCreateConcierge(&req); len(errs) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": errs,
		})
		return
	}

	resp, err := c.conciergeService.CreateConcierge(ctx.Request.Context(), req)
	if err != nil {
		writeServiceError(ctx, err, errorContext{nameValue: req.Name})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"data": resp,
	})
}

// ListConcierges handles GET /api/v1/concierges
func (c *ConciergeController) ListConcierges(ctx *gin.Context) {
	concierges, err := c.conciergeService.ListConcierges(ctx.Request.Context())
	if err != nil {
		writeServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": concierges,
	})
}

// GetConciergeByID handles GET /api/v1/concierges/:conciergeId
func (c *ConciergeController) GetConciergeByID(ctx *gin.Context) {
	id := ctx.Param("conciergeId")

	resp, err := c.conciergeService.GetConciergeByID(ctx.Request.Context(), id)
	if err != nil {
		writeServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": resp,
	})
}
