package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port         string
	DatabaseURL  string
	GeminiAPIKey string
	GeminiModel  string
}

func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.5-flash"
	}

	return Config{
		Port:         port,
		DatabaseURL:  dbURL,
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		GeminiModel:  model,
	}, nil
}
