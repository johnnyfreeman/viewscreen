package terminal

import (
	"os"
	"regexp"
	"strings"

	"golang.org/x/term"
)

// DefaultWidth is the fallback terminal width when detection fails
const DefaultWidth = 80

// Width returns the current terminal width, or DefaultWidth if detection fails.
// This centralizes terminal width detection to avoid duplication across packages.
func Width() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	return DefaultWidth
}

// Truncate shortens a string to maxLen characters, adding "..." if truncated
func Truncate(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

// systemReminderRegex matches <system-reminder>...</system-reminder> blocks
var systemReminderRegex = regexp.MustCompile(`(?s)<system-reminder>.*?</system-reminder>\s*`)

// StripSystemReminders removes <system-reminder> blocks from content
func StripSystemReminders(s string) string {
	return strings.TrimSpace(systemReminderRegex.ReplaceAllString(s, ""))
}

// lineNumberRegex matches line number prefixes like "     1→" or "    10→"
var lineNumberRegex = regexp.MustCompile(`(?m)^\s*\d+→`)

// StripLineNumbers removes line number prefixes (e.g., "     1→") from content
func StripLineNumbers(s string) string {
	return lineNumberRegex.ReplaceAllString(s, "")
}

// DefaultMaxLines is the default number of lines to show before truncating
const DefaultMaxLines = 15

// TruncateLines limits output to maxLines and returns the truncated content
// along with the number of remaining lines (0 if not truncated).
func TruncateLines(content string, maxLines int) (truncated string, remaining int) {
	lines := strings.Split(content, "\n")

	// Remove trailing empty line if present (common from string operations)
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	if len(lines) <= maxLines {
		return content, 0
	}

	truncated = strings.Join(lines[:maxLines], "\n")
	remaining = len(lines) - maxLines
	return truncated, remaining
}
