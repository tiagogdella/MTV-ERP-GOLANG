package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"mtv-erp/purchasing-service/internal/db"
	catalogv1 "mtv-erp/purchasing-service/internal/pb/catalog/v1"
	inventoryv1 "mtv-erp/purchasing-service/internal/pb/inventory/v1"
	"github.com/google/uuid"
)

func setupTestService(t *testing.T) (*PurchaseService, *fakeCatalogClient, *fakeInventoryClient) {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16",
		postgres.WithDatabase("purchasing"),
		postgres.WithUsername("purchasing_service"),
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

	purchaseRepo := db.NewPurchaseRepository(database)
	catalogClient := &fakeCatalogClient{}
	inventoryClient := &fakeInventoryClient{}
	svc := NewPurchaseService(purchaseRepo, catalogClient, inventoryClient)

	return svc, catalogClient, inventoryClient
}

func TestCreatePurchase_Success(t *testing.T) {
	svc, catalogClient, inventoryClient := setupTestService(t)
	ctx := context.Background()

	supplierID := uuid.NewString()
	productID := uuid.NewString()
	unitID := uuid.NewString()
	lotID := uuid.NewString()

	catalogClient.On("GetSupplier", mock.Anything, mock.Anything).Return(&catalogv1.GetSupplierResponse{
		Supplier: &catalogv1.Supplier{Id: supplierID, Name: "Fornecedor Teste", Active: true},
	}, nil)

	catalogClient.On("GetProduct", mock.Anything, mock.Anything).Return(&catalogv1.GetProductResponse{
		Product: &catalogv1.Product{Id: productID, Name: "Arroz tipo 1", Active: true},
	}, nil)

	catalogClient.On("ConvertToKg", mock.Anything, mock.Anything).Return(&catalogv1.ConvertToKgResponse{
		Kg: "300",
	}, nil)

	inventoryClient.On("CreateLot", mock.Anything, mock.Anything).Return(&inventoryv1.CreateLotResponse{
		Lot: &inventoryv1.Lot{Id: lotID, ProductId: productID},
	}, nil)

	result, err := svc.CreatePurchase(ctx, CreatePurchaseInput{
		SupplierID:    supplierID,
		InvoiceNumber: "NF-001",
		InvoiceDate:   time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC),
		InvoiceValue:  decimal.NewFromInt(1000),
		Items: []CreatePurchaseItemInput{
			{
				ProductID: productID,
				UnitID:    unitID,
				Safra:     "2026",
				Quantity:  decimal.NewFromInt(10),
			},
		},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, result.Purchase.ID)
	assert.Equal(t, "NF-001", result.Purchase.InvoiceNumber)
	require.Len(t, result.Items, 1)
	assert.Equal(t, lotID, result.Items[0].LotID)
	assert.NotEmpty(t, result.Items[0].Item.ID)

	catalogClient.AssertExpectations(t)
	inventoryClient.AssertExpectations(t)
}

func TestCreatePurchase_SupplierNotFound(t *testing.T) {
	svc, catalogClient, _ := setupTestService(t)
	ctx := context.Background()

	catalogClient.On("GetSupplier", mock.Anything, mock.Anything).Return(nil, status.Error(codes.NotFound, "fornecedor não encontrado"))

	_, err := svc.CreatePurchase(ctx, CreatePurchaseInput{
		SupplierID:    "supplier-inexistente",
		InvoiceNumber: "NF-002",
		InvoiceDate:   time.Now(),
		InvoiceValue:  decimal.NewFromInt(500),
		Items: []CreatePurchaseItemInput{
			{ProductID: "product-1", UnitID: "unit-1", Safra: "2026", Quantity: decimal.NewFromInt(5)},
		},
	})

	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}
