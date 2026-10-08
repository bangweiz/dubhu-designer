package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/bangweiz/dubhu-designer/internal/controller"
	"github.com/bangweiz/dubhu-designer/internal/db"
	"github.com/bangweiz/dubhu-designer/internal/repository"
	"github.com/bangweiz/dubhu-designer/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	database, err := db.ConnectDB()
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	ctx := context.Background()

	// Dependency Injection: Repositories
	authRepo := repository.NewAuthRepository(database)
	if err := authRepo.InitIndexes(ctx); err != nil {
		panic(fmt.Sprintf("Failed to initialize auth indexes: %v", err))
	}
	variableRepo := repository.NewVariableRepository(database)
	if err := variableRepo.InitIndexes(ctx); err != nil {
		panic(fmt.Sprintf("Failed to initialize variable indexes: %v", err))
	}
	environmentRepo := repository.NewEnvironmentRepository(database)
	if err := environmentRepo.InitIndexes(ctx); err != nil {
		panic(fmt.Sprintf("Failed to initialize environment indexes: %v", err))
	}
	toolRepo := repository.NewToolRepository(database)
	if err := toolRepo.InitIndexes(ctx); err != nil {
		panic(fmt.Sprintf("Failed to initialize tool indexes: %v", err))
	}

	instructionRepo := repository.NewInstructionRepository(database)
	if err := instructionRepo.InitIndexes(ctx); err != nil {
		panic(fmt.Sprintf("Failed to initialize instruction indexes: %v", err))
	}

	conciergeRepo := repository.NewConciergeRepository(database)
	if err := conciergeRepo.InitIndexes(ctx); err != nil {
		panic(fmt.Sprintf("Failed to initialize concierge indexes: %v", err))
	}

	agentRepo := repository.NewAgentRepository(database)
	if err := agentRepo.InitIndexes(ctx); err != nil {
		panic(fmt.Sprintf("Failed to initialize agent indexes: %v", err))
	}
	savedConciergeRepo := repository.NewSavedConciergeVersionRepository(database)
	if err := savedConciergeRepo.InitIndexes(ctx); err != nil {
		panic(fmt.Sprintf("Failed to initialize saved concierge indexes: %v", err))
	}

	// Dependency Injection: Services
	authService, err := service.NewAuthService(authRepo)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize authentication: %v", err))
	}
	variableService := service.NewVariableService(variableRepo)
	environmentService := service.NewEnvironmentService(environmentRepo)
	toolService := service.NewToolService(toolRepo)
	instructionService := service.NewInstructionService(instructionRepo, toolRepo)
	conciergeService := service.NewConciergeService(conciergeRepo, instructionRepo, toolRepo, savedConciergeRepo)
	agentService := service.NewAgentService(agentRepo, conciergeRepo, instructionRepo, toolRepo)

	// Dependency Injection: Controllers
	authController := controller.NewAuthController(authService)
	variableController := controller.NewVariableController(variableService)
	environmentController := controller.NewEnvironmentController(environmentService)
	toolController := controller.NewToolController(toolService)
	instructionController := controller.NewInstructionController(instructionService)
	conciergeController := controller.NewConciergeController(conciergeService)
	agentController := controller.NewAgentController(agentService)

	// Setup Gin router
	router := gin.Default()
	if err := router.SetTrustedProxies(nil); err != nil {
		panic(err)
	}
	router.Use(func(ctx *gin.Context) {
		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 1<<20)
		ctx.Next()
	})

	// API versioning group
	apiV1 := router.Group("/api/v1", authController.RequirePermissions())
	authController.RegisterRoutes(apiV1)
	organisationAPI := apiV1.Group("/organisations/:organisationId")
	variableController.RegisterRoutes(organisationAPI)
	environmentController.RegisterRoutes(organisationAPI)
	toolController.RegisterRoutes(organisationAPI)
	instructionController.RegisterRoutes(organisationAPI)
	conciergeController.RegisterRoutes(organisationAPI)
	agentController.RegisterRoutes(organisationAPI)

	log.Println("Server running on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
