package tui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/johnnyfreeman/viewscreen/style"
)

// Search holds the state for the search feature.
type Search struct {
	Active       bool   // Whether search input is active
	Query        string // Current search query
	matchLines   []int  // Line numbers (0-indexed) that match the query
	currentMatch int    // Index into matchLines (-1 if no matches)
}

// NewSearch creates a new Search with default state.
func NewSearch() Search {
	return Search{currentMatch: -1}
}

// Enter activates search mode.
func (s *Search) Enter() {
	s.Active = true
	s.Query = ""
	s.matchLines = nil
	s.currentMatch = -1
}

// Exit deactivates search mode but keeps the query visible.
func (s *Search) Exit() {
	s.Active = false
}

// Clear deactivates search mode and clears the query entirely.
func (s *Search) Clear() {
	s.Active = false
	s.Query = ""
	s.matchLines = nil
	s.currentMatch = -1
}

// TypeRune appends a rune to the query.
func (s *Search) TypeRune(r rune) {
	s.Query += string(r)
}

// Backspace removes the last character from the query.
func (s *Search) Backspace() {
	if len(s.Query) > 0 {
		s.Query = s.Query[:len(s.Query)-1]
	}
}

// UpdateMatches finds all lines in content that match the current query.
// Content is the raw viewport content (may contain ANSI escape sequences).
func (s *Search) UpdateMatches(content string) {
	s.matchLines = nil
	s.currentMatch = -1

	if s.Query == "" {
		return
	}

	query := strings.ToLower(s.Query)
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		// Strip ANSI sequences before matching so we search visible text only
		plain := strings.ToLower(ansi.Strip(line))
		if strings.Contains(plain, query) {
			s.matchLines = append(s.matchLines, i)
		}
	}

	if len(s.matchLines) > 0 {
		s.currentMatch = 0
	}
}

// NextMatch advances to the next match, wrapping around.
func (s *Search) NextMatch() {
	if len(s.matchLines) == 0 {
		return
	}
	s.currentMatch = (s.currentMatch + 1) % len(s.matchLines)
}

// PrevMatch goes to the previous match, wrapping around.
func (s *Search) PrevMatch() {
	if len(s.matchLines) == 0 {
		return
	}
	s.currentMatch = (s.currentMatch - 1 + len(s.matchLines)) % len(s.matchLines)
}

// CurrentLine returns the line number of the current match, or -1 if none.
func (s *Search) CurrentLine() int {
	if s.currentMatch < 0 || s.currentMatch >= len(s.matchLines) {
		return -1
	}
	return s.matchLines[s.currentMatch]
}

// MatchCount returns the total number of matches.
func (s *Search) MatchCount() int {
	return len(s.matchLines)
}

// CurrentMatchIndex returns the 1-based index of the current match, or 0 if none.
func (s *Search) CurrentMatchIndex() int {
	if s.currentMatch < 0 {
		return 0
	}
	return s.currentMatch + 1
}

// HasQuery returns true if there is a non-empty search query (active or not).
func (s *Search) HasQuery() bool {
	return s.Query != ""
}

// RenderSearchBar renders the search input bar at the bottom of the viewport.
func RenderSearchBar(s Search, width int) string {
	if !s.Active && !s.HasQuery() {
		return ""
	}

	var sb strings.Builder

	if s.Active {
		sb.WriteString(style.AccentText("/"))
		sb.WriteString(s.Query)
		sb.WriteString(style.MutedText("█")) // cursor
	} else {
		sb.WriteString(style.MutedText("/"))
		sb.WriteString(style.MutedText(s.Query))
	}

	// Show match count
	if s.HasQuery() {
		count := s.MatchCount()
		if count == 0 {
			sb.WriteString("  ")
			sb.WriteString(style.ErrorText("no matches"))
		} else {
			sb.WriteString("  ")
			sb.WriteString(style.MutedText(
				strings.Join([]string{
					itoa(s.CurrentMatchIndex()),
					"/",
					itoa(count),
				}, ""),
			))
		}
	}

	// Pad to full width to create a visual bar
	line := sb.String()
	visibleLen := len(ansi.Strip(line))
	if visibleLen < width {
		line += strings.Repeat(" ", width-visibleLen)
	}

	return line
}

// itoa is a simple int-to-string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
