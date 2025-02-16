package agent

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/Grifonhard/Practicum-metrics/internal/grpc"
	"github.com/Grifonhard/Practicum-metrics/internal/logger"
	metgen "github.com/Grifonhard/Practicum-metrics/internal/met_gen"
	webclient "github.com/Grifonhard/Practicum-metrics/internal/web_client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Agent struct {
	addr string
	wg   *sync.WaitGroup
	url  string
	gen  *metgen.MetGen
	ctx  context.Context
}

func New(addr string, wg *sync.WaitGroup, gen *metgen.MetGen, realIP string) *Agent {
	return &Agent{
		wg:  wg,
		url: addr,
		gen: gen,
		ctx: metadata.AppendToOutgoingContext(context.Background(), "X-Real-IP", realIP),
	}
}

func (a *Agent) PushBulk() error {
	a.wg.Add(1)
	defer a.wg.Done()
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan *webclient.Metrics)

	// подготовка данных
	gauge, counter, err := a.gen.Collect()
	if err != nil {
		logger.Error(fmt.Sprintf("fail collect metrics: %s", err.Error()))
	}
	var items []*webclient.Metrics

	go webclient.PrepareDataToSend(gauge, counter, ch, cancel)
	for {
		select {
		case item := <-ch:
			items = append(items, item)
		case <-ctx.Done():
			conn, err := grpc.NewClient(
				a.addr,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				return err
			}
			defer conn.Close()

			client := pb.NewMetricsServiceClient(conn)

			var pbReqMet pb.PushBulkRequest
			for i := range items {
				var value float64
				if (items[i].Value == nil || *items[i].Value == 0) && items[i].Delta != nil {
					value = float64(*items[i].Delta)
				} else if items[i].Value == nil {
					value = *items[i].Value
				}
				pbReqMet.Metrics = append(pbReqMet.Metrics, &pb.Metric{
					Name:  items[i].ID,
					Type:  items[i].MType,
					Value: value,
				})
			}

			pushBulkResp, err := client.PushBulk(a.ctx, &pbReqMet)
			if err != nil {
				return fmt.Errorf("message: %s, is success:  %t", pushBulkResp.Message, pushBulkResp.Success)
			}
			return nil
		}
	}
}

func (a *Agent) PushStream() error {
	a.wg.Add(1)
	defer a.wg.Done()
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan *webclient.Metrics)

	// подготовка данных
	gauge, counter, err := a.gen.Collect()
	if err != nil {
		logger.Error(fmt.Sprintf("fail collect metrics: %s", err.Error()))
	}
	var items []*webclient.Metrics

	go webclient.PrepareDataToSend(gauge, counter, ch, cancel)
	for {
		select {
		case item := <-ch:
			items = append(items, item)
		case <-ctx.Done():
			conn, err := grpc.NewClient(
				a.addr,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				return err
			}
			defer conn.Close()

			client := pb.NewMetricsServiceClient(conn)

			var pbReqMet []*pb.Metric
			for i := range items {
				var value float64
				if (items[i].Value == nil || *items[i].Value == 0) && items[i].Delta != nil {
					value = float64(*items[i].Delta)
				} else if items[i].Value == nil {
					value = *items[i].Value
				}
				pbReqMet = append(pbReqMet, &pb.Metric{
					Name:  items[i].ID,
					Type:  items[i].MType,
					Value: value,
				})
			}

			stream, err := client.PushStream(a.ctx)
			if err != nil {
				return err
			}

			for i := range pbReqMet {
				err := stream.Send(pbReqMet[i])
				if err != nil {
					return err
				}
			}

			pushStreamResp, err := stream.CloseAndRecv()
			if err != nil {
				return err
			}

			logger.Info(fmt.Sprintf("grpc stream success %t message: %s", pushStreamResp.Success, pushStreamResp.Message))

			return nil
		}
	}
}
