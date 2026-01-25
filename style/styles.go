package style

import (
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
)

// ANSI escape codes for extended underline styles
// These work in modern terminals (kitty, iTerm2, WezTerm, etc.)
const (
	ansiDottedUnderline = "\x1b[4:4m"
	ansiUnderlineReset  = "\x1b[24m"
)

var (
	// Text modifiers
	Bold      lipgloss.Style
	Dim       lipgloss.Style
	Italic    lipgloss.Style
	Underline lipgloss.Style

	// Semantic styles
	Error   lipgloss.Style
	Success lipgloss.Style
	Warning lipgloss.Style
	Info    lipgloss.Style
	Muted   lipgloss.Style

	// Component styles
	ToolHeader    lipgloss.Style
	SessionHeader lipgloss.Style

	// Diff styles (delta-like)
	DiffAdd       lipgloss.Style
	DiffRemove    lipgloss.Style
	DiffAddBg     lipgloss.Color
	DiffRemoveBg  lipgloss.Color
	LineNumber    lipgloss.Style
	LineNumberSep lipgloss.Style

	noColor bool
)

// Init initializes styles based on color settings
func Init(disableColor bool) {
	noColor = disableColor

	if disableColor {
		CurrentTheme = NoColorTheme
	} else {
		CurrentTheme = DefaultTheme
		// Use TrueColor (24-bit) for richer color output
		lipgloss.SetColorProfile(termenv.TrueColor)
	}

	if noColor {
		// All styles are no-ops when color is disabled
		Bold = lipgloss.NewStyle()
		Dim = lipgloss.NewStyle()
		Italic = lipgloss.NewStyle()
		Underline = lipgloss.NewStyle()
		Error = lipgloss.NewStyle()
		Success = lipgloss.NewStyle()
		Warning = lipgloss.NewStyle()
		Info = lipgloss.NewStyle()
		Muted = lipgloss.NewStyle()
		ToolHeader = lipgloss.NewStyle()
		SessionHeader = lipgloss.NewStyle()
		DiffAdd = lipgloss.NewStyle()
		DiffRemove = lipgloss.NewStyle()
		DiffAddBg = lipgloss.Color("")
		DiffRemoveBg = lipgloss.Color("")
		LineNumber = lipgloss.NewStyle()
		LineNumberSep = lipgloss.NewStyle()
		return
	}

	t := CurrentTheme

	// Text modifiers
	Bold = lipgloss.NewStyle().Bold(true)
	Dim = lipgloss.NewStyle().Faint(true)
	Italic = lipgloss.NewStyle().Italic(true)
	Underline = lipgloss.NewStyle().Underline(true)

	// Semantic styles using theme colors
	Error = lipgloss.NewStyle().Foreground(t.Error)
	Success = lipgloss.NewStyle().Foreground(t.Success)
	Warning = lipgloss.NewStyle().Foreground(t.Warning)
	Info = lipgloss.NewStyle().Foreground(t.Info)
	Muted = lipgloss.NewStyle().Foreground(t.FgMuted)

	// Component styles
	ToolHeader = lipgloss.NewStyle().Bold(true).Foreground(t.Accent)
	SessionHeader = lipgloss.NewStyle().Bold(true).Foreground(t.Info)

	// Diff styles (delta-like with subtle backgrounds)
	DiffAddBg = t.DiffAddBg
	DiffRemoveBg = t.DiffRemoveBg
	DiffAdd = lipgloss.NewStyle().Background(DiffAddBg)
	DiffRemove = lipgloss.NewStyle().Background(DiffRemoveBg)

	// Line number styles
	LineNumber = lipgloss.NewStyle().Foreground(t.FgSubtle)
	LineNumberSep = lipgloss.NewStyle().Foreground(t.Info)
}

// NoColor returns whether color output is disabled
func NoColor() bool {
	return noColor
}

// DottedUnderline applies a dotted underline to text.
// Falls back to regular underline in terminals that don't support it.
func DottedUnderline(text string) string {
	if noColor {
		return text
	}
	return ansiDottedUnderline + text + ansiUnderlineReset
}
