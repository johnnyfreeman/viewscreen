package style

import (
	"github.com/charmbracelet/colorprofile"

	"charm.land/lipgloss/v2"
)

const (
	// Bullet is the prefix for tool/action lines
	Bullet = "● "
	// OutputPrefix is the prefix for tool output lines
	OutputPrefix = "  ⎿  "
	// OutputContinue is the prefix for continued output lines
	OutputContinue = "     "

	// Nested prefixes for sub-agent tool calls
	NestedPrefix         = "  │ "     // Prefix for nested tool headers (before bullet)
	NestedOutputPrefix   = "  │   ⎿  " // Prefix for nested tool results
	NestedOutputContinue = "  │      " // Continued nested output
)

var (
	// Diff background colors used by Ultraviolet for composition-safe styling.
	// These are passed to HighlightFileWithBg() for syntax-highlighted diffs.
	DiffAddBg    Color
	DiffRemoveBg Color

	noColor bool
)

// Init initializes styles based on color settings
func Init(disableColor bool) {
	noColor = disableColor

	if disableColor {
		CurrentTheme = NoColorTheme
	} else {
		CurrentTheme = DefaultTheme

		// Force TrueColor output even when stdout is piped (not a TTY).
		//
		// In Lipgloss v2, color handling is done through the Writer variable
		// which uses colorprofile. We set TrueColor directly to ensure colors
		// work in pipelines.
		lipgloss.Writer.Profile = colorprofile.TrueColor
	}

	if noColor {
		DiffAddBg = ""
		DiffRemoveBg = ""
		return
	}

	t := CurrentTheme
	DiffAddBg = t.DiffAddBg
	DiffRemoveBg = t.DiffRemoveBg
}

// NoColor returns whether color output is disabled
func NoColor() bool {
	return noColor
}
