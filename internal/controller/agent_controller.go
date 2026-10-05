package controller

import (
	"net/http"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/etag"
	"github.com/bangweiz/dubhu-designer/internal/service"
	"github.com/bangweiz/dubhu-designer/internal/validator"
	"github.com/gin-gonic/gin"
)

// AgentController handles HTTP requests related to agents.
type AgentController struct {
	agentService *service.AgentService
}

// NewAgentController creates a new AgentController.
func NewAgentController(agentService *service.AgentService) *AgentController {
	return &AgentController{
		agentService: agentService,
	}
}

// RegisterRoutes registers the agent endpoints onto the gin router/group under /concierges/:conciergeId/agents.
func (c *AgentController) RegisterRoutes(rg *gin.RouterGroup) {
	agents := rg.Group("/concierges/:conciergeId/agents")
	{
		agents.POST("", c.CreateAgent)
		agents.GET("", c.ListAgents)
		agents.GET("/:agentId", c.GetAgentByID)
		agents.PUT("/:agentId", c.UpdateAgent)
		agents.POST("/:agentId/instructions/:instructionId", c.AssignInstruction)
		agents.PUT("/:agentId/instructions/:instructionId", c.AssignInstruction)
		agents.DELETE("/:agentId/instructions/:instructionId", c.UnassignInstruction)
	}
}

// CreateAgent handles POST /api/v1/concierges/:conciergeId/agents
func (c *AgentController) CreateAgent(ctx *gin.Context) {
	conciergeID := ctx.Param("conciergeId")

	var req dto.CreateAgentDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	req.Trim()

	if errs := validator.ValidateCreateAgent(&req); len(errs) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": errs,
		})
		return
	}

	resp, err := c.agentService.CreateAgent(ctx.Request.Context(), conciergeID, req)
	if err != nil {
		writeServiceError(ctx, err, errorContext{nameValue: req.Name})
		return
	}

	ctx.Header("ETag", etag.Format(resp.Version))
	ctx.JSON(http.StatusCreated, gin.H{
		"data": resp,
	})
}

// ListAgents handles GET /api/v1/concierges/:conciergeId/agents
func (c *AgentController) ListAgents(ctx *gin.Context) {
	conciergeID := ctx.Param("conciergeId")

	agents, err := c.agentService.ListAgents(ctx.Request.Context(), conciergeID)
	if err != nil {
		writeServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": agents,
	})
}

// GetAgentByID handles GET /api/v1/concierges/:conciergeId/agents/:agentId
func (c *AgentController) GetAgentByID(ctx *gin.Context) {
	conciergeID := ctx.Param("conciergeId")
	agentID := ctx.Param("agentId")

	resp, err := c.agentService.GetAgentByID(ctx.Request.Context(), conciergeID, agentID)
	if err != nil {
		writeServiceError(ctx, err)
		return
	}

	ctx.Header("ETag", etag.Format(resp.Version))
	ctx.JSON(http.StatusOK, gin.H{
		"data": resp,
	})
}

// UpdateAgent handles PUT /api/v1/concierges/:conciergeId/agents/:agentId
func (c *AgentController) UpdateAgent(ctx *gin.Context) {
	conciergeID := ctx.Param("conciergeId")
	agentID := ctx.Param("agentId")
	ifMatch := ctx.GetHeader("If-Match")
	if ifMatch == "" {
		ctx.JSON(http.StatusPreconditionRequired, gin.H{
			"error": "If-Match header is required",
		})
		return
	}

	var req dto.UpdateAgentDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	req.Trim()

	if errs := validator.ValidateUpdateAgent(&req); len(errs) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": errs,
		})
		return
	}

	resp, err := c.agentService.UpdateAgent(ctx.Request.Context(), conciergeID, agentID, ifMatch, req)
	if err != nil {
		writeServiceError(ctx, err, errorContext{nameValue: req.Name})
		return
	}

	ctx.Header("ETag", etag.Format(resp.Version))
	ctx.JSON(http.StatusOK, gin.H{
		"data": resp,
	})
}

// AssignInstruction handles POST/PUT /api/v1/concierges/:conciergeId/agents/:agentId/instructions/:instructionId
func (c *AgentController) AssignInstruction(ctx *gin.Context) {
	conciergeID := ctx.Param("conciergeId")
	agentID := ctx.Param("agentId")
	instructionID := ctx.Param("instructionId")

	err := c.agentService.AssignInstruction(ctx.Request.Context(), conciergeID, agentID, instructionID)
	if err != nil {
		writeServiceError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// UnassignInstruction handles DELETE /api/v1/concierges/:conciergeId/agents/:agentId/instructions/:instructionId
func (c *AgentController) UnassignInstruction(ctx *gin.Context) {
	conciergeID := ctx.Param("conciergeId")
	agentID := ctx.Param("agentId")
	instructionID := ctx.Param("instructionId")

	err := c.agentService.UnassignInstruction(ctx.Request.Context(), conciergeID, agentID, instructionID)
	if err != nil {
		writeServiceError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}
