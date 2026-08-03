// Command mole-sidecar exposes the mole-tui scan/cleanup engine over
// JSON-RPC 2.0 (newline-delimited JSON on stdio) so external GUIs can drive
// the mo CLI without owning subprocesses. It never deletes files itself —
// all destructive actions go through `mo clean` (FR-6).
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/jellydn/mole-tui/internal/cleanup"
	"github.com/jellydn/mole-tui/internal/scanner"
)

// version is injected at build time via -ldflags (like cmd/mole-tui).
var version = "dev"

// ---- JSON-RPC 2.0 wire types ----

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// event is a server-initiated notification (no id).
type event struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

// ---- server ----

type server struct {
	out io.Writer

	mu            sync.Mutex
	moPath        string
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

// serve reads requests from in until EOF, dispatching each one synchronously.
// Long-running work (scan/cleanup) runs in goroutines and reports via events.
func (s *server) serve(in io.Reader) error {
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var req request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.respond(nil, nil, &rpcError{Code: -32700, Message: "parse error"})
			continue
		}
		result, rerr := s.handle(req.Method, req.Params)
		s.respond(req.ID, result, rerr)
	}
	return sc.Err()
}

func (s *server) handle(method string, params json.RawMessage) (any, *rpcError) {
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
		return nil, &rpcError{Code: -32601, Message: "method not found: " + method}
	}
}

func invalidParams(err error) *rpcError {
	return &rpcError{Code: -32602, Message: "invalid params: " + err.Error()}
}

// startScan kicks off `mo clean --dry-run` in a goroutine. Completion is
// reported as scan.done / scan.error / scan.cancelled events.
func (s *server) startScan(p scanParams) (any, *rpcError) {
	s.mu.Lock()
	if s.scanCancel != nil {
		s.mu.Unlock()
		return nil, &rpcError{Code: -32000, Message: "a scan is already running"}
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.scanCancel = cancel
	s.scanDone = make(chan struct{})
	s.mu.Unlock()

	go func() {
		defer close(s.scanDone)
		result, err := scanner.Scan(ctx, s.moPath, p.Sudo)
		s.mu.Lock()
		s.scanCancel = nil
		s.mu.Unlock()
		switch {
		case err != nil && ctx.Err() != nil:
			s.emit("scan.cancelled", map[string]string{"reason": "cancelled"})
		case err != nil:
			s.emit("scan.error", map[string]string{"error": err.Error()})
		default:
			s.emit("scan.done", map[string]any{"result": result})
		}
	}()
	return map[string]bool{"accepted": true}, nil
}

// startCleanup kicks off `mo clean` in a goroutine, streaming stdout/stderr
// chunks as cleanup.line events and completion as cleanup.done.
func (s *server) startCleanup(p cleanupParams) (any, *rpcError) {
	s.mu.Lock()
	if s.cleanupCancel != nil {
		s.mu.Unlock()
		return nil, &rpcError{Code: -32000, Message: "a cleanup is already running"}
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cleanupCancel = cancel
	s.cleanupDone = make(chan struct{})
	s.mu.Unlock()

	go func() {
		defer close(s.cleanupDone)
		opts := cleanup.Options{DryRun: p.DryRun, Sudo: p.Sudo}
		result, err := cleanup.Run(ctx, opts, &eventWriter{s: s}, s.moPath)
		if err != nil && ctx.Err() == nil {
			s.emit("cleanup.error", map[string]string{"error": err.Error()})
			return
		}
		s.emit("cleanup.done", map[string]any{"result": result, "cancelled": ctx.Err() != nil})
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

func (s *server) respond(id json.RawMessage, result any, rerr *rpcError) {
	data, err := json.Marshal(response{JSONRPC: "2.0", ID: id, Result: result, Error: rerr})
	if err != nil {
		data = []byte(`{"jsonrpc":"2.0","error":{"code":-32603,"message":"internal error"}}`)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Fprintln(s.out, string(data))
}

func (s *server) emit(method string, params any) {
	data, err := json.Marshal(event{JSONRPC: "2.0", Method: method, Params: params})
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Fprintln(s.out, string(data))
}

// eventWriter forwards cleanup output chunks to the client as cleanup.line
// events, so a GUI can stream live output without holding a terminal.
type eventWriter struct{ s *server }

func (w *eventWriter) Write(p []byte) (int, error) {
	w.s.emit("cleanup.line", map[string]string{"line": string(p)})
	return len(p), nil
}

func main() {
	moPath, err := exec.LookPath("mo")
	if err != nil {
		fmt.Fprintln(os.Stderr, "mole-sidecar: mo is not on $PATH (see https://github.com/tw93/mole)")
		os.Exit(1)
	}
	s := &server{out: os.Stdout, moPath: moPath}
	if err := s.serve(os.Stdin); err != nil {
		fmt.Fprintf(os.Stderr, "mole-sidecar: %v\n", err)
		os.Exit(1)
	}
}
