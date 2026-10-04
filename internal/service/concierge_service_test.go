package service

import (
	"context"
	"errors"
	"testing"
)

func TestConciergeService_GetConciergeByID_InvalidHex(t *testing.T) {
	svc := NewConciergeService(nil)
	ctx := context.Background()

	_, err := svc.GetConciergeByID(ctx, "invalid-hex")
	if !errors.Is(err, ErrConciergeNotFound) {
		t.Errorf("expected ErrConciergeNotFound, got %v", err)
	}
}

func TestConciergeService_NilConciergeRepo_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic when conciergeRepo is nil, but did not panic")
		}
	}()

	svc := NewConciergeService(nil)
	validHex := "651da3b57f8641a27e36c579"
	_, _ = svc.GetConciergeByID(context.Background(), validHex)
}
