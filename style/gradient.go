package style

import (
	"image/color"
	"strings"

	"github.com/charmbracelet/lipgloss"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/lucasb-eyer/go-colorful"
)

// applyGradientCore applies a horizontal color gradient to text with optional bold.
// Uses HCL color space for perceptually uniform blending.
// Uses Ultraviolet for proper style/content separation - this ensures that
// gradient text can be safely composed with other styles without escape
// sequence conflicts.
func applyGradientCore(text string, from, to lipgloss.Color, bold bool) string {
	if noColor {
		if bold {
			return Bold.Render(text)
		}
		return text
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}

	fromColor, err := colorful.Hex(string(from))
	if err != nil {
		if bold {
			return Bold.Render(text)
		}
		return text
	}
	toColor, err := colorful.Hex(string(to))
	if err != nil {
		if bold {
			return Bold.Render(text)
		}
		return text
	}

	var b strings.Builder
	b.Grow(len(text) * 20) // Estimate for ANSI codes

	var attrs uint8
	if bold {
		attrs = uv.AttrBold
	}

	for i, r := range runes {
		// Calculate interpolation factor
		var t float64
		if len(runes) > 1 {
			t = float64(i) / float64(len(runes)-1)
		}

		// Blend in HCL space for perceptually uniform gradient
		blended := fromColor.BlendHcl(toColor, t).Clamped()

		// Use Ultraviolet for proper style/content separation.
		// This produces cleaner ANSI sequences that won't conflict
		// with surrounding escape codes when gradient text is further
		// processed or wrapped in other styles.
		style := &uv.Style{
			Fg:    colorfulToRGBA(blended),
			Attrs: attrs,
		}
		b.WriteString(style.Styled(string(r)))
	}

	return b.String()
}

// colorfulToRGBA converts a colorful.Color to color.RGBA.
func colorfulToRGBA(c colorful.Color) color.RGBA {
	return color.RGBA{
		R: uint8(c.R * 255),
		G: uint8(c.G * 255),
		B: uint8(c.B * 255),
		A: 255,
	}
}

// ApplyGradient applies a horizontal color gradient to text.
func ApplyGradient(text string, from, to lipgloss.Color) string {
	return applyGradientCore(text, from, to, false)
}

// ApplyThemeGradient applies the current theme's gradient colors to text.
func ApplyThemeGradient(text string) string {
	return ApplyGradient(text, CurrentTheme.GradientStart, CurrentTheme.GradientEnd)
}

// ApplyBoldGradient applies a bold gradient to text.
func ApplyBoldGradient(text string, from, to lipgloss.Color) string {
	return applyGradientCore(text, from, to, true)
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
