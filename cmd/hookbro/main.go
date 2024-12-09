package main

import (
	"context"
	"os"
	"strconv"

	"github.com/markonick/gigs-challenge/config"
	"github.com/markonick/gigs-challenge/internal/controllers"
	controller "github.com/markonick/gigs-challenge/internal/controllers"
	container "github.com/markonick/gigs-challenge/internal/di"
	"github.com/markonick/gigs-challenge/internal/logger"
	"github.com/markonick/gigs-challenge/internal/router"
	services "github.com/markonick/gigs-challenge/internal/services"
	"github.com/markonick/gigs-challenge/internal/svix"
	"github.com/markonick/gigs-challenge/internal/worker"
)

func main() {
	// Load configuration
	config.Load()

	container := container.NewContainer()
	container.Invoke(func(controller *controllers.NotificationController) {
		router := router.Setup(controller)
		router.Run(":8080")
	})
	// Get configuration from environment
	svixToken := os.Getenv("SVIX_AUTH_TOKEN")
	if svixToken == "" {
		logger.Log.Fatal().Msg("SVIX_AUTH_TOKEN is not set")
	}

	maxWorkers, err := strconv.Atoi(os.Getenv("MAX_WORKERS"))
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("MAX_WORKERS is not set up correctly")
	}

	// Initialize the Svix client
	svixClient := svix.NewClient(svixToken)

	// Get the projects from the API. For now lets just use the default gigs project dev.
	projects := []string{"dev"}

	// Just initialize the endpoint
	projectAppIDs, err := svix.InitializeApplications(context.Background(), svixClient, projects)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to initialize Svix applications")
	}
	// Create the worker pool
	workerPool := worker.NewPool(maxWorkers)
	// defer workerPool.Close() // Ensure the worker pool is closed when the program exits

	// Task service to inject into the controller
	taskService := services.NewTaskService(
		svixClient,
		projectAppIDs,
		workerPool,
	)
	// Create the controller
	recv := controller.NewNotificationController(taskService)

	// Register the router and start the server to listen for events
	router := router.Setup(recv)

	// Start the server
	logger.Log.Info().Msg("Starting server and listening on port 8080")
	if err := router.Run(":8080"); err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to start server")
	}
}
