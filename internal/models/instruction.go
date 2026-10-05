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
	Version   int             `bson:"version" json:"-"`
	CreatedAt time.Time       `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time       `bson:"updated_at" json:"updatedAt"`
}

// PopulatedInstruction represents an Instruction with its referenced Tool documents resolved via aggregation.
type PopulatedInstruction struct {
	Instruction   `bson:",inline"`
	ResolvedTools []Tool `bson:"resolved_tools"`
}
