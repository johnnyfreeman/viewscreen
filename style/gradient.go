package style

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// ApplyGradient applies a horizontal color gradient to text.
// Uses HCL color space for perceptually uniform blending.
func ApplyGradient(text string, from, to lipgloss.Color) string {
	if noColor {
		return text
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}

	// Parse colors
	fromColor, err := colorful.Hex(string(from))
	if err != nil {
		return text
	}
	toColor, err := colorful.Hex(string(to))
	if err != nil {
		return text
	}

	var b strings.Builder
	b.Grow(len(text) * 20) // Estimate for ANSI codes

	for i, r := range runes {
		// Calculate interpolation factor
		var t float64
		if len(runes) > 1 {
			t = float64(i) / float64(len(runes)-1)
		}

		// Blend in HCL space for perceptually uniform gradient
		blended := fromColor.BlendHcl(toColor, t).Clamped()
		hex := blended.Hex()

		// Apply color to this character
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(hex))
		b.WriteString(style.Render(string(r)))
	}

	return b.String()
}

// ApplyThemeGradient applies the current theme's gradient colors to text.
func ApplyThemeGradient(text string) string {
	return ApplyGradient(text, CurrentTheme.GradientStart, CurrentTheme.GradientEnd)
}

// ApplyBoldGradient applies a bold gradient to text.
func ApplyBoldGradient(text string, from, to lipgloss.Color) string {
	if noColor {
		return Bold.Render(text)
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}

	fromColor, err := colorful.Hex(string(from))
	if err != nil {
		return Bold.Render(text)
	}
	toColor, err := colorful.Hex(string(to))
	if err != nil {
		return Bold.Render(text)
	}

	var b strings.Builder
	b.Grow(len(text) * 20)

	for i, r := range runes {
		var t float64
		if len(runes) > 1 {
			t = float64(i) / float64(len(runes)-1)
		}

		blended := fromColor.BlendHcl(toColor, t).Clamped()
		hex := blended.Hex()

		style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(hex))
		b.WriteString(style.Render(string(r)))
	}

	return b.String()
}

// ApplyThemeBoldGradient applies a bold gradient using theme colors.
func ApplyThemeBoldGradient(text string) string {
	return ApplyBoldGradient(text, CurrentTheme.GradientStart, CurrentTheme.GradientEnd)
}

// ApplySuccessGradient applies a success (green) gradient to text.
func ApplySuccessGradient(text string) string {
	return ApplyBoldGradient(text, CurrentTheme.SuccessGradientStart, CurrentTheme.SuccessGradientEnd)
}

// ApplyErrorGradient applies an error (red) gradient to text.
func ApplyErrorGradient(text string) string {
	return ApplyBoldGradient(text, CurrentTheme.ErrorGradientStart, CurrentTheme.ErrorGradientEnd)
}
