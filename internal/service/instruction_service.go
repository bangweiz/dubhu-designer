package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/etag"
	"github.com/bangweiz/dubhu-designer/internal/mapper"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"github.com/bangweiz/dubhu-designer/internal/repository"
	"github.com/bangweiz/dubhu-designer/internal/service/util"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// InstructionService handles business logic operations for instructions.
type InstructionService struct {
	instructionRepo *repository.InstructionRepository
	toolRepo        *repository.ToolRepository
}

func (s *InstructionService) ListInstructionUsages(ctx context.Context, idStr string) ([]dto.InstructionUsageResponseDTO, error) {
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrInstructionNotFound
	}
	instruction, err := s.instructionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if instruction == nil {
		return nil, ErrInstructionNotFound
	}
	usages, err := s.instructionRepo.ListUsages(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapper.ToInstructionUsageResponseDTOList(usages), nil
}

// NewInstructionService creates a new InstructionService instance.
func NewInstructionService(
	instructionRepo *repository.InstructionRepository,
	toolRepo *repository.ToolRepository,
) *InstructionService {
	return &InstructionService{
		instructionRepo: instructionRepo,
		toolRepo:        toolRepo,
	}
}

// resolveTools parses tool IDs from content and queries MongoDB to find the matching Tool entities.
// If any referenced tool has an invalid ObjectID format or does not exist, it returns an ErrReferencedToolsNotFound error.
// Returns the resolved ObjectIDs and the matching Tool entities.
func (s *InstructionService) resolveTools(ctx context.Context, content string) ([]bson.ObjectID, []models.Tool, error) {
	idStrs := util.ExtractToolIDs(content)
	if len(idStrs) == 0 {
		return []bson.ObjectID{}, []models.Tool{}, nil
	}

	var invalidIDs []string
	validIDs := make([]bson.ObjectID, 0, len(idStrs))
	for _, idStr := range idStrs {
		objID, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			invalidIDs = append(invalidIDs, idStr)
		} else {
			validIDs = append(validIDs, objID)
		}
	}

	if len(invalidIDs) > 0 {
		return nil, nil, &ErrReferencedToolsNotFound{ToolIDs: invalidIDs}
	}

	tools, err := s.toolRepo.FindByIDs(ctx, validIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query tools: %w", err)
	}

	foundMap := make(map[bson.ObjectID]models.Tool, len(tools))
	for _, t := range tools {
		foundMap[t.ID] = t
	}

	var missing []string
	orderedTools := make([]models.Tool, 0, len(validIDs))
	for _, id := range validIDs {
		if t, ok := foundMap[id]; ok {
			orderedTools = append(orderedTools, t)
		} else {
			missing = append(missing, id.Hex())
		}
	}

	if len(missing) > 0 {
		return nil, nil, &ErrReferencedToolsNotFound{ToolIDs: missing}
	}

	return validIDs, orderedTools, nil
}

// CreateInstruction validates tool references, maps DTO to a new Instruction entity, and persists it.
// Returns the created instruction with fully populated tools.
func (s *InstructionService) CreateInstruction(ctx context.Context, input dto.CreateInstructionDTO) (*dto.InstructionResponseDTO, error) {
	toolIDs, tools, err := s.resolveTools(ctx, input.Content)
	if err != nil {
		return nil, err
	}

	entity := mapper.ToInitialInstructionEntity(input, toolIDs)
	created, err := s.instructionRepo.Create(ctx, entity)
	if err != nil {
		if errors.Is(err, repository.ErrInstructionNameExists) {
			return nil, ErrInstructionNameExists
		}
		return nil, err
	}

	res := mapper.ToInstructionResponseDTO(created, tools)
	return &res, nil
}

// GetInstructionByID retrieves an instruction by its hex ObjectID string and resolves all referenced tools via aggregation.
func (s *InstructionService) GetInstructionByID(ctx context.Context, idStr string) (*dto.InstructionResponseDTO, error) {
	objectID, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrInstructionNotFound
	}

	populated, err := s.instructionRepo.GetByIDWithTools(ctx, objectID)
	if err != nil {
		return nil, err
	}
	if populated == nil {
		return nil, ErrInstructionNotFound
	}

	res := mapper.ToPopulatedInstructionResponseDTO(populated)
	return &res, nil
}

// ListInstructions retrieves all instructions as summaries without embedding full tool specifications.
func (s *InstructionService) ListInstructions(ctx context.Context) ([]dto.InstructionSummaryResponseDTO, error) {
	instructions, err := s.instructionRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.ToInstructionSummaryResponseDTOList(instructions), nil
}

// UpdateInstruction updates an existing instruction with optimistic concurrency control.
// Returns the updated instruction with fully populated tools.
func (s *InstructionService) UpdateInstruction(ctx context.Context, idStr string, ifMatch string, input dto.UpdateInstructionDTO) (*dto.InstructionResponseDTO, error) {
	objectID, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrInstructionNotFound
	}

	expectedUpdatedAt, err := etag.Parse(ifMatch)
	if err != nil {
		return nil, ErrInstructionETagMismatch
	}

	toolIDs, tools, err := s.resolveTools(ctx, input.Content)
	if err != nil {
		return nil, err
	}

	updateDoc := bson.M{
		"name":       input.Name,
		"content":    input.Content,
		"tools":      toolIDs,
		"updated_at": time.Now().UTC(),
	}

	updated, err := s.instructionRepo.Update(ctx, objectID, expectedUpdatedAt, updateDoc)
	if err != nil {
		if errors.Is(err, repository.ErrInstructionConflict) {
			existing, getErr := s.instructionRepo.GetByID(ctx, objectID)
			if getErr == nil && existing == nil {
				return nil, ErrInstructionNotFound
			}
			return nil, ErrInstructionETagMismatch
		}
		if errors.Is(err, repository.ErrInstructionNameExists) {
			return nil, ErrInstructionNameExists
		}
		return nil, err
	}

	res := mapper.ToInstructionResponseDTO(updated, tools)
	return &res, nil
}
