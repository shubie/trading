package grpcserver_test

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/shubie/trading/internal/aggregator"
	"github.com/shubie/trading/internal/grpcserver"

	candlestickpb "github.com/shubie/trading/api/protos/candlestick"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func startTestGRPCServer(t *testing.T, agg *aggregator.Aggregator) *grpcserver.Server {
	lis = bufconn.Listen(bufSize)
	server := grpc.NewServer()

	s := grpcserver.NewServer(0, agg)
	candlestickpb.RegisterCandlestickServiceServer(server, s)
	candlestickpb.RegisterHealthCheckServiceServer(server, s)

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
	return s
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestHealth_Healthy(t *testing.T) {
	agg := aggregator.NewAggregator()
	// I am seeting the last data time to healthy which recent time
	agg.SetLastDataTimeForTesting(time.Now())

	_ = startTestGRPCServer(t, agg)

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := candlestickpb.NewHealthCheckServiceClient(conn)

	resp, err := client.Health(ctx, &candlestickpb.HealthRequest{})
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}

	if resp.Status != candlestickpb.HealthResponse_HEALTHY {
		t.Errorf("Expected HEALTHY, got %v", resp.Status)
	}
}

func TestHealth_Unhealthy(t *testing.T) {
	agg := aggregator.NewAggregator()
	//  I am seeting the last data time to unhealthy which is the old time
	agg.SetLastDataTimeForTesting(time.Now().Add(-10 * time.Minute))

	_ = startTestGRPCServer(t, agg)

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := candlestickpb.NewHealthCheckServiceClient(conn)

	resp, err := client.Health(ctx, &candlestickpb.HealthRequest{})
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}

	if resp.Status != candlestickpb.HealthResponse_UNHEALTHY {
		t.Errorf("Expected UNHEALTHY, got %v", resp.Status)
	}
}
