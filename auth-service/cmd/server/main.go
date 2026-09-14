package main

import (
	"log/slog"
	"os"
	"net/http"
	"net"

	"mtv-erp/auth-service/internal/config"
	"mtv-erp/auth-service/internal/health"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"mtv-erp/auth-service/internal/db"
	"google.golang.org/grpc"
	"mtv-erp/auth-service/internal/grpcserver"
	authpb "mtv-erp/auth-service/internal/authpb/auth/v1"
	"context"
	"mtv-erp/auth-service/internal/observability"
	"mtv-erp/auth-service/internal/interceptors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()

	shutdownTracer, err := observability.InitTracer(ctx, "auth-service")
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
		slog.Error("falha ao conectar no banco", "error", err)
		os.Exit(1)
	}

	slog.Info("conectado ao banco de dados")

	repo := db.NewUserRepository(database)
	
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptors.Recovery(),
			interceptors.Logging(),
		),
	)
	authpb.RegisterAuthServiceServer(grpcServer, grpcserver.NewServer(repo, cfg.JWTSecret))


	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		slog.Error("failed opening GRPC PORT", "error", err)
		os.Exit(1)
	}

	go func() {
		slog.Info("server GRPC initialized", "port", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("server GRPC stopped", "error", err)
			os.Exit(1)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", health.LivenessHandler)
	mux.HandleFunc("/readyz", health.ReadinessHandler)
	mux.Handle("/metrics", promhttp.Handler())

	slog.Info("serviço iniciado", "service", "auth-service", "env", cfg.Environment, "port", cfg.Port)

	addr := ":" + cfg.Port
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("servidor parou", "error", err)
		os.Exit(1)
	}
}