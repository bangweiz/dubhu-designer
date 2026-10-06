package service

import (
	"context"
	"errors"
	"time"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/etag"
	"github.com/bangweiz/dubhu-designer/internal/mapper"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"github.com/bangweiz/dubhu-designer/internal/repository"
	"github.com/bangweiz/dubhu-designer/internal/service/util"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ConciergeService handles business logic operations for concierges.
type ConciergeService struct {
	conciergeRepo        *repository.ConciergeRepository
	conciergeVersionRepo *repository.DraftConciergeVersionRepository
	instructionRepo      *repository.InstructionRepository
	toolRepo             *repository.ToolRepository
	savedConciergeRepo   *repository.SavedConciergeVersionRepository
}

func (s *ConciergeService) UpdateConcierge(ctx context.Context, idStr, ifMatch string, input dto.UpdateConciergeDTO) (*dto.ConciergeResponseDTO, error) {
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrConciergeNotFound
	}
	version, err := etag.Parse(ifMatch)
	if err != nil {
		return nil, ErrConciergeETagMismatch
	}
	c, err := s.conciergeRepo.Update(ctx, id, version, input.Name, input.Description)
	if errors.Is(err, repository.ErrVersionConflict) {
		existing, lookupErr := s.conciergeRepo.GetByID(ctx, id)
		if lookupErr != nil {
			return nil, lookupErr
		}
		if existing == nil {
			return nil, ErrConciergeNotFound
		}
		return nil, ErrConciergeETagMismatch
	}
	if errors.Is(err, repository.ErrConciergeNameExists) {
		return nil, ErrConciergeNameExists
	}
	if err != nil {
		return nil, err
	}
	return s.conciergeResponse(ctx, c)
}

// NewConciergeService creates a new ConciergeService instance.
func NewConciergeService(conciergeRepo *repository.ConciergeRepository, conciergeVersionRepo *repository.DraftConciergeVersionRepository, instructionRepo *repository.InstructionRepository, toolRepo *repository.ToolRepository, savedConciergeRepo *repository.SavedConciergeVersionRepository) *ConciergeService {
	return &ConciergeService{
		conciergeRepo:        conciergeRepo,
		conciergeVersionRepo: conciergeVersionRepo,
		instructionRepo:      instructionRepo,
		toolRepo:             toolRepo,
		savedConciergeRepo:   savedConciergeRepo,
	}
}

// SaveConcierge snapshots the current customer-facing version and advances the draft version.
func (s *ConciergeService) SaveConcierge(ctx context.Context, idStr string) (*models.SavedConciergeVersion, error) {
	conciergeID, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrConciergeNotFound
	}
	return util.RunInTransaction(ctx, s.savedConciergeRepo.Client(), func(txCtx context.Context) (*models.SavedConciergeVersion, error) {
		concierge, err := s.conciergeVersionRepo.GetByConciergeID(txCtx, conciergeID)
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
		if err := s.conciergeVersionRepo.Advance(txCtx, conciergeID); err != nil {
			if errors.Is(err, repository.ErrConciergeNotFound) {
				return nil, ErrConciergeNotFound
			}
			return nil, err
		}
		if err := s.conciergeRepo.RecordSavedVersion(txCtx, conciergeID, concierge.ID, saved.ID, concierge.Version); err != nil {
			return nil, err
		}
		return saved, nil
	})
}

// GetSavedConciergeVersion retrieves an immutable customer-facing version.
func (s *ConciergeService) GetConciergeVersion(ctx context.Context, conciergeIDStr, versionIDStr string) (*models.SavedConciergeVersion, error) {
	id, err := bson.ObjectIDFromHex(conciergeIDStr)
	if err != nil {
		return nil, ErrConciergeNotFound
	}
	if versionIDStr != "draft" {
		versionID, err := bson.ObjectIDFromHex(versionIDStr)
		if err != nil {
			return nil, ErrConciergeVersionNotFound
		}
		saved, err := s.savedConciergeRepo.GetByID(ctx, id, versionID)
		if err != nil {
			return nil, err
		}
		if saved != nil {
			return saved, nil
		}
	}
	draft, err := s.conciergeVersionRepo.GetByConciergeID(ctx, id)
	if err != nil {
		return nil, err
	}
	if draft == nil || (versionIDStr != "draft" && draft.ID.Hex() != versionIDStr) {
		return nil, ErrConciergeVersionNotFound
	}
	response, err := s.buildSnapshot(ctx, draft)
	if err != nil {
		return nil, err
	}
	response.ID = draft.ID
	response.CreatedAt = draft.CreatedAt
	response.UpdatedAt = draft.UpdatedAt
	response.Saved = false
	return response, nil
}

func (s *ConciergeService) buildSnapshot(ctx context.Context, concierge *models.DraftConciergeVersion) (*models.SavedConciergeVersion, error) {
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

	return &models.SavedConciergeVersion{
		ID: bson.NewObjectID(), ConciergeID: concierge.ConciergeID, Version: concierge.Version,
		Agents:       concierge.Agents,
		Instructions: instructions, Tools: tools, CreatedAt: concierge.CreatedAt, UpdatedAt: concierge.UpdatedAt, Saved: true,
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

	return util.RunInTransaction(ctx, s.savedConciergeRepo.Client(), func(txCtx context.Context) (*dto.ConciergeResponseDTO, error) {
		entity := mapper.ToInitialConciergeEntity(input)
		now := time.Now().UTC()
		draft := &models.DraftConciergeVersion{ID: bson.NewObjectID(), ConciergeID: entity.ID, Agents: []models.Agent{}, Version: 1, ETagVersion: 1, CreatedAt: now, UpdatedAt: now}
		entity.ConciergeVersions = []models.ConciergeVersionReference{{ConciergeVersionID: draft.ID, Version: 1}}
		created, err := s.conciergeRepo.Create(txCtx, entity)
		if err != nil {
			if errors.Is(err, repository.ErrConciergeNameExists) {
				return nil, ErrConciergeNameExists
			}
			return nil, err
		}
		if err := s.conciergeVersionRepo.Create(txCtx, draft); err != nil {
			return nil, err
		}
		response := mapper.ToConciergeResponseDTO(created)
		return &response, nil
	})
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

	return s.conciergeResponse(ctx, concierge)
}

// ListConcierges retrieves all concierges and returns them as a slice of ConciergeSummaryResponseDTO.
func (s *ConciergeService) ListConcierges(ctx context.Context) ([]dto.ConciergeSummaryResponseDTO, error) {
	concierges, err := s.conciergeRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.ConciergeSummaryResponseDTO, 0, len(concierges))
	for i := range concierges {
		response, err := s.conciergeResponse(ctx, &concierges[i])
		if err != nil {
			return nil, err
		}
		result = append(result, *response)
	}
	return result, nil
}

func (s *ConciergeService) conciergeResponse(ctx context.Context, c *models.Concierge) (*dto.ConciergeResponseDTO, error) {
	response := mapper.ToConciergeResponseDTO(c)
	return &response, nil
}
