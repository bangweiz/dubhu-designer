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

const collectionInstructions = "instructions"

var (
	ErrInstructionNameExists  = errors.New("instruction name already exists")
	ErrInstructionNotFound    = errors.New("instruction not found")
	ErrInstructionConflict    = errors.New("instruction version conflict")
)

// InstructionRepository manages Instruction persistence in MongoDB.
type InstructionRepository struct {
	collection *mongo.Collection
}

// NewInstructionRepository creates a new InstructionRepository instance.
func NewInstructionRepository(database *mongo.Database) *InstructionRepository {
	return &InstructionRepository{
		collection: database.Collection(collectionInstructions),
	}
}

// InitIndexes creates necessary database indexes for instructions:
// 1. Unique index on name.
func (r *InstructionRepository) InitIndexes(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	if _, err := r.collection.Indexes().CreateOne(ctx, indexModel); err != nil {
		return fmt.Errorf("failed to create index on instructions collection: %w", err)
	}
	return nil
}

// Create inserts an Instruction document into MongoDB.
// Returns ErrInstructionNameExists if an instruction with the same name already exists.
func (r *InstructionRepository) Create(ctx context.Context, inst *models.Instruction) (*models.Instruction, error) {
	if _, err := r.collection.InsertOne(ctx, inst); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrInstructionNameExists
		}
		return nil, fmt.Errorf("failed to insert instruction: %w", err)
	}
	return inst, nil
}

// GetByID finds an Instruction document by its ObjectID.
func (r *InstructionRepository) GetByID(ctx context.Context, id bson.ObjectID) (*models.Instruction, error) {
	var inst models.Instruction
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&inst)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find instruction: %w", err)
	}
	return &inst, nil
}

// List retrieves all Instruction documents from MongoDB without sorting, filtering, or pagination.
func (r *InstructionRepository) List(ctx context.Context) ([]models.Instruction, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to query instructions: %w", err)
	}
	defer cursor.Close(ctx)

	var list []models.Instruction
	if err := cursor.All(ctx, &list); err != nil {
		return nil, fmt.Errorf("failed to decode instructions: %w", err)
	}

	if list == nil {
		list = []models.Instruction{}
	}
	return list, nil
}

// Update updates an existing Instruction document using optimistic concurrency control.
// Matches by _id and version. On success, increments version and returns the updated document.
func (r *InstructionRepository) Update(ctx context.Context, id bson.ObjectID, expectedVersion int, updateDoc bson.M) (*models.Instruction, error) {
	filter := bson.M{
		"_id":     id,
		"version": expectedVersion,
	}

	update := bson.M{
		"$set": updateDoc,
		"$inc": bson.M{"version": 1},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updated models.Instruction
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrInstructionConflict
		}
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrInstructionNameExists
		}
		return nil, fmt.Errorf("failed to update instruction: %w", err)
	}

	return &updated, nil
}


