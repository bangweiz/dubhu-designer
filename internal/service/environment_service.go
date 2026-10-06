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

type EnvironmentService struct {
	environmentRepo *repository.EnvironmentRepository
}

func (s *EnvironmentService) GetEnvironmentByID(ctx context.Context, idStr string) (*dto.EnvironmentResponseDTO, error) {
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrEnvironmentNotFound
	}
	e, err := s.environmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, ErrEnvironmentNotFound
	}
	response := mapper.ToEnvironmentResponseDTO(e)
	return &response, nil
}

func NewEnvironmentService(repo *repository.EnvironmentRepository) *EnvironmentService {
	return &EnvironmentService{environmentRepo: repo}
}

func (s *EnvironmentService) CreateEnvironment(ctx context.Context, input dto.CreateEnvironmentDTO) (*dto.EnvironmentResponseDTO, error) {
	e, err := s.environmentRepo.Create(ctx, mapper.ToInitialEnvironmentEntity(input))
	if errors.Is(err, repository.ErrEnvironmentNameExists) {
		return nil, ErrEnvironmentNameExists
	}
	if err != nil {
		return nil, err
	}
	response := mapper.ToEnvironmentResponseDTO(e)
	return &response, nil
}

func (s *EnvironmentService) ListEnvironments(ctx context.Context) ([]dto.EnvironmentResponseDTO, error) {
	environments, err := s.environmentRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]dto.EnvironmentResponseDTO, 0, len(environments))
	for _, e := range environments {
		result = append(result, mapper.ToEnvironmentResponseDTO(&e))
	}
	return result, nil
}

func (s *EnvironmentService) UpdateEnvironment(ctx context.Context, idStr, ifMatch string, input dto.UpdateEnvironmentDTO) (*dto.EnvironmentResponseDTO, error) {
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrEnvironmentNotFound
	}
	version, err := etag.Parse(ifMatch)
	if err != nil {
		return nil, ErrEnvironmentETagMismatch
	}
	e, err := s.environmentRepo.Update(ctx, id, version, input.Name, input.Description)
	if errors.Is(err, repository.ErrVersionConflict) {
		existing, lookupErr := s.environmentRepo.GetByID(ctx, id)
		if lookupErr != nil {
			return nil, lookupErr
		}
		if existing == nil {
			return nil, ErrEnvironmentNotFound
		}
		return nil, ErrEnvironmentETagMismatch
	}
	if errors.Is(err, repository.ErrEnvironmentNameExists) {
		return nil, ErrEnvironmentNameExists
	}
	if err != nil {
		return nil, err
	}
	response := mapper.ToEnvironmentResponseDTO(e)
	return &response, nil
}
