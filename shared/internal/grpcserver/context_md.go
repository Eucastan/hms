package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func AppendJWTToContext(ctx context.Context, token string) context.Context {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})

	return metadata.NewOutgoingContext(ctx, md)
}
