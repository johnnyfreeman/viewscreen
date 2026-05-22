package tui

import (
	"strings"
	"testing"

	"github.com/johnnyfreeman/viewscreen/style"
)

func init() {
	// Initialize style with noColor for consistent test output
	style.Init(true)
}

func TestNewLogoRenderer(t *testing.T) {
	r := NewLogoRenderer()
	if r == nil {
		t.Fatal("NewLogoRenderer returned nil")
	}
}

func TestLogoRenderer_Render(t *testing.T) {
	r := NewLogoRenderer()
	output := r.Render("")

	t.Run("contains decoration dots", func(t *testing.T) {
		if !strings.Contains(output, "·") {
			t.Error("expected decoration dots in logo")
		}
	})

	t.Run("defaults to claude branding for unknown agent", func(t *testing.T) {
		if !strings.Contains(output, "claude") {
			t.Error("expected 'claude' text in logo")
		}
	})

	t.Run("contains all logo lines", func(t *testing.T) {
		for _, line := range logoLines {
			if !strings.Contains(output, line) {
				t.Errorf("expected logo line %q in output", line)
			}
		}
	})

	t.Run("has multiple lines", func(t *testing.T) {
		lines := strings.Split(output, "\n")
		if len(lines) < 4 {
			t.Errorf("expected at least 4 lines in logo, got %d", len(lines))
		}
	})
}

func TestLogoRenderer_RenderAgentBranding(t *testing.T) {
	r := NewLogoRenderer()

	t.Run("codex agent brands as codex", func(t *testing.T) {
		output := r.Render("codex")
		if !strings.Contains(output, "codex") {
			t.Errorf("expected 'codex' branding, got %q", output)
		}
		if strings.Contains(output, "claude") {
			t.Errorf("did not expect 'claude' branding for codex, got %q", output)
		}
	})

	t.Run("claude agent brands as claude", func(t *testing.T) {
		output := r.Render("claude")
		if !strings.Contains(output, "claude") {
			t.Errorf("expected 'claude' branding, got %q", output)
		}
	})
}

func TestAgentLabel(t *testing.T) {
	tests := []struct {
		agent string
		want  string
	}{
		{"", "claude"},
		{"claude", "claude"},
		{"codex", "codex"},
	}
	for _, tt := range tests {
		if got := agentLabel(tt.agent); got != tt.want {
			t.Errorf("agentLabel(%q) = %q, want %q", tt.agent, got, tt.want)
		}
	}
}

func TestLogoRenderer_RenderTitle(t *testing.T) {
	r := NewLogoRenderer()

	t.Run("contains wordmark", func(t *testing.T) {
		output := r.RenderTitle("")
		if !strings.Contains(output, "VIEWSCREEN") {
			t.Error("expected VIEWSCREEN in title output")
		}
	})

	t.Run("brands for the agent", func(t *testing.T) {
		output := r.RenderTitle("codex")
		if !strings.Contains(output, "codex") {
			t.Errorf("expected 'codex' branding in title, got %q", output)
		}
		if strings.Contains(output, "claude") {
			t.Errorf("did not expect 'claude' branding for codex, got %q", output)
		}
	})

	t.Run("defaults to claude for unknown agent", func(t *testing.T) {
		output := r.RenderTitle("")
		if !strings.Contains(output, "claude") {
			t.Errorf("expected default 'claude' branding, got %q", output)
		}
	})
}

func TestLogoLines(t *testing.T) {
	// Verify logo lines are defined and have content
	if len(logoLines) == 0 {
		t.Fatal("logoLines is empty")
	}

	for i, line := range logoLines {
		if line == "" {
			t.Errorf("logoLines[%d] is empty", i)
		}
	}
}

func TestLogoDecoration(t *testing.T) {
	if logoDecoration == "" {
		t.Error("logoDecoration is empty")
	}
	if !strings.Contains(logoDecoration, "·") {
		t.Error("expected dots in logoDecoration")
	}
}
