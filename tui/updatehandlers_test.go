package tui

import (
	"errors"
	"io"
	"strings"
	"testing"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/johnnyfreeman/viewscreen/assistant"
	"github.com/johnnyfreeman/viewscreen/events"
	"github.com/johnnyfreeman/viewscreen/types"
)

type fakeClaudeProcess struct {
	killCount int
	waitCount int
	stdout    io.ReadCloser
}

func (p *fakeClaudeProcess) Stdout() io.ReadCloser { return p.stdout }
func (p *fakeClaudeProcess) Wait() error {
	p.waitCount++
	return nil
}
func (p *fakeClaudeProcess) Kill() error {
	p.killCount++
	return nil
}

func newTestModel() Model {
	m := NewModel()
	// Simulate window size to initialize viewport
	m.width = 100
	m.height = 40
	m.viewport.SetWidth(100)
	m.viewport.SetHeight(40)
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

	t.Run("quit kills running spawned claude process", func(t *testing.T) {
		proc := &fakeClaudeProcess{}
		m := newTestModel()
		m.claudeProcess = proc
		m.stdinDone = false

		_, cmd := m.handleKeyMsg(tea.KeyPressMsg{Text: "q"})

		if cmd == nil {
			t.Error("expected quit command on q")
		}
		if proc.killCount != 1 {
			t.Errorf("Kill called %d times, want 1", proc.killCount)
		}
	})

	t.Run("quit does not kill completed spawned claude process", func(t *testing.T) {
		proc := &fakeClaudeProcess{}
		m := newTestModel()
		m.claudeProcess = proc
		m.stdinDone = true

		_, cmd := m.handleKeyMsg(tea.KeyPressMsg{Text: "q"})

		if cmd == nil {
			t.Error("expected quit command on q")
		}
		if proc.killCount != 0 {
			t.Errorf("Kill called %d times, want 0", proc.killCount)
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

	t.Run("d does not toggle hidden details while help is open", func(t *testing.T) {
		m := newTestModel()
		m.layoutMode = LayoutHeader
		m.showHelpModal = true
		m.showDetailsModal = false

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "d"})
		if m.showDetailsModal {
			t.Error("expected details modal to remain hidden while help modal has focus")
		}
		if !m.showHelpModal {
			t.Error("expected help modal to remain open")
		}

		m.showDetailsModal = true
		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "d"})
		if !m.showDetailsModal {
			t.Error("expected existing details modal state to remain unchanged behind help")
		}
		if !m.showHelpModal {
			t.Error("expected help modal to remain open")
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

	t.Run("shifted ? toggles help modal", func(t *testing.T) {
		m := newTestModel()

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "?", Mod: tea.ModShift})

		if !m.showHelpModal {
			t.Error("expected showHelpModal to be true after shifted ?")
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

func TestViewTerminalModes(t *testing.T) {
	m := newTestModel()

	view := m.View()

	if !view.AltScreen {
		t.Error("expected TUI view to use alt screen")
	}
	if view.MouseMode != tea.MouseModeNone {
		t.Errorf("expected mouse mode to be disabled, got %v", view.MouseMode)
	}
}

func TestHandleWindowSizeMsg(t *testing.T) {
	t.Run("starts with usable fallback dimensions", func(t *testing.T) {
		m := NewModel()

		if m.width != defaultInitialWidth {
			t.Errorf("width = %d, want %d", m.width, defaultInitialWidth)
		}
		if m.height != defaultInitialHeight {
			t.Errorf("height = %d, want %d", m.height, defaultInitialHeight)
		}
		if m.viewport.Width() <= 0 {
			t.Errorf("viewport width = %d, want positive", m.viewport.Width())
		}
		if m.viewport.Height() <= 0 {
			t.Errorf("viewport height = %d, want positive", m.viewport.Height())
		}
	})

	t.Run("uses provided initial dimensions before first resize message", func(t *testing.T) {
		m := NewModel(WithInitialSize(132, 43))

		if m.width != 132 {
			t.Errorf("width = %d, want 132", m.width)
		}
		if m.height != 43 {
			t.Errorf("height = %d, want 43", m.height)
		}
		if m.viewport.Width() != 132-sidebarWidth-3 {
			t.Errorf("viewport width = %d, want %d", m.viewport.Width(), 132-sidebarWidth-3)
		}
		if m.viewport.Height() != 41 {
			t.Errorf("viewport height = %d, want 41", m.viewport.Height())
		}
	})

	t.Run("provided narrow initial dimensions select header layout", func(t *testing.T) {
		m := NewModel(WithInitialSize(70, 30))

		if m.layoutMode != LayoutHeader {
			t.Errorf("layoutMode = %d, want LayoutHeader (%d)", m.layoutMode, LayoutHeader)
		}
		if m.viewport.Width() != 68 {
			t.Errorf("viewport width = %d, want 68", m.viewport.Width())
		}
	})

	t.Run("sets dimensions on size message", func(t *testing.T) {
		m := NewModel()

		m = m.handleWindowSizeMsg(tea.WindowSizeMsg{Width: 120, Height: 50})

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

		m = m.handleWindowSizeMsg(tea.WindowSizeMsg{Width: 120, Height: 50})

		// Content width should be total - sidebar - border
		expectedWidth := 120 - sidebarWidth - 3
		if m.viewport.Width() != expectedWidth {
			t.Errorf("viewport width = %d, want %d", m.viewport.Width(), expectedWidth)
		}
	})

	t.Run("uses sidebar mode for wide terminals", func(t *testing.T) {
		m := NewModel()

		m = m.handleWindowSizeMsg(tea.WindowSizeMsg{Width: 120, Height: 50})

		if m.layoutMode != LayoutSidebar {
			t.Errorf("layoutMode = %d, want LayoutSidebar (%d)", m.layoutMode, LayoutSidebar)
		}
	})

	t.Run("uses header mode for narrow terminals", func(t *testing.T) {
		m := NewModel()

		m = m.handleWindowSizeMsg(tea.WindowSizeMsg{Width: 70, Height: 50})

		if m.layoutMode != LayoutHeader {
			t.Errorf("layoutMode = %d, want LayoutHeader (%d)", m.layoutMode, LayoutHeader)
		}
	})

	t.Run("header mode uses full width", func(t *testing.T) {
		m := NewModel()

		m = m.handleWindowSizeMsg(tea.WindowSizeMsg{Width: 70, Height: 50})

		// Content width should be total - 2 (padding)
		expectedWidth := 70 - 2
		if m.viewport.Width() != expectedWidth {
			t.Errorf("viewport width = %d, want %d", m.viewport.Width(), expectedWidth)
		}
	})

	t.Run("tiny header mode keeps viewport inside terminal", func(t *testing.T) {
		m := NewModel()

		m = m.handleWindowSizeMsg(tea.WindowSizeMsg{Width: 10, Height: 5})

		if m.layoutMode != LayoutHeader {
			t.Fatalf("layoutMode = %d, want LayoutHeader (%d)", m.layoutMode, LayoutHeader)
		}
		if m.viewport.Width() != 8 {
			t.Errorf("viewport width = %d, want 8", m.viewport.Width())
		}
		if m.viewport.Width() > m.width {
			t.Errorf("viewport width = %d exceeds terminal width %d", m.viewport.Width(), m.width)
		}
	})

	t.Run("header mode adjusts height for header", func(t *testing.T) {
		m := NewModel()

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

		_, cmd := m.handleSpinnerTick(spinner.TickMsg{})

		if cmd == nil {
			t.Error("expected spinner command after tick")
		}
	})
}

func TestHandleStdinClosed(t *testing.T) {
	t.Run("sets stdinDone", func(t *testing.T) {
		m := newTestModel()

		m, _ = m.handleStdinClosed(nil)

		if !m.stdinDone {
			t.Error("expected stdinDone to be true")
		}
	})

	t.Run("starts countdown when autoExit enabled", func(t *testing.T) {
		m := newTestModel()
		m.autoExit = true

		m, cmd := m.handleStdinClosed(nil)

		if !m.stdinDone {
			t.Error("expected stdinDone to be true")
		}
		if m.autoExitRemaining != 5 {
			t.Errorf("expected autoExitRemaining=5, got %d", m.autoExitRemaining)
		}
		if cmd == nil {
			t.Error("expected tick command when autoExit enabled")
		}
	})

	t.Run("skips countdown after user scrolled before stdin closed", func(t *testing.T) {
		m := newTestModel()
		m.autoExit = true
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.viewport.GotoBottom()

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "k"})
		m, cmd := m.handleStdinClosed(nil)

		if m.autoExitRemaining != 0 {
			t.Errorf("expected autoExitRemaining=0 after pre-close scroll, got %d", m.autoExitRemaining)
		}
		if cmd != nil {
			t.Error("expected no tick command after pre-close scroll")
		}
	})

	t.Run("skips countdown when search is already active", func(t *testing.T) {
		m := newTestModel()
		m.autoExit = true

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "/"})
		m, cmd := m.handleStdinClosed(nil)

		if m.autoExitRemaining != 0 {
			t.Errorf("expected autoExitRemaining=0 while search is active, got %d", m.autoExitRemaining)
		}
		if cmd != nil {
			t.Error("expected no tick command while search is active")
		}
	})

	t.Run("no countdown when autoExit disabled", func(t *testing.T) {
		m := newTestModel()
		m.autoExit = false

		m, cmd := m.handleStdinClosed(nil)

		if m.autoExitRemaining != 0 {
			t.Errorf("expected autoExitRemaining=0, got %d", m.autoExitRemaining)
		}
		if cmd != nil {
			t.Error("expected no command when autoExit disabled")
		}
	})

	t.Run("shows read error and skips auto-exit countdown", func(t *testing.T) {
		m := newTestModel()
		m.autoExit = true
		m.content.WriteString("partial output")
		m.viewport.SetContent(m.content.String())
		m.state.SetCurrentTool("Read", "file.txt")
		readErr := errors.New("bufio.Scanner: token too long")

		m, cmd := m.handleStdinClosed(readErr)

		if !m.stdinDone {
			t.Error("expected stdinDone to be true")
		}
		if !errors.Is(m.streamErr, readErr) {
			t.Fatalf("streamErr = %v, want %v", m.streamErr, readErr)
		}
		if cmd != nil {
			t.Error("expected no auto-exit tick command on read error")
		}
		if m.autoExitRemaining != 0 {
			t.Errorf("expected autoExitRemaining=0 on read error, got %d", m.autoExitRemaining)
		}
		if m.state.ToolInProgress {
			t.Error("expected tool progress to clear when input stream fails")
		}
		if got := m.content.String(); !strings.Contains(got, "Input error: bufio.Scanner: token too long") {
			t.Fatalf("content = %q, want visible input error", got)
		}
	})
}

func TestHandleRerun(t *testing.T) {
	t.Run("uses injected starter and resets stream state", func(t *testing.T) {
		oldProc := &fakeClaudeProcess{}
		newProc := &fakeClaudeProcess{stdout: io.NopCloser(strings.NewReader(""))}
		var gotPrompt string

		m := newTestModel()
		m.claudeProcess = oldProc
		m.claudeStarter = func(prompt string) (managedClaudeProcess, error) {
			gotPrompt = prompt
			return newProc, nil
		}
		m.stdinDone = true
		m.streamErr = errors.New("old error")
		m.content.WriteString("old content")
		m.search.Query = "old"
		m.search.UpdateMatches(m.content.String())

		m, cmd := m.handleRerun(RerunMsg{Prompt: "new prompt"})

		if oldProc.killCount != 1 {
			t.Errorf("old Kill called %d times, want 1", oldProc.killCount)
		}
		if oldProc.waitCount != 1 {
			t.Errorf("old Wait called %d times, want 1", oldProc.waitCount)
		}
		if gotPrompt != "new prompt" {
			t.Errorf("starter prompt = %q, want new prompt", gotPrompt)
		}
		if m.claudeProcess != newProc {
			t.Error("expected model to use new process")
		}
		if m.stdinDone {
			t.Error("expected stdinDone to reset for new stream")
		}
		if m.streamErr != nil {
			t.Errorf("streamErr = %v, want nil", m.streamErr)
		}
		if m.content.String() != "" {
			t.Errorf("content = %q, want empty after rerun", m.content.String())
		}
		if m.search.HasQuery() {
			t.Error("expected search to clear on rerun")
		}
		if m.state.Prompt != "new prompt" {
			t.Errorf("state.Prompt = %q, want new prompt", m.state.Prompt)
		}
		if cmd == nil {
			t.Error("expected read command for new process stdout")
		}
	})

	t.Run("reports missing starter instead of spawning from model code", func(t *testing.T) {
		oldProc := &fakeClaudeProcess{}
		m := newTestModel()
		m.claudeProcess = oldProc

		m, cmd := m.handleRerun(RerunMsg{Prompt: "new prompt"})

		if cmd != nil {
			t.Error("expected no read command when starter is missing")
		}
		if oldProc.killCount != 1 {
			t.Errorf("old Kill called %d times, want 1", oldProc.killCount)
		}
		if m.claudeProcess != nil {
			t.Error("expected stale process to clear after failed rerun")
		}
		if !m.stdinDone {
			t.Error("expected stdinDone after failed rerun")
		}
		if got := m.content.String(); !strings.Contains(got, "claude starter unavailable") {
			t.Fatalf("content = %q, want starter error", got)
		}
	})

	t.Run("reports missing stdout from started process", func(t *testing.T) {
		m := newTestModel()
		m.claudeStarter = func(string) (managedClaudeProcess, error) {
			return &fakeClaudeProcess{}, nil
		}

		m, cmd := m.handleRerun(RerunMsg{Prompt: "new prompt"})

		if cmd != nil {
			t.Error("expected no read command when stdout is missing")
		}
		if !m.stdinDone {
			t.Error("expected stdinDone after failed rerun")
		}
		if got := m.content.String(); !strings.Contains(got, "claude stdout unavailable") {
			t.Fatalf("content = %q, want stdout error", got)
		}
	})
}

func TestHandleAutoExitTick(t *testing.T) {
	t.Run("decrements remaining and continues", func(t *testing.T) {
		m := newTestModel()
		m.autoExitRemaining = 3

		m, cmd := m.handleAutoExitTick()

		if m.autoExitRemaining != 2 {
			t.Errorf("expected autoExitRemaining=2, got %d", m.autoExitRemaining)
		}
		if cmd == nil {
			t.Error("expected tick command to continue countdown")
		}
	})

	t.Run("quits at zero", func(t *testing.T) {
		m := newTestModel()
		m.autoExitRemaining = 1

		m, cmd := m.handleAutoExitTick()

		if m.autoExitRemaining != 0 {
			t.Errorf("expected autoExitRemaining=0, got %d", m.autoExitRemaining)
		}
		if cmd == nil {
			t.Error("expected quit command at zero")
		}
	})

	t.Run("no-op when already inactive", func(t *testing.T) {
		m := newTestModel()
		m.autoExitRemaining = 0

		m, cmd := m.handleAutoExitTick()

		if cmd != nil {
			t.Error("expected no command when countdown inactive")
		}
	})
}

func TestAutoExitCancelOnKeyPress(t *testing.T) {
	t.Run("any key cancels countdown", func(t *testing.T) {
		m := newTestModel()
		m.autoExitRemaining = 3

		m, cmd := m.handleKeyMsg(tea.KeyPressMsg{Text: "j"})

		if m.autoExitRemaining != 0 {
			t.Errorf("expected autoExitRemaining=0 after key press, got %d", m.autoExitRemaining)
		}
		if cmd != nil {
			t.Error("expected no command after cancelling countdown")
		}
	})

	t.Run("cancel key still performs normal action", func(t *testing.T) {
		m := newTestModel()
		m.autoExitRemaining = 3

		m, cmd := m.handleKeyMsg(tea.KeyPressMsg{Text: "?"})

		if m.autoExitRemaining != 0 {
			t.Errorf("expected autoExitRemaining=0 after key press, got %d", m.autoExitRemaining)
		}
		if !m.showHelpModal {
			t.Error("expected ? to open help after cancelling countdown")
		}
		if cmd != nil {
			t.Error("expected no command after opening help")
		}
	})

	t.Run("navigation key cancels countdown and scrolls", func(t *testing.T) {
		m := newTestModel()
		m.autoExitRemaining = 3
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.followMode = true

		initialY := m.viewport.YOffset()
		m, cmd := m.handleKeyMsg(tea.KeyPressMsg{Text: "j"})

		if m.autoExitRemaining != 0 {
			t.Errorf("expected autoExitRemaining=0 after key press, got %d", m.autoExitRemaining)
		}
		if m.viewport.YOffset() <= initialY {
			t.Error("expected j to scroll after cancelling countdown")
		}
		if cmd != nil {
			t.Error("expected no command after scrolling")
		}
	})

	t.Run("q still quits during countdown", func(t *testing.T) {
		m := newTestModel()
		m.autoExitRemaining = 3

		_, cmd := m.handleKeyMsg(tea.KeyPressMsg{Text: "q"})

		if cmd == nil {
			t.Error("expected quit command on q during countdown")
		}
	})

	t.Run("ctrl+c still quits during countdown", func(t *testing.T) {
		m := newTestModel()
		m.autoExitRemaining = 3

		_, cmd := m.handleKeyMsg(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})

		if cmd == nil {
			t.Error("expected quit command on ctrl+c during countdown")
		}
	})

	t.Run("space skips countdown and exits", func(t *testing.T) {
		m := newTestModel()
		m.autoExitRemaining = 3

		_, cmd := m.handleKeyMsg(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})

		if cmd == nil {
			t.Error("expected quit command on space during countdown")
		}
	})

	t.Run("enter skips countdown and exits", func(t *testing.T) {
		m := newTestModel()
		m.autoExitRemaining = 3

		_, cmd := m.handleKeyMsg(tea.KeyPressMsg{Code: tea.KeyEnter})

		if cmd == nil {
			t.Error("expected quit command on enter during countdown")
		}
	})
}

func TestMouseWheelScrolling(t *testing.T) {
	t.Run("wheel up scrolls and pauses follow mode", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.viewport.GotoBottom()
		m.followMode = true

		initialY := m.viewport.YOffset()
		updated, cmd := m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
		m = updated.(Model)

		if cmd != nil {
			t.Error("expected no command after mouse wheel scroll")
		}
		if m.viewport.YOffset() >= initialY {
			t.Error("expected mouse wheel up to scroll viewport up")
		}
		if m.followMode {
			t.Error("expected mouse wheel up to pause follow mode")
		}
	})

	t.Run("wheel down to bottom resumes follow mode", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.viewport.GotoBottom()
		m.viewport.ScrollUp(1)
		m.followMode = false

		updated, _ := m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
		m = updated.(Model)

		if !m.viewport.AtBottom() {
			t.Fatal("expected mouse wheel down to reach bottom")
		}
		if !m.followMode {
			t.Error("expected follow mode to resume after wheel down reaches bottom")
		}
	})

	t.Run("wheel down while already paused at bottom stays paused", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.viewport.GotoBottom()
		m.followMode = false

		updated, _ := m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
		m = updated.(Model)

		if m.followMode {
			t.Error("expected follow mode to stay paused when wheeling down at bottom")
		}
	})

	t.Run("wheel cancels auto-exit countdown", func(t *testing.T) {
		m := newTestModel()
		m.autoExitRemaining = 3
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.viewport.GotoBottom()

		updated, _ := m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
		m = updated.(Model)

		if m.autoExitRemaining != 0 {
			t.Errorf("expected autoExitRemaining=0 after mouse wheel, got %d", m.autoExitRemaining)
		}
	})

	t.Run("wheel before stdin closes prevents auto-exit countdown", func(t *testing.T) {
		m := newTestModel()
		m.autoExit = true
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.viewport.GotoBottom()

		updated, _ := m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
		m = updated.(Model)
		m, cmd := m.handleStdinClosed(nil)

		if m.autoExitRemaining != 0 {
			t.Errorf("expected autoExitRemaining=0 after pre-close mouse wheel, got %d", m.autoExitRemaining)
		}
		if cmd != nil {
			t.Error("expected no tick command after pre-close mouse wheel")
		}
	})

	t.Run("wheel does not scroll hidden content behind help modal", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.viewport.GotoBottom()
		m.showHelpModal = true

		initialY := m.viewport.YOffset()
		updated, _ := m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
		m = updated.(Model)

		if m.viewport.YOffset() != initialY {
			t.Errorf("viewport YOffset = %d, want %d", m.viewport.YOffset(), initialY)
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

func TestHandleRawLineShowsParseErrorsInVerboseMode(t *testing.T) {
	m := NewModel(WithVerboseParseErrors(true))
	m, _ = m.handleRawLine(RawLineMsg{Line: "not json"})

	if got := m.content.String(); !strings.Contains(got, "Parse error: not json") {
		t.Fatalf("content = %q, want verbose parse error", got)
	}
}

func TestUpdateSearchMatchesOnNewContent(t *testing.T) {
	m := newTestModel()
	m.search.Query = "needle"
	m.content.WriteString("needle one")
	m.search.UpdateMatches(m.content.String())

	m.content.WriteString("\nneedle two")
	m.updateSearchMatches()

	if m.search.MatchCount() != 2 {
		t.Errorf("MatchCount() = %d, want 2", m.search.MatchCount())
	}
}

func TestProcessEventRefreshesSearchMatches(t *testing.T) {
	m := newTestModel()
	m.content.WriteString("needle one\n")
	m.viewport.SetContent(m.content.String())
	m.search.Query = "needle"
	m.search.UpdateMatches(m.content.String())

	m, _ = m.processEvent(events.AssistantEvent{
		Data: assistant.Event{
			Message: assistant.Message{
				Content: []types.ContentBlock{
					{Type: "text", Text: "needle two"},
				},
			},
		},
	})

	if m.search.MatchCount() != 2 {
		t.Errorf("MatchCount() = %d, want 2", m.search.MatchCount())
	}
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

	t.Run("scroll down away from bottom keeps follow mode off", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.followMode = false

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "j"})
		if m.followMode {
			t.Error("expected follow mode to remain off after scrolling down")
		}
	})

	t.Run("scroll down to bottom re-enables follow mode", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.viewport.GotoBottom()
		m.viewport.ScrollUp(1)
		m.followMode = false

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "j"})
		if !m.followMode {
			t.Error("expected follow mode to turn on after scrolling down to bottom")
		}
	})

	t.Run("page down to bottom re-enables follow mode", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.viewport.GotoBottom()
		m.viewport.ScrollUp(1)
		m.followMode = false

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: tea.KeyPgDown})
		if !m.followMode {
			t.Error("expected follow mode to turn on after paging down to bottom")
		}
	})

	t.Run("scroll down while already at bottom keeps intentional pause", func(t *testing.T) {
		m := newTestModel()
		m.viewport.SetContent(strings.Repeat("line\n", 100))
		m.viewport.GotoBottom()
		m.followMode = false

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "j"})
		if m.followMode {
			t.Error("expected follow mode to stay off when already paused at bottom")
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
