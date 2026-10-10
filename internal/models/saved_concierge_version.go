package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// SavedConciergeVersion is an immutable, self-contained customer-facing concierge snapshot.
type SavedConciergeVersion struct {
	ID             bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	OrganisationID bson.ObjectID   `bson:"organisation_id" json:"-"`
	ConciergeID    bson.ObjectID   `bson:"concierge_id" json:"conciergeId"`
	Name           string          `bson:"name" json:"name"`
	Description    string          `bson:"description" json:"description"`
	Version        int             `bson:"version" json:"version"`
	Agents         []Agent         `bson:"agents" json:"agents"`
	Instructions   []Instruction   `bson:"instructions" json:"instructions"`
	Variables      []Variable      `bson:"variables" json:"variables"`
	Tools          []Tool          `bson:"tools" json:"tools"`
	EnvironmentIDs []bson.ObjectID `bson:"environment_ids" json:"-"`
	AuditFields    `bson:",inline"`
}
