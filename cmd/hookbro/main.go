package main

import (
	"github.com/markonick/gigs-challenge/config"
	"github.com/markonick/gigs-challenge/internal/controllers"
	container "github.com/markonick/gigs-challenge/internal/di"
	"github.com/markonick/gigs-challenge/internal/logger"
	"github.com/markonick/gigs-challenge/internal/router"
)

func main() {
	// Load configuration
	config.Load()

	c := container.NewContainer()

	err := c.Invoke(func(controller *controllers.NotificationController) {
		r := router.Setup(controller)

		logger.Log.Info().Msg("Starting server and listening on port 8080")
		if err := r.Run(":8080"); err != nil {
			logger.Log.Fatal().Err(err).Msg("Failed to start server")
		}
	})

	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to start application")
	}
}
