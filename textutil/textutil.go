// Package textutil provides text processing utilities for cleaning,
// truncating, and formatting text content.
package textutil

import (
	"fmt"
	"io"
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

// PrefixedWriter handles the common pattern of writing lines with different
// prefixes for the first line vs. subsequent lines. This consolidates
// the "first line uses OutputPrefix, rest use OutputContinue" pattern
// found across multiple renderers.
type PrefixedWriter struct {
	w              io.Writer
	firstPrefix    string
	continuePrefix string
	first          bool
}

// NewPrefixedWriter creates a writer that uses firstPrefix for the first line
// and continuePrefix for all subsequent lines.
func NewPrefixedWriter(w io.Writer, firstPrefix, continuePrefix string) *PrefixedWriter {
	return &PrefixedWriter{
		w:              w,
		firstPrefix:    firstPrefix,
		continuePrefix: continuePrefix,
		first:          true,
	}
}

// Prefix returns the current prefix (firstPrefix on first call, continuePrefix after).
// Each call advances the state, so the first call returns firstPrefix and
// subsequent calls return continuePrefix.
func (p *PrefixedWriter) Prefix() string {
	if p.first {
		p.first = false
		return p.firstPrefix
	}
	return p.continuePrefix
}

// WriteLine writes a line with the appropriate prefix and newline.
func (p *PrefixedWriter) WriteLine(line string) {
	fmt.Fprintf(p.w, "%s%s\n", p.Prefix(), line)
}

// WriteLinef writes a formatted line with the appropriate prefix and newline.
func (p *PrefixedWriter) WriteLinef(format string, args ...any) {
	fmt.Fprintf(p.w, "%s"+format+"\n", append([]any{p.Prefix()}, args...)...)
}

// Reset resets the writer so the next line uses firstPrefix again.
func (p *PrefixedWriter) Reset() {
	p.first = true
}

// IsFirst returns true if the next write will use the first prefix.
func (p *PrefixedWriter) IsFirst() bool {
	return p.first
}

// TruncationIndicator formats the standard "… (N more lines)" indicator.
// This consolidates the common pattern of showing how many lines were truncated.
func TruncationIndicator(remaining int) string {
	return fmt.Sprintf("… (%d more lines)", remaining)
}

// ContentCleaner applies a configurable sequence of cleaning operations to text.
// This consolidates scattered cleaning logic (like stripping system reminders
// and line numbers) into a single, composable pipeline.
type ContentCleaner struct {
	cleaners []func(string) string
}

// NewContentCleaner creates an empty content cleaner.
// Use the With* methods to add cleaning operations.
func NewContentCleaner() *ContentCleaner {
	return &ContentCleaner{}
}

// WithSystemReminderStrip adds system reminder stripping to the pipeline.
func (c *ContentCleaner) WithSystemReminderStrip() *ContentCleaner {
	c.cleaners = append(c.cleaners, StripSystemReminders)
	return c
}

// WithLineNumberStrip adds line number stripping to the pipeline.
func (c *ContentCleaner) WithLineNumberStrip() *ContentCleaner {
	c.cleaners = append(c.cleaners, StripLineNumbers)
	return c
}

// WithCustom adds a custom cleaning function to the pipeline.
func (c *ContentCleaner) WithCustom(fn func(string) string) *ContentCleaner {
	c.cleaners = append(c.cleaners, fn)
	return c
}

// Clean applies all configured cleaning operations in order.
func (c *ContentCleaner) Clean(content string) string {
	for _, cleaner := range c.cleaners {
		content = cleaner(content)
	}
	return content
}

// DefaultContentCleaner returns a cleaner with the standard cleaning operations:
// strip system reminders, then strip line numbers.
func DefaultContentCleaner() *ContentCleaner {
	return NewContentCleaner().
		WithSystemReminderStrip().
		WithLineNumberStrip()
}
