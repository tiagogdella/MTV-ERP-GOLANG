package grpcserver

import (
	"context"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"mtv-erp/inventory-service/internal/db"
	inventoryv1 "mtv-erp/inventory-service/internal/pb/inventory/v1"
)

func setupTestServer(t *testing.T) *Server {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16",
		postgres.WithDatabase("inventory"),
		postgres.WithUsername("inventory_service"),
		postgres.WithPassword("senha_teste"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err, "failed to start a docker Postgres container")
	t.Cleanup(func() { pgContainer.Terminate(ctx) })

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	m, err := migrate.New("file://../../migrations", connStr)
	require.NoError(t, err)
	require.NoError(t, m.Up())

	database, err := db.Connect(connStr)
	require.NoError(t, err)

	lotRepo := db.NewLotsRepository(database)
	movementRepo := db.NewStockMovementRepository(database)
	return NewServer(lotRepo, movementRepo)
}

func TestRegisterMovement_SemLoteFalha(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	_, err := server.RegisterMovement(ctx, &inventoryv1.RegisterMovementRequest{
		LotId:      uuid.NewString(),
		Type:       "entrada",
		QuantityKg: "100",
		OccurredAt: time.Now().Format(time.RFC3339),
		Origin:     "teste",
	})

	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestGetStockByProduct_SomaVariosLotes(t *testing.T) {
	server := setupTestServer(t)
	ctx := context.Background()

	productID := uuid.NewString()

	lot1, err := server.CreateLot(ctx, &inventoryv1.CreateLotRequest{
		ProductId:      productID,
		PurchaseItemId: uuid.NewString(),
		Safra:          "2026",
		QuantityKg:     "1000",
		ReceivedAt:     "2026-09-01",
	})
	require.NoError(t, err)

	lot2, err := server.CreateLot(ctx, &inventoryv1.CreateLotRequest{
		ProductId:      productID,
		PurchaseItemId: uuid.NewString(),
		Safra:          "2026",
		QuantityKg:     "500",
		ReceivedAt:     "2026-09-05",
	})
	require.NoError(t, err)

	_, err = server.RegisterMovement(ctx, &inventoryv1.RegisterMovementRequest{
		LotId:      lot1.Lot.Id,
		Type:       "entrada",
		QuantityKg: "1000",
		OccurredAt: time.Now().Format(time.RFC3339),
		Origin:     "compra",
	})
	require.NoError(t, err)

	_, err = server.RegisterMovement(ctx, &inventoryv1.RegisterMovementRequest{
		LotId:      lot1.Lot.Id,
		Type:       "saida",
		QuantityKg: "-300",
		OccurredAt: time.Now().Format(time.RFC3339),
		Origin:     "venda",
	})
	require.NoError(t, err)

	_, err = server.RegisterMovement(ctx, &inventoryv1.RegisterMovementRequest{
		LotId:      lot2.Lot.Id,
		Type:       "entrada",
		QuantityKg: "500",
		OccurredAt: time.Now().Format(time.RFC3339),
		Origin:     "compra",
	})
	require.NoError(t, err)

	resp, err := server.GetStockByProduct(ctx, &inventoryv1.GetStockByProductRequest{
		ProductId: productID,
	})
	require.NoError(t, err)
	assert.Equal(t, "1200", resp.TotalKg)
}
