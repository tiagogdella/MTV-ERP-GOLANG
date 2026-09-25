package main

import (
	"log/slog"
	"os"
	"net/http"

	"mtv-erp/purchasing-service/internal/config"
	"mtv-erp/purchasing-service/internal/health"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc/reflection"
	"mtv-erp/purchasing-service/internal/db"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"net"
	"context"
	"mtv-erp/purchasing-service/internal/observability"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"mtv-erp/purchasing-service/internal/interceptors"
	"google.golang.org/grpc/credentials/insecure"
	catalogv1 "mtv-erp/purchasing-service/internal/pb/catalog/v1"
	inventoryv1 "mtv-erp/purchasing-service/internal/pb/inventory/v1"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()

	shutdownTracer, err := observability.InitTracer(ctx, "purchasing-service")
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
	
	_ = database // temporário, até o repository do Purchase existir

	
	//Abrindo conexão com Catalog
	catalogConn, err := grpc.NewClient(cfg.CatalogServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("falha ao conectar com o catalog-service", "error", err)
		os.Exit(1)
	}
	defer catalogConn.Close()
	catalogClient := catalogv1.NewCatalogServiceClient(catalogConn)

	//Abrindo conexão com inventory
	inventoryConn, err := grpc.NewClient(cfg.InventoryServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("falha ao conectar com o inventory-service", "error", err)
		os.Exit(1)
	}
	defer inventoryConn.Close()
	inventoryClient := inventoryv1.NewInventoryServiceClient(inventoryConn)

	_ = catalogClient
	_ = inventoryClient

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptors.Recovery(),
			interceptors.Logging(),
		),
	)

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

	slog.Info("serviço iniciado", "service", "purchasing-service", "env", cfg.Environment, "port", cfg.Port)

	addr := ":" + cfg.Port
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("servidor parou", "error", err)
		os.Exit(1)
	}

}