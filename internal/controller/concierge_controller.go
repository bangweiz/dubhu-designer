package controller

import (
	"net/http"
	"strings"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/etag"
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
		concierges.PUT("/:conciergeId", c.UpdateConcierge)
		concierges.GET("/:conciergeId/concierge-versions/:conciergeVersionId", c.GetConciergeVersion)
	}
	rg.POST("/concierge/:conciergeAction", c.SaveConcierge)
}

func (c *ConciergeController) UpdateConcierge(ctx *gin.Context) {
	ifMatch := ctx.GetHeader("If-Match")
	if ifMatch == "" {
		ctx.JSON(http.StatusPreconditionRequired, gin.H{"error": "If-Match header is required"})
		return
	}
	var req dto.UpdateConciergeDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	req.Trim()
	if errs := validator.ValidateCreateConcierge(&req); len(errs) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": errs})
		return
	}
	resp, err := c.conciergeService.UpdateConcierge(ctx.Request.Context(), ctx.Param("conciergeId"), ifMatch, req)
	if err != nil {
		writeServiceError(ctx, err, errorContext{nameValue: req.Name})
		return
	}
	ctx.Header("ETag", etag.Format(resp.Version))
	ctx.JSON(http.StatusOK, gin.H{"data": resp})
}

// GetConciergeVersion retrieves the working version or a saved snapshot.
func (c *ConciergeController) GetConciergeVersion(ctx *gin.Context) {
	versionID := ctx.Param("conciergeVersionId")

	saved, err := c.conciergeService.GetConciergeVersion(ctx.Request.Context(), ctx.Param("conciergeId"), versionID)
	if err != nil {
		writeServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": saved})
}

// SaveConcierge handles POST /api/v1/organisations/:organisationId/concierge/:conciergeId:save.
func (c *ConciergeController) SaveConcierge(ctx *gin.Context) {
	conciergeID, ok := strings.CutSuffix(ctx.Param("conciergeAction"), ":save")
	if !ok || conciergeID == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	saved, err := c.conciergeService.SaveConcierge(ctx.Request.Context(), conciergeID)
	if err != nil {
		writeServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"data": saved})
}

// CreateConcierge handles POST /api/v1/organisations/:organisationId/concierges
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

	ctx.Header("ETag", etag.Format(resp.Version))
	ctx.JSON(http.StatusCreated, gin.H{
		"data": resp,
	})
}

// ListConcierges handles GET /api/v1/organisations/:organisationId/concierges
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

// GetConciergeByID handles GET /api/v1/organisations/:organisationId/concierges/:conciergeId
func (c *ConciergeController) GetConciergeByID(ctx *gin.Context) {
	id := ctx.Param("conciergeId")

	resp, err := c.conciergeService.GetConciergeByID(ctx.Request.Context(), id)
	if err != nil {
		writeServiceError(ctx, err)
		return
	}

	ctx.Header("ETag", etag.Format(resp.Version))
	ctx.JSON(http.StatusOK, gin.H{
		"data": resp,
	})
}
