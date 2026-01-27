package tui

import (
	"strings"

	"github.com/johnnyfreeman/viewscreen/style"
)

// logoLines is the ASCII art for "viewscreen"
var logoLines = []string{
	"█ █ █ █▀▀ █ █ █",
	"▀▄▀ █ ██▄ ▀▄▀▄▀",
	"█▀ █▀▀ █▀█ █▀▀ █▀▀ █▄ █",
	"▄█ █▄▄ █▀▄ ██▄ ██▄ █ ▀█",
}

const logoDecoration = "· · · · · · · · · · · · ·"

// LogoRenderer renders the ASCII logo with gradient styling.
// It's a focused component that handles only logo-related rendering.
type LogoRenderer struct{}

// NewLogoRenderer creates a new logo renderer.
func NewLogoRenderer() *LogoRenderer {
	return &LogoRenderer{}
}

// Render renders the ASCII logo with gradient and decorations.
// Uses Ultraviolet for text styling to avoid escape sequence conflicts.
func (r *LogoRenderer) Render() string {
	var sb strings.Builder

	sb.WriteString(style.SidebarDecoText(logoDecoration))
	sb.WriteString("\n")
	sb.WriteString(style.MutedText("claude"))
	sb.WriteString("\n")

	for _, line := range logoLines {
		sb.WriteString(style.ApplyThemeBoldGradient(line))
		sb.WriteString("\n")
	}

	sb.WriteString(style.SidebarDecoText(logoDecoration))
	sb.WriteString("\n")

	return sb.String()
}

// RenderTitle renders just the styled "VIEWSCREEN" title.
// Used by the header in narrow terminal mode.
func (r *LogoRenderer) RenderTitle() string {
	return style.ApplyThemeBoldGradient("VIEWSCREEN")
}
