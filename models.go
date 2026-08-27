package logevent

import (
	"context"
	"sync"
)

type (
	// Logger is the interface that represents a logger.
	Logger any

	// LogEvent is the interface that wraps how to Log the event.
	LogEvent[L Logger] interface {
		// Log the event.
		Log(ctx context.Context, li L)
	}

	// PtrLogEvent helps to make that only a pointer can be passed to the middleware.
	// It is a constraint that ensures PT is a pointer to type T and implements logevent.LogEvent.
	PtrLogEvent[L Logger, T any] interface {
		*T
		LogEvent[L]
	}

	// wrapperLogEvent gives some concurrency support to a logevent.LogEvent.
	// It ensures the Log method is called only once (via [sync.Once]) and protects
	// concurrent updates to the underlying log event with a mutex.
	wrapperLogEvent[L Logger, T any, PT PtrLogEvent[L, T]] struct {
		once sync.Once
		mu   sync.RWMutex
		le   PT
	}
)

// Update add the context to the inner logevent.LogEvent.
func (w *wrapperLogEvent[L, T, PT]) Update(f func(t PT)) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	f(w.le)

	return nil
}

// Log call the inner logevent.LogEvent to log.
func (w *wrapperLogEvent[L, T, PT]) Log(ctx context.Context, li L) {
	w.once.Do(func() {
		w.le.Log(ctx, li)
	})
}
