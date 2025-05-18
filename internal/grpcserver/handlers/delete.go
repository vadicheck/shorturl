// Package handlers defines the Delete method, which handles gRPC requests to delete URLs.
package handlers

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vadicheck/shorturl/internal/constants"
	pbrpc "github.com/vadicheck/shorturl/internal/proto/v1/rpc"
	"github.com/vadicheck/shorturl/pkg/logger/sl"
)

// Delete handles a gRPC request to delete multiple user-specific shortened URLs.
//
// It extracts the user ID from the context, validates the list of IDs,
// and asynchronously invokes the deletion service. The method listens for the context
// cancellation or successful completion of the delete operation and returns an empty response.
//
// Parameters:
// - ctx: Context containing metadata, including user ID injected by middleware.
// - in: The incoming DeleteRequest with a list of short URL IDs to delete.
//
// Returns:
// - *pbrpc.DeleteResponse: An empty response on successful deletion initiation.
// - error: A gRPC error if validation fails or the user ID is invalid.
func (s *ServerAdmin) Delete(ctx context.Context, in *pbrpc.DeleteRequest) (*pbrpc.DeleteResponse, error) {
	userID, ok := ctx.Value(constants.MdUserID).(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid user-id in context")
	}

	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user-id")
	}

	request := in.GetIds()

	errs := s.validator.DeleteShortURLs(&request)
	if len(errs.Errors) != 0 {
		return nil, status.Error(codes.InvalidArgument, errs.Error())
	}

	closeCh := make(chan string)

	go func() {
		defer close(closeCh)

		if err := s.urlService.Delete(ctx, request, userID); err != nil {
			slog.Error("failed to delete URLs", sl.Err(err))
		}
	}()

	select {
	case <-closeCh:
	case <-ctx.Done():
		return &pbrpc.DeleteResponse{}, nil
	}

	return &pbrpc.DeleteResponse{}, nil
}
