package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AccountRole string

const (
	RoleRoot  AccountRole = "root"
	RoleAdmin AccountRole = "admin"
	RoleUser  AccountRole = "user"
)

type Account struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	OrganisationID bson.ObjectID `bson:"organisation_id" json:"organisationId"`
	Name           string        `bson:"name" json:"name"`
	Email          string        `bson:"email" json:"email"`
	PasswordHash   string        `bson:"password_hash" json:"-"`
	Role           AccountRole   `bson:"role" json:"role"`
	AuditFields    `bson:",inline"`
}

type AuthSession struct {
	TokenHash      string        `bson:"_id" json:"-"`
	OrganisationID bson.ObjectID `bson:"organisation_id" json:"-"`
	AccountID      bson.ObjectID `bson:"account_id" json:"-"`
	CreatedAt      time.Time     `bson:"created_at" json:"-"`
	ExpiresAt      time.Time     `bson:"expires_at" json:"-"`
}
