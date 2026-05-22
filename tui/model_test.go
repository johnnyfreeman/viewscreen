package tui

import (
	"strings"
	"testing"
)

func TestWithAgent(t *testing.T) {
	t.Run("seeds state.Agent so branding is correct before any event", func(t *testing.T) {
		m := NewModel(WithAgent("codex"))
		if m.state.Agent != "codex" {
			t.Errorf("state.Agent = %q, want %q", m.state.Agent, "codex")
		}
	})

	t.Run("empty agent is ignored, leaving detection to the stream", func(t *testing.T) {
		m := NewModel(WithAgent(""))
		if m.state.Agent != "" {
			t.Errorf("state.Agent = %q, want empty", m.state.Agent)
		}
	})

	t.Run("seeded codex agent renders codex branding in the sidebar", func(t *testing.T) {
		m := NewModel(WithAgent("codex"), WithInitialSize(120, 40))
		sidebar := RenderSidebar(m.state, m.spinner, 40, m.sidebarStyles,
			m.followMode, m.scrollPosition(), false, 0)
		if !strings.Contains(sidebar, "codex") {
			t.Errorf("expected codex branding in sidebar, got:\n%s", sidebar)
		}
	})
}
