package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bangweiz/dubhu-designer/internal/identity"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// AgentRepository manages Agent embedded subdocuments within Concierge documents.
type AgentRepository struct {
	collection *scopedCollection
}

// NewAgentRepository creates a new AgentRepository instance pointing to the concierges collection.
func NewAgentRepository(database *mongo.Database) *AgentRepository {
	return &AgentRepository{
		collection: newScopedCollection(database, collectionConcierges),
	}
}

// Client returns the MongoDB client used for service transactions.
func (r *AgentRepository) Client() *mongo.Client {
	return r.collection.collection.Database().Client()
}

// InitIndexes initializes any required indexes for agent operations.
func (r *AgentRepository) InitIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "organisation_id", Value: 1},
				{Key: "_id", Value: 1},
				{Key: "agents.name", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{"agents.0": bson.M{"$exists": true}}),
		},
	)
	return err
}

// Create atomically appends an Agent subdocument to a Concierge's agents array.
// Returns ErrConciergeNotFound if the parent concierge does not exist.
// Returns ErrAgentNameExists for duplicate names and ErrAgentLimitReached at capacity.
func (r *AgentRepository) Create(ctx context.Context, conciergeID bson.ObjectID, agent *models.Agent) (*models.Agent, error) {
	principal, ok := identity.FromContext(ctx)
	if !ok {
		return nil, identity.ErrMissingOrganisation
	}

	agent.CreatedBy = principal.AccountID
	agent.UpdatedBy = principal.AccountID
	filter := bson.M{
		"_id":         conciergeID,
		"agents.name": bson.M{"$ne": agent.Name},
		fmt.Sprintf("agents.%d", models.MaxAgentsPerConcierge-1): bson.M{"$exists": false},
	}
	update := bson.M{
		"$push": bson.M{"agents": agent},
		"$set":  bson.M{"updated_at": time.Now().UTC()},
	}

	res, err := r.collection.UpdateOne(ctx, filter, update)
	if mongo.IsDuplicateKeyError(err) {
		return nil, ErrAgentNameExists
	}
	if err != nil {
		return nil, fmt.Errorf("failed to insert agent into concierge: %w", err)
	}

	if res.MatchedCount == 0 {
		var concierge models.Concierge
		err := r.collection.FindOne(ctx, bson.M{"_id": conciergeID}).Decode(&concierge)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrConciergeNotFound
		}
		if err != nil {
			return nil, fmt.Errorf("failed to verify concierge: %w", err)
		}
		if len(concierge.Agents) >= models.MaxAgentsPerConcierge {
			return nil, ErrAgentLimitReached
		}

		return nil, ErrAgentNameExists
	}

	agent.ConciergeID = conciergeID
	return agent, nil
}

// GetByIDAndConciergeID finds an Agent subdocument within a Concierge by conciergeID and agentID.
// Returns ErrConciergeNotFound if concierge does not exist.
// Returns ErrAgentNotFound if agent does not exist in the concierge.
func (r *AgentRepository) GetByIDAndConciergeID(
	ctx context.Context,
	agentID bson.ObjectID,
	conciergeID bson.ObjectID,
) (*models.Agent, error) {
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

// GetByIDPopulated finds an Agent and resolves its assigned instructions and tools in one aggregation.
// Returns ErrConciergeNotFound if the concierge does not exist.
// Returns ErrAgentNotFound if the agent does not exist in the concierge.
func (r *AgentRepository) GetByIDPopulated(
	ctx context.Context,
	agentID bson.ObjectID,
	conciergeID bson.ObjectID,
) (*models.PopulatedAgent, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"_id": conciergeID}}},
		bson.D{{Key: "$set", Value: bson.M{
			"selected_agent": bson.M{
				"$arrayElemAt": []any{
					bson.M{"$filter": bson.M{
						"input": "$agents",
						"as":    "agent",
						"cond":  bson.M{"$eq": []any{"$$agent._id", agentID}},
					}},
					0,
				},
			},
		}}},
		bson.D{{Key: "$project", Value: bson.M{
			"_id":          "$selected_agent._id",
			"concierge_id": "$_id",
			"name":         "$selected_agent.name",
			"description":  "$selected_agent.description",
			"goal":         "$selected_agent.goal",
			"model":        "$selected_agent.model",
			"instructions": "$selected_agent.instructions",
			"tools":        "$selected_agent.tools",
			"created_by":   "$selected_agent.created_by",
			"updated_by":   "$selected_agent.updated_by",
			"created_at":   "$selected_agent.created_at",
			"updated_at":   "$selected_agent.updated_at",
		}}},
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         collectionInstructions,
			"localField":   "instructions",
			"foreignField": "_id",
			"as":           "resolved_instructions",
		}}},
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         collectionTools,
			"localField":   "tools",
			"foreignField": "_id",
			"as":           "resolved_tools",
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate agent with instructions: %w", err)
	}

	defer cursor.Close(ctx)

	var populated models.PopulatedAgent
	if cursor.Next(ctx) {
		if err := cursor.Decode(&populated); err != nil {
			return nil, fmt.Errorf("failed to decode populated agent: %w", err)
		}
		if populated.ID.IsZero() {
			return nil, ErrAgentNotFound
		}
		if populated.ResolvedInstructions == nil {
			populated.ResolvedInstructions = []models.Instruction{}
		}
		if populated.ResolvedTools == nil {
			populated.ResolvedTools = []models.Tool{}
		}

		return &populated, nil
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error aggregating agent with instructions: %w", err)
	}

	return nil, ErrConciergeNotFound
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
// updates agent fields (name, description, goal, model), advances agent updated_at,
// and updates the concierge's updated_at timestamp.
// Returns ErrConciergeNotFound if the parent concierge does not exist.
// Returns ErrAgentNotFound if the agent does not exist in the concierge.
// Returns ErrAgentNameExists if another agent in this concierge already has the new name.
func (r *AgentRepository) Update(
	ctx context.Context,
	conciergeID bson.ObjectID,
	agentID bson.ObjectID,
	expectedUpdatedAt time.Time,
	agent *models.Agent,
) (*models.Agent, error) {
	now := time.Now().UTC()
	filter := bson.M{
		"_id": conciergeID,
		"$and": []bson.M{
			{"agents": bson.M{"$elemMatch": bson.M{"_id": agentID, "updated_at": expectedUpdatedAt}}},
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
			"agents.$[elem].updated_at":  now,
			"updated_at":                 now,
		},
	}

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After).
		SetArrayFilters([]any{bson.M{"elem._id": agentID, "elem.updated_at": expectedUpdatedAt}})

	var concierge models.Concierge
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&concierge)
	if mongo.IsDuplicateKeyError(err) {
		return nil, ErrAgentNameExists
	}
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
			if !existingAgent.UpdatedAt.Equal(expectedUpdatedAt) {
				return nil, ErrAgentUpdateConflict
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
func (r *AgentRepository) AssignInstruction(
	ctx context.Context,
	conciergeID bson.ObjectID,
	agentID bson.ObjectID,
	instructionID bson.ObjectID,
) error {
	now := time.Now().UTC()
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
			"agents.$[elem].updated_at": now,
			"updated_at":                now,
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
func (r *AgentRepository) UnassignInstruction(
	ctx context.Context,
	conciergeID bson.ObjectID,
	agentID bson.ObjectID,
	instructionID bson.ObjectID,
) error {
	now := time.Now().UTC()
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
			"agents.$[elem].updated_at": now,
			"updated_at":                now,
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

// AssignTool adds a tool ObjectID to an Agent's tools array idempotently.
func (r *AgentRepository) AssignTool(ctx context.Context, conciergeID bson.ObjectID, agentID bson.ObjectID, toolID bson.ObjectID) error {
	return r.assignReference(ctx, conciergeID, agentID, toolID, "tools")
}

// UnassignTool removes a tool ObjectID from an Agent's tools array idempotently.
func (r *AgentRepository) UnassignTool(
	ctx context.Context,
	conciergeID bson.ObjectID,
	agentID bson.ObjectID,
	toolID bson.ObjectID,
) error {
	return r.unassignReference(ctx, conciergeID, agentID, toolID, "tools")
}

func (r *AgentRepository) assignReference(ctx context.Context, conciergeID, agentID, referenceID bson.ObjectID, field string) error {
	now := time.Now().UTC()
	filter := bson.M{
		"_id": conciergeID,
		"agents": bson.M{"$elemMatch": bson.M{
			"_id": agentID,
			field: bson.M{"$ne": referenceID},
		}},
	}
	update := bson.M{
		"$addToSet": bson.M{"agents.$[elem]." + field: referenceID},
		"$set": bson.M{
			"agents.$[elem].updated_at": now,
			"updated_at":                now,
		},
	}
	opts := options.UpdateOne().SetArrayFilters([]any{bson.M{"elem._id": agentID}})
	res, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to assign %s to agent: %w", field, err)
	}
	if res.MatchedCount > 0 {
		return nil
	}

	return r.verifyAgentExists(ctx, conciergeID, agentID)
}

func (r *AgentRepository) unassignReference(ctx context.Context, conciergeID, agentID, referenceID bson.ObjectID, field string) error {
	now := time.Now().UTC()
	filter := bson.M{
		"_id": conciergeID,
		"agents": bson.M{"$elemMatch": bson.M{
			"_id": agentID,
			field: referenceID,
		}},
	}
	update := bson.M{
		"$pull": bson.M{"agents.$[elem]." + field: referenceID},
		"$set": bson.M{
			"agents.$[elem].updated_at": now,
			"updated_at":                now,
		},
	}
	opts := options.UpdateOne().SetArrayFilters([]any{bson.M{"elem._id": agentID}})
	res, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to unassign %s from agent: %w", field, err)
	}
	if res.MatchedCount > 0 {
		return nil
	}

	return r.verifyAgentExists(ctx, conciergeID, agentID)
}

func (r *AgentRepository) verifyAgentExists(ctx context.Context, conciergeID, agentID bson.ObjectID) error {
	var concierge models.Concierge
	if err := r.collection.FindOne(ctx, bson.M{"_id": conciergeID}).Decode(&concierge); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrConciergeNotFound
		}

		return fmt.Errorf("failed to verify concierge: %w", err)
	}

	for _, agent := range concierge.Agents {
		if agent.ID == agentID {
			return nil
		}
	}

	return ErrAgentNotFound
}
