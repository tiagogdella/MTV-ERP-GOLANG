package interceptors

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// Logging devolve um interceptor que loga toda chamada gRPC: método,
// duração e o código de status resultante.
func Logging() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)
		
		slog.Info("chamada gRPC",
				"method", info.FullMethod,
				"duration_ms", time.Since(start).Milliseconds(),
				"code", status.Code(err),
		)
	
		return resp, err
	}

	
}