package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

// AuditFields records who created and last updated a model, and when.
type AuditFields struct {
	CreatedBy bson.ObjectID `bson:"created_by" json:"createdBy"`
	UpdatedBy bson.ObjectID `bson:"updated_by" json:"updatedBy"`
	CreatedAt time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time     `bson:"updated_at" json:"updatedAt"`
}
