package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const collectionConcierges = "concierges"

var (
	ErrConciergeNameExists = errors.New("concierge name already exists")
	ErrConciergeNotFound   = errors.New("concierge not found")
)

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
// 2. Multikey index on agents._id for subdocument lookups.
func (r *ConciergeRepository) InitIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "agents._id", Value: 1}},
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
	if concierge.Agents == nil {
		concierge.Agents = []models.Agent{}
	}
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
	if concierge.Agents == nil {
		concierge.Agents = []models.Agent{}
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
	for i := range list {
		if list[i].Agents == nil {
			list[i].Agents = []models.Agent{}
		}
	}
	return list, nil
}
