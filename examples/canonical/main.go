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
	i := -1
	for {
		i++

		if err := handle(i); err != nil {
			return err
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func handle(i int) error {
	// Simulate some background work that needs logging

	// Step 1: Add the log event to the context
	ctx := logevent.AddLogEventToContext[*slog.Logger](context.Background(), examples.MyLogEvent{})

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
	if rand.IntN(10) < 4 { // 40% chance of simulated error
		errProcessing := fmt.Errorf("simulated processing error")
		if updateErr := logevent.UpdateLogEvent(ctx, func(e *examples.MyLogEvent) {
			e.Err = errProcessing
		}); updateErr != nil { // should never happen
			return updateErr
		}
	}
	if updateErr := logevent.UpdateLogEvent(ctx, func(e *examples.MyLogEvent) {
		e.Elapsed = time.Since(start)
		e.TaskID = fmt.Sprintf("%d", i)
	}); updateErr != nil {
		return updateErr
	}
	return nil
}

func getRandomDuration() time.Duration {
	mean := 100.0
	std := 20.0
	val := mean + std*rand.NormFloat64()

	return time.Duration(val) * time.Millisecond
}
