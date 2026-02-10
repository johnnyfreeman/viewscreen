package tui

import (
	"bufio"
	"os"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	followMode        bool // auto-scroll to bottom on new content
	autoExit          bool // --auto-exit flag enabled
	autoExitRemaining int  // seconds left in countdown, 0 = inactive
}

// NewModel creates a new TUI model
func NewModel() Model {
	// Initialize spinner with Dot spinner and lipgloss styling.
	// We use lipgloss here (not Ultraviolet) because the spinner output
	// goes through bubbletea/lipgloss rendering pipeline.
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(string(style.CurrentTheme.Accent)))),
	)

	// Create scanner for stdin with large buffer
	scanner := bufio.NewScanner(os.Stdin)
	const maxCapacity = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	st := state.NewState()
	return Model{
		spinner:       s,
		state:         st,
		content:       &strings.Builder{},
		scanner:       scanner,
		sidebarStyles: NewSidebarStyles(),
		headerStyles:  NewHeaderStyles(),
		layoutMode:    LayoutSidebar, // default to sidebar mode
		processor:     events.NewEventProcessor(st),
		search:        NewSearch(),
		followMode:    true, // auto-scroll to bottom by default
		autoExit:      config.AutoExit,
	}
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
		return RenderHelpModal(m.width, m.height, m.headerStyles, m.autoExitRemaining > 0)
	}

	// Render search bar if active or has a query
	searchBar := RenderSearchBar(m.search, m.viewport.Width())
	scrollPos := m.scrollPosition()

	switch m.layoutMode {
	case LayoutHeader:
		// Header mode: single-line header on top, content below at full width
		header := RenderHeader(m.state, m.width, m.followMode, scrollPos, m.stdinDone, m.autoExitRemaining)
		parts := []string{header, m.viewport.View()}
		if searchBar != "" {
			parts = append(parts, searchBar)
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
		mainContent := lipgloss.JoinVertical(lipgloss.Left, mainParts...)
		return lipgloss.JoinHorizontal(lipgloss.Top, mainContent, sidebar)
	}
}

// Run starts the TUI
func Run() error {
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

	_, err := p.Run()
	return err
}

// isatty returns true if the file descriptor is a terminal.
func isatty(fd uintptr) bool {
	return term.IsTerminal(int(fd))
}
