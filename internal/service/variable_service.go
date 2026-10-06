package service

import (
	"context"
	"errors"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/etag"
	"github.com/bangweiz/dubhu-designer/internal/mapper"
	"github.com/bangweiz/dubhu-designer/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type VariableService struct {
	variableRepo *repository.VariableRepository
}

func (s *VariableService) GetVariableByID(ctx context.Context, idStr string) (*dto.VariableResponseDTO, error) {
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrVariableNotFound
	}
	e, err := s.variableRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, ErrVariableNotFound
	}
	response := mapper.ToVariableResponseDTO(e)
	return &response, nil
}

func NewVariableService(repo *repository.VariableRepository) *VariableService {
	return &VariableService{variableRepo: repo}
}

func (s *VariableService) CreateVariable(ctx context.Context, input dto.CreateVariableDTO) (*dto.VariableResponseDTO, error) {
	e, err := s.variableRepo.Create(ctx, mapper.ToInitialVariableEntity(input))
	if errors.Is(err, repository.ErrVariableNameExists) {
		return nil, ErrVariableNameExists
	}
	if err != nil {
		return nil, err
	}
	response := mapper.ToVariableResponseDTO(e)
	return &response, nil
}

func (s *VariableService) ListVariables(ctx context.Context) ([]dto.VariableResponseDTO, error) {
	variables, err := s.variableRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]dto.VariableResponseDTO, 0, len(variables))
	for _, e := range variables {
		result = append(result, mapper.ToVariableResponseDTO(&e))
	}
	return result, nil
}

func (s *VariableService) UpdateVariable(ctx context.Context, idStr, ifMatch string, input dto.UpdateVariableDTO) (*dto.VariableResponseDTO, error) {
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrVariableNotFound
	}
	version, err := etag.Parse(ifMatch)
	if err != nil {
		return nil, ErrVariableETagMismatch
	}
	e, err := s.variableRepo.Update(ctx, id, version, input.Name, input.Description)
	if errors.Is(err, repository.ErrVersionConflict) {
		existing, lookupErr := s.variableRepo.GetByID(ctx, id)
		if lookupErr != nil {
			return nil, lookupErr
		}
		if existing == nil {
			return nil, ErrVariableNotFound
		}
		return nil, ErrVariableETagMismatch
	}
	if errors.Is(err, repository.ErrVariableNameExists) {
		return nil, ErrVariableNameExists
	}
	if err != nil {
		return nil, err
	}
	response := mapper.ToVariableResponseDTO(e)
	return &response, nil
}
