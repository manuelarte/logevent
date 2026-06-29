package internal

import (
	"context"
	"sync"

	"github.com/manuelarte/logevent"
)

type (
	// LogEventKey represents the key to be used to store the LogEvent in the context.
	// It is a generic type that ensures type-safety when storing and retrieving log events.
	LogEventKey[L any, T any, PT PtrLogEvent[L, T]] struct{}

	// PtrLogEvent helps to make that only a pointer can be passed to the middleware.
	// It is a constraint that ensures PT is a pointer to type T and implements logevent.LogEvent.
	PtrLogEvent[L any, T any] interface {
		*T
		logevent.LogEvent[L]
	}

	// WrapperLogEvent gives some concurrency support to a logevent.LogEvent.
	// It ensures the Log method is called only once (via sync.Once) and protects
	// concurrent updates to the underlying log event with a mutex.
	WrapperLogEvent[L any, T any, PT PtrLogEvent[L, T]] struct {
		once sync.Once
		mu   sync.RWMutex
		le   PT
	}
)

func NewWrapperLogEvent[L, T any, PT PtrLogEvent[L, T]](le PT) *WrapperLogEvent[L, T, PT] {
	return &WrapperLogEvent[L, T, PT]{
		le: le,
	}
}

// Update add the context to the inner logevent.LogEvent.
func (w *WrapperLogEvent[L, T, PT]) Update(f func(t PT)) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	f(w.le)

	return nil
}

// Log call the inner logevent.LogEvent to log.
func (w *WrapperLogEvent[L, T, PT]) Log(ctx context.Context, li L) {
	w.once.Do(func() {
		w.le.Log(ctx, li)
	})
}

// UpdateLogEvent updates the log event with the function f.
func UpdateLogEvent[L, T any, PT PtrLogEvent[L, T]](ctx context.Context, f func(t PT)) error {
	v, ok := ctx.Value(LogEventKey[L, T, PT]{}).(*WrapperLogEvent[L, T, PT])
	if !ok {
		return logevent.ErrLogEventNotInitialized
	}

	return v.Update(f)
}
