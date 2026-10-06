package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Concierge represents a document in the "concierges" collection.
type Concierge struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string        `bson:"name" json:"name"`
	Description string        `bson:"description" json:"description"`
	Agents      []Agent       `bson:"agents" json:"agents"`
	ETagVersion int           `bson:"version" json:"-"`
	Version     int           `bson:"customer_version" json:"version"`
	CreatedAt   time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time     `bson:"updated_at" json:"updatedAt"`
}
