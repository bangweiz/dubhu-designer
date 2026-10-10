package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// VariableType is the declared value type of a variable.
type VariableType string

const (
	VariableTypeString VariableType = "string"
	VariableTypeNumber VariableType = "number"
	VariableTypeBool   VariableType = "bool"
)

func (t VariableType) IsValid() bool {
	return t == VariableTypeString || t == VariableTypeNumber || t == VariableTypeBool
}

// EffectiveType treats legacy variables without a type as strings.
func (v *Variable) EffectiveType() VariableType {
	if v.Type == "" {
		return VariableTypeString
	}
	return v.Type
}

type Variable struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	OrganisationID bson.ObjectID `bson:"organisation_id" json:"-"`
	Name           string        `bson:"name" json:"name"`
	Description    string        `bson:"description" json:"description"`
	Type           VariableType  `bson:"type" json:"type"`
	AuditFields    `bson:",inline"`
}
