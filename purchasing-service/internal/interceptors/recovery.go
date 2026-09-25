package interceptors

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Recovery devolve um interceptor que recupera panics dentro de um handler
// gRPC, evitando que derrubem o processo inteiro, e os transforma num erro
// gRPC normal (codes.Internal) só pra aquela chamada.
func Recovery() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func(){
			if r := recover(); r != nil {
				slog.Error("panic recuperado", "method", info.FullMethod, "panic", r)
				err = status.Errorf(codes.Internal, "erro interno")
			}
		}()

		return handler(ctx, req)
	}
}