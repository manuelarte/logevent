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
	http.Handle("/tasks", logeventmiddleware.AddLogEventMiddleware(examples.MyLogEvent{}, slog.Default())(http.HandlerFunc(processHandler)))

	listener, errPort := net.Listen("tcp", ":0")
	if errPort != nil {
		return fmt.Errorf("error creating listener: %w", errPort)
	}

	slog.InfoContext(context.Background(), "HTTP server listening on", slog.String("address", listener.Addr().String()))
	go func() {
		_ = http.Serve(listener, nil)
	}()

	httpClient := http.DefaultClient
	i := -1
	for {
		delay := time.Tick(500 * time.Millisecond)
		select {
		case <-delay:
			i++
			url := fmt.Sprintf("http://%s/tasks?task_id=%d", listener.Addr().String(), i)
			_, err := httpClient.Get(url)
			if err != nil {
				return err
			}
		}

	}
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	// Read process_id from request query and simulate processing time.
	taskID := r.URL.Query().Get("task_id")
	// Simulate some processing time.
	time.Sleep(getRandomDuration())
	// If something goes wrong, update the error in the log event
	if rand.IntN(10) < 4 { // 40% chance of simulated error
		err := fmt.Errorf("simulated processing error")
		if updateErr := logevent.UpdateLogEvent(r.Context(), func(e *examples.MyLogEvent) {
			e.Err = err
		}); updateErr != nil {
			panic(updateErr) // should never happen
		}
	}
	if updateErr := logevent.UpdateLogEvent(r.Context(), func(t *examples.MyLogEvent) {
		t.TaskID = taskID
		t.Elapsed = time.Since(now)
	}); updateErr != nil {
		panic(updateErr) // should never happen
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
