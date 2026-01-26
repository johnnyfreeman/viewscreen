package render

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

func TestNewSpinner(t *testing.T) {
	tests := []struct {
		name    string
		noColor bool
	}{
		{"with color", false},
		{"without color", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSpinner(tt.noColor)
			if s == nil {
				t.Fatal("NewSpinner returned nil")
			}
			if s.noColor != tt.noColor {
				t.Errorf("noColor = %v, want %v", s.noColor, tt.noColor)
			}
			if len(s.frames) != len(defaultFrames) {
				t.Errorf("frames length = %d, want %d", len(s.frames), len(defaultFrames))
			}
			if s.index != 0 {
				t.Errorf("index = %d, want 0", s.index)
			}
			if s.output == nil {
				t.Error("output should not be nil")
			}
		})
	}
}

func TestSpinner_WithOptions(t *testing.T) {
	t.Run("WithSpinnerOutput", func(t *testing.T) {
		buf := &bytes.Buffer{}
		s := NewSpinner(true, WithSpinnerOutput(buf))

		if s.output != buf {
			t.Error("output should be set to custom buffer")
		}
	})

	t.Run("WithSpinnerFrames", func(t *testing.T) {
		customFrames := []string{"a", "b", "c"}
		s := NewSpinner(true, WithSpinnerFrames(customFrames))

		if len(s.frames) != len(customFrames) {
			t.Errorf("frames length = %d, want %d", len(s.frames), len(customFrames))
		}
		for i, f := range s.frames {
			if f != customFrames[i] {
				t.Errorf("frame[%d] = %q, want %q", i, f, customFrames[i])
			}
		}
	})

	t.Run("WithSpinnerFrames empty does not override", func(t *testing.T) {
		s := NewSpinner(true, WithSpinnerFrames([]string{}))

		if len(s.frames) != len(defaultFrames) {
			t.Errorf("frames should remain default when given empty slice, got length %d", len(s.frames))
		}
	})

	t.Run("multiple options", func(t *testing.T) {
		buf := &bytes.Buffer{}
		customFrames := []string{"x", "y"}
		s := NewSpinner(false, WithSpinnerOutput(buf), WithSpinnerFrames(customFrames))

		if s.output != buf {
			t.Error("output should be custom buffer")
		}
		if len(s.frames) != 2 {
			t.Errorf("frames length = %d, want 2", len(s.frames))
		}
	})
}

func TestSpinner_Frame(t *testing.T) {
	t.Run("cycles through frames in noColor mode", func(t *testing.T) {
		s := NewSpinner(true)

		for i := 0; i < len(defaultFrames); i++ {
			frame := s.Frame()
			if frame != defaultFrames[i] {
				t.Errorf("Frame() at index %d = %q, want %q", i, frame, defaultFrames[i])
			}
		}

		// Should wrap around
		frame := s.Frame()
		if frame != defaultFrames[0] {
			t.Errorf("Frame() after wrap = %q, want %q", frame, defaultFrames[0])
		}
	})

	t.Run("returns frames with color (may include styling)", func(t *testing.T) {
		s := NewSpinner(false)
		frame := s.Frame()

		// The frame should at minimum contain the base character
		// In TTY mode it would have ANSI codes, but in test mode it may not
		if !strings.Contains(frame, defaultFrames[0]) {
			t.Errorf("colored frame should contain base character %q, got %q", defaultFrames[0], frame)
		}
	})

	t.Run("gradient calculation does not panic", func(t *testing.T) {
		s := NewSpinner(false)

		// Cycle through all frames to test gradient at all positions
		for i := 0; i < len(defaultFrames)*2; i++ {
			frame := s.Frame()
			// Frame should always contain one of the valid characters
			found := false
			for _, f := range defaultFrames {
				if strings.Contains(frame, f) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("frame %d should contain a valid frame character, got %q", i, frame)
			}
		}
	})
}

func TestSpinner_Show(t *testing.T) {
	t.Run("writes frame to output", func(t *testing.T) {
		buf := &bytes.Buffer{}
		s := NewSpinner(true, WithSpinnerOutput(buf))

		s.Show()

		output := buf.String()
		if output != defaultFrames[0] {
			t.Errorf("Show() wrote %q, want %q", output, defaultFrames[0])
		}
	})

	t.Run("advances frame on each call", func(t *testing.T) {
		buf := &bytes.Buffer{}
		s := NewSpinner(true, WithSpinnerOutput(buf))

		s.Show()
		s.Show()

		output := buf.String()
		expected := defaultFrames[0] + defaultFrames[1]
		if output != expected {
			t.Errorf("two Show() calls wrote %q, want %q", output, expected)
		}
	})
}

func TestSpinner_Clear(t *testing.T) {
	t.Run("writes clear sequence", func(t *testing.T) {
		buf := &bytes.Buffer{}
		s := NewSpinner(true, WithSpinnerOutput(buf))

		s.Clear()

		output := buf.String()
		expected := "\b \b"
		if output != expected {
			t.Errorf("Clear() wrote %q, want %q", output, expected)
		}
	})
}

func TestSpinner_Reset(t *testing.T) {
	t.Run("resets index to zero", func(t *testing.T) {
		s := NewSpinner(true)

		// Advance several frames
		for i := 0; i < 5; i++ {
			s.Frame()
		}

		s.Reset()

		frame := s.Frame()
		if frame != defaultFrames[0] {
			t.Errorf("after Reset(), Frame() = %q, want %q", frame, defaultFrames[0])
		}
	})
}

func TestSpinner_Concurrency(t *testing.T) {
	t.Run("Frame is thread-safe", func(t *testing.T) {
		s := NewSpinner(true)
		const goroutines = 100
		const framesPerGoroutine = 100

		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < framesPerGoroutine; j++ {
					frame := s.Frame()
					// Should always be one of the valid frames
					found := false
					for _, f := range defaultFrames {
						if frame == f {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("got invalid frame %q", frame)
					}
				}
			}()
		}

		wg.Wait()
	})

	t.Run("Reset is thread-safe with Frame", func(t *testing.T) {
		s := NewSpinner(true)
		const iterations = 1000

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				s.Frame()
			}
		}()

		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				s.Reset()
			}
		}()

		wg.Wait()
		// Test passes if no race condition detected
	})
}

func TestNewStreamingIndicator(t *testing.T) {
	tests := []struct {
		name    string
		noColor bool
	}{
		{"with color", false},
		{"without color", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := NewStreamingIndicator(tt.noColor)
			if i == nil {
				t.Fatal("NewStreamingIndicator returned nil")
			}
			if i.noColor != tt.noColor {
				t.Errorf("noColor = %v, want %v", i.noColor, tt.noColor)
			}
			if i.shown {
				t.Error("shown should be false initially")
			}
			if i.output == nil {
				t.Error("output should not be nil")
			}
		})
	}
}

func TestStreamingIndicator_WithOptions(t *testing.T) {
	t.Run("WithIndicatorOutput", func(t *testing.T) {
		buf := &bytes.Buffer{}
		i := NewStreamingIndicator(true, WithIndicatorOutput(buf))

		if i.output != buf {
			t.Error("output should be set to custom buffer")
		}
	})
}

func TestStreamingIndicator_Show(t *testing.T) {
	t.Run("shows dots in noColor mode", func(t *testing.T) {
		buf := &bytes.Buffer{}
		i := NewStreamingIndicator(true, WithIndicatorOutput(buf))

		i.Show()

		output := buf.String()
		if output != "..." {
			t.Errorf("Show() in noColor mode wrote %q, want %q", output, "...")
		}
	})

	t.Run("shows styled indicator with color", func(t *testing.T) {
		buf := &bytes.Buffer{}
		i := NewStreamingIndicator(false, WithIndicatorOutput(buf))

		i.Show()

		output := buf.String()
		// Should contain the bullet character (with or without ANSI codes depending on TTY)
		if !strings.Contains(output, "●") {
			t.Errorf("Show() with color should contain bullet, got %q", output)
		}
	})

	t.Run("only shows once", func(t *testing.T) {
		buf := &bytes.Buffer{}
		i := NewStreamingIndicator(true, WithIndicatorOutput(buf))

		i.Show()
		i.Show()
		i.Show()

		output := buf.String()
		if output != "..." {
			t.Errorf("multiple Show() calls should only write once, got %q", output)
		}
	})

	t.Run("sets shown flag", func(t *testing.T) {
		buf := &bytes.Buffer{}
		i := NewStreamingIndicator(true, WithIndicatorOutput(buf))

		if i.IsShown() {
			t.Error("IsShown() should be false before Show()")
		}

		i.Show()

		if !i.IsShown() {
			t.Error("IsShown() should be true after Show()")
		}
	})
}

func TestStreamingIndicator_Clear(t *testing.T) {
	t.Run("clears dots in noColor mode", func(t *testing.T) {
		buf := &bytes.Buffer{}
		i := NewStreamingIndicator(true, WithIndicatorOutput(buf))

		i.Show()
		buf.Reset() // Clear the Show output
		i.Clear()

		output := buf.String()
		// Should contain backspaces and spaces to clear "..."
		if !strings.Contains(output, "\b") {
			t.Errorf("Clear() should contain backspaces, got %q", output)
		}
	})

	t.Run("clears single character with color", func(t *testing.T) {
		buf := &bytes.Buffer{}
		i := NewStreamingIndicator(false, WithIndicatorOutput(buf))

		i.Show()
		buf.Reset() // Clear the Show output
		i.Clear()

		output := buf.String()
		expected := "\b \b"
		if output != expected {
			t.Errorf("Clear() with color wrote %q, want %q", output, expected)
		}
	})

	t.Run("does nothing if not shown", func(t *testing.T) {
		buf := &bytes.Buffer{}
		i := NewStreamingIndicator(true, WithIndicatorOutput(buf))

		i.Clear()

		output := buf.String()
		if output != "" {
			t.Errorf("Clear() when not shown should write nothing, got %q", output)
		}
	})

	t.Run("resets shown flag", func(t *testing.T) {
		buf := &bytes.Buffer{}
		i := NewStreamingIndicator(true, WithIndicatorOutput(buf))

		i.Show()
		if !i.IsShown() {
			t.Error("IsShown() should be true after Show()")
		}

		i.Clear()
		if i.IsShown() {
			t.Error("IsShown() should be false after Clear()")
		}
	})
}

func TestStreamingIndicator_ShowClearCycle(t *testing.T) {
	t.Run("can show, clear, and show again", func(t *testing.T) {
		buf := &bytes.Buffer{}
		i := NewStreamingIndicator(true, WithIndicatorOutput(buf))

		i.Show()
		if !i.IsShown() {
			t.Error("should be shown after first Show()")
		}

		i.Clear()
		if i.IsShown() {
			t.Error("should not be shown after Clear()")
		}

		buf.Reset()
		i.Show()
		if !i.IsShown() {
			t.Error("should be shown after second Show()")
		}

		output := buf.String()
		if output != "..." {
			t.Errorf("second Show() should write indicator, got %q", output)
		}
	})
}

func TestStreamingIndicator_IsShown(t *testing.T) {
	tests := []struct {
		name     string
		actions  func(*StreamingIndicator)
		expected bool
	}{
		{
			name:     "initially false",
			actions:  func(i *StreamingIndicator) {},
			expected: false,
		},
		{
			name:     "true after Show",
			actions:  func(i *StreamingIndicator) { i.Show() },
			expected: true,
		},
		{
			name:     "false after Show then Clear",
			actions:  func(i *StreamingIndicator) { i.Show(); i.Clear() },
			expected: false,
		},
		{
			name:     "true after Show-Clear-Show",
			actions:  func(i *StreamingIndicator) { i.Show(); i.Clear(); i.Show() },
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			i := NewStreamingIndicator(true, WithIndicatorOutput(buf))

			tt.actions(i)

			if i.IsShown() != tt.expected {
				t.Errorf("IsShown() = %v, want %v", i.IsShown(), tt.expected)
			}
		})
	}
}

func TestDefaultFrames(t *testing.T) {
	t.Run("has expected braille frames", func(t *testing.T) {
		expected := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

		if len(defaultFrames) != len(expected) {
			t.Errorf("defaultFrames length = %d, want %d", len(defaultFrames), len(expected))
		}

		for i, frame := range defaultFrames {
			if frame != expected[i] {
				t.Errorf("defaultFrames[%d] = %q, want %q", i, frame, expected[i])
			}
		}
	})
}
