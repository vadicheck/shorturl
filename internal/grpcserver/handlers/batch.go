// Package handlers batch provides handlers for processing batch URL shortening requests.
package handlers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vadicheck/shorturl/internal/config"
	"github.com/vadicheck/shorturl/internal/constants"
	"github.com/vadicheck/shorturl/internal/models/shorten"
	pbmodels "github.com/vadicheck/shorturl/internal/proto/v1/models"
	pbrpc "github.com/vadicheck/shorturl/internal/proto/v1/rpc"
)

// Batch handles a gRPC request to shorten multiple URLs at once.
//
// It extracts the user ID from the context, validates each request in the batch,
// calls the URL shortening service to generate short links, and returns a response
// with the correlation IDs and corresponding short URLs.
//
// Parameters:
// - ctx: Context containing metadata, including user ID from middleware.
// - in: The incoming gRPC BatchRequest containing the list of original URLs.
//
// Returns:
// - *pbrpc.BatchResponse: The list of shortened URLs with correlation IDs.
// - error: A gRPC error if validation fails or service returns an error.
func (s *ServerAdmin) Batch(ctx context.Context, in *pbrpc.BatchRequest) (*pbrpc.BatchResponse, error) {
	userID, ok := ctx.Value(constants.MdUserID).(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid user-id in context")
	}

	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user-id")
	}

	var request []shorten.CreateBatchURLRequest

	for _, url := range in.GetUrls() {
		request = append(request, shorten.CreateBatchURLRequest{
			CorrelationID: url.GetCorrelationId(),
			OriginalURL:   url.GetOriginalUrl(),
		})
	}

	errs := s.validator.CreateBatchShortURL(&request)
	if len(errs.Errors) != 0 {
		return nil, status.Error(codes.InvalidArgument, errs.Error())
	}

	batchURL, err := s.urlService.CreateBatch(ctx, request, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, errs.Error())
	}

	var pbUrls []*pbmodels.BatchShortUrl

	for _, url := range *batchURL {
		shortURL := config.Config.BaseURL + "/" + url.ShortCode
		pbUrls = append(pbUrls, &pbmodels.BatchShortUrl{
			CorrelationId: &url.CorrelationID,
			ShortUrl:      &shortURL,
		})
	}

	return &pbrpc.BatchResponse{
		Result: pbUrls,
	}, nil
}
