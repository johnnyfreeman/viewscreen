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
	output := r.Render()

	t.Run("contains decoration dots", func(t *testing.T) {
		if !strings.Contains(output, "·") {
			t.Error("expected decoration dots in logo")
		}
	})

	t.Run("contains claude text", func(t *testing.T) {
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

func TestLogoRenderer_RenderTitle(t *testing.T) {
	r := NewLogoRenderer()
	output := r.RenderTitle()

	if !strings.Contains(output, "VIEWSCREEN") {
		t.Error("expected VIEWSCREEN in title output")
	}
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
