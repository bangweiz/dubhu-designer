package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"github.com/bangweiz/dubhu-designer/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestErrReferencedToolsNotFound_Error(t *testing.T) {
	err := &ErrReferencedToolsNotFound{
		ToolIDs: []string{"651da3b57f8641a27e36c579", "651da3b57f8641a27e36c580"},
	}
	expected := "referenced tools do not exist: 651da3b57f8641a27e36c579, 651da3b57f8641a27e36c580"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestInstructionService_ResolveTools_NoTools(t *testing.T) {
	svc := NewInstructionService(nil, nil)
	ctx := context.Background()

	ids, tools, err := svc.resolveTools(ctx, "Instruction without any tool chips.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected 0 ids, got %d", len(ids))
	}
	if len(tools) != 0 {
		t.Errorf("expected 0 tools, got %d", len(tools))
	}
}

func TestInstructionService_ResolveTools_InvalidHexIDs(t *testing.T) {
	svc := NewInstructionService(nil, nil)
	ctx := context.Background()

	_, _, err := svc.resolveTools(ctx, "Instruction referencing {{tool:invalid_hex_id}} and {{tool:also_invalid}}.")
	if err == nil {
		t.Fatalf("expected ErrReferencedToolsNotFound, got nil")
	}

	var refErr *ErrReferencedToolsNotFound
	if !errors.As(err, &refErr) {
		t.Fatalf("expected error of type *ErrReferencedToolsNotFound, got %T: %v", err, err)
	}

	if len(refErr.ToolIDs) != 2 {
		t.Fatalf("expected 2 missing tool IDs, got %d", len(refErr.ToolIDs))
	}
	if refErr.ToolIDs[0] != "invalid_hex_id" || refErr.ToolIDs[1] != "also_invalid" {
		t.Errorf("unexpected missing tool IDs: %v", refErr.ToolIDs)
	}
}

func TestInstructionService_ResolveTools_MixedValidAndInvalidIDs_NoRepoCall(t *testing.T) {
	// toolRepo is nil, so calling toolRepo.FindByIDs would cause a panic.
	// Because invalid IDs exist, resolveTools must return directly without calling toolRepo.
	svc := NewInstructionService(nil, nil)
	ctx := context.Background()

	content := "Mix of valid {{tool:651da3b57f8641a27e36c579}} and invalid {{tool:invalid_id}}."
	_, _, err := svc.resolveTools(ctx, content)
	if err == nil {
		t.Fatalf("expected ErrReferencedToolsNotFound, got nil")
	}

	var refErr *ErrReferencedToolsNotFound
	if !errors.As(err, &refErr) {
		t.Fatalf("expected error of type *ErrReferencedToolsNotFound, got %T: %v", err, err)
	}

	if len(refErr.ToolIDs) != 1 || refErr.ToolIDs[0] != "invalid_id" {
		t.Errorf("expected [invalid_id], got %v", refErr.ToolIDs)
	}
}

func TestInstructionService_GetInstructionByID_InvalidHex(t *testing.T) {
	svc := NewInstructionService(nil, nil)
	ctx := context.Background()

	_, err := svc.GetInstructionByID(ctx, "not-a-valid-hex")
	if !errors.Is(err, ErrInstructionNotFound) {
		t.Errorf("expected ErrInstructionNotFound, got %v", err)
	}
}

func TestInstructionService_UpdateInstruction_InvalidHex(t *testing.T) {
	svc := NewInstructionService(nil, nil)
	ctx := context.Background()

	_, err := svc.UpdateInstruction(ctx, "not-a-valid-hex", dto.UpdateInstructionDTO{
		Name:    "test",
		Content: "test",
		Version: 1,
	})
	if !errors.Is(err, ErrInstructionNotFound) {
		t.Errorf("expected ErrInstructionNotFound, got %v", err)
	}
}

func TestInstructionService_GetInstructionByID_Integration(t *testing.T) {
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

	testDB := client.Database("dubhu_test_instruction_service")
	defer testDB.Drop(context.Background())

	instRepo := repository.NewInstructionRepository(testDB)
	toolRepo := repository.NewToolRepository(testDB)
	svc := NewInstructionService(instRepo, toolRepo)

	// Create tool
	tool, err := toolRepo.Create(ctx, &models.Tool{
		ID:          bson.NewObjectID(),
		Name:        "service_tool",
		Description: "Service Tool Description",
		Inputs:      []models.ToolInput{},
		Outputs:     []models.ToolOutput{},
		Version:     1,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	// Create instruction
	createDTO := dto.CreateInstructionDTO{
		Name:    "service_test_inst",
		Content: fmt.Sprintf("Use tool {{tool:%s}}", tool.ID.Hex()),
	}
	created, err := svc.CreateInstruction(ctx, createDTO)
	if err != nil {
		t.Fatalf("failed to create instruction: %v", err)
	}

	// Test GetInstructionByID with existing ID
	got, err := svc.GetInstructionByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("unexpected error getting instruction by ID: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, got.ID)
	}
	if got.Name != "service_test_inst" {
		t.Errorf("expected Name 'service_test_inst', got %s", got.Name)
	}
	if len(got.Tools) != 1 || got.Tools[0].ID != tool.ID.Hex() {
		t.Fatalf("expected 1 tool with ID %s, got %v", tool.ID.Hex(), got.Tools)
	}
	if got.Tools[0].Name != "service_tool" {
		t.Errorf("expected tool name 'service_tool', got %s", got.Tools[0].Name)
	}

	// Test GetInstructionByID with non-existent ObjectID
	nonExistentHex := bson.NewObjectID().Hex()
	_, err = svc.GetInstructionByID(ctx, nonExistentHex)
	if !errors.Is(err, ErrInstructionNotFound) {
		t.Errorf("expected ErrInstructionNotFound for non-existent ID, got %v", err)
	}
}
