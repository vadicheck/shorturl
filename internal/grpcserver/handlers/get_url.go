// Package handlers defines the GetURL method, which retrieves the original URL by its short identifier.
package handlers

import (
	"context"
	"fmt"
	"log/slog"

	pbrpc "github.com/vadicheck/shorturl/internal/proto/v1/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetURL handles a gRPC request to retrieve the original URL by its short identifier.
//
// It validates the input, queries the storage for the corresponding URL,
// and returns the original URL if found. If the ID is empty or the URL does not exist,
// it responds with an appropriate gRPC error code.
//
// Parameters:
// - ctx: Context for the request lifecycle.
// - in: The incoming GetUrlRequest containing the short URL identifier.
//
// Returns:
// - *pbrpc.GetUrlResponse: The response containing the original URL.
// - error: A gRPC error if the ID is missing, the URL is not found, or another issue occurs.
func (s *ServerAdmin) GetURL(ctx context.Context, in *pbrpc.GetUrlRequest) (*pbrpc.GetUrlResponse, error) {
	id := in.GetId()

	if id == "" {
		slog.Error("id is empty")
		return nil, status.Error(codes.InvalidArgument, "id is empty")
	}

	slog.Info(fmt.Sprintf("id requested: %s", id))

	mURL, err := s.storage.GetURLByID(ctx, id)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get url by id. id: %s, err: %s", id, err))
		return nil, status.Error(codes.Internal, "Failed to get url")
	}

	if mURL.ID == 0 {
		slog.Error(fmt.Sprintf("URL not found. id: %s", id))
		return nil, status.Error(codes.NotFound, "URL not found")
	}

	resp := &pbrpc.GetUrlResponse{
		Url: &mURL.URL,
	}

	return resp, nil
}
