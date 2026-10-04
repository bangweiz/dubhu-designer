package main

import (
	"context"
	"fmt"
	"log"

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

	// Dependency Injection: Services
	toolService := service.NewToolService(toolRepo)
	instructionService := service.NewInstructionService(instructionRepo, toolRepo)
	conciergeService := service.NewConciergeService(conciergeRepo)
	agentService := service.NewAgentService(agentRepo, conciergeRepo)

	// Dependency Injection: Controllers
	toolController := controller.NewToolController(toolService)
	instructionController := controller.NewInstructionController(instructionService)
	conciergeController := controller.NewConciergeController(conciergeService)
	agentController := controller.NewAgentController(agentService)

	// Setup Gin router
	router := gin.Default()

	// API versioning group
	apiV1 := router.Group("/api/v1")
	toolController.RegisterRoutes(apiV1)
	instructionController.RegisterRoutes(apiV1)
	conciergeController.RegisterRoutes(apiV1)
	agentController.RegisterRoutes(apiV1)

	log.Println("Server running on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
