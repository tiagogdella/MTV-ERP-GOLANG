package handlers

import (
	"net/http"
	"strings"

	"mtv-erp/api-gateway/internal/json"
	"mtv-erp/api-gateway/internal/models"
	catalogv1 "mtv-erp/api-gateway/internal/pb/catalog/v1"
)

func validateUnitInput(in models.UnitInput) (msg string, ok bool) {
	if strings.TrimSpace(in.Name) == "" {
		return "O campo 'name' é obrigatório", false
	}
	if strings.TrimSpace(in.ConversionFactorKg) == "" {
		return "O campo 'conversionFactorKg' é obrigatório", false
	}
	return "", true
}
func toUnitDTO(u *catalogv1.UnitOfMeasure) models.Unit {
	return models.Unit{
		ID:                 u.Id,
		Name:               u.Name,
		ConversionFactorKg: u.ConversionFactorKg,
	}
}

// HandleListUnits godoc
// @Summary Lista as unidades de comercialização
// @Tags units
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Unit
// @Failure 401 {object} json.ErrorResponse
// @Router /units [get]
func HandleListUnits(catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := catalogClient.ListUnitsOfMeasure(r.Context(), &catalogv1.ListUnitsOfMeasureRequest{})
		if err != nil {
			json.WriteGRPCError(w, err)
			return
		}

		units := make([]models.Unit, 0, len(resp.Units))
		for _, u := range resp.Units {
			units = append(units, toUnitDTO(u))
		}

		json.WriteJSON(w, http.StatusOK, units)
	}
}

// HandleCreateUnit godoc
// @Summary Cadastra uma unidade de comercialização
// @Description O fator é travado depois de criado (RF-CAT-6). Máximo 4 casas decimais, maior que zero.
// @Tags units
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param unit body models.UnitInput true "Nome e fator de conversão pra kg (como texto, ex: \"30\")"
// @Success 201 {object} models.Unit
// @Failure 400 {object} json.ErrorResponse
// @Failure 401 {object} json.ErrorResponse
// @Failure 422 {object} json.ErrorResponse
// @Router /units [post]
func HandleCreateUnit(catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in models.UnitInput
		if err := json.DecodeJSON(r, &in); err != nil {
			json.WriteError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
			return
		}

		if msg, ok := validateUnitInput(in); !ok {
			json.WriteError(w, http.StatusUnprocessableEntity, msg)
			return
		}

		resp, err := catalogClient.CreateUnitOfMeasure(r.Context(), &catalogv1.CreateUnitOfMeasureRequest{
			Name:               in.Name,
			ConversionFactorKg: in.ConversionFactorKg,
		})
		if err != nil {
			json.WriteGRPCError(w, err)
			return
		}

		json.WriteJSON(w, http.StatusCreated, toUnitDTO(resp.UnitOfMeasure))
	}
}

// HandleConvertUnit godoc
// @Summary Converte uma quantidade da unidade para kg
// @Tags units
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID da unidade (UUID)"
// @Param quantity query string true "Quantidade na unidade (ex: 10)"
// @Success 200 {object} models.Conversion
// @Failure 400 {object} json.ErrorResponse
// @Failure 401 {object} json.ErrorResponse
// @Failure 404 {object} json.ErrorResponse
// @Router /units/{id}/conversion [get]
func HandleConvertUnit(catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := json.ParseID(r)
		if err != nil {
			json.WriteError(w, http.StatusBadRequest, "O 'id' informado não é um UUID válido")
			return
		}

		quantity := r.URL.Query().Get("quantity")
		if quantity == "" {
			json.WriteError(w, http.StatusBadRequest, "informe a quantidade na URL: ?quantity=10")
			return
		}

		resp, err := catalogClient.ConvertToKg(r.Context(), &catalogv1.ConvertToKgRequest{
			UnitId:   id,
			Quantity: quantity,
		})
		if err != nil {
			json.WriteGRPCError(w, err)
			return
		}

		json.WriteJSON(w, http.StatusOK, models.Conversion{
			UnitID:   id,
			Quantity: quantity,
			Kg:       resp.Kg,
		})
	}
}