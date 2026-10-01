package handlers

import (
	"net/http"

	"mtv-erp/api-gateway/internal/json"
	"mtv-erp/api-gateway/internal/models"
	catalogv1 "mtv-erp/api-gateway/internal/pb/catalog/v1"
)

func HandleListProducts(catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := catalogClient.ListProducts(r.Context(), &catalogv1.ListProductsRequest{})
		if err != nil {
			json.WriteError(w, http.StatusInternalServerError, "erro ao consultar produtos")
			return
		}

		products := make([]models.Product, 0, len(resp.Products))
		for _, p := range resp.Products {
			products = append(products, models.Product{
				ID: p.Id,
				Name: p.Name,
				Active: p.Active,
			})
		}

		json.WriteJSON(w, http.StatusOK, products)
	}
}

func HandleGetProduct(catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := json.ParseId(r)

		resp, err := catalogClient.GetProduct(r.Context(), &catalogv1.GetProductRequest{Id:id})
		if err != nil {
			json.WriteGRPCError(w, err)
			return
		}

		product := models.Product{
			ID: resp.Product.Id,
			Name: resp.Product.Name,
			Active: resp.Product.Active,
		}

		json.WriteJSON(w, http.StatusOK, product)
	}
}