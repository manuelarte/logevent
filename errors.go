package logevent

import (
	"errors"
	"fmt"
	"reflect"
)

var _ error = new(DifferentLogEventTypeError)

var (
	// ErrLogEventNotInitialized error returned when adding context to a log event.
	// but the log event was not initialized.
	ErrLogEventNotInitialized = errors.New("LogEvent not initialized")

	// ErrUpdatingLoggedEvent error returned when trying to update a log event that has already been logged.
	ErrUpdatingLoggedEvent = errors.New("LogEvent already logged")
)

type (
	// DifferentLogEventTypeError is returned when the log event type is different from the previous one.
	DifferentLogEventTypeError struct {
		previousType reflect.Type
		currentType  reflect.Type
	}
)

// Error implements the error interface.
func (l DifferentLogEventTypeError) Error() string {
	return fmt.Sprintf("LogEvent type mismatch: previous type %v, current type %v", l.previousType, l.currentType)
}
