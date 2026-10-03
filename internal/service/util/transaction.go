package util

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// RunInTransaction executes fn within a MongoDB transaction if a client is configured.
// If client is nil (e.g. during certain unit tests), fn is executed directly with ctx.
// Uses Go generics to provide type-safe return values without requiring caller type assertions.
func RunInTransaction[T any](ctx context.Context, client *mongo.Client, fn func(sessCtx context.Context) (T, error)) (T, error) {
	if client == nil {
		return fn(ctx)
	}

	session, err := client.StartSession()
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to start transaction session: %w", err)
	}
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, func(sessCtx context.Context) (any, error) {
		return fn(sessCtx)
	})
	if err != nil {
		var zero T
		return zero, err
	}

	return result.(T), nil
}
