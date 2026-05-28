package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port         string
	DatabaseURL  string
	JWTSecret    string
	LLMProvider  string
	GeminiAPIKey string
	GeminiModel  string
	OllamaURL    string
	OllamaModel  string
	MemoryURL    string
	MemoryToken  string
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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.5-flash"
	}
	llmProvider := os.Getenv("LLM_PROVIDER")
	if llmProvider == "" {
		llmProvider = "ollama"
	}
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "qwen2.5:7b-instruct"
	}

	return Config{
		Port:         port,
		DatabaseURL:  dbURL,
		JWTSecret:    jwtSecret,
		LLMProvider:  llmProvider,
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		GeminiModel:  model,
		OllamaURL:    ollamaURL,
		OllamaModel:  ollamaModel,
		MemoryURL:    os.Getenv("MEMORY_SERVICE_URL"),
		MemoryToken:  os.Getenv("MEMORY_SERVICE_TOKEN"),
	}, nil
}
