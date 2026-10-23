package handlers

import (
	"net/http"

	"mtv-erp/api-gateway/internal/json"
	"mtv-erp/api-gateway/internal/models"
	inventoryv1 "mtv-erp/api-gateway/internal/pb/inventory/v1"
)

func toLotDTO(l *inventoryv1.Lot) models.Lot {
	return models.Lot{
		ID:             l.Id,
		ProductID:      l.ProductId,
		PurchaseItemID: l.PurchaseItemId,
		Harvest:        l.Safra,
		QuantityKg:     l.QuantityKg,
		ReceivedAt:     l.ReceivedAt,
	}
}

// HandleGetProductStock godoc
// @Summary Estoque de um produto, com o saldo de cada lote
// @Tags stock
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do produto (UUID)"
// @Success 200 {object} models.ProductStock
// @Failure 400 {object} json.ErrorResponse
// @Failure 401 {object} json.ErrorResponse
// @Router /products/{id}/stock [get]
func HandleGetProductStock(inventoryClient inventoryv1.InventoryServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := json.ParseID(r)
		if err != nil {
			json.WriteError(w, http.StatusBadRequest, "O 'id' informado não é um UUID válido")
			return
		}

		resp, err := inventoryClient.GetStockByProduct(r.Context(), &inventoryv1.GetStockByProductRequest{ProductId: id})
		if err != nil {
			json.WriteGRPCError(w, err)
			return
		}

		lots := make([]models.LotStock, 0, len(resp.Lots))
		for _, l := range resp.Lots {
			lots = append(lots, models.LotStock{
				Lot:       toLotDTO(l.Lot),
				BalanceKg: l.BalanceKg,
			})
		}

		json.WriteJSON(w, http.StatusOK, models.ProductStock{
			ProductID: id,
			TotalKg:   resp.TotalKg,
			Lots:      lots,
		})
	}
}

// HandleGetLotDetails godoc
// @Summary Rastreabilidade de um lote: dados, saldo e todas as movimentações
// @Tags stock
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do lote (UUID)"
// @Success 200 {object} models.LotDetails
// @Failure 400 {object} json.ErrorResponse
// @Failure 401 {object} json.ErrorResponse
// @Failure 404 {object} json.ErrorResponse
// @Router /lots/{id} [get]
func HandleGetLotDetails(inventoryClient inventoryv1.InventoryServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := json.ParseID(r)
		if err != nil {
			json.WriteError(w, http.StatusBadRequest, "O 'id' informado não é um UUID válido")
			return
		}

		resp, err := inventoryClient.GetLotDetails(r.Context(), &inventoryv1.GetLotDetailsRequest{LotId: id})
		if err != nil {
			json.WriteGRPCError(w, err)
			return
		}

		movements := make([]models.StockMovement, 0, len(resp.Movements))
		for _, m := range resp.Movements {
			movements = append(movements, models.StockMovement{
				ID:         m.Id,
				Type:       m.Type,
				QuantityKg: m.QuantityKg,
				OccurredAt: m.OccurredAt,
				Origin:     m.Origin,
			})
		}

		json.WriteJSON(w, http.StatusOK, models.LotDetails{
			Lot:       toLotDTO(resp.Lot),
			BalanceKg: resp.BalanceKg,
			Movements: movements,
		})
	}
}