package style

// Color is a hex color string (e.g., "#A855F7").
// This type replaces lipgloss.Color to avoid Lipgloss v1 dependency in the theme.
// The actual styling is done by Ultraviolet, which accepts color.RGBA.
type Color string

// Theme defines a color palette for the application
type Theme struct {
	// Foreground colors
	FgBase  Color // Primary text
	FgMuted Color // Secondary/dimmed text
	FgSubtle Color // Very subtle text (line numbers)

	// Background colors
	BgBase    Color // Primary background
	BgSubtle  Color // Subtle background
	BgOverlay Color // Overlay/modal background

	// Semantic colors
	Success Color // Green for success/added
	Error   Color // Red for errors/removed
	Warning Color // Yellow for warnings
	Info    Color // Cyan for info
	Accent  Color // Purple/magenta accent (Claude branding)

	// Diff-specific colors
	DiffAddBg    Color // Background for added lines
	DiffRemoveBg Color // Background for removed lines

	// Gradient colors (for headers)
	GradientStart Color
	GradientEnd   Color

	// Success gradient (for session complete)
	SuccessGradientStart Color
	SuccessGradientEnd   Color

	// Error gradient (for session error)
	ErrorGradientStart Color
	ErrorGradientEnd   Color
}

// DefaultTheme is the default color theme using TrueColor hex values
var DefaultTheme = Theme{
	// Foreground colors
	FgBase:   "#E4E4E7", // Zinc-200
	FgMuted:  "#A1A1AA", // Zinc-400
	FgSubtle: "#71717A", // Zinc-500

	// Background colors
	BgBase:    "#18181B", // Zinc-900
	BgSubtle:  "#27272A", // Zinc-800
	BgOverlay: "#3F3F46", // Zinc-700

	// Semantic colors
	Success: "#4ADE80", // Green-400
	Error:   "#F87171", // Red-400
	Warning: "#FACC15", // Yellow-400
	Info:    "#22D3EE", // Cyan-400
	Accent:  "#A855F7", // Purple-500 (Claude-like)

	// Diff colors (subtle backgrounds)
	DiffAddBg:    "#14532D", // Green-900
	DiffRemoveBg: "#7F1D1D", // Red-900

	// Gradient (purple to violet-blue)
	GradientStart: "#A855F7", // Purple-500
	GradientEnd:   "#818CF8", // Indigo-400

	// Success gradient (green to teal)
	SuccessGradientStart: "#4ADE80", // Green-400
	SuccessGradientEnd:   "#2DD4BF", // Teal-400

	// Error gradient (red to orange)
	ErrorGradientStart: "#F87171", // Red-400
	ErrorGradientEnd:   "#FB923C", // Orange-400
}

// NoColorTheme is used when color output is disabled
var NoColorTheme = Theme{
	FgBase:               "",
	FgMuted:              "",
	FgSubtle:             "",
	BgBase:               "",
	BgSubtle:             "",
	BgOverlay:            "",
	Success:              "",
	Error:                "",
	Warning:              "",
	Info:                 "",
	Accent:               "",
	DiffAddBg:            "",
	DiffRemoveBg:         "",
	GradientStart:        "",
	GradientEnd:          "",
	SuccessGradientStart: "",
	SuccessGradientEnd:   "",
	ErrorGradientStart:   "",
	ErrorGradientEnd:     "",
}

// CurrentTheme holds the active theme
var CurrentTheme = DefaultTheme

// SetTheme sets the current theme
func SetTheme(t Theme) {
	CurrentTheme = t
}
