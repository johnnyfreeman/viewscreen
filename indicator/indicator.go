// Package indicator provides streaming progress indicators for terminal output.
// These are interactive UI components that show real-time feedback during
// streaming operations, separate from the content rendering utilities.
package indicator

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// Spinner provides visual feedback during streaming with cycling braille characters.
type Spinner struct {
	frames    []string
	index     int
	mu        sync.Mutex
	noColor   bool
	gradStart colorful.Color
	gradEnd   colorful.Color
	output    io.Writer
}

// Default spinner frames (braille pattern)
var defaultFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// SpinnerOption is a functional option for configuring a Spinner
type SpinnerOption func(*Spinner)

// WithSpinnerOutput sets a custom output writer for the Spinner
func WithSpinnerOutput(w io.Writer) SpinnerOption {
	return func(s *Spinner) {
		s.output = w
	}
}

// WithSpinnerFrames sets custom frames for the Spinner
func WithSpinnerFrames(frames []string) SpinnerOption {
	return func(s *Spinner) {
		if len(frames) > 0 {
			s.frames = frames
		}
	}
}

// NewSpinner creates a new spinner
func NewSpinner(noColor bool, opts ...SpinnerOption) *Spinner {
	// Gradient colors: purple to cyan
	start, _ := colorful.Hex("#A855F7")
	end, _ := colorful.Hex("#22D3EE")

	s := &Spinner{
		frames:    defaultFrames,
		noColor:   noColor,
		gradStart: start,
		gradEnd:   end,
		output:    os.Stdout,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
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

// Show writes the current spinner frame to output (overwrites previous)
func (s *Spinner) Show() {
	fmt.Fprint(s.output, s.Frame())
}

// Clear clears the spinner from the line
func (s *Spinner) Clear() {
	// Move back one character and clear
	fmt.Fprint(s.output, "\b \b")
}

// Reset resets the spinner to the first frame
func (s *Spinner) Reset() {
	s.mu.Lock()
	s.index = 0
	s.mu.Unlock()
}

// StreamingIndicator provides a simpler streaming indicator (a single dot or ellipsis).
type StreamingIndicator struct {
	shown   bool
	noColor bool
	output  io.Writer
}

// StreamingOption is a functional option for configuring a StreamingIndicator
type StreamingOption func(*StreamingIndicator)

// WithStreamingOutput sets a custom output writer for the StreamingIndicator
func WithStreamingOutput(w io.Writer) StreamingOption {
	return func(i *StreamingIndicator) {
		i.output = w
	}
}

// NewStreamingIndicator creates a new streaming indicator
func NewStreamingIndicator(noColor bool, opts ...StreamingOption) *StreamingIndicator {
	i := &StreamingIndicator{
		noColor: noColor,
		output:  os.Stdout,
	}

	for _, opt := range opts {
		opt(i)
	}

	return i
}

// Show displays the streaming indicator
func (i *StreamingIndicator) Show() {
	if i.shown {
		return
	}

	var ind string
	if i.noColor {
		ind = "..."
	} else {
		// Subtle pulsing dot
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#A855F7"))
		ind = style.Render("●")
	}
	fmt.Fprint(i.output, ind)
	i.shown = true
}

// Clear removes the streaming indicator
func (i *StreamingIndicator) Clear() {
	if !i.shown {
		return
	}

	// Clear the indicator character(s)
	if i.noColor {
		fmt.Fprint(i.output, "\b\b\b   \b\b\b")
	} else {
		fmt.Fprint(i.output, "\b \b")
	}
	i.shown = false
}

// IsShown returns whether the indicator is currently displayed
func (i *StreamingIndicator) IsShown() bool {
	return i.shown
}
