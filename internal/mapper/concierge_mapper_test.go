package mapper

import (
	"testing"
	"time"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestToInitialConciergeEntity(t *testing.T) {
	req := dto.CreateConciergeDTO{
		Name:        "concierge_agent",
		Description: "General customer concierge agent",
	}

	entity := ToInitialConciergeEntity(req)

	if entity.Name != req.Name {
		t.Errorf("expected name %s, got %s", req.Name, entity.Name)
	}
	if entity.Description != req.Description {
		t.Errorf("expected description %s, got %s", req.Description, entity.Description)
	}
	if entity.Version != 1 {
		t.Errorf("expected version 1, got %d", entity.Version)
	}
	if entity.ID.IsZero() {
		t.Errorf("expected non-zero ObjectID")
	}
	if entity.Agents == nil {
		t.Errorf("expected non-nil Agents slice")
	}
	if len(entity.Agents) != 0 {
		t.Errorf("expected empty Agents, got %d", len(entity.Agents))
	}
	if entity.CreatedAt.IsZero() || entity.UpdatedAt.IsZero() {
		t.Errorf("expected non-zero timestamps")
	}
}

func TestToConciergeResponseDTO(t *testing.T) {
	id := bson.NewObjectID()
	agentID := bson.NewObjectID()
	now := time.Now().UTC()

	entity := &models.Concierge{
		ID:          id,
		Name:        "front_desk",
		Description: "Front desk routing",
		Agents: []models.Agent{
			{
				ID:          agentID,
				ConciergeID: id,
				Name:        "support_agent",
				Description: "Handles support",
				Model:       models.ModelGemini35Flash,
				Version:     1,
			},
		},
		Version:   2,
		CreatedAt: now,
		UpdatedAt: now,
	}

	dtoRes := ToConciergeResponseDTO(entity)

	if dtoRes.ID != id.Hex() {
		t.Errorf("expected ID %s, got %s", id.Hex(), dtoRes.ID)
	}
	if dtoRes.Name != "front_desk" {
		t.Errorf("expected Name 'front_desk', got %s", dtoRes.Name)
	}
	if dtoRes.Description != "Front desk routing" {
		t.Errorf("expected Description 'Front desk routing', got %s", dtoRes.Description)
	}
	if len(dtoRes.Agents) != 1 || dtoRes.Agents[0].ID != agentID.Hex() {
		t.Errorf("expected Agents [%s], got %v", agentID.Hex(), dtoRes.Agents)
	}
	if dtoRes.Agents[0].Name != "support_agent" {
		t.Errorf("expected agent name 'support_agent', got %s", dtoRes.Agents[0].Name)
	}
	if dtoRes.Version != 2 {
		t.Errorf("expected Version 2, got %d", dtoRes.Version)
	}
}

func TestToConciergeResponseDTO_EmptyAgents(t *testing.T) {
	id := bson.NewObjectID()
	now := time.Now().UTC()

	entity := &models.Concierge{
		ID:          id,
		Name:        "front_desk",
		Description: "Front desk routing",
		Agents:      []models.Agent{},
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	dtoRes := ToConciergeResponseDTO(entity)
	if dtoRes.Agents == nil {
		t.Errorf("expected non-nil Agents slice")
	}
	if len(dtoRes.Agents) != 0 {
		t.Errorf("expected empty Agents slice, got %d", len(dtoRes.Agents))
	}
}

func TestToConciergeResponseDTOList(t *testing.T) {
	now := time.Now().UTC()
	concierges := []models.Concierge{
		{
			ID:          bson.NewObjectID(),
			Name:        "concierge_1",
			Description: "desc 1",
			Agents:      []models.Agent{},
			Version:     1,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          bson.NewObjectID(),
			Name:        "concierge_2",
			Description: "desc 2",
			Agents:      []models.Agent{},
			Version:     2,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	list := ToConciergeResponseDTOList(concierges)
	if len(list) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list))
	}
	if list[0].Name != "concierge_1" || list[1].Name != "concierge_2" {
		t.Errorf("unexpected list content: %v", list)
	}
}

func TestToConciergeSummaryResponseDTO(t *testing.T) {
	id := bson.NewObjectID()
	now := time.Now().UTC()

	entity := &models.Concierge{
		ID:          id,
		Name:        "front_desk",
		Description: "Front desk routing",
		Agents: []models.Agent{
			{ID: bson.NewObjectID()},
		},
		Version:   3,
		CreatedAt: now,
		UpdatedAt: now,
	}

	summary := ToConciergeSummaryResponseDTO(entity)

	if summary.ID != id.Hex() {
		t.Errorf("expected ID %s, got %s", id.Hex(), summary.ID)
	}
	if summary.Name != "front_desk" {
		t.Errorf("expected Name 'front_desk', got %s", summary.Name)
	}
	if summary.Description != "Front desk routing" {
		t.Errorf("expected Description 'Front desk routing', got %s", summary.Description)
	}
	if summary.Version != 3 {
		t.Errorf("expected Version 3, got %d", summary.Version)
	}
	if summary.CreatedAt != now || summary.UpdatedAt != now {
		t.Errorf("expected timestamps to match")
	}
}

func TestToConciergeSummaryResponseDTOList(t *testing.T) {
	now := time.Now().UTC()
	concierges := []models.Concierge{
		{
			ID:          bson.NewObjectID(),
			Name:        "concierge_1",
			Description: "desc 1",
			Version:     1,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          bson.NewObjectID(),
			Name:        "concierge_2",
			Description: "desc 2",
			Version:     2,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	list := ToConciergeSummaryResponseDTOList(concierges)
	if len(list) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list))
	}
	if list[0].Name != "concierge_1" || list[1].Name != "concierge_2" {
		t.Errorf("unexpected list content: %v", list)
	}
	if list[0].ID != concierges[0].ID.Hex() || list[1].ID != concierges[1].ID.Hex() {
		t.Errorf("unexpected IDs in list: %v", list)
	}
}
