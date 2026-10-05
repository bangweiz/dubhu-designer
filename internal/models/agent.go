package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Model represents supported foundation models.
type Model string

// Supported agent model constants.
const (
	ModelGemini35Flash     Model = "gemini-3.5-flash"
	ModelGemini35FlashLite Model = "gemini-3.5-flash-lite"
)

// ValidModels holds all supported agent models at this stage.
var ValidModels = []Model{
	ModelGemini35Flash,
	ModelGemini35FlashLite,
}

// IsValid checks whether the model is one of the supported enum values.
func (m Model) IsValid() bool {
	return m == ModelGemini35Flash || m == ModelGemini35FlashLite
}

// String returns the string representation of Model.
func (m Model) String() string {
	return string(m)
}

// IsValidAgentModel checks whether the given model string is a supported model enum.
func IsValidAgentModel(model string) bool {
	return Model(model).IsValid()
}

// Agent represents an embedded subdocument within a Concierge.
type Agent struct {
	ID           bson.ObjectID   `bson:"_id" json:"id"`
	ConciergeID  bson.ObjectID   `bson:"concierge_id,omitempty" json:"conciergeId,omitempty"`
	Name         string          `bson:"name" json:"name"`
	Description  string          `bson:"description" json:"description"`
	Goal         string          `bson:"goal" json:"goal"`
	Model        Model           `bson:"model" json:"model"`
	Instructions []bson.ObjectID `bson:"instructions" json:"instructions"`
	Version      int             `bson:"version" json:"-"`
}
