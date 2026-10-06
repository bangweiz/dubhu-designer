package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Concierge is the identity shared by its working and saved versions.
type Concierge struct {
	ID                bson.ObjectID               `bson:"_id,omitempty" json:"id"`
	Name              string                      `bson:"name" json:"name"`
	Description       string                      `bson:"description" json:"description"`
	Version           int                         `bson:"version" json:"-"`
	ConciergeVersions []ConciergeVersionReference `bson:"concierge_versions" json:"conciergeVersions"`
	CreatedAt         time.Time                   `bson:"created_at" json:"createdAt"`
	UpdatedAt         time.Time                   `bson:"updated_at" json:"updatedAt"`
}

type ConciergeVersionReference struct {
	ConciergeVersionID bson.ObjectID `bson:"concierge_version_id" json:"conciergeVersionId"`
	Version            int           `bson:"version" json:"version"`
}
