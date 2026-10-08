package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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

// prepareUpdate advances timestamps and guards against concurrent writes to the parent.
// This also protects concierge ETags when only an embedded agent changes.
func (c *scopedCollection) prepareUpdate(ctx context.Context, filter bson.M, update any) (bson.M, any, error) {
	var current struct {
		ID        bson.ObjectID `bson:"_id"`
		UpdatedAt time.Time     `bson:"updated_at"`
		Agents    []struct {
			UpdatedAt time.Time `bson:"updated_at"`
		} `bson:"agents"`
	}
	err := c.collection.FindOne(ctx, filter, options.FindOne().SetProjection(bson.M{"_id": 1, "updated_at": 1, "agents.updated_at": 1})).Decode(&current)
	if err != nil {
		return nil, nil, err
	}
	audited, err := auditUpdate(ctx, update)
	if err != nil {
		return nil, nil, err
	}
	document, ok := audited.(bson.M)
	if !ok {
		return nil, nil, fmt.Errorf("timestamp updates require an operator document")
	}
	set := document["$set"].(bson.M)
	now := time.Now().UTC().Truncate(time.Millisecond)
	latest := current.UpdatedAt
	for _, agent := range current.Agents {
		if agent.UpdatedAt.After(latest) {
			latest = agent.UpdatedAt
		}
	}
	if !now.After(latest) {
		now = latest.Add(time.Millisecond)
	}
	set["updated_at"] = now
	for key := range set {
		if strings.HasSuffix(key, ".updated_at") {
			set[key] = now
		}
	}
	var previous any = current.UpdatedAt
	if current.UpdatedAt.IsZero() {
		previous = bson.M{"$exists": false}
	}
	guarded := bson.M{"$and": bson.A{filter, bson.M{"_id": current.ID, "updated_at": previous}}}
	return guarded, document, nil
}

func (c *scopedCollection) UpdateOne(ctx context.Context, filter, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	f, err := scopedFilter(ctx, filter)
	if err != nil {
		return nil, err
	}
	for attempts := 0; attempts < 8; attempts++ {
		guarded, stamped, err := c.prepareUpdate(ctx, f, update)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &mongo.UpdateResult{}, nil
		}
		if err != nil {
			return nil, err
		}
		result, err := c.collection.UpdateOne(ctx, guarded, stamped, opts...)
		if err != nil || result.MatchedCount > 0 {
			return result, err
		}
	}
	return nil, ErrUpdateConflict
}
func (c *scopedCollection) FindOneAndUpdate(ctx context.Context, filter, update any, opts ...options.Lister[options.FindOneAndUpdateOptions]) *mongo.SingleResult {
	f, err := scopedFilter(ctx, filter)
	if err != nil {
		return mongo.NewSingleResultFromDocument(bson.M{}, err, nil)
	}
	for attempts := 0; attempts < 8; attempts++ {
		guarded, stamped, err := c.prepareUpdate(ctx, f, update)
		if err != nil {
			return mongo.NewSingleResultFromDocument(bson.M{}, err, nil)
		}
		result := c.collection.FindOneAndUpdate(ctx, guarded, stamped, opts...)
		if !errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return result
		}
	}
	return mongo.NewSingleResultFromDocument(bson.M{}, ErrUpdateConflict, nil)
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
