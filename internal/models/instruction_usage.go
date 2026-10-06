package models

import "go.mongodb.org/mongo-driver/v2/bson"

// InstructionUsage identifies one agent's reference to an instruction in a version.
type InstructionUsage struct {
	ConciergeID        bson.ObjectID `bson:"concierge_id"`
	ConciergeName      string        `bson:"concierge_name"`
	ConciergeVersionID bson.ObjectID `bson:"concierge_version_id"`
	Version            int           `bson:"version"`
	AgentID            bson.ObjectID `bson:"agent_id"`
	AgentName          string        `bson:"agent_name"`
}
