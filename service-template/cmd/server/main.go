package main

import (
	"log/slog"
	"os"

	"mtv-erp/service-template/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("falha ao carregar configuração", "error", err)
		os.Exit(1)
	}

	slog.Info("serviço iniciado", "service", "service-template", "env", cfg.Environment, "port", cfg.Port)
}