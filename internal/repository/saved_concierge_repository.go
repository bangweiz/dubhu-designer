package repository

import (
	"context"
	"fmt"

	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const collectionSavedConcierges = "saved_concierges"

// SavedConciergeRepository stores immutable concierge snapshots.
type SavedConciergeRepository struct {
	collection *mongo.Collection
	client     *mongo.Client
}

func NewSavedConciergeRepository(database *mongo.Database) *SavedConciergeRepository {
	return &SavedConciergeRepository{
		collection: database.Collection(collectionSavedConcierges),
		client:     database.Client(),
	}
}

func (r *SavedConciergeRepository) Client() *mongo.Client {
	return r.client
}

// GetByID loads a snapshot scoped to its parent concierge in one query.
func (r *SavedConciergeRepository) GetByID(ctx context.Context, conciergeID, versionID bson.ObjectID) (*models.SavedConcierge, error) {
	var saved models.SavedConcierge
	err := r.collection.FindOne(ctx, bson.M{"_id": versionID, "concierge_id": conciergeID}).Decode(&saved)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find concierge version: %w", err)
	}
	return &saved, nil
}

func (r *SavedConciergeRepository) InitIndexes(ctx context.Context) error {
	index := mongo.IndexModel{
		Keys:    bson.D{{Key: "concierge_id", Value: 1}, {Key: "version", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	if _, err := r.collection.Indexes().CreateOne(ctx, index); err != nil {
		return fmt.Errorf("failed to create saved concierge indexes: %w", err)
	}
	return nil
}

func (r *SavedConciergeRepository) Create(ctx context.Context, saved *models.SavedConcierge) error {
	if _, err := r.collection.InsertOne(ctx, saved); err != nil {
		return fmt.Errorf("failed to save concierge snapshot: %w", err)
	}
	return nil
}

// GetByConciergeIDAndVersion retrieves an immutable snapshot by its public composite identity.
func (r *SavedConciergeRepository) GetByConciergeIDAndVersion(ctx context.Context, conciergeID bson.ObjectID, version int) (*models.SavedConcierge, error) {
	var saved models.SavedConcierge
	err := r.collection.FindOne(ctx, bson.M{"concierge_id": conciergeID, "version": version}).Decode(&saved)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find saved concierge: %w", err)
	}
	return &saved, nil
}
