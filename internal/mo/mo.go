// Package mo owns the mo CLI boundary: resolving the binary, wrapping sudo,
// and the Runner seam consumed by scanner and cleanup. Keeping the seam here
// makes the mo subprocess injectable for tests (see internal/mo/motest).
package mo

import (
	"context"
	"io"
	"os/exec"
	"time"
)

// Runner executes mo commands. The production implementation spawns the real
// mo binary; tests substitute a stub (internal/mo/motest).
//
// Contract: stdout and stderr are streamed to the given writers as output
// arrives. The returned exit code is the process exit code; a non-zero exit
// does not set err. err is non-nil only when the command could not be started
// or waited on for another reason (missing binary, cancellation, ...).
type Runner interface {
	Run(ctx context.Context, sudo bool, stdoutW, stderrW io.Writer, args ...string) (exitCode int, err error)
}

// execRunner is the production Runner backed by os/exec.
type execRunner struct{ moPath string }

// commandWaitDelay bounds how long os/exec waits for descendant processes to
// close inherited output pipes after the command is cancelled.
const commandWaitDelay = time.Second

func (r *execRunner) Run(ctx context.Context, sudo bool, stdoutW, stderrW io.Writer, args ...string) (int, error) {
	var cmd *exec.Cmd
	if sudo {
		cmd = exec.CommandContext(ctx, "sudo", append([]string{r.moPath}, args...)...)
	} else {
		cmd = exec.CommandContext(ctx, r.moPath, args...)
	}
	cmd.WaitDelay = commandWaitDelay
	cmd.Stdout = stdoutW
	cmd.Stderr = stderrW
	if err := cmd.Run(); err != nil {
		// CommandContext reports a killed process as an ExitError. Preserve the
		// context cause so callers can distinguish cancellation from a failed
		// mo invocation.
		if ctxErr := ctx.Err(); ctxErr != nil {
			return -1, ctxErr
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return -1, err
	}
	return 0, nil
}

// Resolve returns the absolute path of the mo binary, or an error if it is
// not on $PATH. Resolution lives here, at the boundary where it is consumed.
func Resolve() (string, error) {
	return exec.LookPath("mo")
}

// NewRunner returns the production Runner for the given resolved mo path.
func NewRunner(moPath string) Runner {
	return &execRunner{moPath: moPath}
}
