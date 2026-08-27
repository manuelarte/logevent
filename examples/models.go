package examples

import (
	"context"
	"log/slog"
	"time"

	"github.com/manuelarte/logevent"
)

var _ = new(logevent.LogEvent[*slog.Logger])

type MyLogEvent struct {
	ProcessID string
	Elapsed   time.Duration
	Err       error
}

func (e MyLogEvent) Log(ctx context.Context, logger *slog.Logger) {
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
