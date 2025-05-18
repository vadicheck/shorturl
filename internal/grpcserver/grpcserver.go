// Package grpcserver provides the gRPC server implementation for ShortURL service.
//
// It wraps the ServerAdmin handlers and exposes them as gRPC service methods,
// handling routing of requests and authentication via interceptors.
package grpcserver

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/vadicheck/shorturl/internal/constants"
	"github.com/vadicheck/shorturl/internal/grpcserver/handlers"
	pb "github.com/vadicheck/shorturl/internal/proto/v1"
	pbrpc "github.com/vadicheck/shorturl/internal/proto/v1/rpc"
)

// GRPCServer implements the ShortURLServer interface, forwarding calls to ServerAdmin handlers.
type GRPCServer struct {
	pb.UnimplementedShortURLServer
	*handlers.ServerAdmin
}

// NewGRPCServer creates a new gRPC server wrapping the given ServerAdmin handlers.
//
// Parameters:
// - admin: the ServerAdmin instance containing the business logic handlers.
//
// Returns:
// - pb.ShortURLServer: a gRPC server implementing the ShortURL service interface.
func NewGRPCServer(admin *handlers.ServerAdmin) pb.ShortURLServer {
	return &GRPCServer{
		ServerAdmin: admin,
	}
}

func (s *GRPCServer) Ping(ctx context.Context, in *pbrpc.PingRequest) (*pbrpc.PingResponse, error) {
	return s.ServerAdmin.Ping(ctx, in)
}

func (s *GRPCServer) Shorten(ctx context.Context, in *pbrpc.ShortenRequest) (*pbrpc.ShortenResponse, error) {
	return s.ServerAdmin.Shorten(ctx, in)
}

func (s *GRPCServer) Batch(ctx context.Context, in *pbrpc.BatchRequest) (*pbrpc.BatchResponse, error) {
	return s.ServerAdmin.Batch(ctx, in)
}

func (s *GRPCServer) Delete(ctx context.Context, in *pbrpc.DeleteRequest) (*pbrpc.DeleteResponse, error) {
	return s.ServerAdmin.Delete(ctx, in)
}

func (s *GRPCServer) GetURL(ctx context.Context, in *pbrpc.GetUrlRequest) (*pbrpc.GetUrlResponse, error) {
	return s.ServerAdmin.GetURL(ctx, in)
}

func (s *GRPCServer) GetURLs(ctx context.Context, in *pbrpc.GetUrlsRequest) (*pbrpc.GetUrlsResponse, error) {
	return s.ServerAdmin.GetURLs(ctx, in)
}

func (s *GRPCServer) InternalStats(ctx context.Context, in *pbrpc.StatRequest) (*pbrpc.StatResponse, error) {
	return s.ServerAdmin.InternalStats(ctx, in)
}

// AuthUnaryInterceptor returns a gRPC unary interceptor that enforces authentication
// on protected RPC methods based on presence of user ID metadata.
//
// Parameters:
// - protected: a map of full method names which require authentication.
//
// Returns:
//   - grpc.UnaryServerInterceptor: an interceptor that checks for user ID metadata and
//     injects it into the context, or returns an Unauthenticated error otherwise.
func AuthUnaryInterceptor(protected map[string]bool) grpc.UnaryServerInterceptor {
	slog.Info("Auth interceptor enabled")

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if !protected[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		userIDs := md.Get(string(constants.MdUserID))
		if len(userIDs) == 0 || userIDs[0] == "" {
			return nil, status.Error(codes.Unauthenticated, "missing x-user-id")
		}

		ctx = context.WithValue(ctx, constants.MdUserID, userIDs[0])

		return handler(ctx, req)
	}
}

// LoggingInterceptor returns a gRPC unary interceptor that logs requests and responses.
//
// It logs the full method name, duration, status code, and any error message.
func LoggingInterceptor() grpc.UnaryServerInterceptor {
	slog.Info("gRPC logger interceptor enabled")

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		st := status.Convert(err)

		message := fmt.Sprintf("method: %s, duration: %s, status: %s",
			info.FullMethod,
			duration,
			st.Code().String(),
		)

		if err != nil {
			message += fmt.Sprintf(", error: %s", st.Message())
		}

		slog.Info(message)

		return resp, err
	}
}

// GzipUnaryInterceptor decompresses gzip requests and compresses responses if client accepts it.
func GzipUnaryInterceptor() grpc.UnaryServerInterceptor {
	slog.Info("gRPC gzip interceptor enabled")

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			return resp, err
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return resp, nil
		}

		encodingsRaw := md.Get("grpc-accept-encoding")

		if len(encodingsRaw) == 0 {
			return resp, nil
		}

		encodings := splitCommaHeader(encodingsRaw[0])

		clientAcceptsGzip := false
		for _, enc := range encodings {
			if enc == "gzip" {
				clientAcceptsGzip = true
				break
			}
		}

		if !clientAcceptsGzip {
			return resp, nil
		}

		if message, ok := resp.(proto.Message); ok {
			data, errMarshal := proto.Marshal(message)
			if errMarshal != nil {
				slog.Error("failed to marshal proto message for gzip", "err", errMarshal)
				return resp, nil
			}

			var buf bytes.Buffer
			gw := gzip.NewWriter(&buf)
			if _, errWrite := gw.Write(data); errWrite != nil {
				slog.Error("failed to gzip response", "err", errWrite)
				return resp, nil
			}
			_ = gw.Close()

			return resp, nil
		}

		return resp, nil
	}
}

// splitCommaHeader разбивает заголовок по запятым и чистит пробелы
func splitCommaHeader(header string) []string {
	parts := strings.Split(header, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
