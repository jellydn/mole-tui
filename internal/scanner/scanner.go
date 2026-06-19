// Package scanner invokes `mo clean --dry-run` and parses the output into
// structured sections with a best-effort total reclaimable size.
package scanner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Section represents a named group of related cleanup items.
type Section struct {
	Name  string
	Lines []string
}

// ScanSummary holds metadata parsed from the dry-run output header.
type ScanSummary struct {
	Header              string
	FreeSpace           string
	TotalReclaimable    float64 // best-effort sum of all "dry" sizes in bytes
	SystemCachesSkipped bool
}

// ScanResult holds the full parsed result of a dry-run scan.
type ScanResult struct {
	Sections []Section
	Summary  ScanSummary
	Raw      string // original raw output (ANSI stripped)
}

// reANSI matches ANSI / CSI escape sequences.
var reANSI = regexp.MustCompile(`\033\[[0-9;]*[A-Za-z]`)

// reSection matches section header lines like "➤ User essentials".
var reSection = regexp.MustCompile(`^➤\s+(.+)$`)

// reSize matches size patterns like "466.6MB", "331KB", "0B" followed by "dry".
var reSize = regexp.MustCompile(`([0-9.]+)\s*(KB|MB|GB|B)\s+dry`)

// reFreeSpace matches the free space indicator in the header.
var reFreeSpace = regexp.MustCompile(`Free space:\s*([0-9.]+)(GB|MB|KB|B)`)

// reSystemCachesSkipped matches the sudo-needed hint.
var reSystemCachesSkipped = regexp.MustCompile(`System caches need sudo`)

// stripANSI removes ANSI escape sequences from s.
func stripANSI(s string) string {
	return reANSI.ReplaceAllString(s, "")
}

// ParseSize converts a size string like "466.6MB" to bytes.
func ParseSize(val float64, unit string) float64 {
	switch strings.ToUpper(unit) {
	case "KB":
		return val * 1024
	case "MB":
		return val * 1024 * 1024
	case "GB":
		return val * 1024 * 1024 * 1024
	default:
		return val // B
	}
}

// Parse reads a dry-run output string and returns a ScanResult.
func Parse(output string) ScanResult {
	output = stripANSI(output)
	result := ScanResult{Raw: output}
	lines := strings.Split(output, "\n")

	var currentSection *Section

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Check for section header
		if m := reSection.FindStringSubmatch(trimmed); len(m) > 0 {
			if currentSection != nil {
				result.Sections = append(result.Sections, *currentSection)
			}
			currentSection = &Section{
				Name: strings.TrimSpace(m[1]),
			}
			continue
		}

		// Check for system caches skipped
		if reSystemCachesSkipped.MatchString(trimmed) {
			result.Summary.SystemCachesSkipped = true
		}

		// Check for header info (⚙ line)
		if strings.Contains(trimmed, "Apple Silicon") || strings.Contains(trimmed, "Free space:") {
			result.Summary.Header = trimmed
			if m := reFreeSpace.FindStringSubmatch(trimmed); len(m) > 0 {
				val, err := strconv.ParseFloat(m[1], 64)
				if err == nil {
					result.Summary.FreeSpace = fmt.Sprintf("%.2f%s", val, m[2])
				}
			}
			// Don't add header to a section — it's before any section starts
			if currentSection == nil {
				continue
			}
		}

		// Check for "Clean Your Mac" or "Dry Run Mode" lines — skip them
		if strings.Contains(trimmed, "Clean Your Mac") || strings.Contains(trimmed, "Dry Run Mode") || strings.Contains(trimmed, "Preview only") {
			continue
		}

		// Extract sizes from item lines
		if sizeMatches := reSize.FindAllStringSubmatch(trimmed, -1); len(sizeMatches) > 0 {
			for _, sm := range sizeMatches {
				val, err := strconv.ParseFloat(sm[1], 64)
				if err == nil {
					result.Summary.TotalReclaimable += ParseSize(val, sm[2])
				}
			}
		}

		// Add line to current section, or start a catch-all section
		if currentSection != nil {
			currentSection.Lines = append(currentSection.Lines, trimmed)
		}
	}

	// Append last section
	if currentSection != nil {
		result.Sections = append(result.Sections, *currentSection)
	}

	return result
}

// Scan invokes `mo clean --dry-run` (with optional sudo elevation) and returns
// the parsed result. The context controls cancellation / timeout.
func Scan(ctx context.Context, sudo bool) (ScanResult, error) {
	var cmd *exec.Cmd
	if sudo {
		// sudo -v refreshes the credential cache; actual command is mo clean --dry-run
		// We combine them so the TUI only needs to manage one subprocess.
		cmd = exec.CommandContext(ctx, "sudo", "sh", "-c", "mo clean --dry-run")
	} else {
		cmd = exec.CommandContext(ctx, "mo", "clean", "--dry-run")
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Check for context cancellation before wrapping
		if ctx.Err() != nil {
			return ScanResult{}, ctx.Err()
		}
		errOutput := strings.TrimSpace(stderr.String())
		if errOutput == "" {
			errOutput = err.Error()
		}
		return ScanResult{}, fmt.Errorf("mo clean --dry-run failed: %s", errOutput)
	}

	return Parse(stdout.String()), nil
}
