package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port          string
	DatabaseURL   string
	OpenAIAPIKey  string
	OpenAIModel   string
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

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}

	return Config{
		Port:         port,
		DatabaseURL:  dbURL,
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:  model,
	}, nil
}
