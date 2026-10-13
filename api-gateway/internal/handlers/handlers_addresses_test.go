package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"mtv-erp/api-gateway/internal/clients/viacep"
)

type fakeAddressLookup struct {
	addr *viacep.Address
	err  error
}

func (f fakeAddressLookup) Lookup(ctx context.Context, cep string) (*viacep.Address, error) {
	return f.addr, f.err
}

func TestHandleGetAddress(t *testing.T) {
	tests := []struct {
		name         string
		fake         fakeAddressLookup
		wantStatus   int
		wantContains string
	}{
		{"CEP encontrado", fakeAddressLookup{addr: &viacep.Address{CEP: "01001-000", City: "São Paulo"}}, http.StatusOK, `"city":"São Paulo"`},
		{"CEP inválido (negócio)", fakeAddressLookup{err: viacep.ErrInvalidCEP}, http.StatusBadRequest, "CEP inválido"},
		{"CEP não encontrado (negócio)", fakeAddressLookup{err: viacep.ErrNotFound}, http.StatusNotFound, "CEP não encontrado"},
		{"ViaCEP lento (infra)", fakeAddressLookup{err: viacep.ErrTimeout}, http.StatusGatewayTimeout, "não respondeu a tempo"},
		{"ViaCEP fora do ar (infra)", fakeAddressLookup{err: fmt.Errorf("%w: connection refused", viacep.ErrUnavailable)}, http.StatusBadGateway, "indisponível"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Get("/addresses/{cep}", HandleGetAddress(tt.fake))

			req := httptest.NewRequest(http.MethodGet, "/addresses/01001000", nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.wantContains)
		})
	}
}
