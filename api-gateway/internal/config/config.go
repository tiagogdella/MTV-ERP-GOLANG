package config

import (
	"fmt"
	"os"
)

// Struct
type Config struct {
	Port               string
	Environment        string
	CatalogServiceAddr string
	ViaCEPBaseURL      string
	AuthServiceAddr    string
	PurchasingServiceAddr string
}

func Load() (Config, error) {
	cfg := Config{
		Port:               	getEnv("PORT", "8080"),
		Environment:        	getEnv("ENVIRONMENT", "local"),
		CatalogServiceAddr: 	getEnv("CATALOG_SERVICE_ADDR", ""), // Sem default
		ViaCEPBaseURL:      	getEnv("VIACEP_BASE_URL", "https://viacep.com.br"),
		AuthServiceAddr:     	getEnv("AUTH_SERVICE_ADDR", ""),
		PurchasingServiceAddr:  getEnv("PURCHASING_SERVICE_ADDR", ""),	
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Validação com fail-fast
func (c Config) validate() error {
	if c.Port == "" {
		return fmt.Errorf("PORT não pode ser vazio")
	}
	if c.CatalogServiceAddr == "" {
		return fmt.Errorf("CATALOG_SERVICE_ADDR não pode ser vazio")
	}
	if c.AuthServiceAddr == "" {
		return fmt.Errorf("AUTH_SERVICE_ADDR não pode ser vazio")
	}
	if c.PurchasingServiceAddr == "" {
		return fmt.Errorf("PURCHASING_SERVICE_ADDR não pode ser vazio")
	}

	return nil
}

// Lê a variável de ambiente
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
