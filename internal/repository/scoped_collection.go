package repository

import (
	"context"
	"fmt"

	"github.com/bangweiz/dubhu-designer/internal/identity"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// scopedCollection is the persistence boundary for tenant-owned application data.
// Missing authenticated context fails closed, including calls outside HTTP handlers.
type scopedCollection struct{ collection *mongo.Collection }

func newScopedCollection(db *mongo.Database, name string) *scopedCollection {
	return &scopedCollection{collection: db.Collection(name)}
}
func (c *scopedCollection) sibling(name string) *scopedCollection {
	return newScopedCollection(c.collection.Database(), name)
}
func scopedFilter(ctx context.Context, filter any) (bson.M, error) {
	id, err := identity.OrganisationID(ctx)
	if err != nil {
		return nil, err
	}
	return bson.M{"$and": bson.A{bson.M{"organisation_id": id}, filter}}, nil
}
func (c *scopedCollection) Find(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) (*mongo.Cursor, error) {
	f, err := scopedFilter(ctx, filter)
	if err != nil {
		return nil, err
	}
	return c.collection.Find(ctx, f, opts...)
}
func (c *scopedCollection) FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) *mongo.SingleResult {
	f, err := scopedFilter(ctx, filter)
	if err != nil {
		return mongo.NewSingleResultFromDocument(bson.M{}, err, nil)
	}
	return c.collection.FindOne(ctx, f, opts...)
}
func (c *scopedCollection) CountDocuments(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	f, err := scopedFilter(ctx, filter)
	if err != nil {
		return 0, err
	}
	return c.collection.CountDocuments(ctx, f, opts...)
}
func (c *scopedCollection) UpdateOne(ctx context.Context, filter, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	f, err := scopedFilter(ctx, filter)
	if err != nil {
		return nil, err
	}
	audited, err := auditUpdate(ctx, update)
	if err != nil {
		return nil, err
	}
	return c.collection.UpdateOne(ctx, f, audited, opts...)
}
func (c *scopedCollection) FindOneAndUpdate(ctx context.Context, filter, update any, opts ...options.Lister[options.FindOneAndUpdateOptions]) *mongo.SingleResult {
	f, err := scopedFilter(ctx, filter)
	if err != nil {
		return mongo.NewSingleResultFromDocument(bson.M{}, err, nil)
	}
	audited, err := auditUpdate(ctx, update)
	if err != nil {
		return mongo.NewSingleResultFromDocument(bson.M{}, err, nil)
	}
	return c.collection.FindOneAndUpdate(ctx, f, audited, opts...)
}
func (c *scopedCollection) InsertOne(ctx context.Context, document any, opts ...options.Lister[options.InsertOneOptions]) (*mongo.InsertOneResult, error) {
	id, err := identity.OrganisationID(ctx)
	if err != nil {
		return nil, err
	}
	encoded, err := bson.Marshal(document)
	if err != nil {
		return nil, err
	}
	var doc bson.M
	if err := bson.Unmarshal(encoded, &doc); err != nil {
		return nil, err
	}
	if existing, ok := doc["organisation_id"].(bson.ObjectID); ok && !existing.IsZero() && existing != id {
		return nil, fmt.Errorf("cannot insert another organisation's document")
	}
	doc["organisation_id"] = id
	principal, _ := identity.FromContext(ctx)
	doc["created_by"] = principal.AccountID
	doc["updated_by"] = principal.AccountID
	encoded, err = bson.Marshal(doc)
	if err != nil {
		return nil, err
	}
	if err := bson.Unmarshal(encoded, document); err != nil {
		return nil, err
	}
	return c.collection.InsertOne(ctx, doc, opts...)
}

// Scope the source collection and every lookup, even after projections discard organisation_id.
func (c *scopedCollection) Aggregate(ctx context.Context, pipeline mongo.Pipeline, opts ...options.Lister[options.AggregateOptions]) (*mongo.Cursor, error) {
	id, err := identity.OrganisationID(ctx)
	if err != nil {
		return nil, err
	}
	match := bson.D{{Key: "$match", Value: bson.M{"organisation_id": id}}}
	scoped := mongo.Pipeline{match}
	for _, stage := range pipeline {
		copyStage := append(bson.D(nil), stage...)
		for i, element := range copyStage {
			switch element.Key {
			case "$lookup":
				lookup, ok := element.Value.(bson.M)
				if !ok {
					return nil, fmt.Errorf("unsupported scoped lookup")
				}
				cloned := bson.M{}
				for k, v := range lookup {
					cloned[k] = v
				}
				nested := mongo.Pipeline{}
				if raw, ok := cloned["pipeline"]; ok {
					var valid bool
					nested, valid = raw.(mongo.Pipeline)
					if !valid {
						return nil, fmt.Errorf("unsupported scoped lookup pipeline")
					}
				}
				cloned["pipeline"] = append(mongo.Pipeline{match}, nested...)
				copyStage[i].Value = cloned
			case "$unionWith", "$graphLookup", "$out", "$merge":
				return nil, fmt.Errorf("aggregation stage %s needs explicit tenant scoping", element.Key)
			}
		}
		scoped = append(scoped, copyStage)
	}
	return c.collection.Aggregate(ctx, scoped, opts...)
}

type scopedIndexView struct{ collection *mongo.Collection }

func (c *scopedCollection) Indexes() scopedIndexView {
	return scopedIndexView{collection: c.collection}
}
func (v scopedIndexView) CreateOne(ctx context.Context, index mongo.IndexModel) (string, error) {
	names, err := v.CreateMany(ctx, []mongo.IndexModel{index})
	if err != nil {
		return "", err
	}
	return names[0], nil
}
func (v scopedIndexView) CreateMany(ctx context.Context, indexes []mongo.IndexModel) ([]string, error) {
	scoped := make([]mongo.IndexModel, len(indexes))
	for i, index := range indexes {
		keys, ok := index.Keys.(bson.D)
		if !ok {
			return nil, fmt.Errorf("tenant index keys must be bson.D")
		}
		if len(keys) == 0 || keys[0].Key != "organisation_id" {
			index.Keys = append(bson.D{{Key: "organisation_id", Value: 1}}, keys...)
		}
		scoped[i] = index
	}
	names, err := v.collection.Indexes().CreateMany(ctx, scoped)
	if err != nil {
		return nil, err
	}
	// Replace the legacy global name constraint, without modifying or assigning any data.
	cursor, err := v.collection.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var legacy []string
	for cursor.Next(ctx) {
		var index struct {
			Name   string `bson:"name"`
			Key    bson.D `bson:"key"`
			Unique bool   `bson:"unique"`
		}
		if err := cursor.Decode(&index); err != nil {
			return nil, err
		}
		if index.Unique && len(index.Key) == 1 && index.Key[0].Key == "name" {
			legacy = append(legacy, index.Name)
		}
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	for _, name := range legacy {
		if err := v.collection.Indexes().DropOne(ctx, name); err != nil {
			return nil, err
		}
	}
	return names, nil
}
