package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ToolInput represents an input parameter required by a tool.
type ToolInput struct {
	Name        string `bson:"name" json:"name"`
	Description string `bson:"description" json:"description"`
	Required    bool   `bson:"required" json:"required"`
}

// ToolOutput represents an output parameter returned by a tool.
type ToolOutput struct {
	Name        string `bson:"name" json:"name"`
	Description string `bson:"description" json:"description"`
}

// Tool represents a document in the "tools" collection.
type Tool struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	OrganisationID bson.ObjectID `bson:"organisation_id" json:"-"`
	Name           string        `bson:"name" json:"name"`
	Description    string        `bson:"description" json:"description"`
	Inputs         []ToolInput   `bson:"inputs" json:"inputs"`
	Outputs        []ToolOutput  `bson:"outputs" json:"outputs"`
	AuditFields    `bson:",inline"`
}
