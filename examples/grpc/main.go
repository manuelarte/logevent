package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/manuelarte/logevent/middlewares"
	logeventgrpc "github.com/manuelarte/logevent/middlewares/grpc"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func run() error {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return fmt.Errorf("error creating listener: %w", err)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			logeventgrpc.UnaryServerInterceptor(myLogEvent{}, slog.Default()),
		),
	)
	healthgrpc.RegisterHealthServer(server, new(healthServer))

	slog.InfoContext(context.Background(), "gRPC server listening on", slog.String("address", listener.Addr().String()))
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Stop()

	conn, err := grpc.NewClient(
		listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("error creating client connection: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	client := healthgrpc.NewHealthClient(conn)
	for {
		time.Sleep(500 * time.Millisecond)
		_, errCheck := client.Check(context.Background(), &healthgrpc.HealthCheckRequest{Service: "events"})
		if errCheck != nil {
			return fmt.Errorf("error calling health check: %w", errCheck)
		}
	}
}

type healthServer struct {
	healthgrpc.UnimplementedHealthServer
}

func (s healthServer) Check(ctx context.Context, req *healthgrpc.HealthCheckRequest) (*healthgrpc.HealthCheckResponse, error) {
	start := time.Now()

	// Simulate processing time.
	time.Sleep(getRandomDuration())

	updateErr := middlewares.UpdateLogEvent(ctx, func(e *myLogEvent) {
		e.Service = req.GetService()
		e.Elapsed = time.Since(start)
	})
	if updateErr != nil {
		return nil, updateErr
	}

	return &healthgrpc.HealthCheckResponse{Status: healthgrpc.HealthCheckResponse_SERVING}, nil
}

func (s healthServer) Watch(*healthgrpc.HealthCheckRequest, healthgrpc.Health_WatchServer) error {
	return fmt.Errorf("watch is not implemented in this example")
}

type myLogEvent struct {
	Service string
	Elapsed time.Duration
}

func (e myLogEvent) Log(ctx context.Context, li *slog.Logger) {
	li.InfoContext(
		ctx,
		"rpc handled",
		slog.String("service", e.Service),
		slog.Int64("elapsed_ms", e.Elapsed.Milliseconds()),
	)
}

func getRandomDuration() time.Duration {
	mean := 100.0
	std := 20.0
	val := mean + std*rand.NormFloat64()

	return time.Duration(val) * time.Millisecond
}
