package ui

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jellydn/mole-tui/internal/mo/motest"
)

func TestOperationControllerScanCancellation(t *testing.T) {
	stub := motest.New()
	stub.Sleep = time.Hour
	o := newOperationController(false, stub)

	cmd := o.startScan(false)
	if !o.cancelScan() {
		t.Fatal("cancelScan() = false, want true")
	}

	msg := cmd()
	if _, ok := msg.(scanCancelledMsg); !ok {
		t.Fatalf("scan message = %T, want scanCancelledMsg", msg)
	}
	o.clearScan()
	if o.cancelScan() {
		t.Fatal("cancelScan() after clear = true, want false")
	}
}

func TestOperationControllerCleanupTranslatesEvents(t *testing.T) {
	stub := motest.New()
	stub.CleanOutput = "cleaning\nfinished\n"
	o := newOperationController(false, stub)

	o.startCleanup(false)

	var lines []string
	for i := 0; i < 4; i++ {
		msg := o.readCleanupStreamCmd()()
		switch msg := msg.(type) {
		case cleanupStreamMsg:
			lines = append(lines, msg.line)
		case cleanupCompleteMsg:
			if len(lines) != 2 {
				t.Fatalf("lines at completion = %#v, want two lines", lines)
			}
			if msg.err != nil {
				t.Fatalf("cleanup completion error = %v", msg.err)
			}
			o.clearCleanup()
			if o.cancelCleanup() {
				t.Fatal("cancelCleanup() after clear = true, want false")
			}
			return
		default:
			t.Fatalf("cleanup message = %T, want stream or completion", msg)
		}
	}
	t.Fatal("cleanup stream did not produce completion")
}

func TestOperationControllerCleanupCancellation(t *testing.T) {
	stub := motest.New()
	stub.Sleep = time.Hour
	o := newOperationController(false, stub)

	o.startCleanup(false)
	if !o.cancelCleanup() {
		t.Fatal("cancelCleanup() = false, want true")
	}

	msg := o.readCleanupStreamCmd()()
	completion, ok := msg.(cleanupCompleteMsg)
	if !ok {
		t.Fatalf("cleanup message = %T, want cleanupCompleteMsg", msg)
	}
	if !errors.Is(completion.err, context.Canceled) {
		t.Fatalf("cleanup completion error = %v, want context.Canceled", completion.err)
	}
	o.clearCleanup()
}
