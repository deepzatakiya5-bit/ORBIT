package main

import (
	"fmt"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"orbit/apps/api/handler"
	"orbit/pkg/auth"
	"orbit/pkg/config"
	"orbit/pkg/db"
	"orbit/pkg/llm"
	"orbit/pkg/store"
)

func main() {
	_ = loadEnv()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	pool, err := db.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	var provider llm.Provider
	if cfg.GeminiAPIKey != "" {
		provider = llm.NewGemini(cfg.GeminiAPIKey, cfg.GeminiModel)
		log.Printf("LLM: Gemini (%s)", cfg.GeminiModel)
	} else {
		provider = llm.NewStub()
		log.Print("LLM: stub mode (set GEMINI_API_KEY for real replies)")
	}

	tokens, err := auth.NewTokenService(cfg.JWTSecret, 0)
	if err != nil {
		log.Fatalf("auth: %v", err)
	}

	st := store.New(pool)
	h := handler.New(st, provider, tokens)
	srv := NewServer(pool, h, tokens)
	addr := fmt.Sprintf(":%s", cfg.Port)

	log.Printf("ORBIT API listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Routes()); err != nil {
		log.Fatalf("server: %v", err)
	}
}
