// Package jsonrpc provides a newline-delimited JSON-RPC 2.0 transport over
// arbitrary readers and writers. mole-sidecar uses it to expose the scan and
// cleanup engine over stdio (ADR-016): one JSON object per line, requests read
// from the input, responses and server-initiated notifications written to the
// output. The transport owns framing and wire encoding; callers only implement
// the narrow Handler interface and call Notify for asynchronous events.
package jsonrpc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
)

// Standard JSON-RPC 2.0 error codes.
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
	CodeServerError    = -32000
)

// Error is a JSON-RPC 2.0 error object.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Response is a JSON-RPC 2.0 response object. ID is null for requests that
// carried no id (e.g. parse errors), matching the JSON-RPC spec.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// request is a decoded JSON-RPC 2.0 request.
type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// notification is a server-initiated message with no id.
type notification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

// Handler processes a single JSON-RPC method call. Params is the raw request
// params object (may be nil); the handler decodes it into a typed struct and
// returns a result and/or an error.
type Handler interface {
	Handle(method string, params json.RawMessage) (any, *Error)
}

// Server is a newline-delimited JSON-RPC 2.0 transport. Serve reads requests
// from its input, dispatches them to the Handler, and writes responses to its
// output. Notify can be called concurrently from other goroutines while Serve
// runs to emit server-initiated events.
type Server struct {
	in  io.Reader
	out io.Writer
	h   Handler

	mu sync.Mutex // serializes writes to out
}

// NewServer returns a Server that reads from in, writes to out, and dispatches
// to h.
func NewServer(in io.Reader, out io.Writer, h Handler) *Server {
	return &Server{in: in, out: out, h: h}
}

// Serve reads newline-delimited requests until EOF, dispatching each one to
// the Handler and writing its response. It returns the read error, if any.
func (s *Server) Serve() error {
	sc := bufio.NewScanner(s.in)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var req request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.write(Response{JSONRPC: "2.0", Error: &Error{Code: CodeParseError, Message: "parse error"}})
			continue
		}
		result, rerr := s.h.Handle(req.Method, req.Params)
		s.write(Response{JSONRPC: "2.0", ID: req.ID, Result: result, Error: rerr})
	}
	return sc.Err()
}

// Notify writes a server-initiated notification as a newline-delimited JSON
// object. It is safe to call from any goroutine while Serve runs. Messages
// that cannot be encoded are dropped.
func (s *Server) Notify(method string, params any) {
	data, err := json.Marshal(notification{JSONRPC: "2.0", Method: method, Params: params})
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Fprintln(s.out, string(data))
}

// write encodes and writes a single response line, falling back to an internal
// error response if encoding fails.
func (s *Server) write(resp Response) {
	data, err := json.Marshal(resp)
	if err != nil {
		data = []byte(`{"jsonrpc":"2.0","error":{"code":-32603,"message":"internal error"}}`)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Fprintln(s.out, string(data))
}
