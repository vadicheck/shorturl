// Package handlers provides gRPC handlers for administrative operations,
// including retrieving internal service statistics such as total URL and user counts.
package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/vadicheck/shorturl/internal/config"
	pbrpc "github.com/vadicheck/shorturl/internal/proto/v1/rpc"
)

// InternalStats handles a gRPC request to retrieve internal service statistics,
// including the total number of shortened URLs and registered users.
//
// This method is restricted to trusted networks. It validates the client's IP address
// using the X-Real-IP metadata header and compares it to a configured trusted subnet.
// If the request comes from an untrusted source, access is denied.
//
// Parameters:
// - ctx: The gRPC context, expected to contain metadata with the X-Real-IP header.
// - in: The StatRequest payload (empty).
//
// Returns:
// - *pbrpc.StatResponse: A response containing total counts of URLs and users.
// - error: A gRPC error if authentication fails, permission is denied, or a storage error occurs.
func (s *ServerAdmin) InternalStats(ctx context.Context, in *pbrpc.StatRequest) (*pbrpc.StatResponse, error) {
	var trustedNet *net.IPNet
	if config.Config.TrustedSubnet != "" {
		_, netParsed, err := net.ParseCIDR(config.Config.TrustedSubnet)
		if err != nil {
			slog.Error(fmt.Sprintf("Invalid CIDR in trusted_subnet: %v", err))
			return nil, status.Error(codes.Internal, "invalid CIDR in trusted subnet")
		}
		trustedNet = netParsed
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	clientIPStr := md.Get("X-Real-IP")
	if len(clientIPStr) == 0 || clientIPStr[0] == "" {
		return nil, status.Error(codes.Unauthenticated, "missing X-Real-IP")
	}

	clientIP := net.ParseIP(clientIPStr[0])

	if trustedNet == nil || clientIP == nil || !trustedNet.Contains(clientIP) {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	countURLs, err := s.storage.GetCountURLs(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get count urls. err: %s", err))
		return nil, status.Error(codes.Internal, "failed to get count urls")
	}

	countUsers, err := s.storage.GetCountUsers(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get count users. err: %s", err))
		return nil, status.Error(codes.Internal, "failed to get count users")
	}

	return &pbrpc.StatResponse{
		Urls:  Int32Ptr(countURLs),
		Users: Int32Ptr(countUsers),
	}, nil
}

func Int32Ptr(v int) *int32 {
	v2 := int32(v)
	return &v2
}
