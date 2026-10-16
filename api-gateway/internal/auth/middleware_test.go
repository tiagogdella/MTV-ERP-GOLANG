package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authv1 "mtv-erp/api-gateway/internal/pb/auth/v1"
)

type fakeAuthClient struct {
	authv1.AuthServiceClient
	resp   *authv1.ValidateTokenResponse
	err    error
	called bool
}

func (f *fakeAuthClient) ValidateToken(ctx context.Context, in *authv1.ValidateTokenRequest, opts ...grpc.CallOption) (*authv1.ValidateTokenResponse, error) {
	f.called = true
	return f.resp, f.err
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	user, _ := UserFromContext(r.Context())
	w.Header().Set("X-User-Role", user.Role) // COnfere se o usuario chegou certo no context
	w.WriteHeader(http.StatusOK)
}

func TestRequireAuth(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		fake       *fakeAuthClient
		wantStatus int
		wantCalled bool
		wantRole   string
	}{
		{"sem header", "", &fakeAuthClient{}, http.StatusUnauthorized, false, ""},
		{"Bearer sem token", "Bearer ", &fakeAuthClient{}, http.StatusUnauthorized, false, ""},
		{"outro esquema (Basic)", "Basic abc", &fakeAuthClient{}, http.StatusUnauthorized, false, ""},
		{"token inválido ou expirado", "Bearer xyz", &fakeAuthClient{resp: &authv1.ValidateTokenResponse{Valid: false}}, http.StatusUnauthorized, true, ""},
		{"auth-service fora do ar (infra)", "Bearer xyz", &fakeAuthClient{err: status.Error(codes.Unavailable, "connection refused")}, http.StatusServiceUnavailable, true, ""},
		{"token válido", "Bearer xyz", &fakeAuthClient{resp: &authv1.ValidateTokenResponse{Valid: true, UserId: "u-1", Role: "operador"}}, http.StatusOK, true, "operador"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/products", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()

			RequireAuth(tt.fake)(http.HandlerFunc(okHandler)).ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantCalled, tt.fake.called)
			assert.Equal(t, tt.wantRole, rec.Header().Get("X-User-Role"))
		})
	}
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		wantStatus int
	}{
		{"admin acessa", "admin", http.StatusOK},
		{"operador barrado", "operador", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeAuthClient{resp: &authv1.ValidateTokenResponse{Valid: true, UserId: "u-1", Role: tt.role}}
			handler := RequireAuth(fake)(RequireRole("admin")(http.HandlerFunc(okHandler)))

			req := httptest.NewRequest(http.MethodPost, "/users", nil)
			req.Header.Set("Authorization", "Bearer xyz")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}

	t.Run("sem RequireAuth antes nega por padrão", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/users", nil)
		rec := httptest.NewRecorder()

		RequireRole("admin")(http.HandlerFunc(okHandler)).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
