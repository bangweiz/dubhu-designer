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
	environmentRepo    *repository.EnvironmentRepository
	conciergeRepo      *repository.ConciergeRepository
	instructionRepo    *repository.InstructionRepository
	toolRepo           *repository.ToolRepository
	variableRepo       *repository.VariableRepository
	savedConciergeRepo *repository.SavedConciergeVersionRepository
}

func (s *ConciergeService) UpdateConcierge(
	ctx context.Context,
	idStr, ifMatch string,
	input dto.UpdateConciergeDTO,
) (*dto.ConciergeResponseDTO, error) {
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrConciergeNotFound
	}

	updatedAt, err := etag.Parse(ifMatch)
	if err != nil {
		return nil, ErrConciergeETagMismatch
	}

	c, err := s.conciergeRepo.Update(ctx, id, updatedAt, input.Name, input.Description)
	if errors.Is(err, repository.ErrUpdateConflict) {
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
func NewConciergeService(
	conciergeRepo *repository.ConciergeRepository,
	instructionRepo *repository.InstructionRepository,
	toolRepo *repository.ToolRepository,
	variableRepo *repository.VariableRepository,
	savedConciergeRepo *repository.SavedConciergeVersionRepository,
	environmentRepo *repository.EnvironmentRepository,
) *ConciergeService {
	return &ConciergeService{
		environmentRepo:    environmentRepo,
		conciergeRepo:      conciergeRepo,
		instructionRepo:    instructionRepo,
		toolRepo:           toolRepo,
		variableRepo:       variableRepo,
		savedConciergeRepo: savedConciergeRepo,
	}
}

// SaveConcierge snapshots the current customer-facing version and advances the snapshot sequence.
func (s *ConciergeService) SaveConcierge(ctx context.Context, idStr string) (*dto.SavedConciergeVersionResponseDTO, error) {
	conciergeID, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrConciergeNotFound
	}

	return util.RunInTransaction(ctx, s.savedConciergeRepo.Client(), func(txCtx context.Context) (*dto.SavedConciergeVersionResponseDTO, error) {
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
		if err := s.conciergeRepo.RecordSavedVersion(txCtx, conciergeID, saved.ID, concierge.NextVersion, concierge.UpdatedAt); err != nil {
			return nil, err
		}

		response := mapper.ToSavedConciergeVersionResponseDTO(saved)
		return &response, nil
	})
}

// GetConciergeVersion retrieves an immutable customer-facing version.
func (s *ConciergeService) GetConciergeVersion(
	ctx context.Context,
	conciergeIDStr, versionIDStr string,
) (*dto.SavedConciergeVersionResponseDTO, error) {
	id, err := bson.ObjectIDFromHex(conciergeIDStr)
	if err != nil {
		return nil, ErrConciergeNotFound
	}

	versionID, err := bson.ObjectIDFromHex(versionIDStr)
	if err != nil {
		return nil, ErrConciergeVersionNotFound
	}

	saved, err := s.savedConciergeRepo.GetByID(ctx, id, versionID)
	if err != nil {
		return nil, err
	}
	if saved == nil {
		return nil, ErrConciergeVersionNotFound
	}

	response := mapper.ToSavedConciergeVersionResponseDTO(saved)
	return &response, nil
}

func (s *ConciergeService) buildSnapshot(ctx context.Context, concierge *models.Concierge) (*models.SavedConciergeVersion, error) {
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

	variableIDs := []bson.ObjectID{}
	for _, instruction := range instructions {
		variableIDs = append(variableIDs, instruction.Variables...)
	}

	variableIDs = uniqueObjectIDs(variableIDs)
	foundVariables, err := s.variableRepo.FindByIDs(ctx, variableIDs)
	if err != nil {
		return nil, err
	}

	variableMap := make(map[bson.ObjectID]models.Variable, len(foundVariables))
	for _, variable := range foundVariables {
		variableMap[variable.ID] = variable
	}

	variables := make([]models.Variable, 0, len(variableIDs))
	for _, id := range variableIDs {
		variable, ok := variableMap[id]
		if !ok {
			return nil, &ErrReferencedVariablesNotFound{VariableIDs: []string{id.Hex()}}
		}

		variable.Type = variable.EffectiveType()
		variables = append(variables, variable)
	}

	now := time.Now().UTC()
	return &models.SavedConciergeVersion{
		Variables:   variables,
		ID:          bson.NewObjectID(),
		ConciergeID: concierge.ID,
		Name:        concierge.Name, Description: concierge.Description,
		Version:      concierge.NextVersion,
		Agents:       concierge.Agents,
		Instructions: instructions,
		Tools:        tools,
		AuditFields:  models.AuditFields{CreatedAt: now, UpdatedAt: now},
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

// CreateConcierge maps DTO to a new Concierge entity (with audit timestamps), persists it through repository,
// and returns the mapped ConciergeResponseDTO.
func (s *ConciergeService) CreateConcierge(ctx context.Context, input dto.CreateConciergeDTO) (*dto.ConciergeResponseDTO, error) {

	entity := mapper.ToInitialConciergeEntity(input)
	created, err := s.conciergeRepo.Create(ctx, entity)
	if errors.Is(err, repository.ErrConciergeNameExists) {
		return nil, ErrConciergeNameExists
	}
	if err != nil {
		return nil, err
	}

	response := mapper.ToConciergeResponseDTO(created)
	return &response, nil
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

// SetConciergeVersionDeployment changes only deployment metadata, preserving snapshot content.
func (s *ConciergeService) SetConciergeVersionDeployment(
	ctx context.Context,
	conciergeIDStr, versionIDStr, environmentIDStr string,
	deploy bool,
) (*dto.SavedConciergeVersionResponseDTO, error) {
	conciergeID, err := bson.ObjectIDFromHex(conciergeIDStr)
	if err != nil {
		return nil, ErrConciergeNotFound
	}

	versionID, err := bson.ObjectIDFromHex(versionIDStr)
	if err != nil {
		return nil, ErrConciergeVersionNotFound
	}

	environmentID, err := bson.ObjectIDFromHex(environmentIDStr)
	if err != nil {
		return nil, ErrEnvironmentNotFound
	}

	environment, err := s.environmentRepo.GetByID(ctx, environmentID)
	if err != nil {
		return nil, err
	}
	if environment == nil {
		return nil, ErrEnvironmentNotFound
	}

	saved, err := s.savedConciergeRepo.SetDeployment(ctx, conciergeID, versionID, environmentID, deploy)
	if err != nil {
		return nil, err
	}
	if saved == nil {
		return nil, ErrConciergeVersionNotFound
	}

	response := mapper.ToSavedConciergeVersionResponseDTO(saved)
	return &response, nil
}
