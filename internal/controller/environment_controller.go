package controller

import (
	"net/http"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/etag"
	"github.com/bangweiz/dubhu-designer/internal/service"
	"github.com/bangweiz/dubhu-designer/internal/validator"
	"github.com/gin-gonic/gin"
)

type EnvironmentController struct{ environmentService *service.EnvironmentService }

func NewEnvironmentController(s *service.EnvironmentService) *EnvironmentController {
	return &EnvironmentController{environmentService: s}
}

func (c *EnvironmentController) RegisterRoutes(rg *gin.RouterGroup) {
	environments := rg.Group("/environments")
	environments.POST("", c.CreateEnvironment)
	environments.GET("", c.ListEnvironments)
	environments.GET("/:environmentId", c.GetEnvironmentByID)
	environments.PUT("/:environmentId", c.UpdateEnvironment)
}

func (c *EnvironmentController) CreateEnvironment(ctx *gin.Context) {
	var req dto.CreateEnvironmentDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	req.Trim()
	if errs := validator.ValidateCreateEnvironment(&req); len(errs) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": errs})
		return
	}
	resp, err := c.environmentService.CreateEnvironment(ctx.Request.Context(), req)
	if err != nil {
		writeServiceError(ctx, err, errorContext{nameValue: req.Name})
		return
	}
	ctx.Header("ETag", etag.Format(resp.UpdatedAt))
	ctx.JSON(http.StatusCreated, gin.H{"data": resp})
}

func (c *EnvironmentController) ListEnvironments(ctx *gin.Context) {
	resp, err := c.environmentService.ListEnvironments(ctx.Request.Context())
	if err != nil {
		writeServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": resp})
}

func (c *EnvironmentController) GetEnvironmentByID(ctx *gin.Context) {
	resp, err := c.environmentService.GetEnvironmentByID(ctx.Request.Context(), ctx.Param("environmentId"))
	if err != nil {
		writeServiceError(ctx, err)
		return
	}
	ctx.Header("ETag", etag.Format(resp.UpdatedAt))
	ctx.JSON(http.StatusOK, gin.H{"data": resp})
}

func (c *EnvironmentController) UpdateEnvironment(ctx *gin.Context) {
	ifMatch := ctx.GetHeader("If-Match")
	if ifMatch == "" {
		ctx.JSON(http.StatusPreconditionRequired, gin.H{"error": "If-Match header is required"})
		return
	}
	var req dto.UpdateEnvironmentDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	req.Trim()
	if errs := validator.ValidateUpdateEnvironment(&req); len(errs) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": errs})
		return
	}
	resp, err := c.environmentService.UpdateEnvironment(ctx.Request.Context(), ctx.Param("environmentId"), ifMatch, req)
	if err != nil {
		writeServiceError(ctx, err, errorContext{nameValue: req.Name})
		return
	}
	ctx.Header("ETag", etag.Format(resp.UpdatedAt))
	ctx.JSON(http.StatusOK, gin.H{"data": resp})
}
