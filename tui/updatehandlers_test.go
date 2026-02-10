package tui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
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
		// In v2, KeyPressMsg is used with Text field for printable characters
		_, cmd := m.handleKeyMsg(tea.KeyPressMsg{Text: "q"})

		if cmd == nil {
			t.Error("expected quit command on 'q' key")
		}
	})

	t.Run("quit on ctrl+c", func(t *testing.T) {
		m := newTestModel()
		// In v2, ctrl+c uses Code 'c' with ModCtrl modifier
		_, cmd := m.handleKeyMsg(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})

		if cmd == nil {
			t.Error("expected quit command on ctrl+c")
		}
	})

	t.Run("scroll up on k", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.viewport.GotoBottom()

		initialY := m.viewport.YOffset()
		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "k"})

		if m.viewport.YOffset() >= initialY {
			t.Error("expected viewport to scroll up on 'k' key")
		}
	})

	t.Run("scroll down on j", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))

		initialY := m.viewport.YOffset()
		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "j"})

		if m.viewport.YOffset() <= initialY {
			t.Error("expected viewport to scroll down on 'j' key")
		}
	})

	t.Run("go to top on g", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.viewport.GotoBottom()

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "g"})

		if m.viewport.YOffset() != 0 {
			t.Error("expected viewport to go to top on 'g' key")
		}
	})

	t.Run("go to bottom on G", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "G"})

		if m.viewport.YOffset() == 0 {
			t.Error("expected viewport to go to bottom on 'G' key")
		}
	})

	t.Run("no command on other keys", func(t *testing.T) {
		m := newTestModel()
		_, cmd := m.handleKeyMsg(tea.KeyPressMsg{Text: "x"})

		if cmd != nil {
			t.Error("expected no command on unknown key")
		}
	})

	t.Run("d toggles details modal in header mode", func(t *testing.T) {
		m := newTestModel()
		m.layoutMode = LayoutHeader
		m.showDetailsModal = false

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "d"})
		if !m.showDetailsModal {
			t.Error("expected showDetailsModal to be true after pressing d")
		}

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "d"})
		if m.showDetailsModal {
			t.Error("expected showDetailsModal to be false after pressing d again")
		}
	})

	t.Run("d does nothing in sidebar mode", func(t *testing.T) {
		m := newTestModel()
		m.layoutMode = LayoutSidebar
		m.showDetailsModal = false

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "d"})
		if m.showDetailsModal {
			t.Error("expected showDetailsModal to remain false in sidebar mode")
		}
	})

	t.Run("esc closes details modal", func(t *testing.T) {
		m := newTestModel()
		m.layoutMode = LayoutHeader
		m.showDetailsModal = true

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: tea.KeyEscape})
		if m.showDetailsModal {
			t.Error("expected showDetailsModal to be false after pressing Esc")
		}
	})

	t.Run("scroll keys disabled when modal open", func(t *testing.T) {
		m := newTestModel()
		m.layoutMode = LayoutHeader
		m.showDetailsModal = true
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		initialY := m.viewport.YOffset()

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "j"})
		if m.viewport.YOffset() != initialY {
			t.Error("expected scroll to be disabled when modal is open")
		}
	})

	t.Run("? toggles help modal", func(t *testing.T) {
		m := newTestModel()
		m.showHelpModal = false

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "?"})
		if !m.showHelpModal {
			t.Error("expected showHelpModal to be true after pressing ?")
		}

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "?"})
		if m.showHelpModal {
			t.Error("expected showHelpModal to be false after pressing ? again")
		}
	})

	t.Run("? works in both layout modes", func(t *testing.T) {
		m := newTestModel()
		m.layoutMode = LayoutSidebar

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "?"})
		if !m.showHelpModal {
			t.Error("expected help modal to work in sidebar mode")
		}

		m = newTestModel()
		m.layoutMode = LayoutHeader

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "?"})
		if !m.showHelpModal {
			t.Error("expected help modal to work in header mode")
		}
	})

	t.Run("esc closes help modal", func(t *testing.T) {
		m := newTestModel()
		m.showHelpModal = true

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: tea.KeyEscape})
		if m.showHelpModal {
			t.Error("expected showHelpModal to be false after pressing Esc")
		}
	})

	t.Run("esc prioritizes help modal over details modal", func(t *testing.T) {
		m := newTestModel()
		m.showHelpModal = true
		m.showDetailsModal = true

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: tea.KeyEscape})
		if m.showHelpModal {
			t.Error("expected help modal to close first")
		}
		if !m.showDetailsModal {
			t.Error("expected details modal to remain open")
		}
	})

	t.Run("scroll keys disabled when help modal open", func(t *testing.T) {
		m := newTestModel()
		m.showHelpModal = true
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		initialY := m.viewport.YOffset()

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "j"})
		if m.viewport.YOffset() != initialY {
			t.Error("expected scroll to be disabled when help modal is open")
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
		if m.viewport.Width() != expectedWidth {
			t.Errorf("viewport width = %d, want %d", m.viewport.Width(), expectedWidth)
		}
	})

	t.Run("uses sidebar mode for wide terminals", func(t *testing.T) {
		m := NewModel()
		m.ready = false

		m = m.handleWindowSizeMsg(tea.WindowSizeMsg{Width: 120, Height: 50})

		if m.layoutMode != LayoutSidebar {
			t.Errorf("layoutMode = %d, want LayoutSidebar (%d)", m.layoutMode, LayoutSidebar)
		}
	})

	t.Run("uses header mode for narrow terminals", func(t *testing.T) {
		m := NewModel()
		m.ready = false

		m = m.handleWindowSizeMsg(tea.WindowSizeMsg{Width: 70, Height: 50})

		if m.layoutMode != LayoutHeader {
			t.Errorf("layoutMode = %d, want LayoutHeader (%d)", m.layoutMode, LayoutHeader)
		}
	})

	t.Run("header mode uses full width", func(t *testing.T) {
		m := NewModel()
		m.ready = false

		m = m.handleWindowSizeMsg(tea.WindowSizeMsg{Width: 70, Height: 50})

		// Content width should be total - 2 (padding)
		expectedWidth := 70 - 2
		if m.viewport.Width() != expectedWidth {
			t.Errorf("viewport width = %d, want %d", m.viewport.Width(), expectedWidth)
		}
	})

	t.Run("header mode adjusts height for header", func(t *testing.T) {
		m := NewModel()
		m.ready = false

		m = m.handleWindowSizeMsg(tea.WindowSizeMsg{Width: 70, Height: 50})

		// Content height should be total - headerHeight - 1
		expectedHeight := 50 - headerHeight - 1
		if m.viewport.Height() != expectedHeight {
			t.Errorf("viewport height = %d, want %d", m.viewport.Height(), expectedHeight)
		}
	})

	t.Run("switches mode at breakpoint", func(t *testing.T) {
		m := NewModel()

		// At breakpoint - should be sidebar mode
		m = m.handleWindowSizeMsg(tea.WindowSizeMsg{Width: 80, Height: 50})
		if m.layoutMode != LayoutSidebar {
			t.Errorf("at breakpoint (80): layoutMode = %d, want LayoutSidebar", m.layoutMode)
		}

		// Just below breakpoint - should be header mode
		m = m.handleWindowSizeMsg(tea.WindowSizeMsg{Width: 79, Height: 50})
		if m.layoutMode != LayoutHeader {
			t.Errorf("below breakpoint (79): layoutMode = %d, want LayoutHeader", m.layoutMode)
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

func TestFollowMode(t *testing.T) {
	t.Run("default follow mode is on", func(t *testing.T) {
		m := newTestModel()
		if !m.followMode {
			t.Error("expected follow mode to be on by default")
		}
	})

	t.Run("f toggles follow mode off", func(t *testing.T) {
		m := newTestModel()
		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "f"})
		if m.followMode {
			t.Error("expected follow mode to be off after pressing f")
		}
	})

	t.Run("f toggles follow mode back on", func(t *testing.T) {
		m := newTestModel()
		m.followMode = false
		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "f"})
		if !m.followMode {
			t.Error("expected follow mode to be on after pressing f")
		}
	})

	t.Run("f disabled when help modal open", func(t *testing.T) {
		m := newTestModel()
		m.showHelpModal = true
		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "f"})
		if !m.followMode {
			t.Error("expected follow mode unchanged when help modal is open")
		}
	})

	t.Run("f disabled when details modal open", func(t *testing.T) {
		m := newTestModel()
		m.showDetailsModal = true
		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "f"})
		if !m.followMode {
			t.Error("expected follow mode unchanged when details modal is open")
		}
	})

	t.Run("scroll up disables follow mode", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.viewport.GotoBottom()
		m.followMode = true

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "k"})
		if m.followMode {
			t.Error("expected follow mode to be off after scrolling up")
		}
	})

	t.Run("scroll down does not change follow mode", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.followMode = false

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "j"})
		if m.followMode {
			t.Error("expected follow mode to remain off after scrolling down")
		}
	})

	t.Run("G re-enables follow mode", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.followMode = false

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "G"})
		if !m.followMode {
			t.Error("expected follow mode to be on after pressing G")
		}
	})

	t.Run("g disables follow mode", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.followMode = true

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "g"})
		if m.followMode {
			t.Error("expected follow mode to be off after pressing g")
		}
	})

	t.Run("pgup disables follow mode", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.viewport.GotoBottom()
		m.followMode = true

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: tea.KeyPgUp})
		if m.followMode {
			t.Error("expected follow mode to be off after page up")
		}
	})
}
