// Package internal contains the interfaces used by logevent.
package internal

import (
	"context"
	"sync"

	"github.com/manuelarte/logevent"
)

type (
	// PtrLogEvent helps to make that only a pointer can be passed to the middleware.
	// It is a constraint that ensures PT is a pointer to type T and implements logevent.LogEvent.
	PtrLogEvent[L logevent.Logger, T any] interface {
		*T
		logevent.LogEvent[L]
	}

	// WrapperLogEvent gives some concurrency support to a logevent.LogEvent.
	// It ensures the Log method is called only once (via [sync.Once]) and protects
	// concurrent updates to the underlying log event with a mutex.
	WrapperLogEvent[L logevent.Logger, T any, PT PtrLogEvent[L, T]] struct {
		once sync.Once
		mu   sync.RWMutex
		le   PT
	}
)

// NewWrapperLogEvent creates a new WrapperLogEvent.
func NewWrapperLogEvent[L logevent.Logger, T any, PT PtrLogEvent[L, T]](le PT) *WrapperLogEvent[L, T, PT] {
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
