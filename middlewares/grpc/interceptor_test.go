package grpc

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc"

	"codeberg.org/manuelarte/logevent"
	"codeberg.org/manuelarte/logevent/middlewares"
)

type testLogInterface struct {
	entries *[]string
}

func (l testLogInterface) Info(msg string, _ ...any) {
	*l.entries = append(*l.entries, "info:"+msg)
}

func (l testLogInterface) Warn(msg string, _ ...any) {
	*l.entries = append(*l.entries, "warn:"+msg)
}

func (l testLogInterface) Debug(msg string, _ ...any) {
	*l.entries = append(*l.entries, "debug:"+msg)
}

func (l testLogInterface) Error(msg string, _ ...any) {
	*l.entries = append(*l.entries, "error:"+msg)
}

type testLogEvent struct {
	events *[]string
	value  string
}

func (e *testLogEvent) Log(_ context.Context, li logevent.LogInterface) {
	*e.events = append(*e.events, "log:"+e.value)
	li.Info(e.value)
}

func TestUnaryServerInterceptorLogsAfterHandler(t *testing.T) {
	got := make([]string, 0)
	li := testLogInterface{entries: &got}

	interceptor := UnaryServerInterceptor(testLogEvent{events: &got}, func(context.Context) logevent.LogInterface {
		got = append(got, "factory")

		return li
	})

	handler := func(ctx context.Context, req any) (any, error) {
		got = append(got, "handler")

		err := middlewares.UpdateLogEvent(ctx, func(e *testLogEvent) {
			e.value = "updated"
		})
		if err != nil {
			t.Fatalf("UpdateLogEvent() error = %v", err)
		}

		return "response", nil
	}

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, handler)
	if err != nil {
		t.Fatalf("interceptor() error = %v", err)
	}

	want := []string{"handler", "factory", "log:updated", "info:updated"}
	if len(got) != len(want) {
		t.Fatalf("got %d entries, want %d: %v", len(got), len(want), got)
	}

	for i, entry := range got {
		if entry != want[i] {
			t.Fatalf("got entry[%d]=%s, want %s", i, entry, want[i])
		}
	}
}

func TestUnaryServerInterceptorLogsAfterHandlerError(t *testing.T) {
	events := make([]string, 0)
	li := testLogInterface{entries: &events}
	expectedErr := errors.New("handler error")

	interceptor := UnaryServerInterceptor(testLogEvent{events: &events}, func(context.Context) logevent.LogInterface {
		events = append(events, "factory")

		return li
	})

	handler := func(ctx context.Context, req any) (any, error) {
		events = append(events, "handler")

		err := middlewares.UpdateLogEvent(ctx, func(e *testLogEvent) {
			e.value = "error-update"
		})
		if err != nil {
			t.Fatalf("UpdateLogEvent() error = %v", err)
		}

		return nil, expectedErr
	}

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, handler)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("interceptor() error = %v, want %v", err, expectedErr)
	}

	want := []string{"handler", "factory", "log:error-update", "info:error-update"}
	if len(events) != len(want) {
		t.Fatalf("got %d entries, want %d: %v", len(events), len(want), events)
	}

	for i, entry := range events {
		if entry != want[i] {
			t.Fatalf("got entry[%d]=%s, want %s", i, entry, want[i])
		}
	}
}

func TestUpdateLogEventReturnsErrorWithoutInterceptor(t *testing.T) {
	err := middlewares.UpdateLogEvent(context.Background(), func(*testLogEvent) {})

	if !errors.Is(err, logevent.ErrLogEventNotInitialized) {
		t.Fatalf("UpdateLogEvent() error = %v, want %v", err, logevent.ErrLogEventNotInitialized)
	}
}

func TestUnaryServerInterceptorEachRequestGetsFreshInstance(t *testing.T) {
	entries := make([]string, 0)
	li := testLogInterface{entries: &entries}
	pointers := make([]string, 0, 2)

	interceptor := UnaryServerInterceptor(testLogEvent{events: &entries}, func(context.Context) logevent.LogInterface {
		return li
	})

	handler := func(ctx context.Context, req any) (any, error) {
		err := middlewares.UpdateLogEvent(ctx, func(e *testLogEvent) {
			pointers = append(pointers, fmt.Sprintf("%p", e))
		})
		if err != nil {
			t.Fatalf("UpdateLogEvent() error = %v", err)
		}

		return "response", nil
	}

	interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test1"}, handler)
	interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test2"}, handler)

	if len(pointers) != 2 {
		t.Fatalf("captured pointers = %d, want 2", len(pointers))
	}

	if pointers[0] == pointers[1] {
		t.Fatalf("captured pointers should differ (fresh instance per request): first=%s second=%s", pointers[0], pointers[1])
	}
}
