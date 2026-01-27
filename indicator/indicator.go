// Package indicator provides streaming progress indicators for terminal output.
// These are interactive UI components that show real-time feedback during
// streaming operations, separate from the content rendering utilities.
package indicator

import (
	"fmt"
	"io"
	"os"
	"sync"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/johnnyfreeman/viewscreen/style"
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

// NewSpinner creates a new spinner using theme gradient colors.
func NewSpinner(noColor bool, opts ...SpinnerOption) *Spinner {
	// Use theme colors for the spinner gradient
	start, _ := colorful.Hex(string(style.CurrentTheme.SpinnerGradientStart))
	end, _ := colorful.Hex(string(style.CurrentTheme.SpinnerGradientEnd))

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

// Frame returns the current spinner frame with gradient coloring.
// Uses Ultraviolet for proper style/content separation - this ensures that
// spinner frames can be safely composed with other styled content without
// escape sequence conflicts.
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
	blended := s.gradStart.BlendHcl(s.gradEnd, t).Clamped()

	// Use Ultraviolet for proper style/content separation.
	// This produces cleaner ANSI sequences that won't conflict
	// with surrounding escape codes when spinner frames are
	// composed with other styled text.
	uvStyle := &uv.Style{
		Fg: style.ColorfulToRGBA(blended),
	}
	return uvStyle.Styled(frame)
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

// Show displays the streaming indicator.
// Uses Ultraviolet for proper style/content separation.
func (i *StreamingIndicator) Show() {
	if i.shown {
		return
	}

	var ind string
	if i.noColor {
		ind = "..."
	} else {
		// Subtle pulsing dot - use theme accent color via Ultraviolet for consistent styling
		accent, _ := colorful.Hex(string(style.CurrentTheme.Accent))
		uvStyle := &uv.Style{
			Fg: style.ColorfulToRGBA(accent),
		}
		ind = uvStyle.Styled("●")
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
