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
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string        `bson:"name" json:"name"`
	Description string        `bson:"description" json:"description"`
	Inputs      []ToolInput   `bson:"inputs" json:"inputs"`
	Outputs     []ToolOutput  `bson:"outputs" json:"outputs"`
	Version     int           `bson:"version" json:"version"`
	CreatedAt   time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at" json:"updated_at"`
}
