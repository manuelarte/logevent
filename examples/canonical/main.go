package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/manuelarte/logevent"
	"github.com/manuelarte/logevent/examples"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func run() error {
	// Simulate some background work that needs logging
	// without using HTTP middleware or gRPC interceptors
	for i := 0; i < 10; i++ {
		ctx := context.Background()

		// Step 1: Add the log event to the context
		ctx = logevent.AddLogEventToContext[*slog.Logger](ctx, examples.MyLogEvent{})

		// Step 2: Get the defer function that will log the event
		deferFunc, err := logevent.LogItFunc[*slog.Logger, examples.MyLogEvent, *examples.MyLogEvent](ctx, slog.Default())
		if err != nil {
			return err
		}
		// Step 3: Defer the logging call
		defer deferFunc()

		// Step 4: Process some work and update the log event as needed
		start := time.Now()
		time.Sleep(getRandomDuration())
		elapsed := time.Since(start)

		err = logevent.UpdateLogEvent(ctx, func(e *examples.MyLogEvent) {
			e.Elapsed = elapsed
			e.ProcessID = fmt.Sprintf("%d", i)
		})
		if err != nil {
			return err
		}

		// If something goes wrong, update the error in the log event
		if rand.IntN(10) < 4 { // 20% chance of simulated error
			err := fmt.Errorf("simulated processing error")
			updateErr := logevent.UpdateLogEvent(ctx, func(e *examples.MyLogEvent) {
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

func getRandomDuration() time.Duration {
	mean := 100.0
	std := 20.0
	val := mean + std*rand.NormFloat64()

	return time.Duration(val) * time.Millisecond
}
