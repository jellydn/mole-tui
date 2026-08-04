package mo

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFakeMo writes a stub mo shell script and returns its path. Behaviors
// (selected by args):
//
//	clean --dry-run → dry-run output, exit 0
//	clean           → cleanup output, exit 0
//	clean fail      → stderr "boom", exit 3
func writeFakeMo(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "mo")
	script := `#!/bin/sh
case "$2" in
  --dry-run)
    echo "Section A"
    echo "  -> item, 1.0MB dry"
    exit 0
    ;;
  fail)
    echo "boom" >&2
    exit 3
    ;;
esac
if [ "$1" = "clean" ]; then
  echo "Total freed: 500MB"
  exit 0
fi
echo "unexpected args: $*" >&2
exit 42
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestExecRunnerScan exercises the production seam for a dry-run scan.
func TestExecRunnerScan(t *testing.T) {
	r := NewRunner(writeFakeMo(t))
	var stdout, stderr bytes.Buffer
	exit, err := r.Run(context.Background(), false, &stdout, &stderr, "clean", "--dry-run")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if exit != 0 {
		t.Errorf("exit = %d, want 0", exit)
	}
	if !strings.Contains(stdout.String(), "1.0MB dry") {
		t.Errorf("stdout = %q, want dry-run output", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}
}

// TestExecRunnerCleanup exercises the production seam for a cleanup.
func TestExecRunnerCleanup(t *testing.T) {
	r := NewRunner(writeFakeMo(t))
	var stdout bytes.Buffer
	exit, err := r.Run(context.Background(), false, &stdout, io.Discard, "clean")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if exit != 0 {
		t.Errorf("exit = %d, want 0", exit)
	}
	if !strings.Contains(stdout.String(), "Total freed: 500MB") {
		t.Errorf("stdout = %q", stdout.String())
	}
}

// TestExecRunnerNonZeroExit verifies the exit-code-vs-error contract: a
// non-zero exit returns the code and a nil error.
func TestExecRunnerNonZeroExit(t *testing.T) {
	r := NewRunner(writeFakeMo(t))
	var stdout, stderr bytes.Buffer
	exit, err := r.Run(context.Background(), false, &stdout, &stderr, "clean", "fail")
	if err != nil {
		t.Fatalf("Run: %v (non-zero exit must not set err)", err)
	}
	if exit != 3 {
		t.Errorf("exit = %d, want 3", exit)
	}
	if !strings.Contains(stderr.String(), "boom") {
		t.Errorf("stderr = %q, want boom", stderr.String())
	}
}

// TestExecRunnerCancellation verifies a cancelled context surfaces as an
// error. The context is cancelled before the call so no subprocess is
// spawned — deterministic and orphan-free.
//
// Mid-run cancellation (kill the child) is not asserted here: os/exec's
// internal output pipes can block Wait until orphaned grandchildren close
// their write ends, making such a test timing-dependent. Mid-run cancellation
// is covered at the scanner/cleanup level via the motest stub, and the
// production behaviour is unchanged from the pre-seam code.
func TestExecRunnerCancellation(t *testing.T) {
	r := NewRunner(writeFakeMo(t))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	exit, err := r.Run(ctx, false, io.Discard, io.Discard, "clean")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if exit != -1 {
		t.Errorf("exit = %d, want -1", exit)
	}
}

// TestExecRunnerMissingBinary verifies invocation failures return -1 and err.
func TestExecRunnerMissingBinary(t *testing.T) {
	r := NewRunner(filepath.Join(t.TempDir(), "does-not-exist"))
	exit, err := r.Run(context.Background(), false, io.Discard, io.Discard, "clean")
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
	if exit != -1 {
		t.Errorf("exit = %d, want -1", exit)
	}
}
