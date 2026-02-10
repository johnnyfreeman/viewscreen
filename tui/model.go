package tui

import (
	"bufio"
	"io"
	"os"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	claudepkg "github.com/johnnyfreeman/viewscreen/claude"
	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/events"
	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/state"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/tools"
	"golang.org/x/term"
)

// Model is the main Bubbletea model for the TUI
type Model struct {
	width            int
	height           int
	viewport         viewport.Model
	spinner          spinner.Model
	state            *state.State
	content          *strings.Builder // pointer to avoid copy issues
	stdinDone        bool
	scanner          *bufio.Scanner
	sidebarStyles    SidebarStyles
	headerStyles     HeaderStyles
	layoutMode       LayoutMode
	showDetailsModal bool
	showHelpModal    bool
	ready            bool
	processor        *events.EventProcessor
	search            Search
	promptEditor      PromptEditor
	followMode        bool // auto-scroll to bottom on new content
	autoExit          bool // --auto-exit flag enabled
	autoExitRemaining int  // seconds left in countdown, 0 = inactive
	claudeProcess     *claudepkg.Process // non-nil when we spawned claude
	prompt            string             // the prompt used to spawn claude
	inputReader       io.Reader          // where to read stream-json lines (defaults to os.Stdin)
}

// ModelOption is a functional option for configuring the Model.
type ModelOption func(*Model)

// WithInputReader sets the reader for stream-json input (defaults to os.Stdin).
func WithInputReader(r io.Reader) ModelOption {
	return func(m *Model) {
		m.inputReader = r
	}
}

// WithClaudeProcess attaches a spawned claude subprocess to the model.
func WithClaudeProcess(p *claudepkg.Process) ModelOption {
	return func(m *Model) {
		m.claudeProcess = p
	}
}

// WithPrompt sets the initial prompt.
func WithPrompt(prompt string) ModelOption {
	return func(m *Model) {
		m.prompt = prompt
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
	m := Model{
		spinner:       s,
		state:         st,
		content:       &strings.Builder{},
		sidebarStyles: NewSidebarStyles(),
		headerStyles:  NewHeaderStyles(),
		layoutMode:    LayoutSidebar, // default to sidebar mode
		processor:     events.NewEventProcessor(st),
		search:        NewSearch(),
		promptEditor:  NewPromptEditor(),
		followMode:    true, // auto-scroll to bottom by default
		autoExit:      config.AutoExit,
	}

	for _, opt := range opts {
		opt(&m)
	}

	// Set prompt on state if provided
	if m.prompt != "" {
		m.state.Prompt = m.prompt
	}

	// Default input reader to os.Stdin
	if m.inputReader == nil {
		m.inputReader = os.Stdin
	}

	// Create scanner with large buffer
	m.scanner = bufio.NewScanner(m.inputReader)
	const maxCapacity = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxCapacity)
	m.scanner.Buffer(buf, maxCapacity)

	return m
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
	case tea.KeyMsg:
		m, cmd = m.handleKeyMsg(msg)
		if cmd != nil {
			return m, cmd // KeyMsg may return tea.Quit
		}

	case tea.WindowSizeMsg:
		m = m.handleWindowSizeMsg(msg)

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
		m, cmd = m.handleStdinClosed()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case AutoExitTickMsg:
		m, cmd = m.handleAutoExitTick()
		if cmd != nil {
			return m, cmd
		}

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
	v.MouseMode = tea.MouseModeCellMotion

	if !m.ready {
		v.SetContent("Initializing...")
		return v
	}
	v.SetContent(m.renderLayout())
	return v
}

// updateViewportWithPendingTools updates the viewport content, rendering pending tools with spinner
func (m *Model) updateViewportWithPendingTools() {
	content := m.content.String()
	// Render pending tools with spinner instead of bullet.
	// Apply Ultraviolet styling to the spinner for proper style/content separation.
	m.processor.ForEachPendingTool(func(id string, pending tools.PendingTool) {
		content += m.processor.RenderPendingTool(pending, m.spinner.View())
	})
	m.viewport.SetContent(content)
}

// scrollPosition returns the current scroll position from the viewport.
func (m Model) scrollPosition() ScrollPosition {
	return ScrollPosition{
		AtTop:   m.viewport.AtTop(),
		AtBottom: m.viewport.AtBottom(),
		Percent: m.viewport.ScrollPercent(),
	}
}

// renderLayout composes the main content area and sidebar/header
func (m Model) renderLayout() string {
	// Help modal overlays both layout modes
	if m.showHelpModal {
		return RenderHelpModal(m.width, m.height, m.headerStyles, m.autoExitRemaining > 0, m.claudeProcess != nil)
	}

	// Render search bar and prompt bar if active
	searchBar := RenderSearchBar(m.search, m.viewport.Width())
	promptBar := RenderPromptBar(m.promptEditor, m.viewport.Width())
	scrollPos := m.scrollPosition()

	switch m.layoutMode {
	case LayoutHeader:
		// Header mode: single-line header on top, content below at full width
		header := RenderHeader(m.state, m.width, m.followMode, scrollPos, m.stdinDone, m.autoExitRemaining)
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
			modal := RenderDetailsModal(m.state, m.spinner, m.width, m.height, m.headerStyles, m.followMode, scrollPos, m.stdinDone, m.autoExitRemaining)
			return modal
		}
		return layout
	default:
		// Sidebar mode: content left, sidebar right
		sidebar := RenderSidebar(m.state, m.spinner, m.height, m.sidebarStyles, m.followMode, scrollPos, m.stdinDone, m.autoExitRemaining)
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

// Run starts the TUI and returns the final rendered content for optional dumping.
func Run() (string, error) {
	// Initialize styles (needed for renderers)
	cfg := config.DefaultProvider{}
	render.NewMarkdownRenderer(cfg.NoColor(), 80)

	// AltScreen and MouseMode are now set declaratively in View()
	var opts []tea.ProgramOption

	// When stdin is not a TTY (e.g., piped input), we need to read keyboard
	// input from /dev/tty instead. Otherwise bubbletea tries to read keyboard
	// events from the pipe, which causes terminal escape sequence issues.
	if !isatty(os.Stdin.Fd()) {
		tty, err := os.Open("/dev/tty")
		if err == nil {
			opts = append(opts, tea.WithInput(tty))
			defer tty.Close()
		}
	}

	p := tea.NewProgram(NewModel(), opts...)

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	// Extract final content from the model
	if m, ok := finalModel.(Model); ok {
		return m.content.String(), nil
	}
	return "", nil
}

// RunWithPrompt spawns claude with the given prompt and runs the TUI on its output.
func RunWithPrompt(prompt string) (string, error) {
	// Initialize styles (needed for renderers)
	cfg := config.DefaultProvider{}
	render.NewMarkdownRenderer(cfg.NoColor(), 80)

	proc, err := claudepkg.Start(prompt, nil)
	if err != nil {
		return "", err
	}

	// Keyboard input comes from /dev/tty since stdin is not available
	var teaOpts []tea.ProgramOption
	tty, err := os.Open("/dev/tty")
	if err == nil {
		teaOpts = append(teaOpts, tea.WithInput(tty))
		defer tty.Close()
	}

	model := NewModel(
		WithInputReader(proc.Stdout()),
		WithClaudeProcess(proc),
		WithPrompt(prompt),
	)

	p := tea.NewProgram(model, teaOpts...)

	finalModel, err := p.Run()
	if err != nil {
		proc.Kill()
		return "", err
	}

	// Wait for the subprocess to finish
	proc.Wait()

	if m, ok := finalModel.(Model); ok {
		return m.content.String(), nil
	}
	return "", nil
}

// RunWithStdinPrompt spawns claude with a prompt read from a reader and runs the TUI.
func RunWithStdinPrompt(prompt string) (string, error) {
	return RunWithPrompt(prompt)
}

// isatty returns true if the file descriptor is a terminal.
func isatty(fd uintptr) bool {
	return term.IsTerminal(int(fd))
}
