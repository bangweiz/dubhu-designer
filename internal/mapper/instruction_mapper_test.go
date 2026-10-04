package mapper

import (
	"testing"
	"time"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestToInitialInstructionEntity(t *testing.T) {
	id1 := bson.NewObjectID()
	toolIDs := []bson.ObjectID{id1}
	req := dto.CreateInstructionDTO{
		Name:    "my_instruction",
		Content: "Use {{tool:" + id1.Hex() + "}}",
	}

	entity := ToInitialInstructionEntity(req, toolIDs)

	if entity.Name != req.Name {
		t.Errorf("expected name %s, got %s", req.Name, entity.Name)
	}
	if entity.Content != req.Content {
		t.Errorf("expected content %s, got %s", req.Content, entity.Content)
	}
	if entity.Version != 1 {
		t.Errorf("expected version 1, got %d", entity.Version)
	}
	if len(entity.Tools) != 1 || entity.Tools[0] != id1 {
		t.Errorf("expected tools [%s], got %v", id1.Hex(), entity.Tools)
	}
	if entity.CreatedAt.IsZero() || entity.UpdatedAt.IsZero() {
		t.Errorf("expected non-zero timestamps")
	}
}

func TestToInstructionSummaryResponseDTO(t *testing.T) {
	id := bson.NewObjectID()
	now := time.Now().UTC()

	entity := &models.Instruction{
		ID:        id,
		Name:      "summary_inst",
		Content:   "some content",
		Tools:     []bson.ObjectID{bson.NewObjectID()},
		Version:   3,
		CreatedAt: now,
		UpdatedAt: now,
	}

	dtoRes := ToInstructionSummaryResponseDTO(entity)

	if dtoRes.ID != id.Hex() {
		t.Errorf("expected ID %s, got %s", id.Hex(), dtoRes.ID)
	}
	if dtoRes.Name != "summary_inst" {
		t.Errorf("expected Name 'summary_inst', got %s", dtoRes.Name)
	}
	if dtoRes.Content != "some content" {
		t.Errorf("expected Content 'some content', got %s", dtoRes.Content)
	}
	if dtoRes.Version != 3 {
		t.Errorf("expected Version 3, got %d", dtoRes.Version)
	}
}

func TestToInstructionSummaryResponseDTOList(t *testing.T) {
	instructions := []models.Instruction{
		{
			ID:        bson.NewObjectID(),
			Name:      "inst_1",
			Content:   "content 1",
			Version:   1,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		{
			ID:        bson.NewObjectID(),
			Name:      "inst_2",
			Content:   "content 2",
			Version:   2,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}

	list := ToInstructionSummaryResponseDTOList(instructions)
	if len(list) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list))
	}
	if list[0].Name != "inst_1" || list[1].Name != "inst_2" {
		t.Errorf("unexpected list content: %v", list)
	}
}

func TestToInstructionResponseDTO(t *testing.T) {
	id := bson.NewObjectID()
	toolID := bson.NewObjectID()
	now := time.Now().UTC()

	entity := &models.Instruction{
		ID:        id,
		Name:      "test_inst",
		Content:   "some content",
		Tools:     []bson.ObjectID{toolID},
		Version:   2,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tools := []models.Tool{
		{
			ID:          toolID,
			Name:        "web_search",
			Description: "Searches the web",
			Version:     1,
			Inputs:      []models.ToolInput{},
			Outputs:     []models.ToolOutput{},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	dtoRes := ToInstructionResponseDTO(entity, tools)

	if dtoRes.ID != id.Hex() {
		t.Errorf("expected ID %s, got %s", id.Hex(), dtoRes.ID)
	}
	if dtoRes.Name != "test_inst" {
		t.Errorf("expected Name 'test_inst', got %s", dtoRes.Name)
	}
	if dtoRes.Content != "some content" {
		t.Errorf("expected Content 'some content', got %s", dtoRes.Content)
	}
	if len(dtoRes.Tools) != 1 || dtoRes.Tools[0].ID != toolID.Hex() {
		t.Errorf("expected Tools [%s], got %v", toolID.Hex(), dtoRes.Tools)
	}
	if dtoRes.Tools[0].Name != "web_search" {
		t.Errorf("expected Tool Name 'web_search', got %s", dtoRes.Tools[0].Name)
	}
	if dtoRes.Version != 2 {
		t.Errorf("expected Version 2, got %d", dtoRes.Version)
	}
}

func TestToInstructionResponseDTO_EmptyTools(t *testing.T) {
	id := bson.NewObjectID()
	now := time.Now().UTC()

	entity := &models.Instruction{
		ID:        id,
		Name:      "no_tools",
		Content:   "plain instruction",
		Tools:     []bson.ObjectID{},
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	dtoRes := ToInstructionResponseDTO(entity, []models.Tool{})
	if dtoRes.Tools == nil {
		t.Errorf("expected non-nil empty slice for Tools, got nil")
	}
	if len(dtoRes.Tools) != 0 {
		t.Errorf("expected 0 tools, got %d", len(dtoRes.Tools))
	}
}

func TestToPopulatedInstructionResponseDTO(t *testing.T) {
	id := bson.NewObjectID()
	toolID1 := bson.NewObjectID()
	toolID2 := bson.NewObjectID()
	now := time.Now().UTC()

	populated := &models.PopulatedInstruction{
		Instruction: models.Instruction{
			ID:        id,
			Name:      "ordered_tools_instruction",
			Content:   "Runs tool 2 then tool 1",
			Tools:     []bson.ObjectID{toolID2, toolID1}, // Order: tool2, then tool1
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		ResolvedTools: []models.Tool{
			// Mongo aggregation lookup may return tools in arbitrary/index order (e.g. tool1 first)
			{
				ID:          toolID1,
				Name:        "tool_one",
				Description: "First tool",
				Inputs:      []models.ToolInput{},
				Outputs:     []models.ToolOutput{},
				Version:     1,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			{
				ID:          toolID2,
				Name:        "tool_two",
				Description: "Second tool",
				Inputs:      []models.ToolInput{},
				Outputs:     []models.ToolOutput{},
				Version:     1,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
	}

	dtoRes := ToPopulatedInstructionResponseDTO(populated)

	if dtoRes.ID != id.Hex() {
		t.Errorf("expected ID %s, got %s", id.Hex(), dtoRes.ID)
	}
	if len(dtoRes.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(dtoRes.Tools))
	}
	// Verify deterministic ordering matches Instruction.Tools order (toolID2 first, toolID1 second)
	if dtoRes.Tools[0].ID != toolID2.Hex() || dtoRes.Tools[0].Name != "tool_two" {
		t.Errorf("expected first tool to be tool_two, got %v", dtoRes.Tools[0])
	}
	if dtoRes.Tools[1].ID != toolID1.Hex() || dtoRes.Tools[1].Name != "tool_one" {
		t.Errorf("expected second tool to be tool_one, got %v", dtoRes.Tools[1])
	}
}

func TestToPopulatedInstructionResponseDTO_EmptyTools(t *testing.T) {
	id := bson.NewObjectID()
	now := time.Now().UTC()

	populated := &models.PopulatedInstruction{
		Instruction: models.Instruction{
			ID:        id,
			Name:      "no_tools_inst",
			Content:   "content",
			Tools:     []bson.ObjectID{},
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		ResolvedTools: []models.Tool{},
	}

	dtoRes := ToPopulatedInstructionResponseDTO(populated)
	if dtoRes.Tools == nil {
		t.Errorf("expected non-nil empty slice for Tools, got nil")
	}
	if len(dtoRes.Tools) != 0 {
		t.Errorf("expected 0 tools, got %d", len(dtoRes.Tools))
	}
}
