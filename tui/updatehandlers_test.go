package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/johnnyfreeman/viewscreen/events"
)

func newTestModel() Model {
	m := NewModel()
	// Simulate window size to initialize viewport
	m.width = 100
	m.height = 40
	m.ready = true
	return m
}

func TestHandleKeyMsg(t *testing.T) {
	t.Run("quit on q", func(t *testing.T) {
		m := newTestModel()
		_, cmd := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

		if cmd == nil {
			t.Error("expected quit command on 'q' key")
		}
	})

	t.Run("quit on ctrl+c", func(t *testing.T) {
		m := newTestModel()
		_, cmd := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyCtrlC})

		if cmd == nil {
			t.Error("expected quit command on ctrl+c")
		}
	})

	t.Run("scroll up on k", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.viewport.GotoBottom()

		initialY := m.viewport.YOffset
		m, _ = m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})

		if m.viewport.YOffset >= initialY {
			t.Error("expected viewport to scroll up on 'k' key")
		}
	})

	t.Run("scroll down on j", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))

		initialY := m.viewport.YOffset
		m, _ = m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

		if m.viewport.YOffset <= initialY {
			t.Error("expected viewport to scroll down on 'j' key")
		}
	})

	t.Run("go to top on g", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.viewport.GotoBottom()

		m, _ = m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})

		if m.viewport.YOffset != 0 {
			t.Error("expected viewport to go to top on 'g' key")
		}
	})

	t.Run("go to bottom on G", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))

		m, _ = m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})

		if m.viewport.YOffset == 0 {
			t.Error("expected viewport to go to bottom on 'G' key")
		}
	})

	t.Run("no command on other keys", func(t *testing.T) {
		m := newTestModel()
		_, cmd := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

		if cmd != nil {
			t.Error("expected no command on unknown key")
		}
	})
}

func TestHandleWindowSizeMsg(t *testing.T) {
	t.Run("initializes viewport on first size message", func(t *testing.T) {
		m := NewModel()
		m.ready = false

		m = m.handleWindowSizeMsg(tea.WindowSizeMsg{Width: 120, Height: 50})

		if !m.ready {
			t.Error("expected ready to be true after window size")
		}
		if m.width != 120 {
			t.Errorf("width = %d, want 120", m.width)
		}
		if m.height != 50 {
			t.Errorf("height = %d, want 50", m.height)
		}
	})

	t.Run("updates dimensions on resize", func(t *testing.T) {
		m := newTestModel()

		m = m.handleWindowSizeMsg(tea.WindowSizeMsg{Width: 200, Height: 80})

		if m.width != 200 {
			t.Errorf("width = %d, want 200", m.width)
		}
		if m.height != 80 {
			t.Errorf("height = %d, want 80", m.height)
		}
	})

	t.Run("calculates content width", func(t *testing.T) {
		m := NewModel()
		m.ready = false

		m = m.handleWindowSizeMsg(tea.WindowSizeMsg{Width: 120, Height: 50})

		// Content width should be total - sidebar - border
		expectedWidth := 120 - sidebarWidth - 3
		if m.viewport.Width != expectedWidth {
			t.Errorf("viewport width = %d, want %d", m.viewport.Width, expectedWidth)
		}
	})
}

func TestHandleSpinnerTick(t *testing.T) {
	t.Run("updates spinner", func(t *testing.T) {
		m := newTestModel()

		m, cmd := m.handleSpinnerTick(spinner.TickMsg{})

		if cmd == nil {
			t.Error("expected spinner command after tick")
		}
	})
}

func TestHandleStdinClosed(t *testing.T) {
	t.Run("sets stdinDone", func(t *testing.T) {
		m := newTestModel()

		m = m.handleStdinClosed()

		if !m.stdinDone {
			t.Error("expected stdinDone to be true")
		}
	})
}

func TestHandleParseError(t *testing.T) {
	t.Run("no-op in non-verbose mode", func(t *testing.T) {
		m := newTestModel()
		initialContent := m.content.String()

		m = m.handleParseError(events.ParseError{Line: "bad json"})

		if m.content.String() != initialContent {
			t.Error("expected no content change in non-verbose mode")
		}
	})
}
