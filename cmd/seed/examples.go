package main

import (
	"context"
	"fmt"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/identity"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"github.com/bangweiz/dubhu-designer/internal/repository"
	"github.com/bangweiz/dubhu-designer/internal/service"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// seedExamples adds reference examples without resetting data or changing existing records.
func seedExamples(ctx context.Context, database *mongo.Database) {
	authRepo := repository.NewAuthRepository(database)
	must(authRepo.InitIndexes(ctx))
	auth, err := service.NewAuthService(authRepo)
	must(err)
	orgID, err := bson.ObjectIDFromHex("000000000000000000000001")
	must(err)
	org, err := authRepo.GetOrganisation(ctx, orgID)
	must(err)
	createdOrganisation := false
	if org == nil {
		var existing models.Organisation
		err := database.Collection("organisations").FindOne(ctx, bson.M{"name": "Dubhu Variable Demo"}).Decode(&existing)
		if err == nil {
			org = &existing
		} else if err != mongo.ErrNoDocuments {
			must(err)
		}
	}
	if org == nil {
		response, err := auth.CreateOrganisation(ctx, dto.CreateOrganisationDTO{
			Name: "Dubhu Variable Demo", Description: "Local examples of typed variables and instruction references.",
			RootAccount: dto.RootAccountDTO{
				Name:     "Demo administrator",
				Email:    "root@dubhu.test",
				Password: "DubhuTest123!2026",
			},
		})
		must(err)
		orgID, err = bson.ObjectIDFromHex(response.Organisation.ID)
		must(err)
		org, err = authRepo.GetOrganisation(ctx, orgID)
		must(err)
		createdOrganisation = true
	}

	root, err := authRepo.GetAccount(ctx, org.ID, org.RootAccountID)
	must(err)
	if root == nil || root.Role != models.RoleRoot {
		panic("Demo organisation has no valid root account")
	}

	ctx = identity.WithPrincipal(ctx, identity.Principal{OrganisationID: org.ID, AccountID: root.ID, Role: string(root.Role)})
	variableRepo := repository.NewVariableRepository(database)
	instructionRepo := repository.NewInstructionRepository(database)
	toolRepo := repository.NewToolRepository(database)
	conciergeRepo := repository.NewConciergeRepository(database)
	agentRepo := repository.NewAgentRepository(database)
	must(variableRepo.InitIndexes(ctx))
	must(instructionRepo.InitIndexes(ctx))
	must(conciergeRepo.InitIndexes(ctx))
	must(agentRepo.InitIndexes(ctx))
	variableService := service.NewVariableService(variableRepo)
	instructionService := service.NewInstructionService(instructionRepo, toolRepo, variableRepo)
	conciergeService := service.NewConciergeService(
		conciergeRepo,
		instructionRepo,
		toolRepo,
		variableRepo,
		repository.NewSavedConciergeVersionRepository(database),
		repository.NewEnvironmentRepository(database),
	)
	agentService := service.NewAgentService(agentRepo, conciergeRepo, instructionRepo, toolRepo)
	existingVariables, err := variableRepo.List(ctx)
	must(err)
	variableIDs := make(map[string]string)
	for _, variable := range existingVariables {
		variableIDs[variable.Name] = variable.ID.Hex()
	}

	for _, input := range []dto.CreateVariableDTO{
		{
			Name:        "demo_property_name",
			Description: "The property name used to greet guests.",
			Type:        "string",
		},
		{
			Name:        "demo_nightly_rate",
			Description: "The nightly room rate as a numeric amount.",
			Type:        "number",
		},
		{
			Name:        "demo_breakfast_included",
			Description: "Whether breakfast is included with the stay.",
			Type:        "bool",
		},
	} {
		if variableIDs[input.Name] == "" {
			variable, err := variableService.CreateVariable(ctx, input)
			must(err)
			variableIDs[input.Name] = variable.ID
		}

		fmt.Printf("Variable %s (%s): %s\n", input.Name, input.Type, variableIDs[input.Name])
	}

	content := fmt.Sprintf(
		"Welcome guests to {{var:%s}}. Explain the nightly rate using {{var:%s}}. Use {{var:%s}} to determine whether breakfast is included. Mention {{var:%s}} again when closing the conversation.",
		variableIDs["demo_property_name"],
		variableIDs["demo_nightly_rate"],
		variableIDs["demo_breakfast_included"],
		variableIDs["demo_property_name"],
	)
	instructions, err := instructionRepo.List(ctx)
	must(err)
	instructionID := ""
	for _, instruction := range instructions {
		if instruction.Name == "Typed variable demo" {
			instructionID = instruction.ID.Hex()
		}
	}
	if instructionID == "" {
		instruction, err := instructionService.CreateInstruction(ctx, dto.CreateInstructionDTO{Name: "Typed variable demo", Content: content})
		must(err)
		instructionID = instruction.ID
	}

	concierges, err := conciergeRepo.List(ctx)
	must(err)
	conciergeID := ""
	for _, concierge := range concierges {
		if concierge.Name == "Variable reference demo" {
			conciergeID = concierge.ID.Hex()
		}
	}
	if conciergeID == "" {
		concierge, err := conciergeService.CreateConcierge(
			ctx,
			dto.CreateConciergeDTO{
				Name:        "Variable reference demo",
				Description: "A guest assistant demonstrating string, number, and bool variable references.",
			},
		)
		must(err)
		conciergeID = concierge.ID
	}

	agents, err := agentService.ListAgents(ctx, conciergeID)
	must(err)
	agentID := ""
	for _, agent := range agents {
		if agent.Name == "Typed variable assistant" {
			agentID = agent.ID
		}
	}
	if agentID == "" {
		agent, err := agentService.CreateAgent(
			ctx,
			conciergeID,
			dto.CreateAgentDTO{
				Name:        "Typed variable assistant",
				Description: "Uses typed guest-stay variables.",
				Goal:        "Explain the property, nightly rate, and breakfast availability clearly.",
				Model:       string(models.ModelGemini35Flash),
			},
		)
		must(err)
		agentID = agent.ID
	}

	must(agentService.AssignInstruction(ctx, conciergeID, agentID, instructionID))
	instruction, err := instructionService.GetInstructionByID(ctx, instructionID)
	must(err)
	fmt.Printf("Stored instruction resolves %d variable definitions.\n", len(instruction.Variables))
	fmt.Printf("Organisation: %s\nInstruction: %s\nConcierge: %s\nAgent: %s\n", org.ID.Hex(), instructionID, conciergeID, agentID)
	if createdOrganisation {
		fmt.Println("Created root@dubhu.test with password DubhuTest123!2026")
	} else {
		fmt.Println("Existing account credentials preserved.")
	}
}
