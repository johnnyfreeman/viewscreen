package style

import (
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
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
	// Diff styles (Lipgloss for backgrounds, Ultraviolet for foregrounds)
	// These are valid Lipgloss use cases - background colors for diff highlighting.
	DiffAdd      lipgloss.Style
	DiffRemove   lipgloss.Style
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
		// Diff styles are no-ops when color is disabled
		DiffAdd = lipgloss.NewStyle()
		DiffRemove = lipgloss.NewStyle()
		DiffAddBg = ""
		DiffRemoveBg = ""
		return
	}

	t := CurrentTheme

	// Diff styles (delta-like with subtle backgrounds)
	// These are the only Lipgloss styles remaining - backgrounds are a valid use case.
	DiffAddBg = t.DiffAddBg
	DiffRemoveBg = t.DiffRemoveBg
	DiffAdd = lipgloss.NewStyle().Background(lipgloss.Color(string(DiffAddBg)))
	DiffRemove = lipgloss.NewStyle().Background(lipgloss.Color(string(DiffRemoveBg)))
}

// NoColor returns whether color output is disabled
func NoColor() bool {
	return noColor
}
