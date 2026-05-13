package tui

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/johnnyfreeman/viewscreen/events"
	"github.com/johnnyfreeman/viewscreen/state"
)

var errClaudeStarterUnavailable = errors.New("claude starter unavailable")

// handleKeyMsg processes keyboard input and returns the model and any command.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	if isCtrlCKey(msg) {
		return m.quitCommand()
	}

	// During auto-exit countdown:
	// - q: quit immediately
	// - space/enter: skip countdown and exit (continue the loop)
	// - any other key: cancel countdown and browse
	if m.autoExitRemaining > 0 {
		switch {
		case isPlainTextKey(msg, "q"), isEnterKey(msg), isSpaceKey(msg):
			return m.quitCommand()
		default:
			m.cancelAutoExit()
		}
	} else {
		m.noteUserInteraction()
	}

	if isEscKey(msg) {
		return m.handleEscKey()
	}

	// When prompt editor is active, capture all keys for prompt editing
	if m.promptEditor.Active {
		return m.handlePromptEditorKeyMsg(msg)
	}

	// When search input is active, capture all keys for the search query
	if m.search.Active {
		return m.handleSearchKeyMsg(msg)
	}

	// Keys that work regardless of modal state
	switch {
	case isPlainTextKey(msg, "q"):
		return m.quitCommand()
	case isPlainTextKey(msg, "?"):
		m.showHelpModal = !m.showHelpModal
		return m, nil
	case isPlainTextKey(msg, "d"):
		if m.layoutMode == LayoutHeader && !m.showHelpModal {
			m.showDetailsModal = !m.showDetailsModal
		}
		return m, nil
	}

	// All remaining keys are navigation/action keys blocked by modals
	if m.showDetailsModal || m.showHelpModal {
		return m, nil
	}

	switch {
	case isPlainTextKey(msg, "f"):
		m.followMode = !m.followMode
		if m.followMode {
			m.viewport.GotoBottom()
		}
	case isPlainTextKey(msg, "/"):
		m.search.Enter()
		m.updateViewportDimensions()
	case isPlainTextKey(msg, "n"):
		if m.search.HasQuery() {
			m.search.NextMatch()
			m.scrollToSearchMatch()
		}
	case isPlainTextKey(msg, "N"):
		if m.search.HasQuery() {
			m.search.PrevMatch()
			m.scrollToSearchMatch()
		}
	case isPlainTextKey(msg, "e"):
		if m.canEditPrompt() {
			m.promptEditor.Enter(m.state.Prompt)
			m.updateViewportDimensions()
		}
	case msg.String() == "up", isPlainTextKey(msg, "k"):
		m.followMode = false
		m.viewport.ScrollUp(1)
	case msg.String() == "down", isPlainTextKey(msg, "j"):
		wasAtBottom := m.viewport.AtBottom()
		m.viewport.ScrollDown(1)
		m.resumeFollowModeAtBottom(wasAtBottom)
	case msg.String() == "pgup":
		m.followMode = false
		m.viewport.HalfPageUp()
	case msg.String() == "pgdown":
		wasAtBottom := m.viewport.AtBottom()
		m.viewport.HalfPageDown()
		m.resumeFollowModeAtBottom(wasAtBottom)
	case msg.String() == "home", isPlainTextKey(msg, "g"):
		m.followMode = false
		m.viewport.GotoTop()
	case msg.String() == "end", isPlainTextKey(msg, "G"):
		m.followMode = true
		m.viewport.GotoBottom()
	}
	return m, nil
}

func (m Model) handleEscKey() (Model, tea.Cmd) {
	switch {
	case m.promptEditor.Active:
		m.promptEditor.Cancel(m.state.Prompt)
		m.updateViewportDimensions()
	case m.search.Active || m.search.HasQuery():
		m.search.Clear()
		m.updateViewportDimensions()
	case m.showHelpModal:
		m.showHelpModal = false
	case m.showDetailsModal:
		m.showDetailsModal = false
	}
	return m, nil
}

func isCtrlCKey(msg tea.KeyMsg) bool {
	key := msg.Key()
	return msg.String() == "ctrl+c" ||
		key.Code == 0x03 ||
		(key.Mod.Contains(tea.ModCtrl) && (key.Code == 'c' || key.Code == 'C'))
}

func isEscKey(msg tea.KeyMsg) bool {
	return msg.String() == "esc" || msg.String() == "escape" || msg.Key().Code == tea.KeyEscape
}

func isEnterKey(msg tea.KeyMsg) bool {
	return msg.String() == "enter" || msg.Key().Code == tea.KeyEnter
}

func isSpaceKey(msg tea.KeyMsg) bool {
	return msg.String() == "space" || isPlainTextKey(msg, " ")
}

func isPlainTextKey(msg tea.KeyMsg, text string) bool {
	key := msg.Key()
	if !hasOnlyTextModifiers(key.Mod) {
		return false
	}
	if key.Text == text {
		return true
	}
	return key.Text == "" && len(text) == 1 && key.Code == []rune(text)[0]
}

func keyInputText(msg tea.KeyMsg) string {
	key := msg.Key()
	if !hasOnlyTextModifiers(key.Mod) {
		return ""
	}
	return key.Text
}

func hasOnlyTextModifiers(mod tea.KeyMod) bool {
	return mod&^(tea.ModShift|tea.ModCapsLock|tea.ModNumLock|tea.ModScrollLock) == 0
}

func isPrintableInputText(text string) bool {
	if text == "" {
		return false
	}
	for len(text) > 0 {
		r, size := utf8.DecodeRuneInString(text)
		if r == utf8.RuneError && size == 1 {
			return false
		}
		if unicode.IsControl(r) && r != '\r' && r != '\n' {
			return false
		}
		text = text[size:]
	}
	return true
}

// handleMouseWheelMsg keeps mouse scrolling aligned with keyboard navigation.
func (m Model) handleMouseWheelMsg(msg tea.MouseWheelMsg) (Model, tea.Cmd) {
	if m.autoExitRemaining > 0 {
		m.cancelAutoExit()
	} else {
		m.noteUserInteraction()
	}

	if m.showDetailsModal || m.showHelpModal {
		return m, nil
	}

	verticalWheel := !msg.Mod.Contains(tea.ModShift) &&
		(msg.Button == tea.MouseWheelUp || msg.Button == tea.MouseWheelDown)
	wasAtBottom := m.viewport.AtBottom()
	if verticalWheel && msg.Button == tea.MouseWheelUp {
		m.followMode = false
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	if verticalWheel && msg.Button == tea.MouseWheelDown {
		m.resumeFollowModeAtBottom(wasAtBottom)
	}

	return m, cmd
}

// resumeFollowModeAtBottom keeps manual browsing from getting stuck in a
// confusing "paused at bottom" state after downward navigation reaches bottom.
func (m *Model) resumeFollowModeAtBottom(wasAtBottom bool) {
	if !wasAtBottom && m.viewport.AtBottom() {
		m.followMode = true
	}
}

// noteUserInteraction records pre-completion input so loop-friendly auto-exit
// does not later close a TUI the user has started browsing.
func (m *Model) noteUserInteraction() {
	m.autoExitCanceled = true
}

func (m *Model) cancelAutoExit() {
	m.autoExitRemaining = 0
	m.autoExitCanceled = true
}

func (m Model) shouldStartAutoExitCountdown() bool {
	if !m.autoExit || m.autoExitCanceled {
		return false
	}
	return m.followMode &&
		!m.search.Active &&
		!m.search.HasQuery() &&
		!m.promptEditor.Active &&
		!m.showHelpModal &&
		!m.showDetailsModal
}

// canEditPrompt reports whether the prompt editor can perform its advertised
// action. Prompt edits only have an effect after the current stream finishes
// and when this TUI still has a starter capable of launching another run.
func (m Model) canEditPrompt() bool {
	return m.stdinDone && m.claudeStarter != nil
}

// handlePromptEditorKeyMsg processes keyboard input while prompt editing is active.
func (m Model) handlePromptEditorKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case isEnterKey(msg):
		// Confirm the edited prompt
		m.state.Prompt = m.promptEditor.Value
		m.prompt = m.promptEditor.Value
		m.promptEditor.Exit()
		m.updateViewportDimensions()
		if m.claudeStarter != nil {
			return m, func() tea.Msg { return RerunMsg{Prompt: m.state.Prompt} }
		}
	case msg.String() == "backspace":
		m.promptEditor.Backspace()
	case msg.String() == "delete":
		m.promptEditor.Delete()
	case msg.String() == "left":
		m.promptEditor.CursorLeft()
	case msg.String() == "right":
		m.promptEditor.CursorRight()
	case msg.String() == "home", msg.String() == "ctrl+a":
		m.promptEditor.CursorHome()
	case msg.String() == "end", msg.String() == "ctrl+e":
		m.promptEditor.CursorEnd()
	default:
		if text := keyInputText(msg); isPrintableInputText(text) {
			for _, r := range text {
				m.promptEditor.TypeRune(r)
			}
		}
	}
	return m, nil
}

// handleSearchKeyMsg processes keyboard input while search mode is active.
func (m Model) handleSearchKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case isEnterKey(msg):
		// Confirm search, exit input mode but keep query
		m.search.Exit()
		m.updateViewportDimensions()
	case msg.String() == "backspace":
		m.search.Backspace()
		m.search.UpdateMatches(m.visibleContent())
		m.scrollToSearchMatch()
	default:
		// Type character into search query
		if text := keyInputText(msg); isPrintableInputText(text) {
			m.search.TypeText(text)
			m.search.UpdateMatches(m.visibleContent())
			m.scrollToSearchMatch()
		}
	}
	return m, nil
}

// scrollToSearchMatch scrolls the viewport to show the current search match.
func (m *Model) scrollToSearchMatch() {
	line := m.search.CurrentLine()
	if line < 0 {
		return
	}
	m.followMode = false
	m.viewport.SetYOffset(line)
}

// updateViewportDimensions recalculates the viewport size for the current
// terminal dimensions and active bottom bars.
func (m *Model) updateViewportDimensions() {
	if m.width == 0 || m.height == 0 {
		return
	}

	m.updateLayoutMode()

	var contentWidth, contentHeight int
	switch m.layoutMode {
	case LayoutHeader:
		contentWidth = max(m.width-2, 1)
		contentHeight = m.height - headerHeight - 1
	default:
		contentWidth = max(m.width-sidebarWidth-3, 1)
		contentHeight = m.height - 2
	}

	if m.search.Active || m.search.HasQuery() {
		contentHeight--
	}
	if m.promptEditor.Active {
		contentHeight--
	}

	m.viewport.SetWidth(contentWidth)
	m.viewport.SetHeight(max(contentHeight, 1))
	m.processor.SetWidth(contentWidth)

	if m.processor.HasPendingTools() {
		m.updateViewportWithPendingTools()
	} else {
		m.viewport.SetContent(m.content.String())
	}
	if m.followMode {
		m.viewport.GotoBottom()
	}
}

func (m *Model) updateLayoutMode() {
	if m.width < breakpointWidth {
		m.layoutMode = LayoutHeader
		return
	}
	m.layoutMode = LayoutSidebar
}

// handleWindowSizeMsg processes terminal resize events.
func (m Model) handleWindowSizeMsg(msg tea.WindowSizeMsg) Model {
	m.width = msg.Width
	m.height = msg.Height

	m.updateViewportDimensions()
	return m
}

// handleSpinnerTick processes spinner animation ticks.
func (m Model) handleSpinnerTick(msg spinner.TickMsg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)

	// Refresh viewport to animate spinner for pending tools
	if m.processor.HasPendingTools() {
		m.updateViewportWithPendingTools()
	}

	return m, cmd
}

// handleRawLine processes a line read from stdin.
func (m Model) handleRawLine(msg RawLineMsg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Parse the line and dispatch appropriate message
	parsedMsg := ParseEvent(msg.Line)
	if parsedMsg != nil {
		if parseErr, ok := parsedMsg.(events.ParseError); ok {
			m = m.handleParseError(parseErr)
		} else {
			// Process the parsed event immediately
			var cmd tea.Cmd
			m, cmd = m.processEvent(parsedMsg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	// Continue reading stdin
	cmds = append(cmds, ReadStdinLine(m.scanner))

	return m, tea.Batch(cmds...)
}

// handleStdinClosed processes the stdin closed signal.
func (m Model) handleStdinClosed(err error) (Model, tea.Cmd) {
	m.stdinDone = true
	m.streamErr = err
	if err != nil {
		m.autoExitRemaining = 0
		m.state.ClearCurrentTool()
		m.appendStreamError(err)
		return m, nil
	}
	if m.shouldStartAutoExitCountdown() {
		m.autoExitRemaining = 5
		return m, AutoExitTick()
	}
	return m, nil
}

// handleRerun kills the old claude process, resets state, and spawns a new run.
func (m Model) handleRerun(msg RerunMsg) (Model, tea.Cmd) {
	// Kill old process
	if m.claudeProcess != nil {
		m.claudeProcess.Kill()
		m.claudeProcess.Wait()
	}

	// Reset state
	m.content = &strings.Builder{}
	m.stdinDone = false
	m.streamErr = nil
	m.autoExitRemaining = 0
	m.autoExitCanceled = false
	st := state.NewState()
	st.Prompt = msg.Prompt
	m.state = st
	m.processor = events.NewEventProcessor(st)
	m.prompt = msg.Prompt
	m.followMode = true
	m.claudeProcess = nil
	m.search.Clear()
	m.updateViewportDimensions()

	// Spawn new claude process
	if m.claudeStarter == nil {
		m.failRerunStart(errClaudeStarterUnavailable)
		return m, nil
	}
	proc, err := m.claudeStarter(msg.Prompt)
	if err != nil {
		m.failRerunStart(err)
		return m, nil
	}
	stdout := proc.Stdout()
	if stdout == nil {
		_ = proc.Kill()
		_ = proc.Wait()
		m.failRerunStart(errors.New("claude stdout unavailable"))
		return m, nil
	}

	m.claudeProcess = proc
	m.inputReader = stdout
	m.scanner = newStreamScanner(m.inputReader)

	// Clear viewport
	m.viewport.SetContent("")

	return m, ReadStdinLine(m.scanner)
}

func (m *Model) failRerunStart(err error) {
	m.content.WriteString("Error starting claude: " + err.Error() + "\n")
	m.viewport.SetContent(m.content.String())
	m.stdinDone = true
}

// handleAutoExitTick processes a countdown tick for auto-exit.
func (m Model) handleAutoExitTick() (Model, tea.Cmd) {
	if m.autoExitRemaining <= 0 {
		return m, nil
	}
	m.autoExitRemaining--
	if m.autoExitRemaining <= 0 {
		return m, tea.Quit
	}
	return m, AutoExitTick()
}

// handleParseError processes event parsing errors.
func (m Model) handleParseError(msg events.ParseError) Model {
	if m.showParseErrors {
		m.content.WriteString("Parse error: " + msg.Line + "\n")
		m.updateSearchMatches()
		m.viewport.SetContent(m.content.String())
		if m.followMode {
			m.viewport.GotoBottom()
		}
	}
	return m
}
