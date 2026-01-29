// Package testutil provides shared test utilities and mocks for the viewscreen codebase.
package testutil

import (
	"github.com/johnnyfreeman/viewscreen/style"
)

// MockStyleApplier is a test double for render.StyleApplier.
// It provides predictable, inspectable output for tests by wrapping text
// in brackets (e.g., "[SUCCESS:text]") instead of ANSI codes.
type MockStyleApplier struct {
	NoColorVal bool
}

// Text styles (Ultraviolet-based)
func (m MockStyleApplier) SuccessText(text string) string     { return "[SUCCESS:" + text + "]" }
func (m MockStyleApplier) WarningText(text string) string     { return "[WARNING:" + text + "]" }
func (m MockStyleApplier) MutedText(text string) string       { return "[MUTED:" + text + "]" }
func (m MockStyleApplier) ErrorText(text string) string       { return "[ERROR:" + text + "]" }
func (m MockStyleApplier) ErrorBoldText(text string) string   { return "[ERROR_BOLD:" + text + "]" }
func (m MockStyleApplier) SuccessBoldText(text string) string { return "[SUCCESS_BOLD:" + text + "]" }

// Output prefixes
func (m MockStyleApplier) OutputPrefix() string   { return "  ⎿  " }
func (m MockStyleApplier) OutputContinue() string { return "     " }

// Diff-related styles
func (m MockStyleApplier) LineNumberRender(text string) string    { return "[LN:" + text + "]" }
func (m MockStyleApplier) LineNumberSepRender(text string) string { return "│" }
func (m MockStyleApplier) DiffAddBg() style.Color                 { return "#00ff00" }
func (m MockStyleApplier) DiffRemoveBg() style.Color              { return "#ff0000" }

// Session/header styles
func (m MockStyleApplier) SessionHeaderRender(text string) string    { return "[HEADER:" + text + "]" }
func (m MockStyleApplier) ApplyThemeBoldGradient(text string) string { return "[GRADIENT:" + text + "]" }
func (m MockStyleApplier) ApplySuccessGradient(text string) string   { return "[SUCCESS_GRAD:" + text + "]" }
func (m MockStyleApplier) ApplyErrorGradient(text string) string     { return "[ERROR_GRAD:" + text + "]" }

// Color state
func (m MockStyleApplier) NoColor() bool { return m.NoColorVal }

// TrackingStyleApplier wraps MockStyleApplier to track method calls for testing.
// Use this when you need to verify that specific style methods were called.
type TrackingStyleApplier struct {
	MockStyleApplier
	GradientCalls      []string
	SessionHeaderCalls []string
	MutedTextCalls     []string
	ErrorGradientCalls []string
	SuccessGradientCalls []string
}

func (m *TrackingStyleApplier) MutedText(text string) string {
	m.MutedTextCalls = append(m.MutedTextCalls, text)
	return "[MUTED:" + text + "]"
}

func (m *TrackingStyleApplier) ApplyThemeBoldGradient(text string) string {
	m.GradientCalls = append(m.GradientCalls, text)
	return "[GRADIENT:" + text + "]"
}

func (m *TrackingStyleApplier) SessionHeaderRender(text string) string {
	m.SessionHeaderCalls = append(m.SessionHeaderCalls, text)
	return "[HEADER:" + text + "]"
}

func (m *TrackingStyleApplier) ApplySuccessGradient(text string) string {
	m.SuccessGradientCalls = append(m.SuccessGradientCalls, text)
	return "[SUCCESS_GRAD:" + text + "]"
}

func (m *TrackingStyleApplier) ApplyErrorGradient(text string) string {
	m.ErrorGradientCalls = append(m.ErrorGradientCalls, text)
	return "[ERROR_GRAD:" + text + "]"
}

// MockConfigProvider is a test double for config.Provider.
type MockConfigProvider struct {
	VerboseVal   bool
	NoColorVal   bool
	ShowUsageVal bool
}

func (m MockConfigProvider) IsVerbose() bool { return m.VerboseVal }
func (m MockConfigProvider) NoColor() bool   { return m.NoColorVal }
func (m MockConfigProvider) ShowUsage() bool { return m.ShowUsageVal }

// StripANSI removes ANSI escape sequences from a string.
// Useful for testing output that may contain color codes.
func StripANSI(s string) string {
	// Match ANSI escape sequences: ESC [ ... m (SGR sequences)
	// This covers color codes, bold, underline, etc.
	result := make([]byte, 0, len(s))
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Skip until 'm' (end of SGR sequence)
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			if j < len(s) {
				i = j + 1
				continue
			}
		}
		result = append(result, s[i])
		i++
	}
	return string(result)
}
