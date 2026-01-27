// Package textutil provides text processing utilities for cleaning,
// truncating, and formatting text content.
package textutil

import (
	"regexp"
	"strings"
)

// Truncate shortens a string to maxLen characters, adding "..." if truncated.
// If maxLen is too small to fit the ellipsis (<=3), truncates without ellipsis.
func Truncate(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// systemReminderRegex matches <system-reminder>...</system-reminder> blocks
var systemReminderRegex = regexp.MustCompile(`(?s)<system-reminder>.*?</system-reminder>\s*`)

// StripSystemReminders removes <system-reminder> blocks from content.
func StripSystemReminders(s string) string {
	return strings.TrimSpace(systemReminderRegex.ReplaceAllString(s, ""))
}

// lineNumberRegex matches line number prefixes like "     1→" or "    10→"
var lineNumberRegex = regexp.MustCompile(`(?m)^\s*\d+→`)

// StripLineNumbers removes line number prefixes (e.g., "     1→") from content.
func StripLineNumbers(s string) string {
	return lineNumberRegex.ReplaceAllString(s, "")
}

// DefaultMaxLines is the default number of lines to show before truncating.
const DefaultMaxLines = 15

// WrapText wraps text to fit within maxWidth, breaking on word boundaries.
// Limits output to 3 lines maximum, adding "..." if truncated.
func WrapText(s string, maxWidth int) string {
	if len(s) <= maxWidth {
		return s
	}

	var result strings.Builder
	words := strings.Fields(s)
	lineLen := 0

	for i, word := range words {
		wordLen := len(word)

		if lineLen+wordLen+1 > maxWidth && lineLen > 0 {
			result.WriteString("\n")
			lineLen = 0
		}

		if lineLen > 0 {
			result.WriteString(" ")
			lineLen++
		}

		// Truncate very long words
		if wordLen > maxWidth {
			word = word[:maxWidth-3] + "..."
			wordLen = maxWidth
		}

		result.WriteString(word)
		lineLen += wordLen

		// Limit to 3 lines
		if i > 0 && strings.Count(result.String(), "\n") >= 2 && lineLen > 0 {
			if len(s) > result.Len() {
				result.WriteString("...")
			}
			break
		}
	}

	return result.String()
}

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
