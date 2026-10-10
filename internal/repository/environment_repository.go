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

const collectionEnvironments = "environments"

type EnvironmentRepository struct{ collection *scopedCollection }

func NewEnvironmentRepository(db *mongo.Database) *EnvironmentRepository {
	return &EnvironmentRepository{collection: newScopedCollection(db, collectionEnvironments)}
}

func (r *EnvironmentRepository) InitIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "organisation_id", Value: 1}, {Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		return fmt.Errorf("create environment indexes: %w", err)
	}

	return nil
}

func (r *EnvironmentRepository) Create(ctx context.Context, e *models.Environment) (*models.Environment, error) {
	if _, err := r.collection.InsertOne(ctx, e); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrEnvironmentNameExists
		}

		return nil, fmt.Errorf("create environment: %w", err)
	}

	return e, nil
}

func (r *EnvironmentRepository) GetByID(ctx context.Context, id bson.ObjectID) (*models.Environment, error) {
	var e models.Environment
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&e)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find environment: %w", err)
	}

	return &e, nil
}

func (r *EnvironmentRepository) List(ctx context.Context) ([]models.Environment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}, {Key: "_id", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("list environments: %w", err)
	}

	defer cursor.Close(ctx)
	result := []models.Environment{}
	if err := cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("decode environments: %w", err)
	}

	return result, nil
}

func (r *EnvironmentRepository) Update(
	ctx context.Context,
	id bson.ObjectID,
	expectedUpdatedAt time.Time,
	name, description string,
) (*models.Environment, error) {
	var e models.Environment
	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id, "updated_at": expectedUpdatedAt},
		bson.M{"$set": bson.M{
			"name":        name,
			"description": description,
			"updated_at":  time.Now().UTC().Truncate(time.Millisecond),
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&e)
	if err == mongo.ErrNoDocuments {
		return nil, ErrUpdateConflict
	}
	if mongo.IsDuplicateKeyError(err) {
		return nil, ErrEnvironmentNameExists
	}
	if err != nil {
		return nil, fmt.Errorf("update environment: %w", err)
	}

	return &e, nil
}
