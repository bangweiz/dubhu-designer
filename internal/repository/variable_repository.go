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

const collectionVariables = "variables"

type VariableRepository struct{ collection *mongo.Collection }

func NewVariableRepository(db *mongo.Database) *VariableRepository {
	return &VariableRepository{collection: db.Collection(collectionVariables)}
}

func (r *VariableRepository) InitIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "name", Value: 1}}, Options: options.Index().SetUnique(true)})
	if err != nil {
		return fmt.Errorf("create variable indexes: %w", err)
	}
	return nil
}

func (r *VariableRepository) Create(ctx context.Context, e *models.Variable) (*models.Variable, error) {
	if _, err := r.collection.InsertOne(ctx, e); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrVariableNameExists
		}
		return nil, fmt.Errorf("create variable: %w", err)
	}
	return e, nil
}

func (r *VariableRepository) GetByID(ctx context.Context, id bson.ObjectID) (*models.Variable, error) {
	var e models.Variable
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&e)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find variable: %w", err)
	}
	return &e, nil
}

func (r *VariableRepository) List(ctx context.Context) ([]models.Variable, error) {
	cursor, err := r.collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}, {Key: "_id", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("list variables: %w", err)
	}
	defer cursor.Close(ctx)
	result := []models.Variable{}
	if err := cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("decode variables: %w", err)
	}
	return result, nil
}

func (r *VariableRepository) Update(ctx context.Context, id bson.ObjectID, version int, name, description string) (*models.Variable, error) {
	var e models.Variable
	err := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": id, "version": version}, bson.M{"$set": bson.M{"name": name, "description": description, "updated_at": time.Now().UTC().Truncate(time.Millisecond)}, "$inc": bson.M{"version": 1}}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&e)
	if err == mongo.ErrNoDocuments {
		return nil, ErrVersionConflict
	}
	if mongo.IsDuplicateKeyError(err) {
		return nil, ErrVariableNameExists
	}
	if err != nil {
		return nil, fmt.Errorf("update variable: %w", err)
	}
	return &e, nil
}
