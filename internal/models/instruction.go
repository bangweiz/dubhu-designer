package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Instruction represents a document in the "instructions" collection.
type Instruction struct {
	ID             bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	OrganisationID bson.ObjectID   `bson:"organisation_id" json:"-"`
	Name           string          `bson:"name" json:"name"`
	Content        string          `bson:"content" json:"content"`
	Variables      []bson.ObjectID `bson:"variables" json:"variables"`
	Tools          []bson.ObjectID `bson:"tools" json:"tools"`
	AuditFields    `bson:",inline"`
}

// PopulatedInstruction represents an Instruction with its referenced Tool documents resolved via aggregation.
type PopulatedInstruction struct {
	Instruction       `bson:",inline"`
	ResolvedVariables []Variable `bson:"resolved_variables"`
	ResolvedTools     []Tool     `bson:"resolved_tools"`
}
