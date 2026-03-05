package utils

import (
	"errors"
	"fmt"

	"google.golang.org/codes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/status"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrUnauthenticated  = errors.New("unauthenticated")
	ErrPermissionDenied = errors.New("permission denied")
	ErrInternal         = errors.New("internal error")
)

func ToPublicError(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("unexpected error: %w", err)
	}

	switch st.Code() {
	case codes.NotFound:
		return fmt.Errorf("%w: %s", ErrNotFound, st.Message())
	case codes.InvalidArgument:
		return fmt.Errorf("%w: %s", ErrInvalidArgument, st.Message())
	case codes.Unauthenticated:
		return fmt.Errorf("%w: %s", ErrUnauthenticated, st.Message())
	case codes.PermissionDenied:
		return fmt.Errorf("%w: %s", ErrPermissionDenied, st.Message())
	case codes.Internal, codes.Unavailable, codes.DeadlineExceeded:
		return fmt.Errorf("%w: %s", ErrInternal, st.Message())
	default:
		return fmt.Errorf("unexpected gRPC error (%s): %s", st.Code(), st.Message())
	}
}
