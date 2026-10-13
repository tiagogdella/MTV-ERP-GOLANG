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

type AddressLookup interface {
	Lookup(ctx context.Context, cep string) (*viacep.Address, error)
}

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

func writeAddressError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, viacep.ErrInvalidCEP):
		json.WriteError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, viacep.ErrNotFound):
		json.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, viacep.ErrTimeout):
		slog.Error("timeout consultando ViaCEP", "error", err)
		json.WriteError(w, http.StatusGatewayTimeout, "serviço de CEP não respondeu a tempo")
	default:
		slog.Error("falha consultando ViaCEP", "error", err)
		json.WriteError(w, http.StatusBadGateway, "serviço de CEP indisponível")
	}
}
