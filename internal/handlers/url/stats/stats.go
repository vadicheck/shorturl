// Package stats provides a handler for retrieving statistics about stored URLs and users.
package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	httpError "github.com/vadicheck/shorturl/internal/http/error"
	"github.com/vadicheck/shorturl/internal/models/shorten"
	"github.com/vadicheck/shorturl/pkg/logger/sl"
)

// URLStorage defines the interface for interacting with the URL storage system.
type URLStorage interface {
	GetCountURLs(ctx context.Context) (int, error)
	GetCountUsers(ctx context.Context) (int, error)
}

// New creates a new HTTP handler for retrieving service statistics.
//
// This handler serves the /api/internal/stats endpoint, which returns the total number of shortened URLs
// and the number of unique users who have created them. The response is returned in JSON format.
//
// Access to this endpoint is restricted by a trusted subnet. The client’s IP address is extracted from
// the "X-Real-IP" header and checked against the provided subnet (in CIDR format). If the client IP is
// not within the trusted subnet or if the subnet is not defined, access is denied with a 403 Forbidden status.
//
// On success, the handler returns an HTTP 200 OK status with a JSON-encoded StatsResponse.
// If any error occurs during processing (such as database errors or encoding failures),
// the handler returns a 500 Internal Server Error.
//
// Parameters:
//   - ctx: The base context for managing request lifecycle and database operations.
//   - storage: An implementation of the URLStorage interface that provides access to URL and user statistics.
//   - subnet: A string in CIDR notation representing the trusted subnet from which access is allowed.
func New(ctx context.Context, storage URLStorage, subnet string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var trustedNet *net.IPNet
		if subnet != "" {
			_, netParsed, err := net.ParseCIDR(subnet)
			if err != nil {
				slog.Error(fmt.Sprintf("Invalid CIDR in trusted_subnet: %v", err))
				http.Error(w, "Failed to get count urls", http.StatusInternalServerError)
				return
			}
			trustedNet = netParsed
		}

		clientIPStr := r.Header.Get("X-Real-IP")
		clientIP := net.ParseIP(clientIPStr)

		if trustedNet == nil || clientIP == nil || !trustedNet.Contains(clientIP) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		countURLs, err := storage.GetCountURLs(ctx)
		if err != nil {
			slog.Error(
				fmt.Sprintf("Failed to get count urls. err: %s", err),
			)
			http.Error(w, "Failed to get count urls", http.StatusInternalServerError)
			return
		}

		countUsers, err := storage.GetCountUsers(ctx)
		if err != nil {
			slog.Error(
				fmt.Sprintf("Failed to get count users. err: %s", err),
			)
			http.Error(w, "Failed to get count users", http.StatusInternalServerError)
			return
		}

		response := shorten.StatsResponse{
			URLs:  countURLs,
			Users: countUsers,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("error encoding response", sl.Err(err))
			httpError.RespondWithError(w, http.StatusInternalServerError, "Failed encoding response")
			return
		}
	}
}
