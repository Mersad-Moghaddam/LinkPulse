package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPPort        string
	BaseURL         string
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		HTTPPort:        get("HTTP_PORT", "8080"),
		BaseURL:         get("BASE_URL", "http://localhost:8080"),
		ShutdownTimeout: 10 * time.Second,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
	}
	return cfg, nil
}

func get(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
