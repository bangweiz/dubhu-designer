package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// SavedConciergeVersion is an immutable, self-contained customer-facing concierge snapshot.
type SavedConciergeVersion struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	ConciergeID  bson.ObjectID `bson:"concierge_id" json:"conciergeId"`
	Saved        bool          `bson:"saved" json:"saved"`
	UpdatedAt    time.Time     `bson:"updated_at" json:"updatedAt"`
	Version      int           `bson:"version" json:"version"`
	Agents       []Agent       `bson:"agents" json:"agents"`
	Instructions []Instruction `bson:"instructions" json:"instructions"`
	Tools        []Tool        `bson:"tools" json:"tools"`
	CreatedAt    time.Time     `bson:"created_at" json:"createdAt"`
}
