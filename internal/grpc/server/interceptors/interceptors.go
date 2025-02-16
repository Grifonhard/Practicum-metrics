package interceptor

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/Grifonhard/Practicum-metrics/internal/grpc/metrics_grpc.pb.go"
	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	"google.golang.org/protobuf/types/known/emptypb"
)

// unaryAuthInterceptor проверяет аутентификацию в метаданных unary-запроса.
func unaryAuthInterceptor(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {
    // Пример проверки токена в metadata
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "missing metadata")
    }

    tokens := md["authorization"]
    if len(tokens) == 0 || tokens[0] != "Bearer secret-token" {
        return nil, status.Error(codes.Unauthenticated, "invalid token")
    }

    // Если всё ок — вызываем "настоящий" метод
    return handler(ctx, req)
}

// streamAuthInterceptor проверяет аутентификацию в метаданных streaming-запроса.
func streamAuthInterceptor(
    srv interface{},
    ss grpc.ServerStream,
    info *grpc.StreamServerInfo,
    handler grpc.StreamHandler,
) error {
    // Пример: взять metadata из ServerStream
    md, ok := metadata.FromIncomingContext(ss.Context())
    if !ok {
        return status.Error(codes.Unauthenticated, "missing metadata in stream")
    }
    tokens := md["authorization"]
    if len(tokens) == 0 || tokens[0] != "Bearer secret-token" {
        return status.Error(codes.Unauthenticated, "invalid token in stream")
    }

    // Если токен валидный, передаём управление.
    return handler(srv, ss)
}

// ====== MAIN ======
func main() {
    // Инициализируем сервер
    s := &MetricsServer{
        storage: make(map[string]float64),
    }

    // Создаём gRPC-сервер c интерцепторами
    srv := grpc.NewServer(
        grpc.ChainUnaryInterceptor(unaryAuthInterceptor),
        grpc.ChainStreamInterceptor(streamAuthInterceptor),
    )

    // Регистрируем наш сервис
    pb.RegisterMetricsServiceServer(srv, s)

    // Запуск на :8080 (к примеру)
    lis, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    log.Println("Starting gRPC server on :8080...")
    if err := srv.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
