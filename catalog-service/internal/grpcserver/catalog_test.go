package grpcserver

import (
	"context"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"mtv-erp/catalog-service/internal/db"
	catalogv1 "mtv-erp/catalog-service/internal/pb/catalog/v1"
)

func TestCatalogServiceIntegration(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16",
		postgres.WithDatabase("catalog"),
		postgres.WithUsername("catalog_service"),
		postgres.WithPassword("senha_teste"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err, "failed to start a docker Postgres container")
	defer pgContainer.Terminate(ctx)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	m, err := migrate.New("file://../../migrations", connStr)
	require.NoError(t, err)
	require.NoError(t, m.Up())

	database, err := db.Connect(connStr)
	require.NoError(t, err)

	productRepo := db.NewProductRepository(database)
	unitRepo := db.NewUnitOfMeasureRepository(database)
	server := NewServer(productRepo, unitRepo)

	unitResp, err := server.CreateUnitOfMeasure(ctx, &catalogv1.CreateUnitOfMeasureRequest{
		Name:               "fardo 30kg",
		ConversionFactorKg: "30",
	})
	require.NoError(t, err)

	_, err = server.CreateProduct(ctx, &catalogv1.CreateProductRequest{
		Name: "Arroz tipo 1",
	})
	require.NoError(t, err)

	convResp, err := server.ConvertToKg(ctx, &catalogv1.ConvertToKgRequest{
		UnitId:   unitResp.UnitOfMeasure.Id,
		Quantity: "2.5",
	})
	require.NoError(t, err)
	assert.Equal(t, "75", convResp.Kg)

	listResp, err := server.ListProducts(ctx, &catalogv1.ListProductsRequest{})
	require.NoError(t, err)
	require.Len(t, listResp.Products, 1)

	_, err = server.DeactivateProduct(ctx, &catalogv1.DeactivateProductRequest{
		Id: listResp.Products[0].Id,
	})
	require.NoError(t, err)

	listResp2, err := server.ListProducts(ctx, &catalogv1.ListProductsRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp2.Products, 0)
}
