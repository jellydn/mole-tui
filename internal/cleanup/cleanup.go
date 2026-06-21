// Package cleanup invokes `mo clean`, streaming stdout/stderr to the UI, and
// optionally parses a summary line from the output.
package cleanup

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
)

// Options controls cleanup behaviour.
type Options struct {
	// DryRun short-circuits execution with a canned success message.
	DryRun bool
	// Sudo enables elevated execution via sudo.
	Sudo bool
}

// Result holds the post-cleanup information.
type Result struct {
	ExitCode  int
	Stdout    string
	Stderr    string
	FreedText string // best-effort summary, e.g. "Total freed: 22.8 GB" or ""
}

// reFreed matches lines like "Total freed: 22.8 GB" or "22.8GB freed".
var reFreed = regexp.MustCompile(`(?i)(?:freed|cleaned|saved|reclaimed)\s*(?::)?\s*([0-9.]+\s*(?:KB|MB|GB|B))`)

// Run executes `mo clean` with the given options. moPath should be the
// resolved absolute path of the mo binary (from exec.LookPath). Output is
// written to writer as it arrives (for live streaming). Returns a Result with
// the full output.
//
// When DryRun is true, no command is executed — the writer receives a canned
// message and the result indicates success.
func Run(ctx context.Context, opts Options, writer io.Writer, moPath string) (Result, error) {
	if opts.DryRun {
		msg := "Dry run complete — no files were modified\n"
		if _, err := io.WriteString(writer, msg); err != nil {
			return Result{}, fmt.Errorf("write dry-run message: %w", err)
		}
		return Result{
			ExitCode:  0,
			Stdout:    msg,
			FreedText: "Dry run complete — no files were modified",
		}, nil
	}

	var cmd *exec.Cmd
	if opts.Sudo {
		cmd = exec.CommandContext(ctx, "sudo", moPath, "clean")
	} else {
		cmd = exec.CommandContext(ctx, moPath, "clean")
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	// Tee stdout so we capture it while writing to the UI
	stdout := io.MultiWriter(writer, &stdoutBuf)
	cmd.Stdout = stdout
	cmd.Stderr = &stderrBuf

	// Also capture stderr to writer for live streaming
	// exec.Cmd only supports one Stderr writer, so we tee manually
	pr, pw := io.Pipe()
	cmd.Stderr = pw

	err := cmd.Start()
	if err != nil {
		return Result{}, fmt.Errorf("start mo clean: %w", err)
	}

	// Read stderr and tee to both writer and buffer
	stderrDone := make(chan struct{})
	go func() {
		defer close(stderrDone)
		_, _ = io.Copy(io.MultiWriter(writer, &stderrBuf), pr)
	}()

	runErr := cmd.Wait()

	// Close the pipe writer so the goroutine finishes
	pw.Close()
	<-stderrDone

	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	result := Result{
		ExitCode: exitCode,
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
	}

	// Try to extract a freed amount from the full combined output
	result.FreedText = ParseSummary(stdoutBuf.String() + "\n" + stderrBuf.String())

	return result, nil
}

// ParseSummary attempts to extract a "freed" or "cleaned" amount from output.
// Returns a human-readable string, or "" if nothing can be parsed.
func ParseSummary(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if m := reFreed.FindStringSubmatch(line); len(m) > 0 {
			return fmt.Sprintf("Total freed: %s", m[1])
		}
	}
	return ""
}
