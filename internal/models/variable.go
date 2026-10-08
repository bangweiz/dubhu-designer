package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Variable struct {
	AuditFields    `bson:",inline"`
	OrganisationID bson.ObjectID `bson:"organisation_id" json:"-"`
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string        `bson:"name" json:"name"`
	Description    string        `bson:"description" json:"description"`
}
