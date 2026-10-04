package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"github.com/bangweiz/dubhu-designer/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestAgentService_GetAgentByID_InvalidConciergeHex(t *testing.T) {
	svc := NewAgentService(nil, nil)
	ctx := context.Background()

	_, err := svc.GetAgentByID(ctx, "invalid-concierge-hex", "651da3b57f8641a27e36c579")
	if !errors.Is(err, ErrConciergeNotFound) {
		t.Errorf("expected ErrConciergeNotFound, got %v", err)
	}
}

func TestAgentService_GetAgentByID_InvalidAgentHex(t *testing.T) {
	svc := NewAgentService(nil, nil)
	ctx := context.Background()

	validConciergeHex := "651da3b57f8641a27e36c579"
	_, err := svc.GetAgentByID(ctx, validConciergeHex, "invalid-agent-hex")
	if !errors.Is(err, ErrAgentNotFound) {
		t.Errorf("expected ErrAgentNotFound, got %v", err)
	}
}

func TestAgentService_CreateAgent_InvalidConciergeHex(t *testing.T) {
	svc := NewAgentService(nil, nil)
	ctx := context.Background()

	_, err := svc.CreateAgent(ctx, "invalid-concierge-hex", dto.CreateAgentDTO{})
	if !errors.Is(err, ErrConciergeNotFound) {
		t.Errorf("expected ErrConciergeNotFound, got %v", err)
	}
}

func TestAgentService_ListAgents_InvalidConciergeHex(t *testing.T) {
	svc := NewAgentService(nil, nil)
	ctx := context.Background()

	_, err := svc.ListAgents(ctx, "invalid-concierge-hex")
	if !errors.Is(err, ErrConciergeNotFound) {
		t.Errorf("expected ErrConciergeNotFound, got %v", err)
	}
}

func TestAgentService_NilConciergeRepo_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic when conciergeRepo is nil, but did not panic")
		}
	}()

	svc := NewAgentService(nil, nil)
	validConciergeHex := "651da3b57f8641a27e36c579"
	validAgentHex := "651da3b57f8641a27e36c582"
	_, _ = svc.GetAgentByID(context.Background(), validConciergeHex, validAgentHex)
}

func TestAgentService_CreateAgent_NilConciergeRepo_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic when conciergeRepo is nil, but did not panic")
		}
	}()

	svc := NewAgentService(nil, nil)
	validConciergeHex := "651da3b57f8641a27e36c579"
	_, _ = svc.CreateAgent(context.Background(), validConciergeHex, dto.CreateAgentDTO{Name: "agent-1"})
}

func TestAgentService_GetAndListAgents_Integration(t *testing.T) {
	uri := "mongodb://localhost:27017/?directConnection=true"
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration test")
	}
	defer client.Disconnect(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		t.Skip("MongoDB ping failed, skipping integration test")
	}

	testDB := client.Database("dubhu_test_agent_service")
	defer testDB.Drop(context.Background())

	conciergeRepo := repository.NewConciergeRepository(testDB)
	agentRepo := repository.NewAgentRepository(testDB)
	svc := NewAgentService(agentRepo, conciergeRepo)

	// Create concierge
	concierge, err := conciergeRepo.Create(ctx, &models.Concierge{
		ID:          bson.NewObjectID(),
		Name:        "Test Concierge",
		Description: "For Agent Testing",
		Agents:      []models.Agent{},
		Version:     1,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("failed to create concierge: %v", err)
	}

	// 1. List agents when empty
	listEmpty, err := svc.ListAgents(ctx, concierge.ID.Hex())
	if err != nil {
		t.Fatalf("unexpected error listing empty agents: %v", err)
	}
	if len(listEmpty) != 0 {
		t.Errorf("expected 0 agents, got %d", len(listEmpty))
	}

	// 2. Create agent
	createdAgent, err := svc.CreateAgent(ctx, concierge.ID.Hex(), dto.CreateAgentDTO{
		Name:        "agent-alpha",
		Description: "Alpha Agent",
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	// 3. Get agent by ID from concierge.Agents directly in memory
	gotAgent, err := svc.GetAgentByID(ctx, concierge.ID.Hex(), createdAgent.ID)
	if err != nil {
		t.Fatalf("unexpected error getting agent by ID: %v", err)
	}
	if gotAgent.ID != createdAgent.ID {
		t.Errorf("expected agent ID %s, got %s", createdAgent.ID, gotAgent.ID)
	}
	if gotAgent.Name != "agent-alpha" {
		t.Errorf("expected agent name 'agent-alpha', got %s", gotAgent.Name)
	}
	if gotAgent.ConciergeID != concierge.ID.Hex() {
		t.Errorf("expected concierge ID %s, got %s", concierge.ID.Hex(), gotAgent.ConciergeID)
	}

	// 4. Get non-existent agent ID from existing concierge
	nonExistentAgentHex := bson.NewObjectID().Hex()
	_, err = svc.GetAgentByID(ctx, concierge.ID.Hex(), nonExistentAgentHex)
	if !errors.Is(err, ErrAgentNotFound) {
		t.Errorf("expected ErrAgentNotFound, got %v", err)
	}

	// 5. List agents should now return the created agent
	listPopulated, err := svc.ListAgents(ctx, concierge.ID.Hex())
	if err != nil {
		t.Fatalf("unexpected error listing agents: %v", err)
	}
	if len(listPopulated) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(listPopulated))
	}
	if listPopulated[0].ID != createdAgent.ID {
		t.Errorf("expected listed agent ID %s, got %s", createdAgent.ID, listPopulated[0].ID)
	}
}
