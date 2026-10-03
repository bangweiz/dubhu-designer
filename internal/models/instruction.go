package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Instruction represents a document in the "instructions" collection.
type Instruction struct {
	ID        bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name      string          `bson:"name" json:"name"`
	Content   string          `bson:"content" json:"content"`
	Tools     []bson.ObjectID `bson:"tools" json:"tools"`
	Version   int             `bson:"version" json:"version"`
	CreatedAt time.Time       `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time       `bson:"updated_at" json:"updated_at"`
}
