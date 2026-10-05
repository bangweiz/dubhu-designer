package service

import (
	"context"
	"errors"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/mapper"
	"github.com/bangweiz/dubhu-designer/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ConciergeService handles business logic operations for concierges.
type ConciergeService struct {
	conciergeRepo *repository.ConciergeRepository
}

// NewConciergeService creates a new ConciergeService instance.
func NewConciergeService(conciergeRepo *repository.ConciergeRepository) *ConciergeService {
	return &ConciergeService{
		conciergeRepo: conciergeRepo,
	}
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
