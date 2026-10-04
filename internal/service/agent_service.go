package service

import (
	"context"
	"errors"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/mapper"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"github.com/bangweiz/dubhu-designer/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrAgentNotFound = errors.New("agent not found")
)

// AgentService handles business logic operations for agents.
type AgentService struct {
	agentRepo     *repository.AgentRepository
	conciergeRepo *repository.ConciergeRepository
}

// NewAgentService creates a new AgentService instance.
func NewAgentService(agentRepo *repository.AgentRepository, conciergeRepo *repository.ConciergeRepository) *AgentService {
	return &AgentService{
		agentRepo:     agentRepo,
		conciergeRepo: conciergeRepo,
	}
}

// CreateAgent validates concierge existence, maps DTO to a new Agent entity (version 1),
// atomically appends it to the concierge's embedded agents array, and returns the mapped AgentResponseDTO.
func (s *AgentService) CreateAgent(ctx context.Context, conciergeIDStr string, input dto.CreateAgentDTO) (*dto.AgentResponseDTO, error) {
	conciergeID, err := bson.ObjectIDFromHex(conciergeIDStr)
	if err != nil {
		return nil, ErrConciergeNotFound
	}

	concierge, err := s.conciergeRepo.GetByID(ctx, conciergeID)
	if err != nil {
		return nil, err
	}
	if concierge == nil {
		return nil, ErrConciergeNotFound
	}

	entity := mapper.ToInitialAgentEntity(conciergeID, input)
	created, err := s.agentRepo.Create(ctx, conciergeID, entity)
	if err != nil {
		return nil, err
	}

	res := mapper.ToAgentResponseDTO(created)
	return &res, nil
}

// GetAgentByID retrieves an agent by its hex ObjectID string and its concierge hex ObjectID string,
// returning the mapped AgentResponseDTO.
// Returns ErrConciergeNotFound if concierge does not exist.
// Returns ErrAgentNotFound if agent does not exist or does not belong to the concierge.
func (s *AgentService) GetAgentByID(ctx context.Context, conciergeIDStr string, agentIDStr string) (*dto.AgentResponseDTO, error) {
	conciergeID, err := bson.ObjectIDFromHex(conciergeIDStr)
	if err != nil {
		return nil, ErrConciergeNotFound
	}

	agentID, err := bson.ObjectIDFromHex(agentIDStr)
	if err != nil {
		return nil, ErrAgentNotFound
	}

	concierge, err := s.conciergeRepo.GetByID(ctx, conciergeID)
	if err != nil {
		return nil, err
	}
	if concierge == nil {
		return nil, ErrConciergeNotFound
	}

	for _, a := range concierge.Agents {
		if a.ID == agentID {
			a.ConciergeID = conciergeID
			res := mapper.ToAgentResponseDTO(&a)
			return &res, nil
		}
	}

	return nil, ErrAgentNotFound
}

// ListAgents retrieves all agents for a given concierge and returns them as a slice of AgentResponseDTO.
func (s *AgentService) ListAgents(ctx context.Context, conciergeIDStr string) ([]dto.AgentResponseDTO, error) {
	conciergeID, err := bson.ObjectIDFromHex(conciergeIDStr)
	if err != nil {
		return nil, ErrConciergeNotFound
	}

	concierge, err := s.conciergeRepo.GetByID(ctx, conciergeID)
	if err != nil {
		return nil, err
	}
	if concierge == nil {
		return nil, ErrConciergeNotFound
	}

	agents := concierge.Agents
	if agents == nil {
		agents = []models.Agent{}
	}
	for i := range agents {
		agents[i].ConciergeID = conciergeID
	}

	return mapper.ToAgentResponseDTOList(agents), nil
}
