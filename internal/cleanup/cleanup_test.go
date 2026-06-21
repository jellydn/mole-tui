package cleanup

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunDryRun(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer
	result, err := Run(ctx, Options{DryRun: true}, &buf, "mo")
	if err != nil {
		t.Fatalf("Run with DryRun failed: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
	if !strings.Contains(buf.String(), "Dry run complete") {
		t.Errorf("expected dry-run message in output, got %q", buf.String())
	}
	if result.FreedText == "" {
		t.Errorf("expected FreedText to be set")
	}
}

func TestParseSummary(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Total freed: 22.8 GB", "Total freed: 22.8 GB"},
		{"Cleaned: 1.5GB of data", "Total freed: 1.5GB"},
		{"Saved 500MB", "Total freed: 500MB"},
		{"Nothing to report", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := ParseSummary(tt.input)
		if got != tt.want {
			t.Errorf("ParseSummary(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
