package main

import (
	"context"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/markonick/gigs-challenge/config"
	"github.com/markonick/gigs-challenge/internal/logger"
	"github.com/markonick/gigs-challenge/internal/receiver"
	"github.com/markonick/gigs-challenge/internal/svix"
	"github.com/markonick/gigs-challenge/internal/worker"
)

func main() {
	// Load configuration
	config.Load()

	// Get configuration from environment
	svixToken := os.Getenv("SVIX_AUTH_TOKEN")
	if svixToken == "" {
		logger.Log.Fatal().Msg("SVIX_AUTH_TOKEN is not set")
	}

	gigsAPIURL := os.Getenv("GIGS_API_URL")
	if gigsAPIURL == "" {
		logger.Log.Fatal().Msg("GIGS_API_URL is not set")
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

	// Create the receiver
	recv := receiver.NewReceiver(svixClient, projectAppIDs, workerPool)

	// Register the receiver with the router and start the server to listen for events
	router := gin.Default()
	router.POST("/notifications", recv.HandleNotification)

	// Start the server
	logger.Log.Info().Msg("Starting server and listening on port 8080")
	if err := router.Run(":8080"); err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to start server")
	}
}
