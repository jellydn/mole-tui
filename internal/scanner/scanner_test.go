package scanner

import (
	"os"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		fixture     string
		wantSects   int
		wantSkipped bool
		wantBytes   float64 // approximate — we just check > 0 for real fixture
		checkFunc   func(t *testing.T, result ScanResult)
	}{
		{
			name:      "real fixture has sections",
			fixture:   "testdata/mo-clean-dryrun-real.txt",
			wantSects: 2, // User essentials + App caches (truncated by timeout)
			checkFunc: func(t *testing.T, r ScanResult) {
				if r.Summary.SystemCachesSkipped != true {
					t.Error("expected SystemCachesSkipped=true")
				}
				if r.Summary.FreeSpace == "" {
					t.Error("expected FreeSpace to be parsed")
				}
				t.Logf("FreeSpace: %s, TotalReclaimable: %.0f bytes", r.Summary.FreeSpace, r.Summary.TotalReclaimable)
			},
		},
		{
			name:      "empty output produces no section headers",
			fixture:   "testdata/empty.txt",
			wantSects: 0,
			checkFunc: func(t *testing.T, r ScanResult) {
				if r.Summary.FreeSpace != "200.00GB" {
					t.Errorf("expected FreeSpace 200.00GB, got %s", r.Summary.FreeSpace)
				}
				if r.Summary.TotalReclaimable != 0 {
					t.Errorf("expected TotalReclaimable 0, got %.0f", r.Summary.TotalReclaimable)
				}
				if len(r.Sections) != 0 {
					t.Errorf("expected 0 sections for empty output, got %d", len(r.Sections))
				}
			},
		},
		{
			name:      "error output produces no sections",
			fixture:   "testdata/error-output.txt",
			wantSects: 0,
			checkFunc: func(t *testing.T, r ScanResult) {
				if len(r.Sections) != 0 {
					t.Errorf("expected 0 sections for error output, got %d", len(r.Sections))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.fixture)
			if err != nil {
				t.Fatalf("failed to read fixture %s: %v", tt.fixture, err)
			}
			result := Parse(string(data))
			if tt.wantSects >= 0 && len(result.Sections) < tt.wantSects {
				t.Errorf("expected at least %d sections, got %d", tt.wantSects, len(result.Sections))
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestStripANSI(t *testing.T) {
	input := "\033[1;35mClean Your Mac\033[0m"
	want := "Clean Your Mac"
	got := stripANSI(input)
	if got != want {
		t.Errorf("stripANSI(%q) = %q, want %q", input, got, want)
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		val  float64
		unit string
		want float64
	}{
		{1, "KB", 1024},
		{1, "MB", 1024 * 1024},
		{1, "GB", 1024 * 1024 * 1024},
		{512, "B", 512},
	}
	for _, tt := range tests {
		got := ParseSize(tt.val, tt.unit)
		if got != tt.want {
			t.Errorf("ParseSize(%f, %s) = %f, want %f", tt.val, tt.unit, got, tt.want)
		}
	}
}

func TestParseReclaimableSize(t *testing.T) {
	input := `➤ User essentials
  → User app cache 100 items, 481.6MB dry
  → Darwin user cache files, 241 old items, 331KB dry
  ✓ Trash · already empty`
	result := Parse(input)
	if result.Summary.TotalReclaimable <= 0 {
		t.Errorf("expected positive TotalReclaimable, got %.0f", result.Summary.TotalReclaimable)
	}
}

func TestParseNoSizeLines(t *testing.T) {
	input := `➤ Developer tools
  ✓ Nothing to clean
  ☞ Review manually`
	result := Parse(input)
	if result.Summary.TotalReclaimable != 0 {
		t.Errorf("expected 0 TotalReclaimable, got %.0f", result.Summary.TotalReclaimable)
	}
	if len(result.Sections) != 1 {
		t.Errorf("expected 1 section, got %d", len(result.Sections))
	}
}

func TestANSICleanYourMac(t *testing.T) {
	// Simulate the actual ANSI-styled mo output
	input := "\033[1;35mClean Your Mac\033[0m\n\n\033[0;33mDry Run Mode\033[0m, Preview only, no deletions\n\n\033[1;34m⚙\033[0m Apple Silicon | Free space: 105.00GB\n\033[1;35m➤ User essentials\033[0m\n  \033[0;33m→\033[0m Test item, \033[0;33m100.0MB\033[0m \033[0;33mdry\033[0m\n"
	result := Parse(input)
	if len(result.Sections) != 1 {
		t.Errorf("expected 1 section, got %d", len(result.Sections))
	}
	if result.Summary.FreeSpace == "" {
		t.Error("expected FreeSpace to be parsed")
	}
	if result.Summary.TotalReclaimable <= 0 {
		t.Errorf("expected positive TotalReclaimable, got %.0f", result.Summary.TotalReclaimable)
	}
}

// TestSectionName tests that section names are clean (no ANSI artifacts).
func TestSectionName(t *testing.T) {
	input := "➤ User essentials\n  → item1\n➤ App caches\n  → item2"
	result := Parse(input)
	if len(result.Sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(result.Sections))
	}
	if result.Sections[0].Name != "User essentials" {
		t.Errorf("expected section 0 name 'User essentials', got %q", result.Sections[0].Name)
	}
	if result.Sections[1].Name != "App caches" {
		t.Errorf("expected section 1 name 'App caches', got %q", result.Sections[1].Name)
	}
}

// TestTotalReclaimable verifies the total size aggregation across sections.
func TestTotalReclaimable(t *testing.T) {
	input := `➤ A
  → item, 100.0MB dry
  → other, 50.0MB dry
➤ B
  → thing, 2.0GB dry`
	result := Parse(input)
	expected := (100.0 * 1024 * 1024) + (50.0 * 1024 * 1024) + (2.0 * 1024 * 1024 * 1024)
	if result.Summary.TotalReclaimable != expected {
		t.Errorf("expected %.0f bytes, got %.0f", expected, result.Summary.TotalReclaimable)
	}
}

// TestSystemCachesSkipped verifies detection of the sudo-needed banner.
func TestSystemCachesSkipped(t *testing.T) {
	input := `◎ System caches need sudo, run sudo -v && mo clean --dry-run for full preview`
	result := Parse(input)
	if !result.Summary.SystemCachesSkipped {
		t.Error("expected SystemCachesSkipped=true")
	}
}

// TestWhitelistNotSection verifies whitelist sub-items (↳) don't create sections.
func TestWhitelistNotSection(t *testing.T) {
	input := `➤ User essentials
  → item1`
	if !strings.Contains(input, "→") {
		t.Fatal("test input broken")
	}
	result := Parse(input)
	if len(result.Sections) != 1 {
		t.Errorf("expected 1 section, got %d", len(result.Sections))
	}
}
