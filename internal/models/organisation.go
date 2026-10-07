package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Organisation struct {
	ID            bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string        `bson:"name" json:"name"`
	Description   string        `bson:"description" json:"description"`
	RootAccountID bson.ObjectID `bson:"root_account_id" json:"rootAccountId"`
	CreatedBy     bson.ObjectID `bson:"created_by" json:"createdBy"`
	UpdatedBy     bson.ObjectID `bson:"updated_by" json:"updatedBy"`
	CreatedAt     time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt     time.Time     `bson:"updated_at" json:"updatedAt"`
}
