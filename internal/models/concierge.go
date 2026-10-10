package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

const MaxAgentsPerConcierge = 10

// Concierge owns the editable configuration and references its immutable snapshots.
type Concierge struct {
	ID                bson.ObjectID               `bson:"_id,omitempty" json:"id"`
	OrganisationID    bson.ObjectID               `bson:"organisation_id" json:"-"`
	Name              string                      `bson:"name" json:"name"`
	Description       string                      `bson:"description" json:"description"`
	Agents            []Agent                     `bson:"agents" json:"agents"`
	NextVersion       int                         `bson:"next_version" json:"nextVersion"`
	ConciergeVersions []ConciergeVersionReference `bson:"concierge_versions" json:"conciergeVersions"`
	AuditFields       `bson:",inline"`
}

type ConciergeVersionReference struct {
	ConciergeVersionID bson.ObjectID `bson:"concierge_version_id" json:"conciergeVersionId"`
	Version            int           `bson:"version" json:"version"`
}
