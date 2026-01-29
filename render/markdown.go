package render

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
)

// MarkdownRenderer wraps glamour for markdown rendering with multiple styles
type MarkdownRenderer struct {
	full   *glamour.TermRenderer // Full styled markdown
	muted  *glamour.TermRenderer // Muted style for secondary content
	noColor bool
}

// NewMarkdownRenderer creates a new markdown renderer
func NewMarkdownRenderer(noColor bool, width int) *MarkdownRenderer {
	mr := &MarkdownRenderer{noColor: noColor}

	if noColor {
		r, err := glamour.NewTermRenderer(
			glamour.WithStylePath("notty"),
			glamour.WithWordWrap(width),
		)
		if err == nil {
			mr.full = r
			mr.muted = r // Same renderer for no-color mode
		}
		return mr
	}

	// Full styled renderer
	full, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err == nil {
		mr.full = full
	}

	// Muted renderer for secondary content (thinking, etc.)
	mutedStyle := getMutedStyle()
	muted, err := glamour.NewTermRenderer(
		glamour.WithStyles(mutedStyle),
		glamour.WithWordWrap(width),
	)
	if err == nil {
		mr.muted = muted
	} else {
		// Fallback to full renderer if muted fails
		mr.muted = mr.full
	}

	return mr
}

// Render renders markdown content with full styling
func (m *MarkdownRenderer) Render(content string) string {
	if m.full == nil {
		return content
	}

	rendered, err := m.full.Render(content)
	if err != nil {
		return content
	}
	return strings.TrimSpace(rendered) + "\n"
}

// RenderMuted renders markdown content with muted styling
// Useful for thinking blocks, secondary content, etc.
func (m *MarkdownRenderer) RenderMuted(content string) string {
	if m.muted == nil {
		return content
	}

	rendered, err := m.muted.Render(content)
	if err != nil {
		return content
	}
	return strings.TrimSpace(rendered) + "\n"
}

// SetWidth recreates the renderers with a new word-wrap width.
// This is called when the viewport resizes.
func (m *MarkdownRenderer) SetWidth(width int) {
	if m.noColor {
		r, err := glamour.NewTermRenderer(
			glamour.WithStylePath("notty"),
			glamour.WithWordWrap(width),
		)
		if err == nil {
			m.full = r
			m.muted = r
		}
		return
	}

	// Full styled renderer
	full, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err == nil {
		m.full = full
	}

	// Muted renderer for secondary content
	mutedStyle := getMutedStyle()
	muted, err := glamour.NewTermRenderer(
		glamour.WithStyles(mutedStyle),
		glamour.WithWordWrap(width),
	)
	if err == nil {
		m.muted = muted
	} else {
		m.muted = m.full
	}
}

// getMutedStyle returns a glamour style config with muted colors
func getMutedStyle() ansi.StyleConfig {
	// Start with a dark base style and tone down the colors
	muted := "#71717A" // Zinc-500 for muted text
	subtle := "#52525B" // Zinc-600 for very subtle elements

	return ansi.StyleConfig{
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr(muted),
			},
			Margin: uintPtr(0),
		},
		Heading: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr(muted),
				Bold:  boolPtr(true),
			},
		},
		H1: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr(muted),
				Bold:  boolPtr(true),
			},
		},
		H2: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr(muted),
				Bold:  boolPtr(true),
			},
		},
		H3: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr(muted),
				Bold:  boolPtr(true),
			},
		},
		Paragraph: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr(muted),
			},
		},
		Text: ansi.StylePrimitive{
			Color: stringPtr(muted),
		},
		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr(subtle),
			},
		},
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color: stringPtr(subtle),
				},
				Margin: uintPtr(1),
			},
			Chroma: &ansi.Chroma{
				Text: ansi.StylePrimitive{
					Color: stringPtr(subtle),
				},
			},
		},
		Link: ansi.StylePrimitive{
			Color:     stringPtr(subtle),
			Underline: boolPtr(true),
		},
		LinkText: ansi.StylePrimitive{
			Color: stringPtr(muted),
		},
		List: ansi.StyleList{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color: stringPtr(muted),
				},
			},
		},
		Item: ansi.StylePrimitive{
			Color: stringPtr(muted),
		},
		Emph: ansi.StylePrimitive{
			Color:  stringPtr(muted),
			Italic: boolPtr(true),
		},
		Strong: ansi.StylePrimitive{
			Color: stringPtr(muted),
			Bold:  boolPtr(true),
		},
	}
}

// Helper functions for pointers
func stringPtr(s string) *string { return &s }
func boolPtr(b bool) *bool       { return &b }
func uintPtr(u uint) *uint       { return &u }
