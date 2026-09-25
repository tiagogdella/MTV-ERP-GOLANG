package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port                 string
	Environment          string
	DatabaseURL          string
	GRPCPort             string
	CatalogServiceAddr   string
	InventoryServiceAddr string
}

func Load() (Config, error) {
	cfg := Config{
		Port:                 getEnv("PORT", "8080"),
		Environment:          getEnv("ENVIRONMENT", "local"),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		GRPCPort:             getEnv("GRPC_PORT", "9090"),
		CatalogServiceAddr:   getEnv("CATALOG_SERVICE_ADDR", ""),
		InventoryServiceAddr: getEnv("INVENTORY_SERVICE_ADDR", ""),
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

	if c.CatalogServiceAddr == "" {
		return fmt.Errorf("CATALOG_SERVICE_ADDR não pode ser vazio")
	}

	if c.InventoryServiceAddr == "" {
		return fmt.Errorf("INVENTORY_SERVICE_ADDR não pode ser vazio")
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
