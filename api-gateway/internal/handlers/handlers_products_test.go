package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	catalogv1 "mtv-erp/api-gateway/internal/pb/catalog/v1"
)

type fakeCatalogClient struct {
	catalogv1.CatalogServiceClient
	createCalled bool
	createErr    error
	getErr       error
}

func (f *fakeCatalogClient) CreateProduct(ctx context.Context, in *catalogv1.CreateProductRequest, opts ...grpc.CallOption) (*catalogv1.CreateProductResponse, error) {
	f.createCalled = true
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &catalogv1.CreateProductResponse{Product: &catalogv1.Product{Id: "p-1", Name: in.Name, Active: true}}, nil
}

func (f *fakeCatalogClient) GetProduct(ctx context.Context, in *catalogv1.GetProductRequest, opts ...grpc.CallOption) (*catalogv1.GetProductResponse, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &catalogv1.GetProductResponse{Product: &catalogv1.Product{Id: in.Id, Name: "Arroz tipo 1", Active: true}}, nil
}

func TestHandleCreateProduct(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		fake       *fakeCatalogClient
		wantStatus int
		wantCalled bool
	}{
		{"cria produto", `{"name":"Arroz tipo 1"}`, &fakeCatalogClient{}, http.StatusCreated, true},
		{"JSON quebrado", `{"name":`, &fakeCatalogClient{}, http.StatusBadRequest, false},
		{"campo desconhecido", `{"name":"Arroz","id":"x"}`, &fakeCatalogClient{}, http.StatusBadRequest, false},
		{"nome vazio", `{"name":"   "}`, &fakeCatalogClient{}, http.StatusUnprocessableEntity, false},
		{"catalog fora do ar", `{"name":"Arroz"}`, &fakeCatalogClient{createErr: status.Error(codes.Unavailable, "connection refused")}, http.StatusServiceUnavailable, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			HandleCreateProduct(tt.fake).ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantCalled, tt.fake.createCalled)
		})
	}
}

func TestHandleGetProduct(t *testing.T) {
	validID := "5bcb9d1f-304a-4454-9969-1d2fe74e3e55"

	tests := []struct {
		name       string
		id         string
		fake       *fakeCatalogClient
		wantStatus int
	}{
		{"produto encontrado", validID, &fakeCatalogClient{}, http.StatusOK},
		{"id mal formado", "abc", &fakeCatalogClient{}, http.StatusBadRequest},
		{"produto não existe (negócio)", validID, &fakeCatalogClient{getErr: status.Error(codes.NotFound, "produto não encontrado")}, http.StatusNotFound},
		{"falha de banco no catalog (infra)", validID, &fakeCatalogClient{getErr: status.Error(codes.Internal, "conexão com banco caiu")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Get("/products/{id}", HandleGetProduct(tt.fake))

			req := httptest.NewRequest(http.MethodGet, "/products/"+tt.id, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
