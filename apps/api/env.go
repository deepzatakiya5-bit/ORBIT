package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func loadEnv() error {
	if _, err := os.Stat(".env"); err != nil {
		return nil
	}
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: load .env: %v", err)
	}
	return nil
}
