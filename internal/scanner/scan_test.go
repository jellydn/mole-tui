package scanner

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jellydn/mole-tui/internal/mo/motest"
)

// TestScanParsesOutput runs a full Scan through the stub seam and checks the
// parsed result end-to-end.
func TestScanParsesOutput(t *testing.T) {
	stub := motest.New()
	stub.ScanOutput = "Clean Your Mac\n\n" +
		"⚙ Apple Silicon | Free space: 105.27GB\n" +
		"➤ User essentials\n" +
		"  → User app cache 100 items, 481.6MB dry\n" +
		"  ✓ Trash · already empty\n"

	result, err := Scan(context.Background(), stub, false)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(result.Sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(result.Sections))
	}
	if result.Sections[0].Name != "User essentials" {
		t.Errorf("section name = %q, want %q", result.Sections[0].Name, "User essentials")
	}
	if result.Summary.FreeSpace != "105.27GB" {
		t.Errorf("FreeSpace = %q, want 105.27GB", result.Summary.FreeSpace)
	}
	if want := 481.6 * 1024 * 1024; result.Summary.TotalReclaimable != want {
		t.Errorf("TotalReclaimable = %v, want %v", result.Summary.TotalReclaimable, want)
	}
}

// TestScanSudoVariant verifies the sudo flag reaches the seam (ADR-005).
func TestScanSudoVariant(t *testing.T) {
	stub := motest.New()
	stub.ScanOutput = "➤ A\n  → item, 1.0MB dry\n"

	if _, err := Scan(context.Background(), stub, true); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if stub.ScanCalls != 1 {
		t.Errorf("ScanCalls = %d, want 1", stub.ScanCalls)
	}
	if stub.SudoCalls != 1 {
		t.Errorf("SudoCalls = %d, want 1", stub.SudoCalls)
	}
}

// TestScanNoSudoByDefault verifies elevation is opt-in (ADR-005).
func TestScanNoSudoByDefault(t *testing.T) {
	stub := motest.New()
	stub.ScanOutput = "➤ A\n  → item, 1.0MB dry\n"

	if _, err := Scan(context.Background(), stub, false); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if stub.SudoCalls != 0 {
		t.Errorf("SudoCalls = %d, want 0", stub.SudoCalls)
	}
}

// TestScanError verifies non-zero exits surface the stderr text.
func TestScanError(t *testing.T) {
	stub := motest.New()
	stub.ScanExit = 1
	stub.ScanStderr = "mo is not installed\n"

	_, err := Scan(context.Background(), stub, false)
	if err == nil {
		t.Fatal("expected error for non-zero exit")
	}
	if !strings.Contains(err.Error(), "mo is not installed") {
		t.Errorf("error = %q, want stderr text", err)
	}
}

// TestScanCancellation verifies a cancelled context short-circuits with
// context.Canceled (mapped by the UI to scanCancelledMsg).
func TestScanCancellation(t *testing.T) {
	stub := motest.New()
	stub.Sleep = time.Hour

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Scan(ctx, stub, false)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}
