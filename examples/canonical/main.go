package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/manuelarte/logevent/mw"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func run() error {
	// Simulate some background work that needs logging
	// without using HTTP middleware or gRPC interceptors
	for range 10 {
		ctx := context.Background()

		// Step 1: Add the log event to the context
		ctx = mw.AddLogEventToContext(ctx, myLogEvent{})

		// Step 2: Get the defer function that will log the event
		deferFunc, err := mw.LogEventFunc[*slog.Logger, myLogEvent, *myLogEvent](ctx, slog.Default())
		if err != nil {
			return err
		}
		// Step 3: Defer the logging call
		defer deferFunc()

		// Step 4: Process some work and update the log event as needed
		start := time.Now()
		time.Sleep(getRandomDuration())
		elapsed := time.Since(start)

		err = mw.UpdateLogEvent(ctx, func(e *myLogEvent) {
			e.Elapsed = elapsed
			e.ProcessID = "task-123"
		})
		if err != nil {
			return err
		}

		// If something goes wrong, update the error in the log event
		if rand.IntN(10) < 2 { // 20% chance of simulated error
			err := fmt.Errorf("simulated processing error")
			updateErr := mw.UpdateLogEvent(ctx, func(e *myLogEvent) {
				e.Err = err
			})
			if updateErr != nil {
				return updateErr
			}
		}

		// The defer will automatically call Log() on the event after this function returns
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

type myLogEvent struct {
	ProcessID string
	Elapsed   time.Duration
	Err       error
}

func (e myLogEvent) Log(ctx context.Context, logger *slog.Logger) {
	if e.Err != nil {
		logger.ErrorContext(
			ctx,
			"Task processing failed",
			slog.String("process_id", e.ProcessID),
			slog.Int64("elapsed_ms", e.Elapsed.Milliseconds()),
			slog.Any("error", e.Err),
		)
		return
	}

	logger.InfoContext(
		ctx,
		"Task processed successfully",
		slog.String("process_id", e.ProcessID),
		slog.Int64("elapsed_ms", e.Elapsed.Milliseconds()),
	)
}

func getRandomDuration() time.Duration {
	mean := 100.0
	std := 20.0
	val := mean + std*rand.NormFloat64()

	return time.Duration(val) * time.Millisecond
}
