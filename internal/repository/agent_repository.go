package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrAgentNameExists = errors.New("agent name already exists")
	ErrAgentNotFound   = errors.New("agent not found")
)

// AgentRepository manages Agent embedded subdocuments within Concierge documents.
type AgentRepository struct {
	collection *mongo.Collection
}

// NewAgentRepository creates a new AgentRepository instance pointing to the concierges collection.
func NewAgentRepository(database *mongo.Database) *AgentRepository {
	return &AgentRepository{
		collection: database.Collection(collectionConcierges),
	}
}

// InitIndexes initializes any required indexes for agent operations.
func (r *AgentRepository) InitIndexes(ctx context.Context) error {
	return nil
}

// Create atomically appends an Agent subdocument to a Concierge's agents array.
// Returns ErrConciergeNotFound if the parent concierge does not exist.
// Returns ErrAgentNameExists if an agent with the same name already exists in this concierge.
func (r *AgentRepository) Create(ctx context.Context, conciergeID bson.ObjectID, agent *models.Agent) (*models.Agent, error) {
	filter := bson.M{
		"_id":         conciergeID,
		"agents.name": bson.M{"$ne": agent.Name},
	}
	update := bson.M{
		"$push": bson.M{"agents": agent},
		"$set":  bson.M{"updated_at": time.Now().UTC()},
	}

	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("failed to insert agent into concierge: %w", err)
	}

	if res.MatchedCount == 0 {
		count, err := r.collection.CountDocuments(ctx, bson.M{"_id": conciergeID})
		if err != nil {
			return nil, fmt.Errorf("failed to verify concierge: %w", err)
		}
		if count == 0 {
			return nil, ErrConciergeNotFound
		}
		return nil, ErrAgentNameExists
	}

	agent.ConciergeID = conciergeID
	return agent, nil
}

// GetByIDAndConciergeID finds an Agent subdocument within a Concierge by conciergeID and agentID.
// Returns ErrConciergeNotFound if concierge does not exist.
// Returns ErrAgentNotFound if agent does not exist in the concierge.
func (r *AgentRepository) GetByIDAndConciergeID(ctx context.Context, agentID bson.ObjectID, conciergeID bson.ObjectID) (*models.Agent, error) {
	var concierge models.Concierge
	err := r.collection.FindOne(ctx, bson.M{"_id": conciergeID}).Decode(&concierge)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrConciergeNotFound
		}
		return nil, fmt.Errorf("failed to find concierge: %w", err)
	}

	for _, a := range concierge.Agents {
		if a.ID == agentID {
			a.ConciergeID = conciergeID
			return &a, nil
		}
	}
	return nil, ErrAgentNotFound
}

// ListByConciergeID retrieves all Agent subdocuments from a Concierge.
// Returns ErrConciergeNotFound if concierge does not exist.
func (r *AgentRepository) ListByConciergeID(ctx context.Context, conciergeID bson.ObjectID) ([]models.Agent, error) {
	var concierge models.Concierge
	err := r.collection.FindOne(ctx, bson.M{"_id": conciergeID}).Decode(&concierge)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrConciergeNotFound
		}
		return nil, fmt.Errorf("failed to find concierge: %w", err)
	}

	if concierge.Agents == nil {
		return []models.Agent{}, nil
	}
	for i := range concierge.Agents {
		concierge.Agents[i].ConciergeID = conciergeID
	}
	return concierge.Agents, nil
}
