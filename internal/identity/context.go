package identity

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrMissingOrganisation = errors.New("authenticated organisation context is required")

type Principal struct {
	OrganisationID bson.ObjectID
	AccountID      bson.ObjectID
	Role           string
}

type principalKey struct{}

// WithPrincipal is set only after validating a session, never directly from URL parameters.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}
func FromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(Principal)
	return p, ok && !p.OrganisationID.IsZero() && !p.AccountID.IsZero()
}
func OrganisationID(ctx context.Context) (bson.ObjectID, error) {
	p, ok := FromContext(ctx)
	if !ok {
		return bson.NilObjectID, ErrMissingOrganisation
	}
	return p.OrganisationID, nil
}
