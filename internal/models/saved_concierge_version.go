package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// SavedConciergeVersion is an immutable, self-contained customer-facing concierge snapshot.
type SavedConciergeVersion struct {
	AuditFields    `bson:",inline"`
	OrganisationID bson.ObjectID `bson:"organisation_id" json:"-"`
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string        `bson:"name" json:"name"`
	Description    string        `bson:"description" json:"description"`
	ConciergeID    bson.ObjectID `bson:"concierge_id" json:"conciergeId"`
	Version        int           `bson:"version" json:"version"`
	Agents         []Agent       `bson:"agents" json:"agents"`
	Instructions   []Instruction `bson:"instructions" json:"instructions"`
	Tools          []Tool        `bson:"tools" json:"tools"`
}
