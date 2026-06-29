package logevent

import (
	"errors"
	"fmt"
	"reflect"
)

var _ error = new(DifferentLogEventTypeError)

// ErrLogEventNotInitialized error returned when adding context to a log event
// but the log event was not initialized.
var ErrLogEventNotInitialized = errors.New("LogEvent not initialized")

type (
	DifferentLogEventTypeError struct {
		previousType reflect.Type
		currentType  reflect.Type
	}
)

func NewDifferentLogEventTypeError(previousType, currentType reflect.Type) *DifferentLogEventTypeError {
	return &DifferentLogEventTypeError{
		previousType: previousType,
		currentType:  currentType,
	}
}

func (l DifferentLogEventTypeError) Error() string {
	return fmt.Sprintf("LogEvent type mismatch: previous type %v, current type %v", l.previousType, l.currentType)
}
