package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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
	if agent.Instructions == nil {
		agent.Instructions = []bson.ObjectID{}
	}
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

// Update atomically updates an Agent subdocument within a Concierge.
// Matches by conciergeID and agentID, checks that no other agent in this concierge has the same name,
// updates agent fields (name, description, goal, model), increments agent version by 1,
// and updates the concierge's updated_at timestamp.
// Returns ErrConciergeNotFound if the parent concierge does not exist.
// Returns ErrAgentNotFound if the agent does not exist in the concierge.
// Returns ErrAgentNameExists if another agent in this concierge already has the new name.
func (r *AgentRepository) Update(ctx context.Context, conciergeID bson.ObjectID, agentID bson.ObjectID, expectedVersion int, agent *models.Agent) (*models.Agent, error) {
	filter := bson.M{
		"_id": conciergeID,
		"$and": []bson.M{
			{"agents": bson.M{"$elemMatch": bson.M{"_id": agentID, "version": expectedVersion}}},
			{
				"agents": bson.M{
					"$not": bson.M{
						"$elemMatch": bson.M{
							"_id":  bson.M{"$ne": agentID},
							"name": agent.Name,
						},
					},
				},
			},
		},
	}

	update := bson.M{
		"$set": bson.M{
			"agents.$[elem].name":        agent.Name,
			"agents.$[elem].description": agent.Description,
			"agents.$[elem].goal":        agent.Goal,
			"agents.$[elem].model":       agent.Model,
			"updated_at":                 time.Now().UTC(),
		},
		"$inc": bson.M{
			"agents.$[elem].version": 1,
		},
	}

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After).
		SetArrayFilters([]any{bson.M{"elem._id": agentID, "elem.version": expectedVersion}})

	var concierge models.Concierge
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&concierge)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			var existing models.Concierge
			findErr := r.collection.FindOne(ctx, bson.M{"_id": conciergeID}).Decode(&existing)
			if findErr != nil {
				if errors.Is(findErr, mongo.ErrNoDocuments) {
					return nil, ErrConciergeNotFound
				}
				return nil, fmt.Errorf("failed to verify concierge: %w", findErr)
			}

			var existingAgent *models.Agent
			for _, a := range existing.Agents {
				if a.ID == agentID {
					existingAgent = &a
					break
				}
			}
			if existingAgent == nil {
				return nil, ErrAgentNotFound
			}
			if existingAgent.Version != expectedVersion {
				return nil, ErrAgentVersionConflict
			}

			return nil, ErrAgentNameExists
		}
		return nil, fmt.Errorf("failed to update agent: %w", err)
	}

	for _, a := range concierge.Agents {
		if a.ID == agentID {
			a.ConciergeID = conciergeID
			return &a, nil
		}
	}

	return nil, ErrAgentNotFound
}

// AssignInstruction adds an instruction ObjectID to an Agent's instructions array.
// If the instruction is already assigned, it returns idempotently without error.
// Returns ErrConciergeNotFound if concierge does not exist.
// Returns ErrAgentNotFound if agent does not exist in the concierge.
func (r *AgentRepository) AssignInstruction(ctx context.Context, conciergeID bson.ObjectID, agentID bson.ObjectID, instructionID bson.ObjectID) error {
	filter := bson.M{
		"_id": conciergeID,
		"agents": bson.M{
			"$elemMatch": bson.M{
				"_id":          agentID,
				"instructions": bson.M{"$ne": instructionID},
			},
		},
	}

	update := bson.M{
		"$addToSet": bson.M{
			"agents.$[elem].instructions": instructionID,
		},
		"$set": bson.M{
			"updated_at": time.Now().UTC(),
		},
		"$inc": bson.M{
			"agents.$[elem].version": 1,
		},
	}

	opts := options.UpdateOne().
		SetArrayFilters([]any{bson.M{"elem._id": agentID}})

	res, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to assign instruction to agent: %w", err)
	}

	if res.MatchedCount > 0 {
		return nil
	}

	var existing models.Concierge
	findErr := r.collection.FindOne(ctx, bson.M{"_id": conciergeID}).Decode(&existing)
	if findErr != nil {
		if errors.Is(findErr, mongo.ErrNoDocuments) {
			return ErrConciergeNotFound
		}
		return fmt.Errorf("failed to verify concierge: %w", findErr)
	}

	for _, a := range existing.Agents {
		if a.ID == agentID {
			return nil
		}
	}
	return ErrAgentNotFound
}

// UnassignInstruction removes an instruction ObjectID from an Agent's instructions array.
// If the instruction is not assigned, it returns idempotently without error.
// Returns ErrConciergeNotFound if concierge does not exist.
// Returns ErrAgentNotFound if agent does not exist in the concierge.
func (r *AgentRepository) UnassignInstruction(ctx context.Context, conciergeID bson.ObjectID, agentID bson.ObjectID, instructionID bson.ObjectID) error {
	filter := bson.M{
		"_id": conciergeID,
		"agents": bson.M{
			"$elemMatch": bson.M{
				"_id":          agentID,
				"instructions": instructionID,
			},
		},
	}

	update := bson.M{
		"$pull": bson.M{
			"agents.$[elem].instructions": instructionID,
		},
		"$set": bson.M{
			"updated_at": time.Now().UTC(),
		},
		"$inc": bson.M{
			"agents.$[elem].version": 1,
		},
	}

	opts := options.UpdateOne().
		SetArrayFilters([]any{bson.M{"elem._id": agentID}})

	res, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to unassign instruction from agent: %w", err)
	}

	if res.MatchedCount > 0 {
		return nil
	}

	var existing models.Concierge
	findErr := r.collection.FindOne(ctx, bson.M{"_id": conciergeID}).Decode(&existing)
	if findErr != nil {
		if errors.Is(findErr, mongo.ErrNoDocuments) {
			return ErrConciergeNotFound
		}
		return fmt.Errorf("failed to verify concierge: %w", findErr)
	}

	for _, a := range existing.Agents {
		if a.ID == agentID {
			return nil
		}
	}
	return ErrAgentNotFound
}
