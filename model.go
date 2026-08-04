package logevent

import (
	"context"
)

type (
	// Logger is the interface that represents a logger.
	Logger any

	// LogEvent is the interface that wraps how to Log the event.
	LogEvent[L Logger] interface {
		// Log the event..
		Log(ctx context.Context, li L)
	}
)
