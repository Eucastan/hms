package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func AppendJWTToContext(ctx context.Context, tokenStr string) context.Context {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + tokenStr,
	})

	return metadata.NewOutgoingContext(ctx, md)
}
