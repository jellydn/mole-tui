package cleanup

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/jellydn/mole-tui/internal/mo/motest"
)

// TestRunStreamsOutput verifies cleanup output is streamed live to the writer
// and the freed summary is parsed from it.
func TestRunStreamsOutput(t *testing.T) {
	stub := motest.New()
	stub.CleanOutput = "cleaning...\nTotal freed: 22.8 GB\n"

	var buf bytes.Buffer
	result, err := Run(context.Background(), Options{}, &buf, stub)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if got := buf.String(); got != stub.CleanOutput {
		t.Errorf("streamed output = %q, want %q", got, stub.CleanOutput)
	}
	if result.FreedText != "Total freed: 22.8 GB" {
		t.Errorf("FreedText = %q, want %q", result.FreedText, "Total freed: 22.8 GB")
	}
	if stub.CleanCalls != 1 {
		t.Errorf("CleanCalls = %d, want 1", stub.CleanCalls)
	}
}

// TestRunExitCode verifies non-zero exits surface on Result, not as an error
// (the UI renders the exit code banner itself).
func TestRunExitCode(t *testing.T) {
	stub := motest.New()
	stub.CleanExit = 3
	stub.CleanStderr = "boom\n"

	result, err := Run(context.Background(), Options{}, io.Discard, stub)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.ExitCode != 3 {
		t.Errorf("ExitCode = %d, want 3", result.ExitCode)
	}
	if result.Stderr != "boom\n" {
		t.Errorf("Stderr = %q, want %q", result.Stderr, "boom\n")
	}
}

// TestRunStreamsStderr verifies stderr is streamed live to the writer too.
func TestRunStreamsStderr(t *testing.T) {
	stub := motest.New()
	stub.CleanStderr = "warning: something\n"

	var buf bytes.Buffer
	if _, err := Run(context.Background(), Options{}, &buf, stub); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(buf.String(), "warning: something") {
		t.Errorf("writer did not receive stderr: %q", buf.String())
	}
}

// TestRunSudo verifies sudo intent reaches the seam (ADR-005).
func TestRunSudo(t *testing.T) {
	stub := motest.New()

	if _, err := Run(context.Background(), Options{Sudo: true}, io.Discard, stub); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if stub.SudoCalls != 1 {
		t.Errorf("SudoCalls = %d, want 1", stub.SudoCalls)
	}
}

// TestRunCancellation verifies a cancelled context surfaces as an error.
func TestRunCancellation(t *testing.T) {
	stub := motest.New()
	stub.Sleep = time.Hour

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Run(ctx, Options{}, io.Discard, stub)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}
