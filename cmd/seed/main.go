// Local development fixtures. The optional -reset flag clears the local database.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/bangweiz/dubhu-designer/internal/db"
	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"github.com/bangweiz/dubhu-designer/internal/repository"
	"github.com/bangweiz/dubhu-designer/internal/service"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func main() {
	reset := flag.Bool("reset", false, "Delete all documents in local dubhu database before seeding")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	database, err := db.ConnectDB()
	must(err)
	defer database.Client().Disconnect(context.Background())
	if !*reset {
		seedExamples(ctx, database)
		return
	}
	// Prepare hashes before removing any existing documents.
	password := "DubhuTest123!"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	must(err)
	names, err := database.ListCollectionNames(ctx, bson.M{})
	must(err)
	for _, name := range names {
		if len(name) >= 7 && name[:7] == "system." {
			continue
		}
		result, err := database.Collection(name).DeleteMany(ctx, bson.M{})
		must(err)
		fmt.Printf("Cleared %s: %d documents\n", name, result.DeletedCount)
	}
	must(repository.NewAgentRepository(database).InitIndexes(ctx))
	authRepo := repository.NewAuthRepository(database)
	must(authRepo.InitIndexes(ctx))
	must(repository.NewInstructionRepository(database).InitIndexes(ctx))
	must(repository.NewToolRepository(database).InitIndexes(ctx))
	must(repository.NewConciergeRepository(database).InitIndexes(ctx))
	must(repository.NewSavedConciergeVersionRepository(database).InitIndexes(ctx))
	must(repository.NewVariableRepository(database).InitIndexes(ctx))
	must(repository.NewEnvironmentRepository(database).InitIndexes(ctx))
	now := time.Now().UTC().Truncate(time.Millisecond)
	orgID, err := bson.ObjectIDFromHex("000000000000000000000001")
	must(err)
	rootID := bson.NewObjectID()
	insert := func(collection string, documents ...any) {
		_, err := database.Collection(collection).InsertMany(ctx, documents)
		must(err)
	}
	insert("organisations", models.Organisation{ID: orgID, Name: "Dubhu Test Workspace", Description: "Development fixtures for the designer UI", RootAccountID: rootID, AuditFields: models.AuditFields{CreatedBy: rootID, UpdatedBy: rootID, CreatedAt: now, UpdatedAt: now}})
	for _, a := range []struct {
		id          bson.ObjectID
		name, email string
		role        models.AccountRole
	}{{rootID, "Alex Morgan", "root@dubhu.test", models.RoleRoot}, {bson.NewObjectID(), "Jamie Chen", "admin@dubhu.test", models.RoleAdmin}, {bson.NewObjectID(), "Taylor Smith", "user@dubhu.test", models.RoleUser}} {
		insert("accounts", models.Account{ID: a.id, OrganisationID: orgID, Name: a.name, Email: a.email, Role: a.role, PasswordHash: string(hash), AuditFields: models.AuditFields{CreatedBy: rootID, UpdatedBy: rootID, CreatedAt: now, UpdatedAt: now}})
	}
	toolIDs := []bson.ObjectID{bson.NewObjectID(), bson.NewObjectID(), bson.NewObjectID()}
	toolNames := []string{"Find a reservation", "Explore local places", "Create support ticket"}
	toolDescriptions := []string{"Look up a guest reservation using a confirmation number.", "Discover nearby restaurants, experiences, and hidden gems.", "Escalate a customer request to the support team."}
	for i, name := range toolNames {
		insert("tools", models.Tool{ID: toolIDs[i], OrganisationID: orgID, Name: name, Description: toolDescriptions[i], Inputs: []models.ToolInput{{Name: []string{"confirmation_number", "location", "message"}[i], Description: "Information provided by the customer", Required: true}}, Outputs: []models.ToolOutput{{Name: "result", Description: "The result of the action"}}, AuditFields: models.AuditFields{CreatedBy: rootID, UpdatedBy: rootID, CreatedAt: now, UpdatedAt: now}})
	}
	instructionIDs := []bson.ObjectID{bson.NewObjectID(), bson.NewObjectID(), bson.NewObjectID()}
	for i, v := range []struct{ name, content string }{{"A warm welcome", "Welcome each guest with warmth. Be clear, considerate, and concise. Use <tool:Find a reservation> when a guest asks about their booking."}, {"Brand voice", "Use a friendly, confident tone. Avoid jargon and make every interaction feel personal."}, {"Support escalation", "Listen carefully, summarise the issue, and use <tool:Create support ticket> when human support is needed."}} {
		if i == 0 {
			v.content = "Welcome each guest with warmth. Be clear, considerate, and concise. Use {{tool:" + toolIDs[0].Hex() + "}} when a guest asks about their booking."
		}
		if i == 2 {
			v.content = "Listen carefully, summarise the issue, and use {{tool:" + toolIDs[2].Hex() + "}} when human support is needed."
		}
		refs := []bson.ObjectID{}
		if i == 0 {
			refs = append(refs, toolIDs[0])
		}
		if i == 2 {
			refs = append(refs, toolIDs[2])
		}
		insert("instructions", models.Instruction{ID: instructionIDs[i], OrganisationID: orgID, Name: v.name, Content: v.content, Tools: refs, AuditFields: models.AuditFields{CreatedBy: rootID, UpdatedBy: rootID, CreatedAt: now, UpdatedAt: now}})
	}
	for i, v := range []struct{ name, description string }{{"Guest experience", "A warm welcome, local recommendations, and effortless stays."}, {"Customer care", "Helpful answers and a human touch, around the clock."}, {"Sales companion", "Turn good conversations into lasting customer relationships."}} {
		id := bson.NewObjectID()
		agent := models.Agent{ID: bson.NewObjectID(), ConciergeID: id, Name: v.name + " assistant", Description: v.description, Goal: "Help customers with clear, useful answers.", Model: models.ModelGemini35Flash, Instructions: []bson.ObjectID{instructionIDs[i], instructionIDs[1]}, Tools: []bson.ObjectID{toolIDs[i]}, AuditFields: models.AuditFields{CreatedBy: rootID, UpdatedBy: rootID, CreatedAt: now, UpdatedAt: now}}
		if i == 1 {
			agent.Instructions = []bson.ObjectID{instructionIDs[1], instructionIDs[2]}
		}
		insert("concierges", models.Concierge{ID: id, OrganisationID: orgID, Name: v.name, Description: v.description, Agents: []models.Agent{agent}, NextVersion: 1, ConciergeVersions: []models.ConciergeVersionReference{}, AuditFields: models.AuditFields{CreatedBy: rootID, UpdatedBy: rootID, CreatedAt: now, UpdatedAt: now}})
	}
	for _, v := range []struct{ name, description string }{{"property_name", "The name of the property welcoming your guests."}, {"support_email", "The email address for the customer care team."}, {"check_in_time", "The standard check-in time for guest reservations."}} {
		insert("variables", models.Variable{Type: models.VariableTypeString, ID: bson.NewObjectID(), OrganisationID: orgID, Name: v.name, Description: v.description, AuditFields: models.AuditFields{CreatedBy: rootID, UpdatedBy: rootID, CreatedAt: now, UpdatedAt: now}})
	}
	for _, v := range []struct{ name, description string }{{"Development", "A workspace for trying new ideas."}, {"Staging", "Review and test changes before going live."}, {"Production", "The live workspace for customer-facing concierges."}} {
		insert("environments", models.Environment{ID: bson.NewObjectID(), OrganisationID: orgID, Name: v.name, Description: v.description, AuditFields: models.AuditFields{CreatedBy: rootID, UpdatedBy: rootID, CreatedAt: now, UpdatedAt: now}})
	}
	auth, err := service.NewAuthService(authRepo)
	must(err)
	for _, email := range []string{"root@dubhu.test", "admin@dubhu.test", "user@dubhu.test"} {
		login, err := auth.Login(ctx, orgID.Hex(), dto.LoginDTO{Email: email, Password: password})
		must(err)
		_, err = auth.Authenticate(ctx, login.AccessToken)
		must(err)
	}
	_, err = database.Collection("auth_sessions").DeleteMany(ctx, bson.M{})
	must(err)
	for _, name := range []string{"organisations", "accounts", "concierges", "instructions", "tools", "variables", "environments"} {
		count, err := database.Collection(name).CountDocuments(ctx, bson.M{})
		must(err)
		fmt.Printf("%s: %d\n", name, count)
	}
	fmt.Printf("Verified all three logins.\nOrganisation ID: %s\nPassword: %s\n", orgID.Hex(), password)
}
