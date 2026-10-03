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

	// Dependency Injection using concrete structs: Repository -> Service -> Controller
	toolRepo := repository.NewToolRepository(database)
	if err := toolRepo.InitIndexes(context.Background()); err != nil {
		panic(fmt.Sprintf("Failed to initialize tool indexes: %v", err))
	}

	toolService := service.NewToolService(toolRepo)
	toolController := controller.NewToolController(toolService)

	// Setup Gin router
	router := gin.Default()

	// API versioning group
	apiV1 := router.Group("/api/v1")
	toolController.RegisterRoutes(apiV1)

	log.Println("Server running on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
