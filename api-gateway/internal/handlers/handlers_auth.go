package handlers

import (
	"net/http"
	"strings"

	"mtv-erp/api-gateway/internal/json"
	"mtv-erp/api-gateway/internal/models"
	authv1 "mtv-erp/api-gateway/internal/pb/auth/v1"
)

func HandleLogin(authClient authv1.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in models.LoginInput
		if err := json.DecodeJSON(r, &in); err != nil {
			json.WriteError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
			return
		}

		if msg, ok := validateLoginInput(in); !ok {
			json.WriteError(w, http.StatusUnprocessableEntity, msg)
			return
		}

		resp, err := authClient.Login(r.Context(), &authv1.LoginRequest{
			Email:    in.Email,
			Password: in.Password,
		})
		if err != nil {
			json.WriteGRPCError(w, err)
			return
		}

		json.WriteJSON(w, http.StatusOK, models.LoginOutput{Token: resp.Token})
	}
}

func validateLoginInput(in models.LoginInput) (msg string, ok bool) {
	if strings.TrimSpace(in.Email) == "" {
		return "O campo 'email' é obrigatório", false
	}
	if in.Password == "" {
		return "O campo 'password' é obrigatório", false
	}
	return "", true
}
