package models

import (
	"time"

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
	OrganisationID bson.ObjectID `bson:"organisation_id" json:"-"`
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string        `bson:"name" json:"name"`
	Description    string        `bson:"description" json:"description"`
	Inputs         []ToolInput   `bson:"inputs" json:"inputs"`
	Outputs        []ToolOutput  `bson:"outputs" json:"outputs"`
	Version        int           `bson:"version" json:"-"`
	CreatedBy      bson.ObjectID `bson:"created_by" json:"createdBy"`
	UpdatedBy      bson.ObjectID `bson:"updated_by" json:"updatedBy"`
	CreatedAt      time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt      time.Time     `bson:"updated_at" json:"updatedAt"`
}
