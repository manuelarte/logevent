package logevent

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHandleWithLogEventLogsAfterHandler(t *testing.T) {
	t.Parallel()

	got := make([]string, 0)
	li := testLogInterface{entries: &got}
	le := testLogEvent{events: &got}

	ctx := t.Context()
	h := func(ctx context.Context) {
		got = append(got, "handler")

		if err := UpdateLogEvent(ctx, func(le *testLogEvent) {
			le.value = "updated"
		}); err != nil {
			t.Fatalf("UpdateLogEvent() error = %v", err)
		}
	}
	HandleWithLogEvent(ctx, le, li, h)

	want := []string{"handler", "log:updated", "info:updated"}
	if !cmp.Equal(got, want) {
		t.Fatalf("events = %v, want %v", got, want)
	}
}

func TestHandleWithLogEventLogsAfterPanic(t *testing.T) {
	t.Parallel()

	events := make([]string, 0)
	li := testLogInterface{entries: &events}
	le := testLogEvent{events: &events}
	panicValue := "boom"

	ctx := t.Context()
	h := func(ctx context.Context) {
		events = append(events, "handler")

		if err := UpdateLogEvent(ctx, func(le *testLogEvent) {
			le.value = "panic-update"
		}); err != nil {
			t.Fatalf("UpdateLogEvent() error = %v", err)
		}

		panic(panicValue)
	}

	defer func() {
		if recovered := recover(); recovered != panicValue {
			t.Fatalf("recover() = %v, want %v", recovered, panicValue)
		}

		want := []string{"handler", "log:panic-update", "info:panic-update"}
		if !cmp.Equal(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()

	HandleWithLogEvent(ctx, le, li, h)
}

func TestHandleWithLogEventLogsUpdateOnDefer(t *testing.T) {
	t.Parallel()

	got := make([]string, 0)
	li := testLogInterface{entries: &got}
	le := testLogEvent{events: &got}

	ctx := t.Context()
	wg := sync.WaitGroup{}
	wg.Add(1)

	h := func(ctx context.Context) {
		got = append(got, "handler")

		go func() {
			defer wg.Done()

			err := UpdateLogEvent(ctx, func(le *testLogEvent) {
				le.value = "updated"
			})
			if !errors.Is(err, ErrUpdatingLoggedEvent) {
				t.Errorf("UpdateLogEvent() error = %v, want %v", err, ErrUpdatingLoggedEvent)
			}
		}()
	}
	HandleWithLogEvent(ctx, le, li, h)
	wg.Wait()
}

func TestHandleWithLogEventLogsOnlyOnce(t *testing.T) {
	t.Parallel()

	got := make([]string, 0)
	li := testLogInterface{entries: &got}
	le := testLogEvent{events: &got}

	ctx := t.Context()
	_, deferFunc := AddLogEventToContext[testLogInterface, testLogEvent, *testLogEvent](ctx, le)

	const numCalls = 10

	wg := sync.WaitGroup{}
	wg.Add(numCalls)

	for range numCalls {
		deferFunc(li)
		wg.Done()
	}

	wg.Wait()

	want := []string{"log:", "info:"}
	if !cmp.Equal(got, want) {
		t.Fatalf("events = %v, want %v (Log should have been called exactly once)", got, want)
	}
}

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
