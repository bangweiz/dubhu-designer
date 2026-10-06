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

// RecordSavedVersion replaces the old draft reference with its snapshot and appends the advanced draft.
// The caller must run this alongside snapshot creation and draft advancement in one transaction.
func (r *ConciergeRepository) RecordSavedVersion(ctx context.Context, id, draftID, savedID bson.ObjectID, version int) error {
	refs := "$concierge_versions"
	replacement := bson.M{"concierge_version_id": savedID, "version": version}
	draft := models.ConciergeVersionReference{ConciergeVersionID: draftID, Version: version + 1}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "concierge_versions": bson.M{"$elemMatch": bson.M{"concierge_version_id": draftID, "version": version}}}, mongo.Pipeline{bson.D{{Key: "$set", Value: bson.M{
		"concierge_versions": bson.M{"$concatArrays": bson.A{
			bson.M{"$map": bson.M{"input": refs, "as": "ref", "in": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$$ref.concierge_version_id", draftID}}, replacement, "$$ref"}}}},
			bson.A{draft},
		}},
		"version": bson.M{"$add": bson.A{"$version", 1}}, "updated_at": time.Now().UTC(),
	}}}})
	if err != nil {
		return fmt.Errorf("record saved concierge version: %w", err)
	}
	if result.MatchedCount == 0 {
		return ErrVersionConflict
	}
	return nil
}

func (r *ConciergeRepository) Update(ctx context.Context, id bson.ObjectID, version int, name, description string) (*models.Concierge, error) {
	var c models.Concierge
	err := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": id, "version": version}, bson.M{"$set": bson.M{"name": name, "description": description, "updated_at": time.Now().UTC()}, "$inc": bson.M{"version": 1}}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&c)
	if err == mongo.ErrNoDocuments {
		return nil, ErrVersionConflict
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
	collection *mongo.Collection
}

// NewConciergeRepository creates a new ConciergeRepository instance.
func NewConciergeRepository(database *mongo.Database) *ConciergeRepository {
	return &ConciergeRepository{
		collection: database.Collection(collectionConcierges),
	}
}

// InitIndexes creates necessary database indexes for concierges:
// 1. Unique index on name.
func (r *ConciergeRepository) InitIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
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
