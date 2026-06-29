package logevent

import (
	"context"
)

type (
	LogEvent[L any] interface {
		Log(ctx context.Context, li L)
	}
)
