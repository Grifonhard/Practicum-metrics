package server

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryInterceptor для вейтгруппы и проверки real ip
func (s *MetricsServer) UnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// вейтгруппа
	s.WG.Add(1)

	// проверка Real IP
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "missing metadata in stream")
	}

	if s.TrS != nil {
		realIP := md["X-Real-IP"]

		if len(realIP) == 0 {
			return nil, status.Error(codes.InvalidArgument, "missing real ip in metadata")
		}

		agentIP := net.ParseIP(realIP[0])
		if agentIP == nil {
			return nil, status.Error(codes.PermissionDenied, "missing real ip in metadata")
		}

		if !s.TrS.Contains(agentIP) {
			return nil, status.Error(codes.PermissionDenied, "bad trusted subnet")
		}
	}

	return handler(ctx, req)
}

// streamAuthInterceptor проверяет аутентификацию в метаданных streaming-запроса.
func (s *MetricsServer) StreamAuthInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	// вейтгруппа
	s.WG.Add(1)

	// проверка Real IP
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return status.Error(codes.InvalidArgument, "missing metadata in stream")
	}

	if s.TrS != nil {
		realIP := md["X-Real-IP"]

		if len(realIP) == 0 {
			return status.Error(codes.InvalidArgument, "missing real ip in metadata")
		}

		agentIP := net.ParseIP(realIP[0])
		if agentIP == nil {
			return status.Error(codes.PermissionDenied, "missing real ip in metadata")
		}

		if !s.TrS.Contains(agentIP) {
			return status.Error(codes.PermissionDenied, "bad trusted subnet")
		}
	}

	return handler(srv, ss)
}
