package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Instruction represents a document in the "instructions" collection.
type Instruction struct {
	Variables      []bson.ObjectID `bson:"variables" json:"variables"`
	AuditFields    `bson:",inline"`
	OrganisationID bson.ObjectID   `bson:"organisation_id" json:"-"`
	ID             bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name           string          `bson:"name" json:"name"`
	Content        string          `bson:"content" json:"content"`
	Tools          []bson.ObjectID `bson:"tools" json:"tools"`
}

// PopulatedInstruction represents an Instruction with its referenced Tool documents resolved via aggregation.
type PopulatedInstruction struct {
	ResolvedVariables []Variable `bson:"resolved_variables"`
	Instruction       `bson:",inline"`
	ResolvedTools     []Tool `bson:"resolved_tools"`
}
