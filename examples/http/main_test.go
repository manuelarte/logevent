package main

import (
	"log/slog"
	"net/http"
	"testing"

	"github.com/manuelarte/logevent/middlewares"
)

// TestEventHandler shows how to add the log event in case the handler needs to be unit tested.
func TestEventHandler(t *testing.T) {
	w := myResponseWriter{}
	ctx := middlewares.AddLogEventToContext[*slog.Logger](t.Context(), myLogEvent{})
	r, err := http.NewRequestWithContext(ctx, "GET", "/events", nil)
	if err != nil {
		t.Fatal(err)
	}
	eventHandler(w, r)
}

type myResponseWriter struct{}

func (m myResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (m myResponseWriter) Write(bytes []byte) (int, error) {
	return len(bytes), nil
}

func (m myResponseWriter) WriteHeader(statusCode int) {
}
