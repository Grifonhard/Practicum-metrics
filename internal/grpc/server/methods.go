package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Grifonhard/Practicum-metrics/internal/drivers/psql"
	pb "github.com/Grifonhard/Practicum-metrics/internal/grpc"
	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MetricsServer реализует интерфейс MetricsServiceServer
type MetricsServer struct {
	pb.UnimplementedMetricsServiceServer
	Storage *storage.MemStorage
	DB      *psql.DB
	WG      *sync.WaitGroup
	TrS     *net.IPNet
}

func New(wg *sync.WaitGroup, TrS *net.IPNet, db *psql.DB, stor *storage.MemStorage) *MetricsServer {
	return &MetricsServer{
		WG:      wg,
		TrS:     TrS,
		DB:      db,
		Storage: stor,
	}
}

// PushStream — client streaming: клиент шлёт метрики по одной.
func (s *MetricsServer) PushStream(stream pb.MetricsService_PushStreamServer) error {
	defer s.WG.Done()
	count := 0
	for {
		metric, err := stream.Recv()
		if err == io.EOF {
			// Когда клиент заканчивает стрим, отправим результат
			return stream.SendAndClose(&pb.PushResponse{
				Success: true,
				Message: fmt.Sprintf("Received %d metrics via stream", count),
			})
		}
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "stream recv error: %v", err)
		}

		err = s.Storage.Push(&storage.Metric{
			Type:  metric.Type,
			Name:  metric.Name,
			Value: metric.Value,
		})
		if err != nil {
			return status.Errorf(codes.Internal, "stream recv error: %v", err)
		}
	}
}

// PushBulk — принимает массив метрик (unary).
func (s *MetricsServer) PushBulk(ctx context.Context, in *pb.PushBulkRequest) (*pb.PushResponse, error) {
	defer s.WG.Done()
	var err error
	for _, m := range in.GetMetrics() {
		err = s.Storage.Push(&storage.Metric{
			Type:  m.Type,
			Name:  m.Name,
			Value: m.Value,
		})
		if err != nil {
			return &pb.PushResponse{
				Success: false,
				Message: fmt.Sprintf("fail"),
			}, status.Errorf(codes.Internal, "stream recv error: %v", err)
		}
	}
	return &pb.PushResponse{
		Success: true,
		Message: fmt.Sprintf("Received %d metrics in bulk", len(in.GetMetrics())),
	}, nil
}

// Get — запрос на получение значения метрики.
func (s *MetricsServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	defer s.WG.Done()
	key := fmt.Sprintf("%s:%s", req.GetType(), req.GetName())
	m, err := s.Storage.Get(&storage.Metric{
		Type: req.GetType(),
		Name: req.GetName(),
	})
	if errors.Is(err, storage.ErrMetricNoData) || errors.Is(err, psql.ErrNoData) {
		return nil, status.Errorf(codes.NotFound, "metric not found: %s", key)
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "problem with get: %s", err.Error())
	}
	return &pb.GetResponse{Value: m}, nil
}

// List — вернуть список метрик строками.
func (s *MetricsServer) List(ctx context.Context, _ *emptypb.Empty) (*pb.ListResponse, error) {
	defer s.WG.Done()
	list, err := s.Storage.List()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "problem with list: %s", err.Error())
	}
	return &pb.ListResponse{
		Metrics: list,
	}, nil
}

// PingDB — простая проверка связи (ping).
func (s *MetricsServer) PingDB(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	defer s.WG.Done()
	err := s.DB.PingDB()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "not pong: %s", err.Error())
	}
	// Возвращаем пустой ответ — означает "OK"
	return &emptypb.Empty{}, nil
}
