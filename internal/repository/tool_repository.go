package repository

import (
	"context"
	"fmt"

	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const collectionTools = "tools"

// ToolRepository manages Tool persistence in MongoDB.
type ToolRepository struct {
	collection *mongo.Collection
}

// NewToolRepository creates a new ToolRepository instance.
func NewToolRepository(database *mongo.Database) *ToolRepository {
	return &ToolRepository{
		collection: database.Collection(collectionTools),
	}
}

// InitIndexes creates necessary database indexes for tools (e.g. unique index on name).
func (r *ToolRepository) InitIndexes(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	if _, err := r.collection.Indexes().CreateOne(ctx, indexModel); err != nil {
		return fmt.Errorf("failed to create unique index on tool name: %w", err)
	}
	return nil
}

// Create inserts a Tool document into MongoDB.
// Returns ErrToolNameExists if a tool with the same name already exists.
func (r *ToolRepository) Create(ctx context.Context, tool *models.Tool) (*models.Tool, error) {
	if _, err := r.collection.InsertOne(ctx, tool); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrToolNameExists
		}
		return nil, fmt.Errorf("failed to insert tool: %w", err)
	}
	return tool, nil
}

// GetByID finds a Tool document by its ObjectID.
func (r *ToolRepository) GetByID(ctx context.Context, id bson.ObjectID) (*models.Tool, error) {
	var tool models.Tool
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&tool)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find tool: %w", err)
	}
	return &tool, nil
}

// List retrieves all Tool documents from MongoDB without sorting, filtering, or pagination.
func (r *ToolRepository) List(ctx context.Context) ([]models.Tool, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to query tools: %w", err)
	}
	defer cursor.Close(ctx)

	var tools []models.Tool
	if err := cursor.All(ctx, &tools); err != nil {
		return nil, fmt.Errorf("failed to decode tools: %w", err)
	}

	if tools == nil {
		tools = []models.Tool{}
	}
	return tools, nil
}

// Update updates an existing Tool document using optimistic concurrency control.
// The document must match both id and expectedVersion. On match, version increments by 1.
// Returns the updated Tool document after update.
func (r *ToolRepository) Update(ctx context.Context, id bson.ObjectID, expectedVersion int, updateDoc bson.M) (*models.Tool, error) {
	filter := bson.M{
		"_id":     id,
		"version": expectedVersion,
	}

	update := bson.M{
		"$set": updateDoc,
		"$inc": bson.M{"version": 1},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedTool models.Tool
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedTool)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrVersionConflict
		}
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrToolNameExists
		}
		return nil, fmt.Errorf("failed to update tool: %w", err)
	}

	return &updatedTool, nil
}

// FindByIDs finds all tools matching any of the given ObjectIDs.
func (r *ToolRepository) FindByIDs(ctx context.Context, ids []bson.ObjectID) ([]models.Tool, error) {
	if len(ids) == 0 {
		return []models.Tool{}, nil
	}

	filter := bson.M{
		"_id": bson.M{"$in": ids},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query tools by IDs: %w", err)
	}
	defer cursor.Close(ctx)

	var tools []models.Tool
	if err := cursor.All(ctx, &tools); err != nil {
		return nil, fmt.Errorf("failed to decode tools: %w", err)
	}

	if tools == nil {
		tools = []models.Tool{}
	}
	return tools, nil
}
