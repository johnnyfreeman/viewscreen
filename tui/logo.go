package tui

import (
	"strings"

	"github.com/johnnyfreeman/viewscreen/config"
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

// Render renders the ASCII logo with gradient and decorations. The agent
// argument is the active CLI ("claude" or "codex"); it is shown as a muted
// sub-label above the wordmark so Codex streams brand as "codex". An empty
// agent falls back to "claude" branding.
// Uses Ultraviolet for text styling to avoid escape sequence conflicts.
func (r *LogoRenderer) Render(agent string) string {
	var sb strings.Builder

	sb.WriteString(style.SidebarDecoText(logoDecoration))
	sb.WriteString("\n")
	sb.WriteString(style.MutedText(agentLabel(agent)))
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

// agentLabel returns the sub-label shown under the logo for the given agent.
// An unknown or empty agent defaults to Claude branding, preserving the
// original behavior for streams whose origin has not yet been detected.
func agentLabel(agent string) string {
	if agent == "" {
		return config.AgentClaude
	}
	return agent
}
