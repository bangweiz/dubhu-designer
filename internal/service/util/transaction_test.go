package util

import (
	"context"
	"errors"
	"testing"
)

func TestRunInTransaction_NilClient(t *testing.T) {
	ctx := context.Background()

	t.Run("successful execution with string", func(t *testing.T) {
		res, err := RunInTransaction(ctx, nil, func(sessCtx context.Context) (string, error) {
			return "success", nil
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res != "success" {
			t.Errorf("expected 'success', got %v", res)
		}
	})

	t.Run("successful execution with int", func(t *testing.T) {
		val, err := RunInTransaction(ctx, nil, func(sessCtx context.Context) (int, error) {
			return 42, nil
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if val != 42 {
			t.Errorf("expected 42, got %d", val)
		}
	})

	t.Run("propagates error", func(t *testing.T) {
		expectedErr := errors.New("something went wrong")
		res, err := RunInTransaction(ctx, nil, func(sessCtx context.Context) (string, error) {
			return "", expectedErr
		})
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected error %v, got %v", expectedErr, err)
		}
		if res != "" {
			t.Errorf("expected empty string result, got %v", res)
		}
	})
}
