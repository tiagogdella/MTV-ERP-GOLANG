package json

import (
	"log/slog"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Item 4.1
func WriteGRPCError(w http.ResponseWriter, err error) {
	st := status.Convert(err)

	switch st.Code() {
	// Tipo: Negócio | status: 404 | Log: false
	case codes.NotFound:
		WriteError(w, http.StatusNotFound, st.Message())
	// Tipo: Negócio | status: 400 | Log: false
	case codes.InvalidArgument:
		WriteError(w, http.StatusBadRequest, st.Message())
	// Tipo: Negócio | status: 409 | Log: false
	case codes.AlreadyExists:
		WriteError(w, http.StatusConflict, st.Message())
	// Tipo: Negócio | status: 422 | Log: false
	case codes.FailedPrecondition:
		WriteError(w, http.StatusUnprocessableEntity, st.Message())
	// Tipo: Infra | status: 503 | Log: true
	case codes.Unavailable:
		slog.Error("serviço interno indisponível", "error", err)
		WriteError(w, http.StatusServiceUnavailable, "serviço temporariamente indisponível")
	// Tipo: Infra | status: 504 | Log: true
	case codes.DeadlineExceeded:
		slog.Error("serviço interno não respondeu a tempo", "error", err)
		WriteError(w, http.StatusGatewayTimeout, "serviço não respondeu a tempo")
	// Tipo: Infra (inesperado) | status: 500 | Log: true
	default:
		slog.Error("erro inesperado de um serviço interno", "error", err)
		WriteError(w, http.StatusInternalServerError, "erro inesperado")
	}
}
