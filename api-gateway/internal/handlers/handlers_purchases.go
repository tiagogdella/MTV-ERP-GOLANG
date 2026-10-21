package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"mtv-erp/api-gateway/internal/json"
	"mtv-erp/api-gateway/internal/models"
	purchasingv1 "mtv-erp/api-gateway/internal/pb/purchasing/v1"
)

func HandleCreatePurchase(purchasingClient purchasingv1.PurchasingServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in models.PurchaseInput
		if err := json.DecodeJSON(r, &in); err != nil {
			json.WriteError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
			return
		}

		if msg, ok := validatePurchaseInput(in); !ok {
			json.WriteError(w, http.StatusUnprocessableEntity, msg)
			return
		}

		items := make([]*purchasingv1.PurchaseItemInput, 0, len(in.Items))
		for _, item := range in.Items {
			items = append(items, &purchasingv1.PurchaseItemInput{
				ProductId: item.ProductID,
				UnitId:    item.UnitID,
				Safra:     item.Harvest,
				Quantity:  item.Quantity,
			})
		}

		resp, err := purchasingClient.CreatePurchase(r.Context(), &purchasingv1.CreatePurchaseRequest{
			SupplierId:    in.SupplierID,
			InvoiceNumber: in.InvoiceNumber,
			InvoiceDate:   in.InvoiceDate,
			InvoiceValue:  in.InvoiceValue,
			Items:         items,
		})
		if err != nil {
			json.WriteGRPCError(w, err)
			return
		}

		json.WriteJSON(w, http.StatusCreated, toPurchaseDTO(resp.Purchase))
	}
}

func toPurchaseDTO(p *purchasingv1.Purchase) models.Purchase {
	items := make([]models.PurchaseItem, 0, len(p.Items))
	for _, item := range p.Items {
		items = append(items, models.PurchaseItem{
			ID:        item.Id,
			ProductID: item.ProductId,
			UnitID:    item.UnitId,
			Harvest:   item.Safra,
			Quantity:  item.Quantity,
			LotID:     item.LotId,
		})
	}

	return models.Purchase{
		ID:            p.Id,
		SupplierID:    p.SupplierId,
		InvoiceNumber: p.InvoiceNumber,
		InvoiceDate:   p.InvoiceDate,
		InvoiceValue:  p.InvoiceValue,
		Items:         items,
	}
}

func validatePurchaseInput(in models.PurchaseInput) (msg string, ok bool) {
	required := map[string]string{
		"supplierId":    in.SupplierID,
		"invoiceNumber": in.InvoiceNumber,
		"invoiceDate":   in.InvoiceDate,
		"invoiceValue":  in.InvoiceValue,
	}
	for _, field := range []string{"supplierId", "invoiceNumber", "invoiceDate", "invoiceValue"} {
		if strings.TrimSpace(required[field]) == "" {
			return fmt.Sprintf("O campo '%s' é obrigatório", field), false
		}
	}

	if len(in.Items) == 0 {
		return "A compra precisa de pelo menos um item em 'items'", false
	}

	for i, item := range in.Items {
		itemFields := map[string]string{
			"productId": item.ProductID,
			"unitId":    item.UnitID,
			"harvest":   item.Harvest,
			"quantity":  item.Quantity,
		}
		for _, field := range []string{"productId", "unitId", "harvest", "quantity"} {
			if strings.TrimSpace(itemFields[field]) == "" {
				return fmt.Sprintf("items[%d]: o campo '%s' é obrigatório", i, field), false
			}
		}
	}

	return "", true
}
