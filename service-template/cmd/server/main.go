package main

import (
	"log/slog"
	"os"
	"net/http"

	"mtv-erp/service-template/internal/config"
	"mtv-erp/service-template/internal/health"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("falha ao carregar configuração", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", health.LivenessHandler)
	mux.HandleFunc("/readyz", health.ReadinessHandler)

	slog.Info("serviço iniciado", "service", "service-template", "env", cfg.Environment, "port", cfg.Port)

	addr := ":" + cfg.Port
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("servidor parou", "error", err)
		os.Exit(1)
	}
}