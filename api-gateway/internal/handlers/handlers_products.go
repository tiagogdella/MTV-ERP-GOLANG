package handlers

import (
	"net/http"
	"strings"

	"mtv-erp/api-gateway/internal/json"
	"mtv-erp/api-gateway/internal/models"
	catalogv1 "mtv-erp/api-gateway/internal/pb/catalog/v1"
)

// HandleListProducts godoc
// @Summary Lista os produtos ativos
// @Tags products
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Product
// @Failure 401 {object} json.ErrorResponse
// @Failure 503 {object} json.ErrorResponse
// @Router /products [get]
func HandleListProducts(catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := catalogClient.ListProducts(r.Context(), &catalogv1.ListProductsRequest{})
		if err != nil {
			json.WriteGRPCError(w, err)
			return
		}

		products := make([]models.Product, 0, len(resp.Products))
		for _, p := range resp.Products {
			products = append(products, models.Product{
				ID:     p.Id,
				Name:   p.Name,
				Active: p.Active,
			})
		}

		json.WriteJSON(w, http.StatusOK, products)
	}
}

// Itens 3.2/3.3

// HandleGetProduct godoc
// @Summary Busca um produto pelo ID
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do produto (UUID)"
// @Success 200 {object} models.Product
// @Failure 400 {object} json.ErrorResponse
// @Failure 401 {object} json.ErrorResponse
// @Failure 404 {object} json.ErrorResponse
// @Router /products/{id} [get]
func HandleGetProduct(catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := json.ParseID(r)
		// 400, o ID não é UUID.
		if err != nil {
			json.WriteError(w, http.StatusBadRequest, "O 'id' informado não é um UUID válido")
			return
		}

		resp, err := catalogClient.GetProduct(r.Context(), &catalogv1.GetProductRequest{Id: id})
		if err != nil {
			json.WriteGRPCError(w, err) // 404 se não existe, 500 se o banco falhou, 503 se o catalog está fora.
			return
		}

		product := models.Product{
			ID:     resp.Product.Id,
			Name:   resp.Product.Name,
			Active: resp.Product.Active,
		}

		// 200, StatusOK
		json.WriteJSON(w, http.StatusOK, product)
	}
}

// Itens 3.2/3.3

// HandleCreateProduct godoc
// @Summary Cadastra um produto
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product body models.ProductInput true "Dados do produto"
// @Success 201 {object} models.Product
// @Failure 400 {object} json.ErrorResponse
// @Failure 401 {object} json.ErrorResponse
// @Failure 422 {object} json.ErrorResponse
// @Router /products [post]
func HandleCreateProduct(catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in models.ProductInput
		// 400, JSON mal formatado
		if err := json.DecodeJSON(r, &in); err != nil {
			json.WriteError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
			return
		}

		// 422, regra de negocio violada
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
			ID:     resp.Product.Id,
			Name:   resp.Product.Name,
			Active: resp.Product.Active,
		}

		// 201 Created
		json.WriteJSON(w, http.StatusCreated, product)
	}
}

// HandleDeactivateProduct godoc
// @Summary Desativa um produto (soft delete)
// @Description Só aceita {"active": false}. Produto nunca é apagado (RF-CAT-2).
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do produto (UUID)"
// @Param body body models.ProductPatchInput true "Sempre {\"active\": false}"
// @Success 200 {object} models.Product
// @Failure 400 {object} json.ErrorResponse
// @Failure 401 {object} json.ErrorResponse
// @Failure 404 {object} json.ErrorResponse
// @Failure 422 {object} json.ErrorResponse
// @Router /products/{id} [patch]
func HandleDeactivateProduct(catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := json.ParseID(r)
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
