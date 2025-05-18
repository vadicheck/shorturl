// Package handlers defines the GetURLs method, which retrieves all shortened URLs for a specific user.
package handlers

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vadicheck/shorturl/internal/constants"
	pbmodels "github.com/vadicheck/shorturl/internal/proto/v1/models"
	pbrpc "github.com/vadicheck/shorturl/internal/proto/v1/rpc"
)

// GetURLs handles a gRPC request to retrieve all shortened URLs associated with the authenticated user.
//
// It extracts the user ID from the context, queries the storage for the user's URLs,
// and returns a list of original and shortened URLs. If the user ID is not found
// or if an error occurs during retrieval, it responds with a suitable gRPC error.
//
// Parameters:
// - ctx: Context for the request lifecycle.
// - in: The incoming GetUrlsRequest (empty payload).
//
// Returns:
// - *pbrpc.GetUrlsResponse: The response containing the user's list of shortened URLs.
// - error: A gRPC error if user authentication fails or if the storage query fails.
func (s *ServerAdmin) GetURLs(ctx context.Context, in *pbrpc.GetUrlsRequest) (*pbrpc.GetUrlsResponse, error) {
	userID, ok := ctx.Value(constants.MdUserID).(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid user ID in context")
	}

	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user-id")
	}

	mURLs, err := s.storage.GetUserURLs(ctx, userID)
	if err != nil {
		slog.Error(
			fmt.Sprintf("Failed to get user urls. userID: %s, err: %s", userID, err),
		)
		return nil, status.Error(codes.Internal, "Failed to get urls")
	}

	var pbUrls []*pbmodels.ShortUrl

	for _, url := range mURLs {
		pbUrls = append(pbUrls, &pbmodels.ShortUrl{
			ShortUrl:    &url.Code,
			OriginalUrl: &url.URL,
		})
	}

	return &pbrpc.GetUrlsResponse{
		Urls: pbUrls,
	}, nil
}
