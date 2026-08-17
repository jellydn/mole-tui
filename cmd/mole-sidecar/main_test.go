package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/jellydn/mole-tui/internal/jsonrpc"
)

// TestServePing verifies a request/response round-trip without a mo binary.
func TestServePing(t *testing.T) {
	in := `{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n"
	var out bytes.Buffer
	s := newServer(strings.NewReader(in), &out, "/usr/local/bin/mo", nil)
	if err := s.transport.Serve(); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 response line, got %d: %q", len(lines), out.String())
	}
	var resp jsonrpc.Response
	if err := json.Unmarshal([]byte(lines[0]), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	res, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("unexpected result shape: %v", resp.Result)
	}
	if res["moPath"] != "/usr/local/bin/mo" {
		t.Errorf("moPath = %v, want /usr/local/bin/mo", res["moPath"])
	}
}

// TestServeUnknownMethod verifies the -32601 error code.
func TestServeUnknownMethod(t *testing.T) {
	in := `{"jsonrpc":"2.0","id":2,"method":"nope"}` + "\n"
	var out bytes.Buffer
	s := newServer(strings.NewReader(in), &out, "mo", nil)
	_ = s.transport.Serve()
	var resp jsonrpc.Response
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error == nil || resp.Error.Code != jsonrpc.CodeMethodNotFound {
		t.Fatalf("expected -32601, got %+v", resp.Error)
	}
}

// TestServeMalformedJSON verifies the -32700 parse-error code.
func TestServeMalformedJSON(t *testing.T) {
	in := "not json\n"
	var out bytes.Buffer
	s := newServer(strings.NewReader(in), &out, "mo", nil)
	_ = s.transport.Serve()
	var resp jsonrpc.Response
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error == nil || resp.Error.Code != jsonrpc.CodeParseError {
		t.Fatalf("expected -32700, got %+v", resp.Error)
	}
}

// TestServeCleanupDryRun exercises the async cleanup path end-to-end without
// a mo binary: a response, a cleanup.line event, then a cleanup.done event.
// Event/response ordering is not assumed — lines are matched by shape.
func TestServeCleanupDryRun(t *testing.T) {
	in := `{"jsonrpc":"2.0","id":3,"method":"cleanup.start","params":{"dryRun":true}}` + "\n"
	var out bytes.Buffer
	s := newServer(strings.NewReader(in), &out, "mo", nil)
	if err := s.transport.Serve(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-s.cleanupDone:
	case <-time.After(2 * time.Second):
		t.Fatal("cleanup did not finish")
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	var gotResp, gotLine, gotDone bool
	for _, l := range lines {
		var raw map[string]any
		if err := json.Unmarshal([]byte(l), &raw); err != nil {
			t.Fatalf("bad line %q: %v", l, err)
		}
		if _, hasID := raw["id"]; hasID {
			result, _ := raw["result"].(map[string]any)
			if result["accepted"] == true {
				gotResp = true
			}
			continue
		}
		switch raw["method"] {
		case "cleanup.line":
			gotLine = true
		case "cleanup.done":
			gotDone = true
		}
	}
	if !gotResp {
		t.Error("missing accepted response")
	}
	if !gotLine {
		t.Error("missing cleanup.line event")
	}
	if !gotDone {
		t.Error("missing cleanup.done event")
	}
}
