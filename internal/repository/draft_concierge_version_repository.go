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

const collectionDraftConciergeVersions = "draft_concierge_versions"

// DraftConciergeVersionRepository enforces one draft per concierge with a unique concierge_id index.
type DraftConciergeVersionRepository struct{ collection *mongo.Collection }

func NewDraftConciergeVersionRepository(db *mongo.Database) *DraftConciergeVersionRepository {
	return &DraftConciergeVersionRepository{collection: db.Collection(collectionDraftConciergeVersions)}
}
func (r *DraftConciergeVersionRepository) InitIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "concierge_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "agents.instructions", Value: 1}}},
		{Keys: bson.D{{Key: "agents.tools", Value: 1}}},
	})
	return err
}
func (r *DraftConciergeVersionRepository) Create(ctx context.Context, v *models.DraftConciergeVersion) error {
	_, err := r.collection.InsertOne(ctx, v)
	return err
}
func (r *DraftConciergeVersionRepository) GetByConciergeID(ctx context.Context, id bson.ObjectID) (*models.DraftConciergeVersion, error) {
	var v models.DraftConciergeVersion
	err := r.collection.FindOne(ctx, bson.M{"concierge_id": id}).Decode(&v)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find concierge version: %w", err)
	}
	return &v, nil
}
func (r *DraftConciergeVersionRepository) Advance(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.UpdateOne(ctx, bson.M{"concierge_id": id}, bson.M{"$inc": bson.M{"version": 1, "etag_version": 1}, "$set": bson.M{"updated_at": time.Now().UTC()}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrConciergeNotFound
	}
	return nil
}
