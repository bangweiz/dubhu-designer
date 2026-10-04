package mapper

import (
	"testing"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestToInitialAgentEntity(t *testing.T) {
	conciergeID := bson.NewObjectID()
	req := dto.CreateAgentDTO{
		Name:        "researcher",
		Description: "Research agent",
		Model:       string(models.ModelGemini35Flash),
	}

	entity := ToInitialAgentEntity(conciergeID, req)

	if entity.ConciergeID != conciergeID {
		t.Errorf("expected conciergeID %s, got %s", conciergeID.Hex(), entity.ConciergeID.Hex())
	}
	if entity.Name != req.Name {
		t.Errorf("expected name %s, got %s", req.Name, entity.Name)
	}
	if entity.Description != req.Description {
		t.Errorf("expected description %s, got %s", req.Description, entity.Description)
	}
	if entity.Model != models.ModelGemini35Flash {
		t.Errorf("expected model %s, got %s", req.Model, entity.Model)
	}
	if entity.Version != 1 {
		t.Errorf("expected version 1, got %d", entity.Version)
	}
	if entity.ID.IsZero() {
		t.Errorf("expected non-zero ObjectID")
	}
}

func TestToAgentResponseDTO(t *testing.T) {
	id := bson.NewObjectID()
	conciergeID := bson.NewObjectID()

	entity := &models.Agent{
		ID:          id,
		ConciergeID: conciergeID,
		Name:        "coder",
		Description: "Coding agent",
		Model:       models.ModelGemini35FlashLite,
		Version:     3,
	}

	dtoRes := ToAgentResponseDTO(entity)

	if dtoRes.ID != id.Hex() {
		t.Errorf("expected ID %s, got %s", id.Hex(), dtoRes.ID)
	}
	if dtoRes.ConciergeID != conciergeID.Hex() {
		t.Errorf("expected ConciergeID %s, got %s", conciergeID.Hex(), dtoRes.ConciergeID)
	}
	if dtoRes.Name != "coder" {
		t.Errorf("expected Name 'coder', got %s", dtoRes.Name)
	}
	if dtoRes.Description != "Coding agent" {
		t.Errorf("expected Description 'Coding agent', got %s", dtoRes.Description)
	}
	if dtoRes.Model != string(models.ModelGemini35FlashLite) {
		t.Errorf("expected Model '%s', got %s", models.ModelGemini35FlashLite, dtoRes.Model)
	}
	if dtoRes.Version != 3 {
		t.Errorf("expected Version 3, got %d", dtoRes.Version)
	}
}

func TestToAgentResponseDTOList(t *testing.T) {
	conciergeID := bson.NewObjectID()
	agents := []models.Agent{
		{
			ID:          bson.NewObjectID(),
			ConciergeID: conciergeID,
			Name:        "agent_1",
			Description: "desc 1",
			Model:       models.ModelGemini35Flash,
			Version:     1,
		},
		{
			ID:          bson.NewObjectID(),
			ConciergeID: conciergeID,
			Name:        "agent_2",
			Description: "desc 2",
			Model:       models.ModelGemini35FlashLite,
			Version:     2,
		},
	}

	list := ToAgentResponseDTOList(agents)
	if len(list) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list))
	}
	if list[0].Name != "agent_1" || list[1].Name != "agent_2" {
		t.Errorf("unexpected list content: %v", list)
	}
	if list[0].ConciergeID != conciergeID.Hex() || list[1].ConciergeID != conciergeID.Hex() {
		t.Errorf("unexpected ConciergeID in list: %v", list)
	}
}
