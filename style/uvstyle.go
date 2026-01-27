// Package style provides styling utilities.
//
// This file contains Ultraviolet-based styling functions that properly combine
// multiple style attributes (colors, underlines, bold, etc.) into single ANSI
// sequences, avoiding the broken UI issues that occur when escape sequences
// are embedded in strings and then wrapped with more sequences.
package style

import (
	"image/color"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/lucasb-eyer/go-colorful"
)

// TextStyle defines a text style with foreground color and optional attributes.
// This allows defining styles declaratively and applying them consistently.
type TextStyle struct {
	// FgColor is either a theme Color (hex string) or a direct color.RGBA.
	// Use ColorFromTheme() or ColorRGBA() to set this.
	fgHex  *Color      // Hex color from theme
	fgRGBA *color.RGBA // Direct RGBA color

	Attrs     uint8          // uv.AttrBold, uv.AttrItalic, etc.
	Underline ansi.Underline // Underline style (e.g., ansi.UnderlineDotted)
}

// styled applies a TextStyle to text. This is the core styling function.
func styled(text string, s TextStyle) string {
	if noColor {
		return text
	}

	var fg color.RGBA
	if s.fgHex != nil {
		fg = hexToRGBA(string(*s.fgHex))
	} else if s.fgRGBA != nil {
		fg = *s.fgRGBA
	}

	style := &uv.Style{
		Fg:        fg,
		Attrs:     s.Attrs,
		Underline: s.Underline,
	}
	return style.Styled(text)
}

// themeStyle creates a TextStyle from a theme color.
func themeStyle(c Color, attrs uint8) TextStyle {
	return TextStyle{fgHex: &c, Attrs: attrs}
}

// rgbaStyle creates a TextStyle from an RGBA color.
func rgbaStyle(c color.RGBA, attrs uint8) TextStyle {
	return TextStyle{fgRGBA: &c, Attrs: attrs}
}

// Common RGBA colors for sidebar (matching ANSI 24x colors).
var (
	colorANSI245 = color.RGBA{R: 142, G: 142, B: 142, A: 255} // Gray, ANSI 245
	colorANSI255 = color.RGBA{R: 238, G: 238, B: 238, A: 255} // Bright white, ANSI 255
	colorANSI241 = color.RGBA{R: 102, G: 102, B: 102, A: 255} // Dim gray, ANSI 241
	colorANSI242 = color.RGBA{R: 108, G: 108, B: 108, A: 255} // Dark gray, ANSI 242
)

// Semantic text styling functions using theme colors.

// MutedDottedUnderline applies both muted foreground color and dotted underline
// to text in a single ANSI sequence.
func MutedDottedUnderline(text string) string {
	if noColor {
		return text
	}
	return styled(text, TextStyle{
		fgHex:     &CurrentTheme.FgMuted,
		Underline: ansi.UnderlineDotted,
	})
}

// SuccessText applies success (green) foreground color.
func SuccessText(text string) string {
	return styled(text, themeStyle(CurrentTheme.Success, 0))
}

// WarningText applies warning (yellow) foreground color.
func WarningText(text string) string {
	return styled(text, themeStyle(CurrentTheme.Warning, 0))
}

// MutedText applies muted foreground color.
func MutedText(text string) string {
	return styled(text, themeStyle(CurrentTheme.FgMuted, 0))
}

// ErrorText applies error (red) foreground color.
func ErrorText(text string) string {
	return styled(text, themeStyle(CurrentTheme.Error, 0))
}

// ErrorBoldText applies error (red) foreground color with bold.
func ErrorBoldText(text string) string {
	return styled(text, themeStyle(CurrentTheme.Error, uv.AttrBold))
}

// SuccessBoldText applies success (green) foreground color with bold.
func SuccessBoldText(text string) string {
	return styled(text, themeStyle(CurrentTheme.Success, uv.AttrBold))
}

// InfoBoldText applies info (cyan) foreground color with bold.
func InfoBoldText(text string) string {
	return styled(text, themeStyle(CurrentTheme.Info, uv.AttrBold))
}

// BoldText applies bold styling without any color.
func BoldText(text string) string {
	if noColor {
		return text
	}
	return styled(text, TextStyle{Attrs: uv.AttrBold})
}

// AccentText applies accent (purple) foreground color.
func AccentText(text string) string {
	return styled(text, themeStyle(CurrentTheme.Accent, 0))
}

// SpinnerText applies spinner styling (accent color).
func SpinnerText(text string) string {
	return styled(text, themeStyle(CurrentTheme.Accent, 0))
}

// LineNumberText applies line number styling (subtle foreground).
func LineNumberText(text string) string {
	return styled(text, themeStyle(CurrentTheme.FgSubtle, 0))
}

// LineNumberSepText applies line number separator styling (info/cyan foreground).
func LineNumberSepText(text string) string {
	return styled(text, themeStyle(CurrentTheme.Info, 0))
}

// Sidebar styling functions using fixed RGBA colors.

// SidebarHeaderText applies sidebar header/label color (gray, ANSI 245).
func SidebarHeaderText(text string) string {
	return styled(text, rgbaStyle(colorANSI245, 0))
}

// SidebarValueText applies sidebar value color (bright white, ANSI 255).
func SidebarValueText(text string) string {
	return styled(text, rgbaStyle(colorANSI255, 0))
}

// SidebarTodoPendingText applies pending todo color (dim gray, ANSI 241).
func SidebarTodoPendingText(text string) string {
	return styled(text, rgbaStyle(colorANSI241, 0))
}

// SidebarTodoActiveText applies active/in-progress todo color (white, ANSI 255).
func SidebarTodoActiveText(text string) string {
	return styled(text, rgbaStyle(colorANSI255, 0))
}

// SidebarTodoDoneText applies completed todo color (muted, ANSI 245).
func SidebarTodoDoneText(text string) string {
	return styled(text, rgbaStyle(colorANSI245, 0))
}

// SidebarPromptText applies prompt text style (italic, gray ANSI 245).
func SidebarPromptText(text string) string {
	return styled(text, rgbaStyle(colorANSI245, uv.AttrItalic))
}

// SidebarDecoText applies decoration text color (dark gray, ANSI 242).
func SidebarDecoText(text string) string {
	return styled(text, rgbaStyle(colorANSI242, 0))
}

// hexToRGBA converts a hex color string to color.RGBA.
func hexToRGBA(hex string) color.RGBA {
	c, err := colorful.Hex(hex)
	if err != nil {
		return color.RGBA{}
	}
	return color.RGBA{
		R: uint8(c.R * 255),
		G: uint8(c.G * 255),
		B: uint8(c.B * 255),
		A: 255,
	}
}
