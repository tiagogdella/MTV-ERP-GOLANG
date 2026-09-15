package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	Environment string
	DatabaseURL string
	GRPCPort    string
}

func Load() (Config, error) {
	cfg := Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "local"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		GRPCPort:    getEnv("GRPC_PORT", "9090"),
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	if c.Port == "" {
		return fmt.Errorf("PORT não pode ser vazio")
	}

	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL não pode ser vazio")
	}

	return nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
