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
		if errors.Is(err, service.ErrConciergeNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "Concierge not found",
			})
			return
		}
		if errors.Is(err, repository.ErrAgentNameExists) {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "Conflict",
				"details": []validator.FieldError{
					{
						Field:  "name",
						Reason: "agent name already exists for this concierge",
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
		"data": resp,
	})
}

// ListAgents handles GET /api/v1/concierges/:conciergeId/agents
func (c *AgentController) ListAgents(ctx *gin.Context) {
	conciergeID := ctx.Param("conciergeId")

	agents, err := c.agentService.ListAgents(ctx.Request.Context(), conciergeID)
	if err != nil {
		if errors.Is(err, service.ErrConciergeNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "Concierge not found",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
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
		if errors.Is(err, service.ErrConciergeNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "Concierge not found",
			})
			return
		}
		if errors.Is(err, service.ErrAgentNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "Agent not found",
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
