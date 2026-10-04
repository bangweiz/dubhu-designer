package repository

import (
	"context"
	"testing"
	"time"

	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestPopulatedInstruction_BSONDecode(t *testing.T) {
	instID := bson.NewObjectID()
	toolID := bson.NewObjectID()
	now := time.Now().UTC().Truncate(time.Millisecond)

	rawDoc := bson.M{
		"_id":        instID,
		"name":       "test_lookup_inst",
		"content":    "Use tool {{tool:" + toolID.Hex() + "}}",
		"tools":      []bson.ObjectID{toolID},
		"version":    2,
		"created_at": now,
		"updated_at": now,
		"resolved_tools": []bson.M{
			{
				"_id":         toolID,
				"name":        "search_tool",
				"description": "Searches data",
				"inputs":      []bson.M{},
				"outputs":     []bson.M{},
				"version":     1,
				"created_at":  now,
				"updated_at":  now,
			},
		},
	}

	rawBytes, err := bson.Marshal(rawDoc)
	if err != nil {
		t.Fatalf("failed to marshal rawDoc: %v", err)
	}

	var populated models.PopulatedInstruction
	if err := bson.Unmarshal(rawBytes, &populated); err != nil {
		t.Fatalf("failed to unmarshal into PopulatedInstruction: %v", err)
	}

	if populated.ID != instID {
		t.Errorf("expected ID %s, got %s", instID.Hex(), populated.ID.Hex())
	}
	if populated.Name != "test_lookup_inst" {
		t.Errorf("expected name 'test_lookup_inst', got %s", populated.Name)
	}
	if populated.Version != 2 {
		t.Errorf("expected version 2, got %d", populated.Version)
	}
	if len(populated.ResolvedTools) != 1 {
		t.Fatalf("expected 1 resolved tool, got %d", len(populated.ResolvedTools))
	}
	if populated.ResolvedTools[0].ID != toolID {
		t.Errorf("expected resolved tool ID %s, got %s", toolID.Hex(), populated.ResolvedTools[0].ID.Hex())
	}
	if populated.ResolvedTools[0].Name != "search_tool" {
		t.Errorf("expected resolved tool name 'search_tool', got %s", populated.ResolvedTools[0].Name)
	}
}

func TestInstructionRepository_GetByIDWithTools_Integration(t *testing.T) {
	uri := "mongodb://localhost:27017/?directConnection=true"
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
	}
	defer client.Disconnect(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		t.Skip("MongoDB ping failed, skipping integration test")
	}

	testDBName := "dubhu_test_instruction_lookup"
	database := client.Database(testDBName)
	defer database.Drop(context.Background())

	instRepo := NewInstructionRepository(database)
	toolRepo := NewToolRepository(database)

	// Create 2 tools
	t1, err := toolRepo.Create(ctx, &models.Tool{
		ID:          bson.NewObjectID(),
		Name:        "tool_alpha",
		Description: "Alpha tool",
		Inputs:      []models.ToolInput{},
		Outputs:     []models.ToolOutput{},
		Version:     1,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("failed to insert tool 1: %v", err)
	}

	t2, err := toolRepo.Create(ctx, &models.Tool{
		ID:          bson.NewObjectID(),
		Name:        "tool_beta",
		Description: "Beta tool",
		Inputs:      []models.ToolInput{},
		Outputs:     []models.ToolOutput{},
		Version:     1,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("failed to insert tool 2: %v", err)
	}

	// Case 1: Instruction with tools
	inst1, err := instRepo.Create(ctx, &models.Instruction{
		ID:        bson.NewObjectID(),
		Name:      "inst_with_tools",
		Content:   "Runs tool alpha and beta",
		Tools:     []bson.ObjectID{t1.ID, t2.ID},
		Version:   1,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("failed to insert inst1: %v", err)
	}

	// Case 2: Instruction with empty tools
	inst2, err := instRepo.Create(ctx, &models.Instruction{
		ID:        bson.NewObjectID(),
		Name:      "inst_empty_tools",
		Content:   "No tools needed",
		Tools:     []bson.ObjectID{},
		Version:   1,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("failed to insert inst2: %v", err)
	}

	t.Run("returns instruction with resolved tools", func(t *testing.T) {
		pop, err := instRepo.GetByIDWithTools(ctx, inst1.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pop == nil {
			t.Fatalf("expected non-nil populated instruction")
		}
		if pop.ID != inst1.ID {
			t.Errorf("expected ID %s, got %s", inst1.ID.Hex(), pop.ID.Hex())
		}
		if pop.Name != "inst_with_tools" {
			t.Errorf("expected name 'inst_with_tools', got %s", pop.Name)
		}
		if len(pop.ResolvedTools) != 2 {
			t.Fatalf("expected 2 resolved tools, got %d", len(pop.ResolvedTools))
		}
	})

	t.Run("returns instruction with empty resolved tools when no tools referenced", func(t *testing.T) {
		pop, err := instRepo.GetByIDWithTools(ctx, inst2.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pop == nil {
			t.Fatalf("expected non-nil populated instruction")
		}
		if len(pop.ResolvedTools) != 0 {
			t.Errorf("expected 0 resolved tools, got %d", len(pop.ResolvedTools))
		}
	})

	t.Run("returns nil, nil when instruction does not exist", func(t *testing.T) {
		nonExistentID := bson.NewObjectID()
		pop, err := instRepo.GetByIDWithTools(ctx, nonExistentID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pop != nil {
			t.Errorf("expected nil for non-existent instruction, got %v", pop)
		}
	})
}
