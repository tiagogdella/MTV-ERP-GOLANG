package main

import (
	"log/slog"
	"os"
	"net/http"

	"mtv-erp/service-template/internal/config"
	"mtv-erp/service-template/internal/health"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc/reflection"
	"mtv-erp/service-template/internal/db"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"net"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("falha ao carregar configuração", "error", err)
		os.Exit(1)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Falha ao conectar com o banco", "error", err)
		os.Exit(1)
	}

	slog.Info("Conectado ao banco de dados")
	_ = database // rep e ser gRPC do servicço entram aqui
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	healthServer := grpchealth.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		slog.Error("falha ao abrir porta gRPC", "error", err)
		os.Exit(1)
	}

	go func() {
		slog.Info("servidor gRPC iniciado", "port", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil{
			slog.Error("servidor gRPC parou", "error", err)
			os.Exit(1)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", health.LivenessHandler)
	mux.HandleFunc("/readyz", health.ReadinessHandler)
	mux.Handle("/metrics", promhttp.Handler())

	slog.Info("serviço iniciado", "service", "service-template", "env", cfg.Environment, "port", cfg.Port)

	addr := ":" + cfg.Port
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("servidor parou", "error", err)
		os.Exit(1)
	}

}