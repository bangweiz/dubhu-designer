package service

import (
	"context"
	"errors"
	"time"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/mapper"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"github.com/bangweiz/dubhu-designer/internal/repository"
	"github.com/bangweiz/dubhu-designer/internal/service/util"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ConciergeService handles business logic operations for concierges.
type ConciergeService struct {
	conciergeRepo      *repository.ConciergeRepository
	instructionRepo    *repository.InstructionRepository
	toolRepo           *repository.ToolRepository
	savedConciergeRepo *repository.SavedConciergeRepository
}

// NewConciergeService creates a new ConciergeService instance.
func NewConciergeService(conciergeRepo *repository.ConciergeRepository, instructionRepo *repository.InstructionRepository, toolRepo *repository.ToolRepository, savedConciergeRepo *repository.SavedConciergeRepository) *ConciergeService {
	return &ConciergeService{
		conciergeRepo:      conciergeRepo,
		instructionRepo:    instructionRepo,
		toolRepo:           toolRepo,
		savedConciergeRepo: savedConciergeRepo,
	}
}

// SaveConcierge snapshots the current customer-facing version and advances the live version.
func (s *ConciergeService) SaveConcierge(ctx context.Context, idStr string) (*models.SavedConcierge, error) {
	conciergeID, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrConciergeNotFound
	}
	return util.RunInTransaction(ctx, s.savedConciergeRepo.Client(), func(txCtx context.Context) (*models.SavedConcierge, error) {
		concierge, err := s.conciergeRepo.GetByID(txCtx, conciergeID)
		if err != nil {
			return nil, err
		}
		if concierge == nil {
			return nil, ErrConciergeNotFound
		}
		saved, err := s.buildSnapshot(txCtx, concierge)
		if err != nil {
			return nil, err
		}
		if err := s.savedConciergeRepo.Create(txCtx, saved); err != nil {
			return nil, err
		}
		if _, err := s.conciergeRepo.IncrementVersion(txCtx, conciergeID); err != nil {
			if errors.Is(err, repository.ErrConciergeNotFound) {
				return nil, ErrConciergeNotFound
			}
			return nil, err
		}
		return saved, nil
	})
}

// GetSavedConcierge retrieves an immutable customer-facing version.
func (s *ConciergeService) GetConciergeVersion(ctx context.Context, conciergeIDStr, versionIDStr string) (*models.SavedConcierge, error) {
	conciergeID, err := bson.ObjectIDFromHex(conciergeIDStr)
	if err != nil {
		return nil, ErrConciergeNotFound
	}
	versionID, err := bson.ObjectIDFromHex(versionIDStr)
	if err != nil {
		return nil, ErrConciergeVersionNotFound
	}
	saved, err := s.savedConciergeRepo.GetByID(ctx, conciergeID, versionID)
	if err != nil {
		return nil, err
	}
	if saved == nil {
		return nil, ErrConciergeVersionNotFound
	}
	return saved, nil
}

func (s *ConciergeService) GetSavedConcierge(ctx context.Context, idStr string, version int) (*models.SavedConcierge, error) {
	conciergeID, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrConciergeNotFound
	}
	saved, err := s.savedConciergeRepo.GetByConciergeIDAndVersion(ctx, conciergeID, version)
	if err != nil {
		return nil, err
	}
	if saved == nil {
		return nil, ErrConciergeNotFound
	}
	return saved, nil
}

func (s *ConciergeService) buildSnapshot(ctx context.Context, concierge *models.Concierge) (*models.SavedConcierge, error) {
	instructionIDs := make([]bson.ObjectID, 0)
	for _, agent := range concierge.Agents {
		instructionIDs = append(instructionIDs, agent.Instructions...)
	}
	instructionIDs = uniqueObjectIDs(instructionIDs)
	foundInstructions, err := s.instructionRepo.FindByIDs(ctx, instructionIDs)
	if err != nil {
		return nil, err
	}
	instructionMap := make(map[bson.ObjectID]models.Instruction, len(foundInstructions))
	toolIDs := make([]bson.ObjectID, 0)
	for _, instruction := range foundInstructions {
		instructionMap[instruction.ID] = instruction
		toolIDs = append(toolIDs, instruction.Tools...)
	}
	instructions := make([]models.Instruction, 0, len(instructionIDs))
	for _, instructionID := range instructionIDs {
		instruction, ok := instructionMap[instructionID]
		if !ok {
			return nil, ErrInstructionNotFound
		}
		instructions = append(instructions, instruction)
	}
	for _, agent := range concierge.Agents {
		toolIDs = append(toolIDs, agent.Tools...)
	}
	toolIDs = uniqueObjectIDs(toolIDs)
	foundTools, err := s.toolRepo.FindByIDs(ctx, toolIDs)
	if err != nil {
		return nil, err
	}
	toolMap := make(map[bson.ObjectID]models.Tool, len(foundTools))
	for _, tool := range foundTools {
		toolMap[tool.ID] = tool
	}
	tools := make([]models.Tool, 0, len(toolIDs))
	for _, toolID := range toolIDs {
		tool, ok := toolMap[toolID]
		if !ok {
			return nil, ErrToolNotFound
		}
		tools = append(tools, tool)
	}

	return &models.SavedConcierge{
		ID: bson.NewObjectID(), ConciergeID: concierge.ID, Version: concierge.Version,
		Name: concierge.Name, Description: concierge.Description, Agents: concierge.Agents,
		Instructions: instructions, Tools: tools, CreatedAt: time.Now().UTC(),
	}, nil
}

func uniqueObjectIDs(ids []bson.ObjectID) []bson.ObjectID {
	seen := make(map[bson.ObjectID]struct{}, len(ids))
	unique := make([]bson.ObjectID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}

// CreateConcierge maps DTO to a new Concierge entity (version 1), persists it through repository,
// and returns the mapped ConciergeResponseDTO.
func (s *ConciergeService) CreateConcierge(ctx context.Context, input dto.CreateConciergeDTO) (*dto.ConciergeResponseDTO, error) {
	entity := mapper.ToInitialConciergeEntity(input)
	created, err := s.conciergeRepo.Create(ctx, entity)
	if err != nil {
		if errors.Is(err, repository.ErrConciergeNameExists) {
			return nil, ErrConciergeNameExists
		}
		return nil, err
	}

	res := mapper.ToConciergeResponseDTO(created)
	return &res, nil
}

// GetConciergeByID retrieves a concierge by its hex ObjectID string and returns the mapped ConciergeResponseDTO.
// Returns ErrConciergeNotFound if the concierge does not exist.
func (s *ConciergeService) GetConciergeByID(ctx context.Context, idStr string) (*dto.ConciergeResponseDTO, error) {
	objectID, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrConciergeNotFound
	}

	concierge, err := s.conciergeRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, err
	}
	if concierge == nil {
		return nil, ErrConciergeNotFound
	}

	res := mapper.ToConciergeResponseDTO(concierge)
	return &res, nil
}

// ListConcierges retrieves all concierges and returns them as a slice of ConciergeSummaryResponseDTO.
func (s *ConciergeService) ListConcierges(ctx context.Context) ([]dto.ConciergeSummaryResponseDTO, error) {
	concierges, err := s.conciergeRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.ToConciergeSummaryResponseDTOList(concierges), nil
}
