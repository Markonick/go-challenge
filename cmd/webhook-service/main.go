package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/markonick/gigs-challenge/config"
	"github.com/markonick/gigs-challenge/internal/handlers"
	"github.com/markonick/gigs-challenge/internal/logger"
	"github.com/markonick/gigs-challenge/internal/receiver"
	"github.com/markonick/gigs-challenge/internal/svix"
)

func main() {
	// Load configuration
	config.Load()

	// Get configuration from environment
	svixAPIKey := os.Getenv("SVIX_API_KEY")
	if svixAPIKey == "" {
		logger.Log.Fatal().Msg("SVIX_API_KEY is not set")
	}

	// Initialize the Svix client
	svixClient := svix.NewClient(svixAPIKey)

	// Create the handler
	handler := handlers.NewHandler(svixClient)

	// Create the receiver
	recv := receiver.NewReceiver(handler)

	// Register the receiver with the router and start the server to listen for events
	router := gin.Default()
	router.POST("/notifications", recv.HandleGigsEvent)

	// Start the server
	logger.Log.Info().Msg("Starting server and listening on port 8080")
	if err := router.Run(":8080"); err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to start server")
	}
}
