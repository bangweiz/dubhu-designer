package service

import (
	"context"
	"errors"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/etag"
	"github.com/bangweiz/dubhu-designer/internal/mapper"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"github.com/bangweiz/dubhu-designer/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// AgentService handles business logic operations for agents.
type AgentService struct {
	agentRepo       *repository.AgentRepository
	conciergeRepo   *repository.ConciergeRepository
	instructionRepo *repository.InstructionRepository
}

// NewAgentService creates a new AgentService instance.
func NewAgentService(agentRepo *repository.AgentRepository, conciergeRepo *repository.ConciergeRepository, instructionRepo *repository.InstructionRepository) *AgentService {
	return &AgentService{
		agentRepo:       agentRepo,
		conciergeRepo:   conciergeRepo,
		instructionRepo: instructionRepo,
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
		switch {
		case errors.Is(err, repository.ErrConciergeNotFound):
			return nil, ErrConciergeNotFound
		case errors.Is(err, repository.ErrAgentNameExists):
			return nil, ErrAgentNameExists
		}
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

// UpdateAgent updates an existing agent within a concierge.
// Name, description, goal, and model are updated, and the agent version is incremented.
// Returns ErrConciergeNotFound if concierge ID is invalid or concierge does not exist.
// Returns ErrAgentNotFound if agent ID is invalid or agent does not exist in the concierge.
// Returns ErrAgentNameExists if an agent with the same name already exists in this concierge.
func (s *AgentService) UpdateAgent(ctx context.Context, conciergeIDStr string, agentIDStr string, ifMatch string, input dto.UpdateAgentDTO) (*dto.AgentResponseDTO, error) {
	conciergeID, err := bson.ObjectIDFromHex(conciergeIDStr)
	if err != nil {
		return nil, ErrConciergeNotFound
	}

	agentID, err := bson.ObjectIDFromHex(agentIDStr)
	if err != nil {
		return nil, ErrAgentNotFound
	}

	expectedVersion, err := etag.Parse(ifMatch)
	if err != nil {
		return nil, ErrAgentETagMismatch
	}

	entity := &models.Agent{
		ID:          agentID,
		ConciergeID: conciergeID,
		Name:        input.Name,
		Description: input.Description,
		Goal:        input.Goal,
		Model:       models.Model(input.Model),
	}

	updated, err := s.agentRepo.Update(ctx, conciergeID, agentID, expectedVersion, entity)
	if err != nil {
		if errors.Is(err, repository.ErrConciergeNotFound) {
			return nil, ErrConciergeNotFound
		}
		if errors.Is(err, repository.ErrAgentNotFound) {
			return nil, ErrAgentNotFound
		}
		if errors.Is(err, repository.ErrAgentVersionConflict) {
			return nil, ErrAgentETagMismatch
		}
		if errors.Is(err, repository.ErrAgentNameExists) {
			return nil, ErrAgentNameExists
		}
		return nil, err
	}

	res := mapper.ToAgentResponseDTO(updated)
	return &res, nil
}

// AssignInstruction assigns an instruction to an agent within a concierge.
// Validates that the concierge, agent, and instruction all exist in the database.
// Returns ErrConciergeNotFound if concierge ID is invalid or concierge does not exist.
// Returns ErrAgentNotFound if agent ID is invalid or agent does not exist in the concierge.
// Returns ErrInstructionNotFound if instruction ID is invalid or instruction does not exist.
func (s *AgentService) AssignInstruction(ctx context.Context, conciergeIDStr string, agentIDStr string, instructionIDStr string) error {
	conciergeID, err := bson.ObjectIDFromHex(conciergeIDStr)
	if err != nil {
		return ErrConciergeNotFound
	}

	agentID, err := bson.ObjectIDFromHex(agentIDStr)
	if err != nil {
		return ErrAgentNotFound
	}

	instructionID, err := bson.ObjectIDFromHex(instructionIDStr)
	if err != nil {
		return ErrInstructionNotFound
	}

	instruction, err := s.instructionRepo.GetByID(ctx, instructionID)
	if err != nil {
		return err
	}
	if instruction == nil {
		return ErrInstructionNotFound
	}

	err = s.agentRepo.AssignInstruction(ctx, conciergeID, agentID, instructionID)
	if err != nil {
		if errors.Is(err, repository.ErrConciergeNotFound) {
			return ErrConciergeNotFound
		}
		if errors.Is(err, repository.ErrAgentNotFound) {
			return ErrAgentNotFound
		}
		return err
	}

	return nil
}

// UnassignInstruction removes an instruction from an agent within a concierge.
// Returns ErrConciergeNotFound if concierge ID is invalid or concierge does not exist.
// Returns ErrAgentNotFound if agent ID is invalid or agent does not exist in the concierge.
// Returns ErrInstructionNotFound if instruction ID is invalid.
func (s *AgentService) UnassignInstruction(ctx context.Context, conciergeIDStr string, agentIDStr string, instructionIDStr string) error {
	conciergeID, err := bson.ObjectIDFromHex(conciergeIDStr)
	if err != nil {
		return ErrConciergeNotFound
	}

	agentID, err := bson.ObjectIDFromHex(agentIDStr)
	if err != nil {
		return ErrAgentNotFound
	}

	instructionID, err := bson.ObjectIDFromHex(instructionIDStr)
	if err != nil {
		return ErrInstructionNotFound
	}

	err = s.agentRepo.UnassignInstruction(ctx, conciergeID, agentID, instructionID)
	if err != nil {
		if errors.Is(err, repository.ErrConciergeNotFound) {
			return ErrConciergeNotFound
		}
		if errors.Is(err, repository.ErrAgentNotFound) {
			return ErrAgentNotFound
		}
		return err
	}

	return nil
}
