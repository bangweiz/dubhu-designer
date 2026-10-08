package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Organisation struct {
	AuditFields   `bson:",inline"`
	ID            bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string        `bson:"name" json:"name"`
	Description   string        `bson:"description" json:"description"`
	RootAccountID bson.ObjectID `bson:"root_account_id" json:"rootAccountId"`
}
