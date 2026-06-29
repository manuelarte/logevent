package internal

import (
	"context"
	"sync"

	"codeberg.org/manuelarte/logevent"
)

type (
	// LogEventKey represents the key to be used to store the LogEvent in the context.
	// It is a generic type that ensures type-safety when storing and retrieving log events.
	LogEventKey[T any, PT PtrLogEvent[T]] struct{}

	// PtrLogEvent helps to make that only a pointer can be passed to the middleware.
	// It is a constraint that ensures PT is a pointer to type T and implements logevent.LogEvent.
	PtrLogEvent[T any] interface {
		*T
		logevent.LogEvent
	}

	// WrapperLogEvent gives some concurrency support to a logevent.LogEvent.
	// It ensures the Log method is called only once (via sync.Once) and protects
	// concurrent updates to the underlying log event with a mutex.
	WrapperLogEvent[T any, PT PtrLogEvent[T]] struct {
		once sync.Once
		mu   sync.RWMutex
		le   PT
	}
)

func NewWrapperLogEvent[T any, PT PtrLogEvent[T]](le PT) *WrapperLogEvent[T, PT] {
	return &WrapperLogEvent[T, PT]{
		le: le,
	}
}

// Update add the context to the inner logevent.LogEvent.
func (w *WrapperLogEvent[T, PT]) Update(f func(t PT)) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	f(w.le)

	return nil
}

// Log call the inner logevent.LogEvent to log.
func (w *WrapperLogEvent[T, PT]) Log(ctx context.Context, li logevent.LogInterface) {
	w.once.Do(func() {
		w.le.Log(ctx, li)
	})
}

// UpdateLogEvent updates the log event with the function f.
func UpdateLogEvent[T any, PT PtrLogEvent[T]](ctx context.Context, f func(t PT)) error {
	v, ok := ctx.Value(LogEventKey[T, PT]{}).(*WrapperLogEvent[T, PT])
	if !ok {
		return logevent.ErrLogEventNotInitialized
	}

	return v.Update(f)
}
