package config

import (
	"github.com/joho/godotenv"
	"github.com/markonick/gigs-challenge/internal/logger"
)

func Load() {
	err := godotenv.Load()
	if err != nil {
		logger.Log.Fatal().Msg("Error loading .env file")
	}
}
