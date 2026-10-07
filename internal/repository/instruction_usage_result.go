package repository

import "go.mongodb.org/mongo-driver/v2/bson"

// InstructionUsageResult is an aggregation projection of an agent reference.
// It is derived from concierge drafts and agents, never stored as a document.
type InstructionUsageResult struct {
	ConciergeID        bson.ObjectID `bson:"concierge_id"`
	ConciergeName      string        `bson:"concierge_name"`
	ConciergeVersionID bson.ObjectID `bson:"concierge_version_id"`
	Version            int           `bson:"version"`
	AgentID            bson.ObjectID `bson:"agent_id"`
	AgentName          string        `bson:"agent_name"`
}
