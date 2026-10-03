package service

import (
	"context"
	"errors"
	"testing"

	"github.com/bangweiz/dubhu-designer/internal/dto"
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
