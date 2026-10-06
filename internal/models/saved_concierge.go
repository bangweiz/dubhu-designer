package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// SavedConcierge is an immutable, self-contained customer-facing concierge snapshot.
type SavedConcierge struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"conciergeVersionId"`
	ConciergeID  bson.ObjectID `bson:"concierge_id" json:"id"`
	Version      int           `bson:"version" json:"version"`
	Name         string        `bson:"name" json:"name"`
	Description  string        `bson:"description" json:"description"`
	Agents       []Agent       `bson:"agents" json:"agents"`
	Instructions []Instruction `bson:"instructions" json:"instructions"`
	Tools        []Tool        `bson:"tools" json:"tools"`
	CreatedAt    time.Time     `bson:"created_at" json:"createdAt"`
}
