package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort        string
	DatabaseURL    string
	OllamaHost     string
	OllamaModel    string
	RequestTimeout time.Duration
}

func Load() (*Config, error) {
	// Keep existing environment variables.
	_ = godotenv.Load()

	timeoutSeconds, err := strconv.Atoi(
		getEnv("REQUEST_TIMEOUT_SECONDS", "90"),
	)
	if err != nil || timeoutSeconds <= 0 {
		return nil, fmt.Errorf(
			"invalid REQUEST_TIMEOUT_SECONDS",
		)
	}

	cfg := &Config{
		AppPort:     getEnv("APP_PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		OllamaHost: getEnv(
			"OLLAMA_HOST",
			"http://localhost:11434",
		),
		OllamaModel: getEnv(
			"OLLAMA_MODEL",
			"llama3.2:latest",
		),
		RequestTimeout: time.Duration(
			timeoutSeconds,
		) * time.Second,
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf(
			"DATABASE_URL is required",
		)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
