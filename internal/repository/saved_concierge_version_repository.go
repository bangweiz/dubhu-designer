package repository

import (
	"context"
	"fmt"

	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const collectionSavedConciergeVersions = "saved_concierge_versions"

// SavedConciergeVersionRepository stores immutable concierge version snapshots.
type SavedConciergeVersionRepository struct {
	collection *scopedCollection
	client     *mongo.Client
}

func NewSavedConciergeVersionRepository(database *mongo.Database) *SavedConciergeVersionRepository {
	return &SavedConciergeVersionRepository{
		collection: newScopedCollection(database, collectionSavedConciergeVersions),
		client:     database.Client(),
	}
}

func (r *SavedConciergeVersionRepository) Client() *mongo.Client {
	return r.client
}

// GetByID loads a snapshot scoped to its parent concierge in one query.
func (r *SavedConciergeVersionRepository) GetByID(
	ctx context.Context,
	conciergeID, versionID bson.ObjectID,
) (*models.SavedConciergeVersion, error) {
	var saved models.SavedConciergeVersion
	err := r.collection.FindOne(ctx, bson.M{"_id": versionID, "concierge_id": conciergeID}).Decode(&saved)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to find concierge version: %w", err)
	}

	return &saved, nil
}

func (r *SavedConciergeVersionRepository) InitIndexes(ctx context.Context) error {
	index := mongo.IndexModel{
		Keys:    bson.D{{Key: "concierge_id", Value: 1}, {Key: "version", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	if _, err := r.collection.Indexes().CreateOne(ctx, index); err != nil {
		return fmt.Errorf("failed to create saved concierge indexes: %w", err)
	}

	return nil
}

func (r *SavedConciergeVersionRepository) Create(ctx context.Context, saved *models.SavedConciergeVersion) error {
	if _, err := r.collection.InsertOne(ctx, saved); err != nil {
		return fmt.Errorf("failed to save concierge snapshot: %w", err)
	}

	return nil
}

// SetDeployment atomically adds or removes a single environment without losing other deployments.
func (r *SavedConciergeVersionRepository) SetDeployment(
	ctx context.Context,
	conciergeID, versionID, environmentID bson.ObjectID,
	deploy bool,
) (*models.SavedConciergeVersion, error) {
	operator := "$pull"
	if deploy {
		operator = "$addToSet"
	}

	var saved models.SavedConciergeVersion
	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": versionID, "concierge_id": conciergeID},
		bson.M{operator: bson.M{"environment_ids": environmentID}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&saved)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update concierge version deployment: %w", err)
	}

	return &saved, nil
}
