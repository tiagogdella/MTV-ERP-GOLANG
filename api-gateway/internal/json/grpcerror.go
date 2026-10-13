package json

import (
	"log/slog"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func WriteGRPCError(w http.ResponseWriter, err error) {
	st := status.Convert(err)

	switch st.Code() {
	case codes.NotFound:
		WriteError(w, http.StatusNotFound, st.Message())
	case codes.InvalidArgument:
		WriteError(w, http.StatusBadRequest, st.Message())
	case codes.AlreadyExists:
		WriteError(w, http.StatusConflict, st.Message())
	case codes.FailedPrecondition:
		WriteError(w, http.StatusUnprocessableEntity, st.Message())
	case codes.Unavailable:
		slog.Error("serviço interno indisponível", "error", err)
		WriteError(w, http.StatusServiceUnavailable, "serviço temporariamente indisponível")
	case codes.DeadlineExceeded:
		slog.Error("serviço interno não respondeu a tempo", "error", err)
		WriteError(w, http.StatusGatewayTimeout, "serviço não respondeu a tempo")
	default:
		slog.Error("erro inesperado de um serviço interno", "error", err)
		WriteError(w, http.StatusInternalServerError, "erro inesperado")
	}
}