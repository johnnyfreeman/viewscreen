package style

import "github.com/charmbracelet/lipgloss"

// Theme defines a color palette for the application
type Theme struct {
	// Foreground colors
	FgBase      lipgloss.Color // Primary text
	FgMuted     lipgloss.Color // Secondary/dimmed text
	FgSubtle    lipgloss.Color // Very subtle text (line numbers)

	// Background colors
	BgBase      lipgloss.Color // Primary background
	BgSubtle    lipgloss.Color // Subtle background
	BgOverlay   lipgloss.Color // Overlay/modal background

	// Semantic colors
	Success     lipgloss.Color // Green for success/added
	Error       lipgloss.Color // Red for errors/removed
	Warning     lipgloss.Color // Yellow for warnings
	Info        lipgloss.Color // Cyan for info
	Accent      lipgloss.Color // Purple/magenta accent (Claude branding)

	// Diff-specific colors
	DiffAddBg   lipgloss.Color // Background for added lines
	DiffRemoveBg lipgloss.Color // Background for removed lines

	// Gradient colors (for headers)
	GradientStart lipgloss.Color
	GradientEnd   lipgloss.Color

	// Success gradient (for session complete)
	SuccessGradientStart lipgloss.Color
	SuccessGradientEnd   lipgloss.Color

	// Error gradient (for session error)
	ErrorGradientStart lipgloss.Color
	ErrorGradientEnd   lipgloss.Color
}

// DefaultTheme is the default color theme using TrueColor hex values
var DefaultTheme = Theme{
	// Foreground colors
	FgBase:      lipgloss.Color("#E4E4E7"), // Zinc-200
	FgMuted:     lipgloss.Color("#A1A1AA"), // Zinc-400
	FgSubtle:    lipgloss.Color("#71717A"), // Zinc-500

	// Background colors
	BgBase:      lipgloss.Color("#18181B"), // Zinc-900
	BgSubtle:    lipgloss.Color("#27272A"), // Zinc-800
	BgOverlay:   lipgloss.Color("#3F3F46"), // Zinc-700

	// Semantic colors
	Success:     lipgloss.Color("#4ADE80"), // Green-400
	Error:       lipgloss.Color("#F87171"), // Red-400
	Warning:     lipgloss.Color("#FACC15"), // Yellow-400
	Info:        lipgloss.Color("#22D3EE"), // Cyan-400
	Accent:      lipgloss.Color("#A855F7"), // Purple-500 (Claude-like)

	// Diff colors (subtle backgrounds)
	DiffAddBg:    lipgloss.Color("#14532D"), // Green-900
	DiffRemoveBg: lipgloss.Color("#7F1D1D"), // Red-900

	// Gradient (purple to violet-blue)
	GradientStart: lipgloss.Color("#A855F7"), // Purple-500
	GradientEnd:   lipgloss.Color("#818CF8"), // Indigo-400

	// Success gradient (green to teal)
	SuccessGradientStart: lipgloss.Color("#4ADE80"), // Green-400
	SuccessGradientEnd:   lipgloss.Color("#2DD4BF"), // Teal-400

	// Error gradient (red to orange)
	ErrorGradientStart: lipgloss.Color("#F87171"), // Red-400
	ErrorGradientEnd:   lipgloss.Color("#FB923C"), // Orange-400
}

// NoColorTheme is used when color output is disabled
var NoColorTheme = Theme{
	FgBase:               lipgloss.Color(""),
	FgMuted:              lipgloss.Color(""),
	FgSubtle:             lipgloss.Color(""),
	BgBase:               lipgloss.Color(""),
	BgSubtle:             lipgloss.Color(""),
	BgOverlay:            lipgloss.Color(""),
	Success:              lipgloss.Color(""),
	Error:                lipgloss.Color(""),
	Warning:              lipgloss.Color(""),
	Info:                 lipgloss.Color(""),
	Accent:               lipgloss.Color(""),
	DiffAddBg:            lipgloss.Color(""),
	DiffRemoveBg:         lipgloss.Color(""),
	GradientStart:        lipgloss.Color(""),
	GradientEnd:          lipgloss.Color(""),
	SuccessGradientStart: lipgloss.Color(""),
	SuccessGradientEnd:   lipgloss.Color(""),
	ErrorGradientStart:   lipgloss.Color(""),
	ErrorGradientEnd:     lipgloss.Color(""),
}

// CurrentTheme holds the active theme
var CurrentTheme = DefaultTheme

// SetTheme sets the current theme
func SetTheme(t Theme) {
	CurrentTheme = t
}
