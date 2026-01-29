package style

import (
	"github.com/charmbracelet/colorprofile"

	"charm.land/lipgloss/v2"
)

const (
	// Bullet is the icon for tool/action lines (without trailing space)
	Bullet = "●"
	// OutputPrefix is the prefix for tool output lines
	OutputPrefix = "  ⎿  "
	// OutputContinue is the prefix for continued output lines
	OutputContinue = "     "

	// Nested prefix pipe character (unstyled, for composition)
	nestedPipe = "│"
)

var (
	// NestedPrefix is the prefix for nested tool headers (before bullet).
	// The pipe character is styled with subtle/dark gray color.
	NestedPrefix string
	// NestedOutputPrefix is the prefix for nested tool results.
	NestedOutputPrefix string
	// NestedOutputContinue is the prefix for continued nested output.
	NestedOutputContinue string
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

	// Initialize nested prefixes with styled pipe character
	initNestedPrefixes()

	if noColor {
		DiffAddBg = ""
		DiffRemoveBg = ""
		return
	}

	t := CurrentTheme
	DiffAddBg = t.DiffAddBg
	DiffRemoveBg = t.DiffRemoveBg
}

// initNestedPrefixes initializes the nested prefix strings with styled pipe characters.
func initNestedPrefixes() {
	styledPipe := SubtleText(nestedPipe)
	NestedPrefix = "  " + styledPipe + " "
	NestedOutputPrefix = "  " + styledPipe + "   ⎿  "
	NestedOutputContinue = "  " + styledPipe + "      "
}

// NoColor returns whether color output is disabled
func NoColor() bool {
	return noColor
}
