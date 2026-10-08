package service

import (
	"context"
	"errors"
	"time"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/etag"
	"github.com/bangweiz/dubhu-designer/internal/mapper"
	"github.com/bangweiz/dubhu-designer/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ToolService handles business logic operations for tools.
type ToolService struct {
	toolRepo *repository.ToolRepository
}

func (s *ToolService) ListToolUsages(ctx context.Context, idStr string) ([]dto.ToolUsageResponseDTO, error) {
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrToolNotFound
	}
	tool, err := s.toolRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tool == nil {
		return nil, ErrToolNotFound
	}
	instructions, err := s.toolRepo.ListReferencingInstructions(ctx, id)
	if err != nil {
		return nil, err
	}
	ids := make([]bson.ObjectID, 0, len(instructions))
	for _, instruction := range instructions {
		ids = append(ids, instruction.ID)
	}
	agents, err := s.toolRepo.ListAgentUsages(ctx, id, ids)
	if err != nil {
		return nil, err
	}
	return mapper.ToInstructionUsageResponseDTOList(agents), nil
}

// NewToolService creates a new ToolService instance.
func NewToolService(toolRepo *repository.ToolRepository) *ToolService {
	return &ToolService{
		toolRepo: toolRepo,
	}
}

// CreateTool maps DTO to a new Tool entity (with audit timestamps), persists it through repository,
// and returns the mapped ToolResponseDTO.
func (s *ToolService) CreateTool(ctx context.Context, input dto.CreateToolDTO) (*dto.ToolResponseDTO, error) {
	tool := mapper.ToInitialToolEntity(input)
	createdTool, err := s.toolRepo.Create(ctx, tool)
	if err != nil {
		if errors.Is(err, repository.ErrToolNameExists) {
			return nil, ErrToolNameExists
		}
		return nil, err
	}

	res := mapper.ToToolResponseDTO(createdTool)
	return &res, nil
}

// GetToolByID retrieves a tool by its hex ObjectID string and returns the mapped ToolResponseDTO.
// Returns ErrToolNotFound if the tool does not exist.
func (s *ToolService) GetToolByID(ctx context.Context, idStr string) (*dto.ToolResponseDTO, error) {
	objectID, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrToolNotFound
	}

	tool, err := s.toolRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, err
	}
	if tool == nil {
		return nil, ErrToolNotFound
	}

	res := mapper.ToToolResponseDTO(tool)
	return &res, nil
}

// ListTools retrieves all tools and returns them as a slice of ToolResponseDTO.
func (s *ToolService) ListTools(ctx context.Context) ([]dto.ToolResponseDTO, error) {
	tools, err := s.toolRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.ToToolResponseDTOList(tools), nil
}

// UpdateTool updates an existing tool with optimistic concurrency control.
// Matches by ID and updated_at, then advances the timestamp.
func (s *ToolService) UpdateTool(ctx context.Context, idStr string, ifMatch string, input dto.UpdateToolDTO) (*dto.ToolResponseDTO, error) {
	objectID, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrToolNotFound
	}

	expectedUpdatedAt, err := etag.Parse(ifMatch)
	if err != nil {
		return nil, ErrToolETagMismatch
	}

	updateDoc := bson.M{
		"name":        input.Name,
		"description": input.Description,
		"inputs":      mapper.ToToolInputs(input.Inputs),
		"outputs":     mapper.ToToolOutputs(input.Outputs),
		"updated_at":  time.Now().UTC(),
	}

	updatedTool, err := s.toolRepo.Update(ctx, objectID, expectedUpdatedAt, updateDoc)
	if err != nil {
		if errors.Is(err, repository.ErrUpdateConflict) {
			// Check existence only after a failed conditional update to distinguish 404 from 412.
			existing, getErr := s.toolRepo.GetByID(ctx, objectID)
			if getErr == nil && existing == nil {
				return nil, ErrToolNotFound
			}
			return nil, ErrToolETagMismatch
		}
		if errors.Is(err, repository.ErrToolNameExists) {
			return nil, ErrToolNameExists
		}
		return nil, err
	}

	res := mapper.ToToolResponseDTO(updatedTool)
	return &res, nil
}
