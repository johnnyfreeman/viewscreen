package tui

import (
	"bufio"
	"io"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/johnnyfreeman/viewscreen/events"
	"github.com/johnnyfreeman/viewscreen/state"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/tools"
)

// Model is the main Bubbletea model for the TUI
type Model struct {
	width             int
	height            int
	viewport          viewport.Model
	spinner           spinner.Model
	state             *state.State
	content           *strings.Builder // pointer to avoid copy issues
	stdinDone         bool
	scanner           *bufio.Scanner
	sidebarStyles     SidebarStyles
	headerStyles      HeaderStyles
	layoutMode        LayoutMode
	showDetailsModal  bool
	showHelpModal     bool
	processor         *events.EventProcessor
	search            Search
	promptEditor      PromptEditor
	followMode        bool                 // auto-scroll to bottom on new content
	autoExit          bool                 // --auto-exit flag enabled
	autoExitRemaining int                  // seconds left in countdown, 0 = inactive
	autoExitCanceled  bool                 // user interacted before auto-exit could start
	showParseErrors   bool                 // show malformed stream-json lines in content
	streamErr         error                // non-nil when stdin ended because of a scanner/read error
	claudeProcess     managedClaudeProcess // non-nil when we spawned claude
	claudeStarter     claudeProcessStarter // starts replacement claude runs for prompt edits
	prompt            string               // the prompt used to spawn claude
	inputReader       io.Reader            // where to read stream-json lines (defaults to os.Stdin)
	ignoreInputUntil  time.Time            // drops startup terminal report bytes parsed as text keys
}

type managedClaudeProcess interface {
	Stdout() io.ReadCloser
	Wait() error
	Kill() error
}

type claudeProcessStarter func(prompt string) (managedClaudeProcess, error)

const (
	defaultInitialWidth  = 80
	defaultInitialHeight = 24
	maxScannerCapacity   = 10 * 1024 * 1024
	startupInputGrace    = 500 * time.Millisecond
)

// ModelOption is a functional option for configuring the Model.
type ModelOption func(*Model)

// WithInputReader sets the reader for stream-json input (defaults to os.Stdin).
func WithInputReader(r io.Reader) ModelOption {
	return func(m *Model) {
		m.inputReader = r
	}
}

// WithClaudeProcess attaches a spawned claude subprocess to the model.
func WithClaudeProcess(p managedClaudeProcess) ModelOption {
	return func(m *Model) {
		m.claudeProcess = p
	}
}

// WithClaudeStarter sets the factory used for prompt-editor re-runs.
func WithClaudeStarter(starter claudeProcessStarter) ModelOption {
	return func(m *Model) {
		m.claudeStarter = starter
	}
}

// WithAutoExit sets whether the model should auto-exit after stream completion.
func WithAutoExit(enabled bool) ModelOption {
	return func(m *Model) {
		m.autoExit = enabled
	}
}

// WithVerboseParseErrors controls whether malformed input appears in the TUI.
func WithVerboseParseErrors(enabled bool) ModelOption {
	return func(m *Model) {
		m.showParseErrors = enabled
	}
}

// WithPrompt sets the initial prompt.
func WithPrompt(prompt string) ModelOption {
	return func(m *Model) {
		m.prompt = prompt
	}
}

// WithInitialSize seeds the model with the terminal size known by the runtime.
// Bubble Tea delivers its own resize message later, but the model's first render
// happens before that message is processed.
func WithInitialSize(width, height int) ModelOption {
	return func(m *Model) {
		if width > 0 {
			m.width = width
		}
		if height > 0 {
			m.height = height
		}
		m.updateLayoutMode()
	}
}

// WithStartupInputGrace ignores printable key input briefly while terminals
// answer Bubble Tea's startup capability probes.
func WithStartupInputGrace(d time.Duration) ModelOption {
	return func(m *Model) {
		if d > 0 {
			m.ignoreInputUntil = time.Now().Add(d)
		}
	}
}

// NewModel creates a new TUI model
func NewModel(opts ...ModelOption) Model {
	// Initialize spinner with Dot spinner and lipgloss styling.
	// We use lipgloss here (not Ultraviolet) because the spinner output
	// goes through bubbletea/lipgloss rendering pipeline.
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(string(style.CurrentTheme.Accent)))),
	)

	st := state.NewState()
	vp := viewport.New()
	vp.KeyMap = viewport.KeyMap{} // Disable viewport key handling; model handles all keys
	m := Model{
		width:         defaultInitialWidth,
		height:        defaultInitialHeight,
		spinner:       s,
		state:         st,
		content:       &strings.Builder{},
		viewport:      vp,
		sidebarStyles: NewSidebarStyles(),
		headerStyles:  NewHeaderStyles(),
		layoutMode:    LayoutSidebar, // default to sidebar mode
		processor:     events.NewEventProcessor(st),
		search:        NewSearch(),
		promptEditor:  NewPromptEditor(),
		followMode:    true, // auto-scroll to bottom by default
	}

	for _, opt := range opts {
		opt(&m)
	}

	m.updateLayoutMode()

	// Set prompt on state if provided
	if m.prompt != "" {
		m.state.Prompt = m.prompt
	}

	// Default input reader to os.Stdin
	if m.inputReader == nil {
		m.inputReader = os.Stdin
	}

	m.updateViewportDimensions()

	m.scanner = newStreamScanner(m.inputReader)

	return m
}

func newStreamScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, maxScannerCapacity)
	scanner.Buffer(buf, maxScannerCapacity)
	return scanner
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		ReadStdinLine(m.scanner),
	)
}

// Update handles messages by dispatching to focused handlers.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		m, cmd = m.handleKeyMsg(msg)
		if cmd != nil {
			return m, cmd // KeyMsg may return tea.Quit
		}

	case tea.KeyReleaseMsg:
		return m, nil

	case tea.WindowSizeMsg:
		m = m.handleWindowSizeMsg(msg)

	case tea.MouseWheelMsg:
		return m.handleMouseWheelMsg(msg)

	case spinner.TickMsg:
		m, cmd = m.handleSpinnerTick(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case RawLineMsg:
		m, cmd = m.handleRawLine(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case StdinClosedMsg:
		m, cmd = m.handleStdinClosed(msg.Err)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case AutoExitTickMsg:
		m, cmd = m.handleAutoExitTick()
		if cmd != nil {
			return m, cmd
		}

	case ClaudeExitedMsg:
		m = m.handleClaudeExited(msg)

	case RerunMsg:
		m, cmd = m.handleRerun(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case events.ParseError:
		m = m.handleParseError(msg)
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// processEvent handles a parsed event message
func (m Model) processEvent(msg tea.Msg) (Model, tea.Cmd) {
	// The message is already an events.Event from ParseEvent
	event, ok := msg.(events.Event)
	if !ok {
		return m, nil
	}

	// Process the event through the EventProcessor
	result := m.processor.Process(event)

	// Append rendered content
	if result.Rendered != "" {
		m.content.WriteString(result.Rendered)
	}
	m.updateSearchMatches()

	// Update viewport based on whether there are pending tools
	if result.HasPendingTools {
		m.updateViewportWithPendingTools()
	} else {
		m.viewport.SetContent(m.content.String())
	}
	if m.followMode {
		m.viewport.GotoBottom()
	}

	return m, nil
}

// View renders the TUI
func (m Model) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true
	v.SetContent(m.renderLayout())
	return v
}

// updateViewportWithPendingTools updates the viewport content, rendering pending tools with spinner
func (m *Model) updateViewportWithPendingTools() {
	m.viewport.SetContent(m.visibleContent())
}

// visibleContent returns the complete text currently shown in the viewport,
// including transient pending tool headers that have not yet resolved.
func (m *Model) visibleContent() string {
	content := m.content.String()
	if !m.processor.HasPendingTools() {
		return content
	}

	var sb strings.Builder
	sb.WriteString(content)
	m.processor.ForEachPendingTool(func(id string, pending tools.PendingTool) {
		// Render pending tools with spinner instead of bullet.
		// Apply Ultraviolet styling to the spinner for proper style/content separation.
		sb.WriteString(m.processor.RenderPendingTool(pending, m.spinner.View()))
	})
	return sb.String()
}

// updateSearchMatches keeps the search status in sync as streamed content grows.
func (m *Model) updateSearchMatches() {
	if !m.search.HasQuery() {
		return
	}
	m.search.UpdateMatchesPreservingSelection(m.visibleContent())
}

func (m *Model) appendStreamError(err error) {
	if err == nil {
		return
	}
	content := m.content.String()
	if content != "" && !strings.HasSuffix(content, "\n") {
		m.content.WriteString("\n")
	}
	m.content.WriteString(style.ErrorText("Input error: "))
	m.content.WriteString(err.Error())
	m.content.WriteString("\n")
	m.updateSearchMatches()
	m.viewport.SetContent(m.content.String())
	if m.followMode {
		m.viewport.GotoBottom()
	}
}

// scrollPosition returns the current scroll position from the viewport.
func (m Model) scrollPosition() ScrollPosition {
	return ScrollPosition{
		AtTop:    m.viewport.AtTop(),
		AtBottom: m.viewport.AtBottom(),
		Percent:  m.viewport.ScrollPercent(),
	}
}

func (m *Model) stopClaudeProcessIfRunning() {
	if m.stdinDone || m.claudeProcess == nil {
		return
	}
	_ = m.claudeProcess.Kill()
}

func (m Model) quitCommand() (Model, tea.Cmd) {
	m.stopClaudeProcessIfRunning()
	return m, tea.Quit
}

// renderLayout composes the main content area and sidebar/header
func (m Model) renderLayout() string {
	// Help modal overlays both layout modes
	if m.showHelpModal {
		return RenderContextualHelpModal(m.width, m.height, m.headerStyles, m.autoExitRemaining > 0, m.layoutMode, m.canEditPrompt())
	}

	// Render search bar and prompt bar if active
	searchBar := RenderSearchBar(m.search, m.viewport.Width())
	promptBar := RenderPromptBar(m.promptEditor, m.viewport.Width())
	scrollPos := m.scrollPosition()

	switch m.layoutMode {
	case LayoutHeader:
		// Header mode: single-line header on top, content below at full width
		header := RenderHeader(m.state, m.width, m.followMode, scrollPos, m.stdinDone, m.autoExitRemaining, m.streamErr)
		parts := []string{header, m.viewport.View()}
		if searchBar != "" {
			parts = append(parts, searchBar)
		}
		if promptBar != "" {
			parts = append(parts, promptBar)
		}
		layout := lipgloss.JoinVertical(lipgloss.Left, parts...)

		// Overlay modal if showing details
		if m.showDetailsModal {
			modal := RenderDetailsModal(m.state, m.spinner, m.width, m.height, m.headerStyles, m.followMode, scrollPos, m.stdinDone, m.autoExitRemaining, m.streamErr)
			return modal
		}
		return layout
	default:
		// Sidebar mode: content left, sidebar right
		sidebar := RenderSidebar(m.state, m.spinner, m.height, m.sidebarStyles, m.followMode, scrollPos, m.stdinDone, m.autoExitRemaining, m.streamErr)
		mainParts := []string{m.viewport.View()}
		if searchBar != "" {
			mainParts = append(mainParts, searchBar)
		}
		if promptBar != "" {
			mainParts = append(mainParts, promptBar)
		}
		mainContent := lipgloss.JoinVertical(lipgloss.Left, mainParts...)
		return lipgloss.JoinHorizontal(lipgloss.Top, mainContent, sidebar)
	}
}

// Prompt returns the current prompt value from the editor (or state).
func (m Model) Prompt() string {
	if m.promptEditor.Value != "" {
		return m.promptEditor.Value
	}
	return m.state.Prompt
}

func (m Model) shouldIgnoreKeyInputNoise(msg tea.KeyMsg) bool {
	if isTerminalReportFragment(keyInputText(msg)) || isCodeOnlyPrintableKey(msg) {
		return true
	}
	return !m.ignoreInputUntil.IsZero() &&
		time.Now().Before(m.ignoreInputUntil) &&
		isPrintableInputText(keyInputText(msg))
}
