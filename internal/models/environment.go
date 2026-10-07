package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Environment struct {
	OrganisationID bson.ObjectID `bson:"organisation_id" json:"-"`
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string        `bson:"name" json:"name"`
	Description    string        `bson:"description" json:"description"`
	Version        int           `bson:"version" json:"-"`
	CreatedBy      bson.ObjectID `bson:"created_by" json:"createdBy"`
	UpdatedBy      bson.ObjectID `bson:"updated_by" json:"updatedBy"`
	CreatedAt      time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt      time.Time     `bson:"updated_at" json:"updatedAt"`
}
