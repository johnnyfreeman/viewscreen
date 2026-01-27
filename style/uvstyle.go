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
