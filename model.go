package logevent

import (
	"context"
)

type (
	LogEvent interface {
		Log(ctx context.Context, li LogInterface)
	}

	LogInterface interface {
		Info(msg string, args ...any)
		Warn(msg string, args ...any)
		Debug(msg string, args ...any)
		Error(msg string, args ...any)
	}
)
