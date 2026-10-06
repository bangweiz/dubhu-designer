package repository

import (
	"context"
	"fmt"

	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const collectionInstructions = "instructions"

// ListUsages reads current instruction references from drafts, excluding immutable snapshots.
func (r *InstructionRepository) ListUsages(ctx context.Context, id bson.ObjectID) ([]models.InstructionUsage, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"agents.instructions": id}}},
		bson.D{{Key: "$unwind", Value: "$agents"}},
		bson.D{{Key: "$match", Value: bson.M{"agents.instructions": id}}},
		bson.D{{Key: "$project", Value: bson.M{
			"_id": 0, "concierge_id": 1, "concierge_version_id": "$_id", "version": 1,
			"agent_id": "$agents._id", "agent_name": "$agents.name",
		}}},
		bson.D{{Key: "$lookup", Value: bson.M{"from": collectionConcierges, "localField": "concierge_id", "foreignField": "_id", "as": "concierge"}}},
		bson.D{{Key: "$unwind", Value: "$concierge"}},
		bson.D{{Key: "$set", Value: bson.M{"concierge_name": "$concierge.name"}}},
		bson.D{{Key: "$unset", Value: "concierge"}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "concierge_name", Value: 1}, {Key: "concierge_id", Value: 1}, {Key: "version", Value: 1}, {Key: "agent_name", Value: 1}, {Key: "agent_id", Value: 1}}}},
	}
	cursor, err := r.collection.Database().Collection(collectionDraftConciergeVersions).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate instruction usages: %w", err)
	}
	defer cursor.Close(ctx)
	usages := []models.InstructionUsage{}
	if err := cursor.All(ctx, &usages); err != nil {
		return nil, fmt.Errorf("decode instruction usages: %w", err)
	}
	return usages, nil
}

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

	if _, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{indexModel, {Keys: bson.D{{Key: "tools", Value: 1}}}}); err != nil {
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

// GetByIDWithTools finds an Instruction by ID and populates its referenced tools via MongoDB aggregation ($lookup).
// Returns nil, nil if the instruction is not found.
func (r *InstructionRepository) GetByIDWithTools(ctx context.Context, id bson.ObjectID) (*models.PopulatedInstruction, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"_id": id}}},
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         collectionTools,
			"localField":   "tools",
			"foreignField": "_id",
			"as":           "resolved_tools",
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate instruction with tools: %w", err)
	}
	defer cursor.Close(ctx)

	var populated models.PopulatedInstruction
	if cursor.Next(ctx) {
		if err := cursor.Decode(&populated); err != nil {
			return nil, fmt.Errorf("failed to decode populated instruction: %w", err)
		}
		return &populated, nil
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error aggregating instruction with tools: %w", err)
	}

	return nil, nil
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

// FindByIDs retrieves all instructions matching the supplied IDs.
func (r *InstructionRepository) FindByIDs(ctx context.Context, ids []bson.ObjectID) ([]models.Instruction, error) {
	if len(ids) == 0 {
		return []models.Instruction{}, nil
	}
	cursor, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, fmt.Errorf("failed to query instructions by IDs: %w", err)
	}
	defer cursor.Close(ctx)
	var instructions []models.Instruction
	if err := cursor.All(ctx, &instructions); err != nil {
		return nil, fmt.Errorf("failed to decode instructions: %w", err)
	}
	if instructions == nil {
		instructions = []models.Instruction{}
	}
	return instructions, nil
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
