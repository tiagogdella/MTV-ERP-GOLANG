package main

import (
	"log/slog"
	"os"
	"net/http"

	"mtv-erp/inventory-service/internal/config"
	"mtv-erp/inventory-service/internal/health"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc/reflection"
	"mtv-erp/inventory-service/internal/db"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"net"
	"context"
	"mtv-erp/inventory-service/internal/observability"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"mtv-erp/inventory-service/internal/interceptors"
	"mtv-erp/inventory-service/internal/grpcserver" 
	inventoryv1 "mtv-erp/inventory-service/internal/pb/inventory/v1"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()

	shutdownTracer, err := observability.InitTracer(ctx, "inventory-service")
	if err != nil {
		slog.Error("falha ao iniciar tracing", "error", err)
		os.Exit(1)
	}
	defer shutdownTracer(ctx)

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

	lotRepo := db.NewLotsRepository(database)
	movementRepo := db.NewStockMovementRepository(database)

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptors.Recovery(),
			interceptors.Logging(),
		),
	)

	inventoryv1.RegisterInventoryServiceServer(grpcServer, grpcserver.NewServer(lotRepo, movementRepo))

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

	slog.Info("serviço iniciado", "service", "inventory-service", "env", cfg.Environment, "port", cfg.Port)

	addr := ":" + cfg.Port
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("servidor parou", "error", err)
		os.Exit(1)
	}

}