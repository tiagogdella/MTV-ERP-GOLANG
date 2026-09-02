package grpcserver

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"mtv-erp/auth-service/internal/db"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"mtv-erp/auth-service/internal/auth"
	"mtv-erp/auth-service/internal/authpb"

)

func TestLoginIntegration(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16",
		postgres.WithDatabase("auth"),
		postgres.WithUsername("auth_service"),
		postgres.WithPassword("senha_teste"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)

	if err != nil {
		t.Fatalf("failed to start a docker Postgres container: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to connection string: %v", err)
	}

	m, err := migrate.New("file://../../migrations", connStr)
	if err != nil {
		t.Fatalf("failed to load migrations: %v", err)
	}
	if err := m.Up(); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	database,err := db.Connect(connStr)
	if err != nil{
		t.Fatalf("failed to connect gorm: %v", err)
	}
	
	repo := db.NewUserRepository(database)

	hash, err := auth.HashPassword("senha123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	user := &db.User{
		Email:        "integration@mtv.com.br",
		PasswordHash: hash,
		Role:         "operador",
	}
	if err := repo.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	server := NewServer(repo, "segredo-de-teste")

	resp, err := server.Login(ctx, &authpb.LoginRequest{
		Email:    "integration@mtv.com.br",
		Password: "senha123",
	})
	if err != nil {
		t.Fatalf("expected successful login, got error: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected a token, got empty string")
	}

	claims, err := auth.ValidateToken(resp.Token, "segredo-de-teste")
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}
	if claims.Subject != user.ID {
		t.Errorf("expected sub '%s', got '%s'", user.ID, claims.Subject)
	}
	if claims.Role != "operador" {
		t.Errorf("expected role 'operador', got '%s'", claims.Role)
	}

}

