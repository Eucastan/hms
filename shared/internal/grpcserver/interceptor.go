package grpc

import (
	"context"
	"strings"

	"github.com/Eucastan/hms/auth/internal/configs"
	"github.com/Eucastan/shared/internal/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthInterceptor(
	cfg *configs.Config,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata in context")
		}

		authHeader := md["authorization"]
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization token")
		}

		token := strings.TrimPrefix(authHeader[0], "Bearer ")

		claims, err := utils.ValidateToken(token, cfg.JWTSecret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, "claims", claims)
		return handler(ctx, req)

	}
}
