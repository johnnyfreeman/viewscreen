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

// MutedDottedUnderline applies both muted foreground color and dotted underline
// to text in a single ANSI sequence. This avoids the broken styling that occurs
// when DottedUnderline() output is passed to Muted.Render().
func MutedDottedUnderline(text string) string {
	if noColor {
		return text
	}

	style := &uv.Style{
		Fg:        hexToRGBA(string(CurrentTheme.FgMuted)),
		Underline: ansi.UnderlineDotted,
	}
	return style.Styled(text)
}

// SuccessText applies success (green) foreground color using Ultraviolet.
// Use this instead of lipgloss when the text might be composed with other styles.
func SuccessText(text string) string {
	if noColor {
		return text
	}
	style := &uv.Style{
		Fg: hexToRGBA(string(CurrentTheme.Success)),
	}
	return style.Styled(text)
}

// WarningText applies warning (yellow) foreground color using Ultraviolet.
// Use this instead of lipgloss when the text might be composed with other styles.
func WarningText(text string) string {
	if noColor {
		return text
	}
	style := &uv.Style{
		Fg: hexToRGBA(string(CurrentTheme.Warning)),
	}
	return style.Styled(text)
}

// MutedText applies muted foreground color using Ultraviolet.
// Use this instead of lipgloss when the text might be composed with other styles.
func MutedText(text string) string {
	if noColor {
		return text
	}
	style := &uv.Style{
		Fg: hexToRGBA(string(CurrentTheme.FgMuted)),
	}
	return style.Styled(text)
}

// ErrorText applies error (red) foreground color using Ultraviolet.
// Use this instead of lipgloss when the text might be composed with other styles.
func ErrorText(text string) string {
	if noColor {
		return text
	}
	style := &uv.Style{
		Fg: hexToRGBA(string(CurrentTheme.Error)),
	}
	return style.Styled(text)
}

// ErrorBoldText applies error (red) foreground color with bold using Ultraviolet.
// Use this for error headers when styling needs to be composition-safe.
func ErrorBoldText(text string) string {
	if noColor {
		return text
	}
	style := &uv.Style{
		Fg:    hexToRGBA(string(CurrentTheme.Error)),
		Attrs: uv.AttrBold,
	}
	return style.Styled(text)
}

// SuccessBoldText applies success (green) foreground color with bold using Ultraviolet.
// Use this for success headers when styling needs to be composition-safe.
func SuccessBoldText(text string) string {
	if noColor {
		return text
	}
	style := &uv.Style{
		Fg:    hexToRGBA(string(CurrentTheme.Success)),
		Attrs: uv.AttrBold,
	}
	return style.Styled(text)
}

// Sidebar styling functions using Ultraviolet for proper style/content separation.
// These use specific color codes matching the original lipgloss SidebarStyles.

// SidebarHeaderText applies sidebar header/label color (gray-ish, #245).
func SidebarHeaderText(text string) string {
	if noColor {
		return text
	}
	style := &uv.Style{
		Fg: color.RGBA{R: 142, G: 142, B: 142, A: 255}, // ANSI 245 equivalent
	}
	return style.Styled(text)
}

// SidebarValueText applies sidebar value color (bright white-ish, #255).
func SidebarValueText(text string) string {
	if noColor {
		return text
	}
	style := &uv.Style{
		Fg: color.RGBA{R: 238, G: 238, B: 238, A: 255}, // ANSI 255 equivalent
	}
	return style.Styled(text)
}

// SidebarTodoPendingText applies pending todo color (dim gray, #241).
func SidebarTodoPendingText(text string) string {
	if noColor {
		return text
	}
	style := &uv.Style{
		Fg: color.RGBA{R: 102, G: 102, B: 102, A: 255}, // ANSI 241 equivalent
	}
	return style.Styled(text)
}

// SidebarTodoActiveText applies active/in-progress todo color (white, #255).
func SidebarTodoActiveText(text string) string {
	if noColor {
		return text
	}
	style := &uv.Style{
		Fg: color.RGBA{R: 238, G: 238, B: 238, A: 255}, // ANSI 255 equivalent
	}
	return style.Styled(text)
}

// SidebarTodoDoneText applies completed todo color (muted, #245).
func SidebarTodoDoneText(text string) string {
	if noColor {
		return text
	}
	style := &uv.Style{
		Fg: color.RGBA{R: 142, G: 142, B: 142, A: 255}, // ANSI 245 equivalent
	}
	return style.Styled(text)
}

// SidebarPromptText applies prompt text style (italic, gray #245).
func SidebarPromptText(text string) string {
	if noColor {
		return text
	}
	style := &uv.Style{
		Fg:    color.RGBA{R: 142, G: 142, B: 142, A: 255}, // ANSI 245 equivalent
		Attrs: uv.AttrItalic,
	}
	return style.Styled(text)
}

// SidebarDecoText applies decoration text color (dark gray, #242).
func SidebarDecoText(text string) string {
	if noColor {
		return text
	}
	style := &uv.Style{
		Fg: color.RGBA{R: 108, G: 108, B: 108, A: 255}, // ANSI 242 equivalent
	}
	return style.Styled(text)
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
