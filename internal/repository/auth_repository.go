package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// These collections are accessed by authentication itself and use explicit organisation filters.
type AuthRepository struct {
	organisations, accounts, sessions *mongo.Collection
	client                            *mongo.Client
}

func NewAuthRepository(db *mongo.Database) *AuthRepository {
	return &AuthRepository{
		organisations: db.Collection("organisations"),
		accounts:      db.Collection("accounts"),
		sessions:      db.Collection("auth_sessions"),
		client:        db.Client(),
	}
}

func (r *AuthRepository) Client() *mongo.Client { return r.client }

func (r *AuthRepository) InitIndexes(ctx context.Context) error {
	if _, err := r.organisations.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	); err != nil {
		return fmt.Errorf("create organisation indexes: %w", err)
	}
	if _, err := r.accounts.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "organisation_id", Value: 1}, {Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "organisation_id", Value: 1}, {Key: "role", Value: 1}},
			Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{"role": models.RoleRoot}),
		},
	}); err != nil {
		return fmt.Errorf("create account indexes: %w", err)
	}

	_, err := r.sessions.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
	)
	return err
}

func (r *AuthRepository) CreateOrganisation(ctx context.Context, o *models.Organisation) error {
	if _, err := r.organisations.InsertOne(ctx, o); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrOrganisationNameExists
		}

		return fmt.Errorf("create organisation: %w", err)
	}

	return nil
}

func (r *AuthRepository) CreateAccount(ctx context.Context, a *models.Account) error {
	if _, err := r.accounts.InsertOne(ctx, a); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrAccountConflict
		}

		return fmt.Errorf("create account: %w", err)
	}

	return nil
}

func (r *AuthRepository) GetOrganisation(ctx context.Context, id bson.ObjectID) (*models.Organisation, error) {
	var o models.Organisation
	err := r.organisations.FindOne(ctx, bson.M{"_id": id}).Decode(&o)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (r *AuthRepository) GetAccountByEmail(ctx context.Context, orgID bson.ObjectID, email string) (*models.Account, error) {
	return r.findAccount(ctx, bson.M{"organisation_id": orgID, "email": email})
}

func (r *AuthRepository) GetAccount(ctx context.Context, orgID, id bson.ObjectID) (*models.Account, error) {
	return r.findAccount(ctx, bson.M{"organisation_id": orgID, "_id": id})
}

func (r *AuthRepository) findAccount(ctx context.Context, filter bson.M) (*models.Account, error) {
	var a models.Account
	err := r.accounts.FindOne(ctx, filter).Decode(&a)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &a, nil
}

func (r *AuthRepository) CreateSession(ctx context.Context, s *models.AuthSession) error {
	_, err := r.sessions.InsertOne(ctx, s)
	return err
}

func (r *AuthRepository) GetSession(ctx context.Context, hash string) (*models.AuthSession, error) {
	var s models.AuthSession
	err := r.sessions.FindOne(ctx, bson.M{"_id": hash, "expires_at": bson.M{"$gt": time.Now().UTC()}}).Decode(&s)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *AuthRepository) DeleteSession(ctx context.Context, orgID, accountID bson.ObjectID, hash string) error {
	_, err := r.sessions.DeleteOne(ctx, bson.M{"_id": hash, "organisation_id": orgID, "account_id": accountID})
	return err
}
