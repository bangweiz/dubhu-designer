package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// DraftConciergeVersion holds the sole mutable working configuration of a concierge.
type DraftConciergeVersion struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	ConciergeID bson.ObjectID `bson:"concierge_id" json:"conciergeId"`
	Agents      []Agent       `bson:"agents" json:"agents"`
	ETagVersion int           `bson:"etag_version" json:"-"`
	Version     int           `bson:"version" json:"version"`
	CreatedAt   time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time     `bson:"updated_at" json:"updatedAt"`
}
