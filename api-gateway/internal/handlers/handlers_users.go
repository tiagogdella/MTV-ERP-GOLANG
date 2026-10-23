package handlers

import (
	"net/http"
	"strings"

	"mtv-erp/api-gateway/internal/json"
	"mtv-erp/api-gateway/internal/models"
	authv1 "mtv-erp/api-gateway/internal/pb/auth/v1"
)

// HandleCreateUser godoc
// @Summary Cadastra um usuário (somente admin)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body models.UserInput true "E-mail, senha e papel (admin ou operador)"
// @Success 201 {object} models.User
// @Failure 400 {object} json.ErrorResponse
// @Failure 401 {object} json.ErrorResponse
// @Failure 403 {object} json.ErrorResponse
// @Failure 409 {object} json.ErrorResponse
// @Failure 422 {object} json.ErrorResponse
// @Router /users [post]
func HandleCreateUser(authClient authv1.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in models.UserInput
		if err := json.DecodeJSON(r, &in); err != nil {
			json.WriteError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
			return
		}

		if msg, ok := validateUserInput(in); !ok {
			json.WriteError(w, http.StatusUnprocessableEntity, msg)
			return
		}

		resp, err := authClient.CreateUser(r.Context(), &authv1.CreateUserRequest{
			Email:    in.Email,
			Password: in.Password,
			Role:     in.Role,
		})
		if err != nil {
			json.WriteGRPCError(w, err)
			return
		}

		json.WriteJSON(w, http.StatusCreated, models.User{
			ID:    resp.UserId,
			Email: in.Email,
			Role:  in.Role,
		})
	}
}

func validateUserInput(in models.UserInput) (msg string, ok bool) {
	if strings.TrimSpace(in.Email) == "" {
		return "O campo 'email' é obrigatório", false
	}
	if in.Password == "" {
		return "O campo 'password' é obrigatório", false
	}
	if in.Role != "admin" && in.Role != "operador" {
		return "O campo 'role' deve ser 'admin' ou 'operador'", false
	}
	return "", true
}
