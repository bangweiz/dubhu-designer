package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Environment struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	OrganisationID bson.ObjectID `bson:"organisation_id" json:"-"`
	Name           string        `bson:"name" json:"name"`
	Description    string        `bson:"description" json:"description"`
	AuditFields    `bson:",inline"`
}
