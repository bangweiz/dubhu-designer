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

const collectionConcierges = "concierges"

// RecordSavedVersion appends a snapshot and advances its sequence atomically.
// Snapshot insertion and this update must run in the same transaction.
func (r *ConciergeRepository) RecordSavedVersion(
	ctx context.Context,
	id, savedID bson.ObjectID,
	nextVersion int,
	expectedUpdatedAt time.Time,
) error {
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "next_version": nextVersion, "updated_at": expectedUpdatedAt}, bson.M{
		"$push": bson.M{"concierge_versions": models.ConciergeVersionReference{ConciergeVersionID: savedID, Version: nextVersion}},
		"$inc":  bson.M{"next_version": 1},
		"$set":  bson.M{"updated_at": time.Now().UTC()},
	})
	if err != nil {
		return fmt.Errorf("record saved concierge version: %w", err)
	}
	if result.MatchedCount == 0 {
		return ErrUpdateConflict
	}

	return nil
}

func (r *ConciergeRepository) Update(
	ctx context.Context,
	id bson.ObjectID,
	expectedUpdatedAt time.Time,
	name, description string,
) (*models.Concierge, error) {
	var c models.Concierge
	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id, "updated_at": expectedUpdatedAt},
		bson.M{"$set": bson.M{"name": name, "description": description, "updated_at": time.Now().UTC()}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&c)
	if err == mongo.ErrNoDocuments {
		return nil, ErrUpdateConflict
	}
	if mongo.IsDuplicateKeyError(err) {
		return nil, ErrConciergeNameExists
	}
	if err != nil {
		return nil, fmt.Errorf("update concierge: %w", err)
	}

	return &c, nil
}

// ConciergeRepository manages Concierge persistence in MongoDB.
type ConciergeRepository struct {
	collection *scopedCollection
}

// NewConciergeRepository creates a new ConciergeRepository instance.
func NewConciergeRepository(database *mongo.Database) *ConciergeRepository {
	return &ConciergeRepository{
		collection: newScopedCollection(database, collectionConcierges),
	}
}

// InitIndexes creates necessary database indexes for concierges:
// 1. Unique index on (organisation_id, name).
func (r *ConciergeRepository) InitIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "agents.instructions", Value: 1}}},
		{Keys: bson.D{{Key: "agents.tools", Value: 1}}},
		{
			Keys:    bson.D{{Key: "organisation_id", Value: 1}, {Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	if _, err := r.collection.Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("failed to create indexes on concierges collection: %w", err)
	}

	return nil
}

// Create inserts a Concierge document into MongoDB.
// Returns ErrConciergeNameExists if a concierge with the same name already exists.
func (r *ConciergeRepository) Create(ctx context.Context, concierge *models.Concierge) (*models.Concierge, error) {

	if _, err := r.collection.InsertOne(ctx, concierge); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrConciergeNameExists
		}

		return nil, fmt.Errorf("failed to insert concierge: %w", err)
	}

	return concierge, nil
}

// GetByID finds a Concierge document by its ObjectID.
func (r *ConciergeRepository) GetByID(ctx context.Context, id bson.ObjectID) (*models.Concierge, error) {
	var concierge models.Concierge
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&concierge)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to find concierge: %w", err)
	}

	return &concierge, nil
}

// List retrieves all Concierge documents from MongoDB without sorting, filtering, or pagination.
func (r *ConciergeRepository) List(ctx context.Context) ([]models.Concierge, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to query concierges: %w", err)
	}

	defer cursor.Close(ctx)

	var list []models.Concierge
	if err := cursor.All(ctx, &list); err != nil {
		return nil, fmt.Errorf("failed to decode concierges: %w", err)
	}

	if list == nil {
		list = []models.Concierge{}
	}

	return list, nil
}
