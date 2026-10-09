package handlers

import (
	"net/http"
	"strings"

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
		id, err := json.ParseId(r)
		if err != nil {
			json.WriteError(w, http.StatusBadRequest, "O 'id' informado não é um UUID válido")
			return
		}


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

func HandleCreateProduct(catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in models.ProductInput
		if err := json.DecodeJSON(r, &in); err != nil{
			json.WriteError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
			return
		}

		if msg, ok := validateProductInput(in); !ok {
			json.WriteError(w, http.StatusUnprocessableEntity, msg)
			return
		}

		resp, err := catalogClient.CreateProduct(r.Context(), &catalogv1.CreateProductRequest{
			Name: in.Name,
		})
		if err != nil {
			json.WriteGRPCError(w, err)
			return
		}

		product := models.Product{
			ID: resp.Product.Id,
			Name: resp.Product.Name,
			Active: resp.Product.Active,
		}

		json.WriteJSON(w, http.StatusCreated, product)
	}
}

func HandleDeactivateProduct(catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := json.ParseId(r)
		if err != nil {
			json.WriteError(w, http.StatusBadRequest, "O 'id' informado não é um UUID válido")
			return
		}


		var in models.ProductPatchInput
		if err := json.DecodeJSON(r, &in); err != nil {
			json.WriteError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
			return
		}

		if msg, ok := validateProductPatchInput(in); !ok {
			json.WriteError(w, http.StatusUnprocessableEntity, msg)
			return
		}

		_, err = catalogClient.DeactivateProduct(r.Context(), &catalogv1.DeactivateProductRequest{Id: id})
		if err != nil {
			json.WriteGRPCError(w, err)
			return
		}

		resp, err := catalogClient.GetProduct(r.Context(), &catalogv1.GetProductRequest{Id: id})
		if err != nil {
			json.WriteGRPCError(w, err)
			return
		}

		product := models.Product{
			ID:     resp.Product.Id,
			Name:   resp.Product.Name,
			Active: resp.Product.Active,
		}

		json.WriteJSON(w, http.StatusOK, product)
	}
}

func validateProductInput(in models.ProductInput) (msg string, ok bool) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return "O campo 'name' é obrigatório", false
	}
	return "", true
}

func validateProductPatchInput(in models.ProductPatchInput) (msg string, ok bool) {
	if in.Active == nil {
		return "O campo 'active' é obrigatório", false
	}
	if *in.Active {
		return "Só é possível desativar um produto (active: false)", false
	}

	return "", true
}
