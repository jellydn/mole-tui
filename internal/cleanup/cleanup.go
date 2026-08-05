// Package cleanup invokes `mo clean`, streaming stdout/stderr to the UI, and
// optionally parses a summary line from the output.
package cleanup

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"

	"github.com/jellydn/mole-tui/internal/mo"
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

// synchronizedWriter serializes writes from the runner's concurrent stdout
// and stderr pumps before they reach the caller's writer.
type synchronizedWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (w *synchronizedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.w.Write(p)
}

// Session owns one Cleanup session's options, runner, context, output
// lifecycle, and result semantics. Consumers choose Run for synchronous use or
// Start for the tea-free event stream without reconstructing those facts.
type Session struct {
	ctx    context.Context
	opts   Options
	runner mo.Runner
}

// NewSession creates a Cleanup session. The session does not start work until
// Run or Start is called.
func NewSession(ctx context.Context, opts Options, runner mo.Runner) Session {
	return Session{ctx: ctx, opts: opts, runner: runner}
}

// Run executes this Cleanup session synchronously. Output is written to writer
// as it arrives (for live streaming), and the returned Result retains the full
// stdout/stderr buffers.
func (s Session) Run(writer io.Writer) (Result, error) {
	if s.opts.DryRun {
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

	var stdoutBuf, stderrBuf bytes.Buffer
	// Both streams go live to the writer while being buffered for the final
	// result; the runner owns all subprocess plumbing. os/exec may write to
	// stdout and stderr concurrently, so protect the caller's writer.
	streamWriter := &synchronizedWriter{w: writer}
	stdout := io.MultiWriter(streamWriter, &stdoutBuf)
	stderr := io.MultiWriter(streamWriter, &stderrBuf)

	exitCode, err := s.runner.Run(s.ctx, s.opts.Sudo, stdout, stderr, "clean")
	if err != nil {
		// Covers both start failures and cancellation mid-run.
		return Result{}, fmt.Errorf("mo clean failed: %w", err)
	}

	result := Result{
		ExitCode: exitCode,
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
	}

	// Try to extract a freed amount from the full combined output.
	result.FreedText = ParseSummary(stdoutBuf.String() + "\n" + stderrBuf.String())
	return result, nil
}

// Run is the compatibility convenience for one synchronous Cleanup session.
func Run(ctx context.Context, opts Options, writer io.Writer, runner mo.Runner) (Result, error) {
	return NewSession(ctx, opts, runner).Run(writer)
}

// EventKind identifies the kind of event in a cleanup stream.
type EventKind uint8

const (
	EventLine EventKind = iota
	EventDone
)

// Event is one tea-free cleanup stream event. EventLine carries one output
// line; EventDone carries the final RunResult.
type Event struct {
	Kind EventKind
	Line string
	Done *RunResult
}

// RunResult is the completion event produced by Start.
type RunResult struct {
	Result Result
	Err    error
}

// Stream is the tea-free event source for an asynchronous cleanup run. Events
// arrive in output order and normally end with one Done event before the
// channel closes. If the consumer abandons a full stream and cancels the
// context, Start may close without delivering Done so its worker can exit.
type Stream struct {
	Events <-chan Event
}

// Start runs cleanup asynchronously and owns the output-pump lifecycle. The
// cleanup package, rather than a UI adapter, owns the pipe, scanner, and
// goroutines; callers only consume plain Go events and translate them into
// their framework's messages.
func (s Session) Start() Stream {
	events := make(chan Event, 256)

	go func() {
		reader, writer := io.Pipe()
		scanFinished := make(chan struct{})

		go func() {
			defer close(scanFinished)
			defer reader.Close()
			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				select {
				case events <- Event{Kind: EventLine, Line: scanner.Text()}:
				case <-s.ctx.Done():
					return
				}
			}
		}()

		result, err := s.Run(writer)
		_ = writer.Close()
		<-scanFinished
		completion := Event{Kind: EventDone, Done: &RunResult{Result: result, Err: err}}
		// Deliver completion whenever there is room. If a consumer abandoned a
		// full stream and cancelled the context, stop instead of leaking this
		// worker forever. A normal consumer (including a cancelled run) gets the
		// completion event before the channel closes.
		select {
		case events <- completion:
		default:
			if s.ctx.Err() != nil {
				close(events)
				return
			}
			select {
			case events <- completion:
			case <-s.ctx.Done():
				close(events)
				return
			}
		}
		close(events)
	}()

	return Stream{Events: events}
}

// Start is the compatibility convenience for one asynchronous Cleanup session.
func Start(ctx context.Context, opts Options, runner mo.Runner) Stream {
	return NewSession(ctx, opts, runner).Start()
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
