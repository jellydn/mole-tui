package cleanup

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jellydn/mole-tui/internal/mo/motest"
)

func TestStartStreamsLinesBeforeCompletion(t *testing.T) {
	stub := motest.New()
	stub.CleanOutput = "cleaning...\nTotal freed: 22.8 GB\nfinished\n"

	stream := Start(context.Background(), Options{}, stub)
	var lines []string
	var completion *RunResult
	for event := range stream.Events {
		if event.Kind == EventDone {
			completion = event.Done
			continue
		}
		lines = append(lines, event.Line)
	}

	if want := []string{"cleaning...", "Total freed: 22.8 GB", "finished"}; !reflect.DeepEqual(lines, want) {
		t.Fatalf("streamed lines = %#v, want %#v", lines, want)
	}
	if completion == nil {
		t.Fatal("stream ended without a completion event")
	}
	if completion.Err != nil {
		t.Fatalf("completion error = %v", completion.Err)
	}
	if completion.Result.FreedText != "Total freed: 22.8 GB" {
		t.Errorf("FreedText = %q, want %q", completion.Result.FreedText, "Total freed: 22.8 GB")
	}
	if stub.CleanCalls != 1 {
		t.Errorf("CleanCalls = %d, want 1", stub.CleanCalls)
	}
}

func TestStartDryRunProducesCompletion(t *testing.T) {
	stub := motest.New()
	stream := Start(context.Background(), Options{DryRun: true}, stub)

	var lines []string
	var completion *RunResult
	for event := range stream.Events {
		if event.Kind == EventDone {
			completion = event.Done
			continue
		}
		lines = append(lines, event.Line)
	}

	if !reflect.DeepEqual(lines, []string{"Dry run complete — no files were modified"}) {
		t.Errorf("dry-run lines = %#v", lines)
	}
	if completion == nil || completion.Err != nil {
		t.Fatalf("dry-run completion = %#v", completion)
	}
	if stub.CleanCalls != 0 {
		t.Errorf("dry-run CleanCalls = %d, want 0", stub.CleanCalls)
	}
}

func TestStartCancellationCompletesWithoutLeaking(t *testing.T) {
	stub := motest.New()
	stub.Sleep = time.Hour
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	stream := Start(ctx, Options{}, stub)
	select {
	case event, ok := <-stream.Events:
		if !ok || event.Kind != EventDone || event.Done == nil {
			t.Fatalf("first event = %#v, open = %v; want completion", event, ok)
		}
		if !errors.Is(event.Done.Err, context.Canceled) {
			t.Fatalf("completion error = %v, want context.Canceled", event.Done.Err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for cancelled cleanup completion")
	}

	if _, ok := <-stream.Events; ok {
		t.Fatal("stream remained open after completion")
	}
}

func TestStartFullBufferStopsAfterCancellation(t *testing.T) {
	var output strings.Builder
	for i := 0; i < 300; i++ {
		fmt.Fprintf(&output, "line %d\n", i)
	}

	stub := motest.New()
	stub.CleanOutput = output.String()
	ctx, cancel := context.WithCancel(context.Background())
	stream := Start(ctx, Options{}, stub)

	deadline := time.Now().Add(time.Second)
	for len(stream.Events) < 256 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if len(stream.Events) < 256 {
		cancel()
		t.Fatal("stream did not fill its bounded event buffer")
	}

	// Abandon the full stream. Cancellation must unblock both the scanner and
	// the completion send rather than leaking the producer goroutine.
	cancel()
	select {
	case _, ok := <-stream.Events:
		if ok {
			for range stream.Events {
			}
		}
	case <-time.After(time.Second):
		t.Fatal("full stream did not close after cancellation")
	}
}

func TestStartAbandonedConsumerStopsAfterCancellation(t *testing.T) {
	stub := motest.New()
	stub.CleanOutput = "line\n"
	ctx, cancel := context.WithCancel(context.Background())
	stream := Start(ctx, Options{}, stub)

	// Do not consume the stream. Cancellation must still let the producer
	// finish instead of blocking forever on an output or completion send.
	cancel()
	select {
	case _, ok := <-stream.Events:
		if ok {
			// A buffered line or completion is fine; the key assertion is that
			// the channel eventually closes.
			for range stream.Events {
			}
		}
	case <-time.After(time.Second):
		t.Fatal("abandoned stream did not close after cancellation")
	}
}
