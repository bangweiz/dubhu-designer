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
func (r *SavedConciergeVersionRepository) GetByID(ctx context.Context, conciergeID, versionID bson.ObjectID) (*models.SavedConciergeVersion, error) {
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

// GetByConciergeIDAndVersion retrieves an immutable snapshot by its public composite identity.
func (r *SavedConciergeVersionRepository) GetByConciergeIDAndVersion(ctx context.Context, conciergeID bson.ObjectID, version int) (*models.SavedConciergeVersion, error) {
	var saved models.SavedConciergeVersion
	err := r.collection.FindOne(ctx, bson.M{"concierge_id": conciergeID, "version": version}).Decode(&saved)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find saved concierge: %w", err)
	}
	return &saved, nil
}

func (r *SavedConciergeVersionRepository) ListByConciergeID(ctx context.Context, id bson.ObjectID) ([]models.SavedConciergeVersion, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"concierge_id": id}, options.Find().SetProjection(bson.M{"_id": 1, "version": 1}).SetSort(bson.D{{Key: "version", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	versions := []models.SavedConciergeVersion{}
	if err := cursor.All(ctx, &versions); err != nil {
		return nil, err
	}
	return versions, nil
}
