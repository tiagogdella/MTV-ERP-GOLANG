package handlers

import (
	"net/http"
	"strings"

	"mtv-erp/api-gateway/internal/json"
	"mtv-erp/api-gateway/internal/models"
	catalogv1 "mtv-erp/api-gateway/internal/pb/catalog/v1"
)

func toSupplierDTO(s *catalogv1.Supplier) models.Supplier {
	return models.Supplier{
		ID:       s.Id,
		Name:     s.Name,
		Document: s.Document,
		Address:  s.Address,
		Active:   s.Active,
	}
}

func HandleListSuppliers(catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := catalogClient.ListSuppliers(r.Context(), &catalogv1.ListSuppliersRequest{})
		if err != nil {
			json.WriteGRPCError(w, err)
			return
		}

		suppliers := make([]models.Supplier, 0, len(resp.Suppliers))
		for _, s := range resp.Suppliers {
			suppliers = append(suppliers, toSupplierDTO(s))
		}

		json.WriteJSON(w, http.StatusOK, suppliers)
	}
}

func HandleGetSupplier(catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := json.ParseId(r)
		if err != nil {
			json.WriteError(w, http.StatusBadRequest, "O 'id' informado não é um UUID válido")
			return
		}


		resp, err := catalogClient.GetSupplier(r.Context(), &catalogv1.GetSupplierRequest{Id: id})
		if err != nil {
			json.WriteGRPCError(w, err)
			return
		}

		json.WriteJSON(w, http.StatusOK, toSupplierDTO(resp.Supplier))
	}
}

func HandleCreateSupplier(catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in models.SupplierInput
		if err := json.DecodeJSON(r, &in); err != nil {
			json.WriteError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
			return
		}

		if msg, ok := validateSupplierInput(in); !ok {
			json.WriteError(w, http.StatusUnprocessableEntity, msg)
			return
		}

		resp, err := catalogClient.CreateSupplier(r.Context(), &catalogv1.CreateSupplierRequest{
			Name:     in.Name,
			Document: in.Document,
			Address:  in.Address,
		})
		if err != nil {
			json.WriteGRPCError(w, err)
			return
		}

		json.WriteJSON(w, http.StatusCreated, toSupplierDTO(resp.Supplier))
	}
}

func HandleDeactivateSupplier(catalogClient catalogv1.CatalogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := json.ParseId(r)
		if err != nil {
			json.WriteError(w, http.StatusBadRequest, "O 'id' informado não é um UUID válido")
			return
		}


		var in models.SupplierPatchInput
		if err := json.DecodeJSON(r, &in); err != nil {
			json.WriteError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
			return
		}

		if msg, ok := validateSupplierPatchInput(in); !ok {
			json.WriteError(w, http.StatusUnprocessableEntity, msg)
			return
		}

		_, err = catalogClient.DeactivateSupplier(r.Context(), &catalogv1.DeactivateSupplierRequest{Id: id})
		if err != nil {
			json.WriteGRPCError(w, err)
			return
		}

		resp, err := catalogClient.GetSupplier(r.Context(), &catalogv1.GetSupplierRequest{Id: id})
		if err != nil {
			json.WriteGRPCError(w, err)
			return
		}

		json.WriteJSON(w, http.StatusOK, toSupplierDTO(resp.Supplier))
	}
}

func validateSupplierInput(in models.SupplierInput) (msg string, ok bool) {
	if strings.TrimSpace(in.Name) == "" {
		return "O campo 'name' é obrigatório", false
	}
	if strings.TrimSpace(in.Document) == "" {
		return "O campo 'document' é obrigatório", false
	}
	if strings.TrimSpace(in.Address) == "" {
		return "O campo 'address' é obrigatório", false
	}
	return "", true
}

func validateSupplierPatchInput(in models.SupplierPatchInput) (msg string, ok bool) {
	if in.Active == nil {
		return "O campo 'active' é obrigatório", false
	}
	if *in.Active {
		return "Só é possível desativar um fornecedor (active: false)", false
	}
	return "", true
}