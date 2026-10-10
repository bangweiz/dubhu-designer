package controller

import (
	"net/http"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/etag"
	"github.com/bangweiz/dubhu-designer/internal/service"
	"github.com/bangweiz/dubhu-designer/internal/validator"
	"github.com/gin-gonic/gin"
)

type VariableController struct{ variableService *service.VariableService }

func NewVariableController(s *service.VariableService) *VariableController {
	return &VariableController{variableService: s}
}

func (c *VariableController) RegisterRoutes(rg *gin.RouterGroup) {
	variables := rg.Group("/variables")
	variables.POST("", c.CreateVariable)
	variables.GET("", c.ListVariables)
	variables.GET("/:variableId", c.GetVariableByID)
	variables.PUT("/:variableId", c.UpdateVariable)
}

func (c *VariableController) CreateVariable(ctx *gin.Context) {
	var req dto.CreateVariableDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	req.Trim()
	if errs := validator.ValidateCreateVariable(&req); len(errs) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": errs})
		return
	}

	resp, err := c.variableService.CreateVariable(ctx.Request.Context(), req)
	if err != nil {
		writeServiceError(ctx, err, errorContext{nameValue: req.Name})
		return
	}

	ctx.Header("ETag", etag.Format(resp.UpdatedAt))
	ctx.JSON(http.StatusCreated, gin.H{"data": resp})
}

func (c *VariableController) ListVariables(ctx *gin.Context) {
	resp, err := c.variableService.ListVariables(ctx.Request.Context())
	if err != nil {
		writeServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": resp})
}

func (c *VariableController) GetVariableByID(ctx *gin.Context) {
	resp, err := c.variableService.GetVariableByID(ctx.Request.Context(), ctx.Param("variableId"))
	if err != nil {
		writeServiceError(ctx, err)
		return
	}

	ctx.Header("ETag", etag.Format(resp.UpdatedAt))
	ctx.JSON(http.StatusOK, gin.H{"data": resp})
}

func (c *VariableController) UpdateVariable(ctx *gin.Context) {
	ifMatch := ctx.GetHeader("If-Match")
	if ifMatch == "" {
		ctx.JSON(http.StatusPreconditionRequired, gin.H{"error": "If-Match header is required"})
		return
	}

	var req dto.UpdateVariableDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	req.Trim()
	if errs := validator.ValidateUpdateVariable(&req); len(errs) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": errs})
		return
	}

	resp, err := c.variableService.UpdateVariable(ctx.Request.Context(), ctx.Param("variableId"), ifMatch, req)
	if err != nil {
		writeServiceError(ctx, err, errorContext{nameValue: req.Name})
		return
	}

	ctx.Header("ETag", etag.Format(resp.UpdatedAt))
	ctx.JSON(http.StatusOK, gin.H{"data": resp})
}
