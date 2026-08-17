package jsonrpc

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"
	"testing"
)

// stubHandler implements Handler with a configurable result/error.
type stubHandler struct {
	result any
	err    *Error
}

func (h *stubHandler) Handle(method string, params json.RawMessage) (any, *Error) {
	return h.result, h.err
}

func decodeLine(t *testing.T, s string) Response {
	t.Helper()
	var resp Response
	if err := json.Unmarshal([]byte(s), &resp); err != nil {
		t.Fatalf("bad response line %q: %v", s, err)
	}
	return resp
}

func TestServeDispatchesAndWritesResponse(t *testing.T) {
	var out bytes.Buffer
	h := &stubHandler{result: map[string]any{"ok": true}}
	s := NewServer(strings.NewReader(`{"jsonrpc":"2.0","id":7,"method":"ping"}`+"\n"), &out, h)
	if err := s.Serve(); err != nil {
		t.Fatal(err)
	}
	resp := decodeLine(t, strings.TrimSpace(out.String()))
	if string(resp.ID) != "7" {
		t.Errorf("id = %s, want 7", resp.ID)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	res, ok := resp.Result.(map[string]any)
	if !ok || res["ok"] != true {
		t.Errorf("result = %v, want {ok:true}", resp.Result)
	}
}

func TestServePropagatesHandlerError(t *testing.T) {
	var out bytes.Buffer
	h := &stubHandler{err: &Error{Code: CodeMethodNotFound, Message: "method not found: nope"}}
	s := NewServer(strings.NewReader(`{"jsonrpc":"2.0","id":2,"method":"nope"}`+"\n"), &out, h)
	if err := s.Serve(); err != nil {
		t.Fatal(err)
	}
	resp := decodeLine(t, strings.TrimSpace(out.String()))
	if resp.Error == nil || resp.Error.Code != CodeMethodNotFound {
		t.Fatalf("expected -32601, got %+v", resp.Error)
	}
	if resp.Result != nil {
		t.Errorf("unexpected result on error: %v", resp.Result)
	}
}

func TestServeParseError(t *testing.T) {
	var out bytes.Buffer
	s := NewServer(strings.NewReader("not json\n"), &out, &stubHandler{})
	if err := s.Serve(); err != nil {
		t.Fatal(err)
	}
	resp := decodeLine(t, strings.TrimSpace(out.String()))
	if resp.Error == nil || resp.Error.Code != CodeParseError {
		t.Fatalf("expected -32700, got %+v", resp.Error)
	}
}

func TestServeSkipsBlankLines(t *testing.T) {
	var out bytes.Buffer
	h := &stubHandler{result: "ok"}
	s := NewServer(strings.NewReader("\n  \n"+`{"jsonrpc":"2.0","id":1,"method":"ping"}`+"\n"), &out, h)
	if err := s.Serve(); err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(strings.TrimSpace(out.String()), "\n"); got != 0 {
		t.Fatalf("expected 1 response line, got %d: %q", got+1, out.String())
	}
}

func TestNotifyWritesNewlineDelimitedEvents(t *testing.T) {
	var out bytes.Buffer
	s := NewServer(strings.NewReader(""), &out, &stubHandler{})
	s.Notify("cleanup.line", map[string]string{"line": "hello"})
	s.Notify("cleanup.done", map[string]any{"done": true})

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 event lines, got %d: %q", len(lines), out.String())
	}
	var ev struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal([]byte(lines[0]), &ev); err != nil {
		t.Fatal(err)
	}
	if ev.Method != "cleanup.line" {
		t.Errorf("method = %s, want cleanup.line", ev.Method)
	}
	var params map[string]string
	if err := json.Unmarshal(ev.Params, &params); err != nil {
		t.Fatal(err)
	}
	if params["line"] != "hello" {
		t.Errorf("params = %v, want line:hello", params)
	}
}

func TestNotifyConcurrentIsSafe(t *testing.T) {
	var out bytes.Buffer
	s := NewServer(strings.NewReader(""), &out, &stubHandler{})
	const n = 10
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Notify("cleanup.line", map[string]string{"line": "x"})
		}()
	}
	wg.Wait()
	if got := strings.Count(strings.TrimSpace(out.String()), "\n"); got != n-1 {
		t.Fatalf("expected %d event lines, got %d: %q", n, got+1, out.String())
	}
}
