package grpcserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	candlestickpb "github.com/shubie/trading/api/protos/candlestick"
	"github.com/shubie/trading/internal/aggregator"
)

type Server struct {
	candlestickpb.UnimplementedCandlestickServiceServer
	candlestickpb.UnimplementedHealthCheckServiceServer
	agg          *aggregator.Aggregator
	grpcServer   *grpc.Server
	port         int
	healthMu     sync.RWMutex
	lastDataTime time.Time
	startupTime  time.Time
}

func NewServer(port int, agg *aggregator.Aggregator) *Server {
	return &Server{
		port:         port,
		agg:          agg,
		startupTime:  time.Now(),
		lastDataTime: time.Now(),
	}
}

func (s *Server) Start() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s.grpcServer = grpc.NewServer()
	candlestickpb.RegisterCandlestickServiceServer(s.grpcServer, s)
	candlestickpb.RegisterHealthCheckServiceServer(s.grpcServer, s)

	log.Printf("gRPC server starting on port %d", s.port)
	if err := s.grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *Server) StreamCandlesticks(req *candlestickpb.StreamRequest, stream candlestickpb.CandlestickService_StreamCandlesticksServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil

		case <-time.After(1 * time.Second):
			for _, symbol := range req.Symbols {
				candle := s.agg.GetCurrentCandle(symbol)
				if candle == nil {
					continue
				}

				err := stream.Send(&candlestickpb.Candlestick{
					Symbol:    candle.Symbol,
					Open:      candle.Open,
					High:      candle.High,
					Low:       candle.Low,
					Close:     candle.Close,
					Volume:    candle.Volume,
					StartTime: candle.StartTime.UnixMilli(),
					EndTime:   candle.EndTime.UnixMilli(),
					IsFinal:   candle.Finalized,
				})

				log.Printf("executed GRPC")

				if err != nil {
					return status.Errorf(codes.Aborted, "stream error: %v", err)
				}

				s.updateDataTime(time.Now())
			}
		}
	}
}

func (s *Server) Health(ctx context.Context, req *candlestickpb.HealthRequest) (*candlestickpb.HealthResponse, error) {
	s.healthMu.RLock()
	defer s.healthMu.RUnlock()

	status := candlestickpb.HealthResponse_HEALTHY
	message := "Service operational"
	lastData := s.agg.GetLastDataTime()

	if time.Since(lastData) > 5*time.Minute {
		status = candlestickpb.HealthResponse_UNHEALTHY
		message = "No data received in last 5 minutes"
	}

	return &candlestickpb.HealthResponse{
		Status:     status,
		Message:    message,
		LastDataMs: lastData.UnixMilli(),
	}, nil
}

func (s *Server) updateDataTime(t time.Time) {
	s.healthMu.Lock()
	defer s.healthMu.Unlock()
	if t.After(s.lastDataTime) {
		s.lastDataTime = t
	}
}

func (s *Server) Stop() {
	log.Println("Initiating gRPC server shutdown...")
	s.grpcServer.GracefulStop()
	log.Println("gRPC server stopped")
}
