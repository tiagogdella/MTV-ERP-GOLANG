package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"mtv-erp/api-gateway/internal/clients/viacep"
	"mtv-erp/api-gateway/internal/config"
	"mtv-erp/api-gateway/internal/handlers"
	"mtv-erp/api-gateway/internal/health"
	"mtv-erp/api-gateway/internal/observability"
	"mtv-erp/api-gateway/internal/auth"
	"mtv-erp/api-gateway/internal/metrics"
	catalogv1 "mtv-erp/api-gateway/internal/pb/catalog/v1"
	authv1 "mtv-erp/api-gateway/internal/pb/auth/v1"
)

func main() {
	// Criando Logger que escreve em JSON
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()

	//Observability
	shutdownTracer, err := observability.InitTracer(ctx, "api-gateway")
	if err != nil {
		slog.Error("falha ao inciar tracing", "error", err)
		os.Exit(1)
	}
	defer shutdownTracer(ctx)

	// Config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("falha ao carregar configuração", "error", err)
		os.Exit(1) // fail-fast
	}

	// Clients
	catalogConn, err := grpc.NewClient( // Conexão
		cfg.CatalogServiceAddr, // Endereço
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Sem TLS/mTLS no momento
	)
	if err != nil {
		slog.Error("falha ao conectar com o catalog-service", "error", err)
		os.Exit(1)
	}
	defer catalogConn.Close()
	catalogClient := catalogv1.NewCatalogServiceClient(catalogConn) // Métodos

	authConn, err := grpc.NewClient(cfg.AuthServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("falha ao conectar com o auth-service", "error", err)
		os.Exit(1)
	}
	defer authConn.Close()
	authClient := authv1.NewAuthServiceClient(authConn)

	viacepClient := viacep.NewClient(cfg.ViaCEPBaseURL, 5*time.Second)

	r := chi.NewRouter()
	// Middlewares
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer) // (um panic vira 500 em vez de derrubar o processo)
	r.Use(middleware.RequestID) // Cria um ID para Reqs
	r.Use(metrics.Middleware)
	r.Use(middleware.Timeout(10 * time.Second)) // Timer
	r.Use(middleware.AllowContentType("application/json")) // Filtra apenas JSON com 415

	r.Get("/healthz", health.LivenessHandler)
	r.Get("/readyz", health.ReadinessHandler)
	r.Handle("/metrics", promhttp.Handler())

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/login", handlers.HandleLogin(authClient))
		
		r.Group(func(r chi.Router){
			r.Use(auth.RequireAuth(authClient)) // Middleware
			r.With(auth.RequireRole("admin")).Post("/users", handlers.HandleCreateUser(authClient))

			r.Route("/products", func(r chi.Router) {
				r.Get("/", handlers.HandleListProducts(catalogClient))
				r.Post("/", handlers.HandleCreateProduct(catalogClient))
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", handlers.HandleGetProduct(catalogClient))
					r.Patch("/", handlers.HandleDeactivateProduct(catalogClient))
				})
			})
			r.Route("/suppliers", func(r chi.Router) {
				r.Get("/", handlers.HandleListSuppliers(catalogClient))
				r.Post("/", handlers.HandleCreateSupplier(catalogClient))
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", handlers.HandleGetSupplier(catalogClient))
					r.Patch("/", handlers.HandleDeactivateSupplier(catalogClient))
				})
			})
			r.Route("/units", func(r chi.Router) {
				r.Get("/", handlers.HandleListUnits(catalogClient))
				r.Post("/", handlers.HandleCreateUnit(catalogClient))
				r.Get("/{id}/conversion", handlers.HandleConvertUnit(catalogClient))
			})
			r.Get("/addresses/{cep}", handlers.HandleGetAddress(viacepClient))
		})	
	})

	slog.Info("service iniciado", "service", "api-gateway", "env", cfg.Environment, "port", cfg.Port)

	// Inicia HTTP
	addr := ":" + cfg.Port
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("servidor parou", "error", err)
		os.Exit(1)
	}
}
