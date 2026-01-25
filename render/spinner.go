package render

import (
	"fmt"
	"os"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// Spinner provides visual feedback during streaming
type Spinner struct {
	frames    []string
	index     int
	mu        sync.Mutex
	noColor   bool
	gradStart colorful.Color
	gradEnd   colorful.Color
}

// Default spinner frames (braille pattern)
var defaultFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NewSpinner creates a new spinner
func NewSpinner(noColor bool) *Spinner {
	// Gradient colors: purple to cyan
	start, _ := colorful.Hex("#A855F7")
	end, _ := colorful.Hex("#22D3EE")

	return &Spinner{
		frames:    defaultFrames,
		noColor:   noColor,
		gradStart: start,
		gradEnd:   end,
	}
}

// Frame returns the current spinner frame with gradient coloring
func (s *Spinner) Frame() string {
	s.mu.Lock()
	frame := s.frames[s.index]
	idx := s.index
	s.index = (s.index + 1) % len(s.frames)
	s.mu.Unlock()

	if s.noColor {
		return frame
	}

	// Calculate gradient color based on position
	t := float64(idx) / float64(len(s.frames)-1)
	color := s.gradStart.BlendHcl(s.gradEnd, t).Clamped()

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color.Hex()))
	return style.Render(frame)
}

// Show writes the current spinner frame to stdout (overwrites previous)
func (s *Spinner) Show() {
	fmt.Fprint(os.Stdout, s.Frame())
}

// Clear clears the spinner from the line
func (s *Spinner) Clear() {
	// Move back one character and clear
	fmt.Fprint(os.Stdout, "\b \b")
}

// Reset resets the spinner to the first frame
func (s *Spinner) Reset() {
	s.mu.Lock()
	s.index = 0
	s.mu.Unlock()
}

// StreamingIndicator provides a simpler streaming indicator
type StreamingIndicator struct {
	shown   bool
	noColor bool
}

// NewStreamingIndicator creates a new streaming indicator
func NewStreamingIndicator(noColor bool) *StreamingIndicator {
	return &StreamingIndicator{noColor: noColor}
}

// Show displays the streaming indicator
func (i *StreamingIndicator) Show() {
	if i.shown {
		return
	}

	var indicator string
	if i.noColor {
		indicator = "..."
	} else {
		// Subtle pulsing dot
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#A855F7"))
		indicator = style.Render("●")
	}
	fmt.Fprint(os.Stdout, indicator)
	i.shown = true
}

// Clear removes the streaming indicator
func (i *StreamingIndicator) Clear() {
	if !i.shown {
		return
	}

	// Clear the indicator character(s)
	if i.noColor {
		fmt.Fprint(os.Stdout, "\b\b\b   \b\b\b")
	} else {
		fmt.Fprint(os.Stdout, "\b \b")
	}
	i.shown = false
}

// IsShown returns whether the indicator is currently displayed
func (i *StreamingIndicator) IsShown() bool {
	return i.shown
}
