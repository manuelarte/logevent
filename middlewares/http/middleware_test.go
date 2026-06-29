package http

import (
	"context"
	"errors"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/manuelarte/logevent"
	"github.com/manuelarte/logevent/middlewares"
)

type testLogInterface struct {
	entries *[]string
}

func (l testLogInterface) Info(msg string, _ ...any) {
	*l.entries = append(*l.entries, "info:"+msg)
}

type testLogEvent struct {
	events *[]string
	value  string
}

func (e *testLogEvent) Log(_ context.Context, li testLogInterface) {
	*e.events = append(*e.events, "log:"+e.value)
	li.Info(e.value)
}

func TestAddLogEventMiddlewareLogsAfterHandler(t *testing.T) {
	got := make([]string, 0)
	li := testLogInterface{entries: &got}
	le := testLogEvent{events: &got}

	middleware := AddLogEventMiddleware(le, li)
	handler := middleware(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		got = append(got, "handler")

		err := middlewares.UpdateLogEvent(r.Context(), func(le *testLogEvent) {
			le.value = "updated"
		})
		if err != nil {
			t.Fatalf("UpdateLogEvent() error = %v", err)
		}

		w.WriteHeader(nethttp.StatusAccepted)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), nethttp.MethodGet, "/", nil))

	if rec.Code != nethttp.StatusAccepted {
		t.Fatalf("status code = %d, want %d", rec.Code, nethttp.StatusAccepted)
	}

	want := []string{"handler", "log:updated", "info:updated"}
	if !cmp.Equal(got, want) {
		t.Fatalf("got = %v, want %v", got, want)
	}
}

func TestAddLogEventMiddlewareLogsAfterPanic(t *testing.T) {
	events := make([]string, 0)
	li := testLogInterface{entries: &events}
	le := testLogEvent{events: &events}
	panicValue := "boom"

	middleware := AddLogEventMiddleware(le, li)
	handler := middleware(nethttp.HandlerFunc(func(_ nethttp.ResponseWriter, r *nethttp.Request) {
		events = append(events, "handler")

		err := middlewares.UpdateLogEvent(r.Context(), func(le *testLogEvent) {
			le.value = "panic-update"
		})
		if err != nil {
			t.Fatalf("UpdateLogEvent() error = %v", err)
		}

		panic(panicValue)
	}))

	defer func() {
		if recovered := recover(); recovered != panicValue {
			t.Fatalf("recover() = %v, want %v", recovered, panicValue)
		}

		want := []string{"handler", "log:panic-update", "info:panic-update"}
		if !cmp.Equal(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), nethttp.MethodGet, "/", nil))
}

func TestUpdateLogEventReturnsErrorWithoutMiddleware(t *testing.T) {
	err := middlewares.UpdateLogEvent(context.Background(), func(*testLogEvent) {})

	if !errors.Is(err, logevent.ErrLogEventNotInitialized) {
		t.Fatalf("UpdateLogEvent() error = %v, want %v", err, logevent.ErrLogEventNotInitialized)
	}
}
