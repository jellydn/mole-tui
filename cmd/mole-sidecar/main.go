// Command mole-sidecar exposes the mole-tui scan/cleanup engine over
// JSON-RPC 2.0 (newline-delimited JSON on stdio) so external GUIs can drive
// the mo CLI without owning subprocesses. It never deletes files itself —
// all destructive actions go through `mo clean` (FR-6).
//
// The JSON-RPC framing and wire encoding live in internal/jsonrpc; this
// package only owns the scan/cleanup lifecycle and translates it into
// responses and notifications.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/jellydn/mole-tui/internal/cleanup"
	"github.com/jellydn/mole-tui/internal/jsonrpc"
	"github.com/jellydn/mole-tui/internal/mo"
	"github.com/jellydn/mole-tui/internal/scanner"
)

// version is injected at build time via -ldflags (like cmd/mole-tui).
var version = "dev"

// server is the JSON-RPC handler for the sidecar. It owns the scan and
// cleanup lifecycle and reports progress via transport notifications; all wire
// framing and encoding lives in the jsonrpc transport.
type server struct {
	transport *jsonrpc.Server

	mu            sync.Mutex
	moPath        string
	moRunner      mo.Runner
	scanCancel    context.CancelFunc
	cleanupCancel context.CancelFunc
	scanDone      chan struct{} // closed when the scan goroutine finishes (test hook)
	cleanupDone   chan struct{} // closed when the cleanup goroutine finishes (test hook)
}

type scanParams struct {
	Sudo bool `json:"sudo"`
}

type cleanupParams struct {
	DryRun bool `json:"dryRun"`
	Sudo   bool `json:"sudo"`
}

// newServer wires the domain server to a jsonrpc transport reading from in and
// writing to out. runner may be nil to resolve mo at runtime from moPath.
func newServer(in io.Reader, out io.Writer, moPath string, runner mo.Runner) *server {
	s := &server{moPath: moPath, moRunner: runner}
	s.transport = jsonrpc.NewServer(in, out, s)
	return s
}

// Handle implements jsonrpc.Handler. Long-running work (scan/cleanup) runs in
// goroutines and reports via notifications.
func (s *server) Handle(method string, params json.RawMessage) (any, *jsonrpc.Error) {
	switch method {
	case "ping":
		return map[string]any{"version": version, "moPath": s.moPath}, nil

	case "scan.start":
		var p scanParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, invalidParams(err)
		}
		return s.startScan(p)

	case "scan.cancel":
		s.cancelScan()
		return map[string]bool{"accepted": true}, nil

	case "cleanup.start":
		var p cleanupParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, invalidParams(err)
		}
		return s.startCleanup(p)

	case "cleanup.cancel":
		s.cancelCleanup()
		return map[string]bool{"accepted": true}, nil

	default:
		return nil, &jsonrpc.Error{Code: jsonrpc.CodeMethodNotFound, Message: "method not found: " + method}
	}
}

func invalidParams(err error) *jsonrpc.Error {
	return &jsonrpc.Error{Code: jsonrpc.CodeInvalidParams, Message: "invalid params: " + err.Error()}
}

// startScan kicks off `mo clean --dry-run` in a goroutine. Completion is
// reported as scan.done / scan.error / scan.cancelled events.
func (s *server) runner() mo.Runner {
	if s.moRunner != nil {
		return s.moRunner
	}
	return mo.NewRunner(s.moPath)
}

func (s *server) startScan(p scanParams) (any, *jsonrpc.Error) {
	s.mu.Lock()
	if s.scanCancel != nil {
		s.mu.Unlock()
		return nil, &jsonrpc.Error{Code: jsonrpc.CodeServerError, Message: "a scan is already running"}
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.scanCancel = cancel
	s.scanDone = make(chan struct{})
	s.mu.Unlock()

	go func() {
		defer close(s.scanDone)
		result, err := scanner.Scan(ctx, s.runner(), p.Sudo)
		s.mu.Lock()
		s.scanCancel = nil
		s.mu.Unlock()
		switch {
		case err != nil && ctx.Err() != nil:
			s.transport.Notify("scan.cancelled", map[string]string{"reason": "cancelled"})
		case err != nil:
			s.transport.Notify("scan.error", map[string]string{"error": err.Error()})
		default:
			s.transport.Notify("scan.done", map[string]any{"result": result})
		}
	}()
	return map[string]bool{"accepted": true}, nil
}

// startCleanup kicks off `mo clean` in a goroutine, streaming stdout/stderr
// chunks as cleanup.line events and completion as cleanup.done.
func (s *server) startCleanup(p cleanupParams) (any, *jsonrpc.Error) {
	s.mu.Lock()
	if s.cleanupCancel != nil {
		s.mu.Unlock()
		return nil, &jsonrpc.Error{Code: jsonrpc.CodeServerError, Message: "a cleanup is already running"}
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cleanupCancel = cancel
	s.cleanupDone = make(chan struct{})
	s.mu.Unlock()

	go func() {
		defer close(s.cleanupDone)
		opts := cleanup.Options{DryRun: p.DryRun, Sudo: p.Sudo}
		result, err := cleanup.Run(ctx, opts, &eventWriter{s: s}, s.runner())
		if err != nil && ctx.Err() == nil {
			s.transport.Notify("cleanup.error", map[string]string{"error": err.Error()})
			return
		}
		s.transport.Notify("cleanup.done", map[string]any{"result": result, "cancelled": ctx.Err() != nil})
	}()
	return map[string]bool{"accepted": true}, nil
}

func (s *server) cancelScan() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.scanCancel != nil {
		s.scanCancel()
	}
}

func (s *server) cancelCleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cleanupCancel != nil {
		s.cleanupCancel()
	}
}

// eventWriter forwards cleanup output chunks to the client as cleanup.line
// events, so a GUI can stream live output without holding a terminal.
type eventWriter struct{ s *server }

func (w *eventWriter) Write(p []byte) (int, error) {
	w.s.transport.Notify("cleanup.line", map[string]string{"line": string(p)})
	return len(p), nil
}

func main() {
	moPath, err := exec.LookPath("mo")
	if err != nil {
		fmt.Fprintln(os.Stderr, "mole-sidecar: mo is not on $PATH (see https://github.com/tw93/mole)")
		os.Exit(1)
	}
	s := newServer(os.Stdin, os.Stdout, moPath, mo.NewRunner(moPath))
	if err := s.transport.Serve(); err != nil {
		fmt.Fprintf(os.Stderr, "mole-sidecar: %v\n", err)
		os.Exit(1)
	}
}
