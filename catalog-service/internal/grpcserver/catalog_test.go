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
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	supplierRepo := db.NewSupplierRepository(database)
	server := NewServer(productRepo, unitRepo, supplierRepo)

	unitResp, err := server.CreateUnitOfMeasure(ctx, &catalogv1.CreateUnitOfMeasureRequest{
		Name:               "fardo 30kg",
		ConversionFactorKg: "30",
	})
	require.NoError(t, err)

	createProductResp, err := server.CreateProduct(ctx, &catalogv1.CreateProductRequest{
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
	assert.True(t, containsProductID(listResp.Products, createProductResp.Product.Id))

	_, err = server.DeactivateProduct(ctx, &catalogv1.DeactivateProductRequest{
		Id: createProductResp.Product.Id,
	})
	require.NoError(t, err)

	getProductResp, err := server.GetProduct(ctx, &catalogv1.GetProductRequest{Id: createProductResp.Product.Id})
	require.NoError(t, err)
	assert.Equal(t, "Arroz tipo 1", getProductResp.Product.Name)

	_, err = server.GetProduct(ctx, &catalogv1.GetProductRequest{Id: uuid.NewString()})
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	listResp2, err := server.ListProducts(ctx, &catalogv1.ListProductsRequest{})
	require.NoError(t, err)
	assert.False(t, containsProductID(listResp2.Products, createProductResp.Product.Id))

	supplierResp, err := server.CreateSupplier(ctx, &catalogv1.CreateSupplierRequest{
		Name:     "Fornecedor Teste Ltda",
		Document: "12.345.678/0001-90",
		Address:  "Rua Exemplo, 123",
	})
	require.NoError(t, err)
	assert.True(t, supplierResp.Supplier.Active)

	getSupplierResp, err := server.GetSupplier(ctx, &catalogv1.GetSupplierRequest{Id: supplierResp.Supplier.Id})
	require.NoError(t, err)
	assert.Equal(t, "Fornecedor Teste Ltda", getSupplierResp.Supplier.Name)

	_, err = server.GetSupplier(ctx, &catalogv1.GetSupplierRequest{Id: uuid.NewString()})
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	listSuppliersResp, err := server.ListSuppliers(ctx, &catalogv1.ListSuppliersRequest{})
	require.NoError(t, err)
	require.Len(t, listSuppliersResp.Suppliers, 1)

	_, err = server.DeactivateSupplier(ctx, &catalogv1.DeactivateSupplierRequest{
		Id: listSuppliersResp.Suppliers[0].Id,
	})
	require.NoError(t, err)

	listSuppliersResp2, err := server.ListSuppliers(ctx, &catalogv1.ListSuppliersRequest{})
	require.NoError(t, err)
	assert.Len(t, listSuppliersResp2.Suppliers, 0)

}

func containsProductID(products []*catalogv1.Product, id string) bool {
	for _, p := range products {
		if p.Id == id {
			return true
		}
	}
	return false
}

