package service

import (
	"context"

	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	catalogv1 "mtv-erp/purchasing-service/internal/pb/catalog/v1"
	inventoryv1 "mtv-erp/purchasing-service/internal/pb/inventory/v1"
)

type fakeCatalogClient struct {
	catalogv1.CatalogServiceClient
	mock.Mock
}

func (f *fakeCatalogClient) GetSupplier(ctx context.Context, in *catalogv1.GetSupplierRequest, opts ...grpc.CallOption) (*catalogv1.GetSupplierResponse, error) {
	args := f.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*catalogv1.GetSupplierResponse), args.Error(1)
}

func (f *fakeCatalogClient) GetProduct(ctx context.Context, in *catalogv1.GetProductRequest, opts ...grpc.CallOption) (*catalogv1.GetProductResponse, error) {
	args := f.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*catalogv1.GetProductResponse), args.Error(1)
}

func (f *fakeCatalogClient) ConvertToKg(ctx context.Context, in *catalogv1.ConvertToKgRequest, opts ...grpc.CallOption) (*catalogv1.ConvertToKgResponse, error) {
	args := f.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*catalogv1.ConvertToKgResponse), args.Error(1)
}

type fakeInventoryClient struct {
	inventoryv1.InventoryServiceClient
	mock.Mock
}

func (f *fakeInventoryClient) CreateLot(ctx context.Context, in *inventoryv1.CreateLotRequest, opts ...grpc.CallOption) (*inventoryv1.CreateLotResponse, error) {
	args := f.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*inventoryv1.CreateLotResponse), args.Error(1)
}
