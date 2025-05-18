// Package handlers provides a handler for creating a shortened URL from a long URL.
package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vadicheck/shorturl/internal/config"
	"github.com/vadicheck/shorturl/internal/constants"
	pbrpc "github.com/vadicheck/shorturl/internal/proto/v1/rpc"
	"github.com/vadicheck/shorturl/internal/services/storage"
	"github.com/vadicheck/shorturl/pkg/validators/url"
)

// Shorten processes a URL shortening request for a given user.
//
// It validates the user ID from the context, checks if the provided URL is valid,
// and then attempts to create a shortened URL via the URL service.
//
// If the URL already exists in storage, it returns the existing shortened URL.
// Otherwise, it creates a new shortened URL and returns it.
//
// Parameters:
// - ctx: Context carrying metadata such as the authenticated user ID.
// - in: The ShortenRequest protobuf message containing the original URL.
//
// Returns:
// - *pbrpc.ShortenResponse containing the shortened URL or the existing shortened URL.
// - error if validation fails or creation encounters an internal error.
func (s *ServerAdmin) Shorten(ctx context.Context, in *pbrpc.ShortenRequest) (*pbrpc.ShortenResponse, error) {
	userID, ok := ctx.Value(constants.MdUserID).(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid user-id in context")
	}

	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user-id")
	}

	reqURL := in.GetUrl()
	_, err := url.IsValid(reqURL)
	if err != nil {
		slog.Error(fmt.Sprintf("URL is invalid: %s", err))
		return nil, status.Error(codes.InvalidArgument, "URL is invalid")
	}

	resp := &pbrpc.ShortenResponse{}

	code, err := s.urlService.Create(ctx, reqURL, userID)
	if err != nil {
		var storageErr *storage.ExistsURLError

		if errors.As(err, &storageErr) {
			r := config.Config.BaseURL + "/" + storageErr.ShortCode
			resp.Result = &r
		} else {
			return nil, status.Error(codes.Internal, "failed to create")
		}
	} else {
		r := config.Config.BaseURL + "/" + code
		resp.Result = &r
	}

	return resp, nil
}
