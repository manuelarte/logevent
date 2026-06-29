package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net"
	"net/http"
	"time"

	"github.com/manuelarte/logevent/middlewares"
	logeventmiddleware "github.com/manuelarte/logevent/middlewares/http"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func run() error {
	http.Handle("/events", logeventmiddleware.AddLogEventMiddleware(myLogEvent{}, slog.Default())(http.HandlerFunc(eventHandler)))

	listener, errPort := net.Listen("tcp", ":0")
	if errPort != nil {
		return fmt.Errorf("error creating listener: %w", errPort)
	}

	slog.InfoContext(context.Background(), "HTTP server listening on", slog.String("address", listener.Addr().String()))
	go func() {
		_ = http.Serve(listener, nil)
	}()

	httpClient := http.DefaultClient
	for {
		delay := time.Tick(500 * time.Millisecond)
		select {
		case <-delay:
			_, err := httpClient.Get("http://" + listener.Addr().String() + "/events")
			if err != nil {
				return err
			}
		}

	}
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	// Simulate some processing time.
	time.Sleep(getRandomDuration())
	elapsed := time.Since(now)
	err := middlewares.UpdateLogEvent(r.Context(), func(t *myLogEvent) {
		t.Elapsed = elapsed
	})
	if err != nil {
		panic(err)
	}
	w.WriteHeader(200)
	_, _ = w.Write([]byte("OK"))
}

type myLogEvent struct {
	Elapsed time.Duration
}

func (e myLogEvent) Log(ctx context.Context, logger *slog.Logger) {
	logger.InfoContext(ctx, "Event handled", slog.Int64("elapsed_ms", e.Elapsed.Milliseconds()))
}

func getRandomDuration() time.Duration {
	mean := 100.0
	std := 20.0
	val := mean + std*rand.NormFloat64()

	return time.Duration(val) * time.Millisecond
}
