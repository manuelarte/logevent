package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net"
	"net/http"
	"time"

	"github.com/manuelarte/logevent"
	"github.com/manuelarte/logevent/examples"
	logeventmiddleware "github.com/manuelarte/logevent/mw/http"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func run() error {
	http.Handle("/process", logeventmiddleware.AddLogEventMiddleware(examples.MyLogEvent{}, slog.Default())(http.HandlerFunc(processHandler)))

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
			_, err := httpClient.Get("http://" + listener.Addr().String() + "/process")
			if err != nil {
				return err
			}
		}

	}
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	// Simulate some processing time.
	time.Sleep(getRandomDuration())
	elapsed := time.Since(now)
	err := logevent.UpdateLogEvent(r.Context(), func(t *examples.MyLogEvent) {
		t.Elapsed = elapsed
	})
	if err != nil {
		panic(err)
	}
	w.WriteHeader(200)
	_, _ = w.Write([]byte("OK"))
}

func getRandomDuration() time.Duration {
	mean := 100.0
	std := 20.0
	val := mean + std*rand.NormFloat64()

	return time.Duration(val) * time.Millisecond
}
