// Package motest provides a stub mo.Runner for tests: canned output, exit
// codes, delays, and invocation counters. It is the second adapter that makes
// the mo subprocess seam real (one adapter = hypothetical, two = real).
package motest

import (
	"context"
	"io"
	"time"

	"github.com/jellydn/mole-tui/internal/mo"
)

// Runner is a programmable stub mo binary.
type Runner struct {
	// Canned behaviors. Empty output means an empty stdout/stderr.
	ScanOutput  string
	ScanStderr  string
	ScanExit    int
	CleanOutput string
	CleanStderr string
	CleanExit   int

	// Sleep simulates a slow invocation (for cancellation tests).
	Sleep time.Duration

	// Recorded invocations.
	ScanCalls  int
	CleanCalls int
	SudoCalls  int
}

// Compile-time assertion that Runner implements the seam.
var _ mo.Runner = (*Runner)(nil)

// New returns a stub that succeeds with empty output on every invocation.
func New() *Runner { return &Runner{} }

// Run implements mo.Runner. Invocations are classified by the args: a dry-run
// scan is `clean --dry-run`; everything else is treated as a cleanup.
func (r *Runner) Run(ctx context.Context, sudo bool, stdoutW, stderrW io.Writer, args ...string) (int, error) {
	if r.Sleep > 0 {
		select {
		case <-time.After(r.Sleep):
		case <-ctx.Done():
			return 130, ctx.Err()
		}
	}
	if sudo {
		r.SudoCalls++
	}

	if isScan(args) {
		r.ScanCalls++
		_, _ = io.WriteString(stdoutW, r.ScanOutput)
		_, _ = io.WriteString(stderrW, r.ScanStderr)
		if r.ScanExit != 0 {
			return r.ScanExit, nil
		}
		return 0, nil
	}

	r.CleanCalls++
	_, _ = io.WriteString(stdoutW, r.CleanOutput)
	_, _ = io.WriteString(stderrW, r.CleanStderr)
	if r.CleanExit != 0 {
		return r.CleanExit, nil
	}
	return 0, nil
}

func isScan(args []string) bool {
	return len(args) == 2 && args[0] == "clean" && args[1] == "--dry-run"
}
