package service

import (
	"context"
	"errors"
	"testing"

	"github.com/bangweiz/dubhu-designer/internal/dto"
)

func TestToolService_GetToolByID_InvalidHex(t *testing.T) {
	svc := NewToolService(nil)
	ctx := context.Background()

	_, err := svc.GetToolByID(ctx, "invalid-hex")
	if !errors.Is(err, ErrToolNotFound) {
		t.Errorf("expected ErrToolNotFound, got %v", err)
	}
}

func TestToolService_UpdateTool_InvalidHex(t *testing.T) {
	svc := NewToolService(nil)
	ctx := context.Background()

	_, err := svc.UpdateTool(ctx, "invalid-hex", dto.UpdateToolDTO{
		Name:    "test",
		Version: 1,
	})
	if !errors.Is(err, ErrToolNotFound) {
		t.Errorf("expected ErrToolNotFound, got %v", err)
	}
}
