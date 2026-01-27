package style

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
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
	DiffAddBg    lipgloss.Color
	DiffRemoveBg lipgloss.Color

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
		// By default, termenv checks isatty(stdout) and returns Ascii profile
		// when piped, causing lipgloss to strip all color codes. Using WithUnsafe()
		// bypasses the TTY check, allowing TrueColor to work in pipelines.
		//
		// Alternative approaches considered:
		//
		// 1. Environment variables (limited effectiveness):
		//    - CLICOLOR_FORCE=1: Only upgrades Ascii->ANSI (16 colors), not TrueColor
		//    - COLORTERM=truecolor: Only checked AFTER the TTY check, so ignored when piped
		//    - NO_COLOR/CLICOLOR: These disable colors, not enable them
		//
		// 2. termenv.WithTTY(true): Similar to WithUnsafe() but slightly less
		//    permissive. WithUnsafe() is preferred for CLI tools that intentionally
		//    output to pipes.
		//
		// 3. lipgloss.SetColorProfile(termenv.TrueColor): Sets the profile on the
		//    default renderer, but the underlying termenv.Output still thinks it's
		//    not a TTY, which can cause issues with some operations.
		//
		// The WithUnsafe() approach is the most robust for tools designed to have
		// their output piped or captured while preserving ANSI color codes.
		output := termenv.NewOutput(os.Stdout, termenv.WithUnsafe(), termenv.WithProfile(termenv.TrueColor))
		renderer := lipgloss.NewRenderer(os.Stdout, termenv.WithUnsafe())
		renderer.SetOutput(output)
		lipgloss.SetDefaultRenderer(renderer)
	}

	if noColor {
		// Diff styles are no-ops when color is disabled
		DiffAdd = lipgloss.NewStyle()
		DiffRemove = lipgloss.NewStyle()
		DiffAddBg = lipgloss.Color("")
		DiffRemoveBg = lipgloss.Color("")
		return
	}

	t := CurrentTheme

	// Diff styles (delta-like with subtle backgrounds)
	// These are the only Lipgloss styles remaining - backgrounds are a valid use case.
	DiffAddBg = t.DiffAddBg
	DiffRemoveBg = t.DiffRemoveBg
	DiffAdd = lipgloss.NewStyle().Background(DiffAddBg)
	DiffRemove = lipgloss.NewStyle().Background(DiffRemoveBg)
}

// NoColor returns whether color output is disabled
func NoColor() bool {
	return noColor
}
