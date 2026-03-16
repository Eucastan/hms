package grpc

import (
	"context"
	"strings"

	"github.com/Eucastan/hms/shared/internal/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthInterceptor(
	cfg string,
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

		claims, err := utils.ValidateToken(token, cfg)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Propagate useful values into context
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "role", claims.Role)
		ctx = context.WithValue(ctx, "jwt_token", token)

		// Also propagate metadata for outgoing calls
		newMD := metadata.New(map[string]string{
			"user_id":       claims.UserID,
			"role":          claims.Role,
			"authorization": "Bearer " + token,
		})

		// Merge with existing outgoing metadata if any
		outgoingMD, _ := metadata.FromOutgoingContext(ctx)
		newMD = metadata.Join(outgoingMD, newMD)

		ctx = metadata.NewOutgoingContext(ctx, newMD)

		return handler(ctx, req)

	}
}
