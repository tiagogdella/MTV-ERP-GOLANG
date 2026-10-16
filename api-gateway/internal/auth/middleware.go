package auth

import (
	"context"
	"net/http"
	"strings"

	"mtv-erp/api-gateway/internal/json"
	authv1 "mtv-erp/api-gateway/internal/pb/auth/v1"
)

type User struct {
	ID   string
	Role string
}

type contextKey struct{}

func RequireAuth(authClient authv1.AuthServiceClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				json.WriteError(w, http.StatusUnauthorized, "token ausente: envie o header 'Authorization: Bearer <token>'")
				return
			}

			resp, err := authClient.ValidateToken(r.Context(), &authv1.ValidateTokenRequest{Token: token})
			if err != nil {
				json.WriteGRPCError(w, err)
				return
			}
			if !resp.Valid {
				json.WriteError(w, http.StatusUnauthorized, "token inválido ou expirado")
				return
			}

			user := User{ID: resp.UserId, Role: resp.Role}
			ctx := context.WithValue(r.Context(), contextKey{}, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok {
				json.WriteError(w, http.StatusUnauthorized, "usuário não autenticado")
				return
			}
			if user.Role != role {
				json.WriteError(w, http.StatusForbidden, "acesso restrito ao papel '"+role+"'")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(contextKey{}).(User)
	return user, ok
}

func bearerToken(r *http.Request) (string, bool) {
	token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !found || strings.TrimSpace(token) == "" {
		return "", false
	}
	return token, true
}
