package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"mtv-erp/api-gateway/internal/clients/viacep"
	"mtv-erp/api-gateway/internal/json"
	"mtv-erp/api-gateway/internal/models"
)
// 5.2 Interface - para trocar por fake nos testes
type AddressLookup interface {
	Lookup(ctx context.Context, cep string) (*viacep.Address, error)
}

// HandleGetAddress godoc
// @Summary Consulta um endereço pelo CEP (ViaCEP)
// @Tags addresses
// @Produce json
// @Security BearerAuth
// @Param cep path string true "CEP com 8 dígitos, com ou sem hífen"
// @Success 200 {object} models.Address
// @Failure 400 {object} json.ErrorResponse
// @Failure 401 {object} json.ErrorResponse
// @Failure 404 {object} json.ErrorResponse
// @Failure 502 {object} json.ErrorResponse
// @Failure 504 {object} json.ErrorResponse
// @Router /addresses/{cep} [get]
func HandleGetAddress(lookup AddressLookup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cep := chi.URLParam(r, "cep")

		addr, err := lookup.Lookup(r.Context(), cep)
		if err != nil {
			writeAddressError(w, err)
			return
		}

		json.WriteJSON(w, http.StatusOK, models.Address{
			CEP:          addr.CEP,
			Street:       addr.Street,
			Neighborhood: addr.Neighborhood,
			City:         addr.City,
			State:        addr.State,
		})
	}
}

// Item 4/ 4.1
func writeAddressError(w http.ResponseWriter, err error) {
	switch {
	// Tipo: Negócio | status: 400 | Log: false
	case errors.Is(err, viacep.ErrInvalidCEP):
		json.WriteError(w, http.StatusBadRequest, err.Error())
	// Tipo: Negócio | status: 404 | Log: false
	case errors.Is(err, viacep.ErrNotFound):
		json.WriteError(w, http.StatusNotFound, err.Error())
	// Tipo: Infra | status: 504 | Log: true
	case errors.Is(err, viacep.ErrTimeout):
		slog.Error("timeout consultando ViaCEP", "error", err)
		json.WriteError(w, http.StatusGatewayTimeout, "serviço de CEP não respondeu a tempo")
	// Tipo: Infra | status: 502 | Log: true
	default:
		slog.Error("falha consultando ViaCEP", "error", err)
		json.WriteError(w, http.StatusBadGateway, "serviço de CEP indisponível")
	}
}
