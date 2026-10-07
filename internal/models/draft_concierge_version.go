package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// DraftConciergeVersion holds the sole mutable working configuration of a concierge.
type DraftConciergeVersion struct {
	OrganisationID bson.ObjectID `bson:"organisation_id" json:"-"`
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	ConciergeID    bson.ObjectID `bson:"concierge_id" json:"conciergeId"`
	Agents         []Agent       `bson:"agents" json:"agents"`
	ETagVersion    int           `bson:"etag_version" json:"-"`
	Version        int           `bson:"version" json:"version"`
	CreatedBy      bson.ObjectID `bson:"created_by" json:"createdBy"`
	UpdatedBy      bson.ObjectID `bson:"updated_by" json:"updatedBy"`
	CreatedAt      time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt      time.Time     `bson:"updated_at" json:"updatedAt"`
}
