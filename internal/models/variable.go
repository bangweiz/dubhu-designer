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
	Type           VariableType `bson:"type" json:"type"`
	AuditFields    `bson:",inline"`
	OrganisationID bson.ObjectID `bson:"organisation_id" json:"-"`
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string        `bson:"name" json:"name"`
	Description    string        `bson:"description" json:"description"`
}
