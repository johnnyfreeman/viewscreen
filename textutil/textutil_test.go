package textutil

import (
	"strings"
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "string shorter than max",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "string exactly at max",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "string longer than max",
			input:    "hello world",
			maxLen:   5,
			expected: "hello...",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "max length zero",
			input:    "hello",
			maxLen:   0,
			expected: "...",
		},
		{
			name:     "string with leading/trailing spaces",
			input:    "  hello  ",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "string with spaces truncated",
			input:    "  hello world  ",
			maxLen:   5,
			expected: "hello...",
		},
		{
			name:     "single character",
			input:    "a",
			maxLen:   1,
			expected: "a",
		},
		{
			name:     "unicode characters - byte counting",
			input:    "hello 世界",
			maxLen:   12, // "hello " is 6 bytes, each Chinese char is 3 bytes = 12 total
			expected: "hello 世界",
		},
		{
			name:     "unicode truncated at byte boundary",
			input:    "hello 世界 test",
			maxLen:   6, // truncate right after "hello "
			expected: "hello ...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Truncate(%q, %d) = %q, expected %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestStripSystemReminders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no system reminder",
			input:    "Hello, this is some content",
			expected: "Hello, this is some content",
		},
		{
			name:     "single system reminder",
			input:    "Before <system-reminder>reminder text</system-reminder> After",
			expected: "Before After",
		},
		{
			name:     "system reminder at start",
			input:    "<system-reminder>reminder</system-reminder>Content here",
			expected: "Content here",
		},
		{
			name:     "system reminder at end",
			input:    "Content here<system-reminder>reminder</system-reminder>",
			expected: "Content here",
		},
		{
			name:     "multiple system reminders",
			input:    "A<system-reminder>r1</system-reminder>B<system-reminder>r2</system-reminder>C",
			expected: "ABC",
		},
		{
			name:     "multiline system reminder",
			input:    "Before\n<system-reminder>\nline1\nline2\n</system-reminder>\nAfter",
			expected: "Before\nAfter",
		},
		{
			name:     "only system reminder",
			input:    "<system-reminder>just a reminder</system-reminder>",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "nested angle brackets in reminder",
			input:    "Text <system-reminder>reminder with <tag> inside</system-reminder> more text",
			expected: "Text more text",
		},
		{
			name: "complex multiline content",
			input: `Hello world
<system-reminder>
This is a reminder
with multiple lines
and some code:
func main() {}
</system-reminder>
Goodbye`,
			expected: "Hello world\nGoodbye",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripSystemReminders(tt.input)
			if result != tt.expected {
				t.Errorf("StripSystemReminders(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStripLineNumbers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no line numbers",
			input:    "just plain text",
			expected: "just plain text",
		},
		{
			name:     "single line with number",
			input:    "     1→hello",
			expected: "hello",
		},
		{
			name:     "double digit line number",
			input:    "    10→content",
			expected: "content",
		},
		{
			name:     "triple digit line number",
			input:    "   100→content",
			expected: "content",
		},
		{
			name: "multiple lines with numbers",
			input: `     1→package main
     2→
     3→import "fmt"
     4→
     5→func main() {
     6→	fmt.Println("hello")
     7→}`,
			expected: `package main

import "fmt"

func main() {
	fmt.Println("hello")
}`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "mixed content",
			input:    "     1→numbered\nplain line\n     3→numbered again",
			expected: "numbered\nplain line\nnumbered again",
		},
		{
			name:     "line number not at start",
			input:    "text     1→not at start",
			expected: "text     1→not at start",
		},
		{
			name:     "just arrow without number",
			input:    "→arrow without number",
			expected: "→arrow without number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripLineNumbers(tt.input)
			if result != tt.expected {
				t.Errorf("StripLineNumbers(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTruncateLines(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		maxLines          int
		expectedContent   string
		expectedRemaining int
	}{
		{
			name:              "fewer lines than max",
			input:             "line1\nline2\nline3",
			maxLines:          5,
			expectedContent:   "line1\nline2\nline3",
			expectedRemaining: 0,
		},
		{
			name:              "exactly max lines",
			input:             "line1\nline2\nline3",
			maxLines:          3,
			expectedContent:   "line1\nline2\nline3",
			expectedRemaining: 0,
		},
		{
			name:              "more lines than max",
			input:             "line1\nline2\nline3\nline4\nline5",
			maxLines:          3,
			expectedContent:   "line1\nline2\nline3",
			expectedRemaining: 2,
		},
		{
			name:              "single line",
			input:             "single line",
			maxLines:          5,
			expectedContent:   "single line",
			expectedRemaining: 0,
		},
		{
			name:              "empty string",
			input:             "",
			maxLines:          5,
			expectedContent:   "",
			expectedRemaining: 0,
		},
		{
			name:              "trailing newline stripped",
			input:             "line1\nline2\nline3\n",
			maxLines:          5,
			expectedContent:   "line1\nline2\nline3\n",
			expectedRemaining: 0,
		},
		{
			name:              "trailing newline with truncation",
			input:             "line1\nline2\nline3\nline4\n",
			maxLines:          2,
			expectedContent:   "line1\nline2",
			expectedRemaining: 2,
		},
		{
			name:              "max lines zero",
			input:             "line1\nline2",
			maxLines:          0,
			expectedContent:   "",
			expectedRemaining: 2,
		},
		{
			name:              "max lines one",
			input:             "line1\nline2\nline3",
			maxLines:          1,
			expectedContent:   "line1",
			expectedRemaining: 2,
		},
		{
			name:              "default max lines value",
			input:             strings.Repeat("line\n", 20),
			maxLines:          DefaultMaxLines,
			expectedContent:   strings.TrimSuffix(strings.Repeat("line\n", DefaultMaxLines), "\n"),
			expectedRemaining: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, remaining := TruncateLines(tt.input, tt.maxLines)
			if content != tt.expectedContent {
				t.Errorf("TruncateLines(%q, %d) content = %q, expected %q", tt.input, tt.maxLines, content, tt.expectedContent)
			}
			if remaining != tt.expectedRemaining {
				t.Errorf("TruncateLines(%q, %d) remaining = %d, expected %d", tt.input, tt.maxLines, remaining, tt.expectedRemaining)
			}
		})
	}
}

func TestDefaultMaxLines(t *testing.T) {
	if DefaultMaxLines != 15 {
		t.Errorf("DefaultMaxLines = %d, expected 15", DefaultMaxLines)
	}
}
