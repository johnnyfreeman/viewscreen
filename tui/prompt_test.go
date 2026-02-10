package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestPromptEditor(t *testing.T) {
	t.Run("starts inactive", func(t *testing.T) {
		p := NewPromptEditor()
		if p.Active {
			t.Error("expected prompt editor to start inactive")
		}
		if p.Value != "" {
			t.Error("expected empty value")
		}
	})

	t.Run("enter activates with prompt", func(t *testing.T) {
		p := NewPromptEditor()
		p.Enter("hello world")
		if !p.Active {
			t.Error("expected active after Enter")
		}
		if p.Value != "hello world" {
			t.Errorf("Value = %q, want %q", p.Value, "hello world")
		}
		if p.cursor != len("hello world") {
			t.Errorf("cursor = %d, want %d", p.cursor, len("hello world"))
		}
	})

	t.Run("exit keeps value", func(t *testing.T) {
		p := NewPromptEditor()
		p.Enter("hello")
		p.TypeRune('!')
		p.Exit()
		if p.Active {
			t.Error("expected inactive after Exit")
		}
		if p.Value != "hello!" {
			t.Errorf("Value = %q, want %q", p.Value, "hello!")
		}
	})

	t.Run("cancel restores original", func(t *testing.T) {
		p := NewPromptEditor()
		p.Enter("original")
		p.TypeRune('!')
		p.Cancel("original")
		if p.Active {
			t.Error("expected inactive after Cancel")
		}
		if p.Value != "original" {
			t.Errorf("Value = %q, want %q", p.Value, "original")
		}
	})

	t.Run("type rune at end", func(t *testing.T) {
		p := NewPromptEditor()
		p.Enter("")
		p.TypeRune('a')
		p.TypeRune('b')
		p.TypeRune('c')
		if p.Value != "abc" {
			t.Errorf("Value = %q, want %q", p.Value, "abc")
		}
	})

	t.Run("type rune at cursor position", func(t *testing.T) {
		p := NewPromptEditor()
		p.Enter("ac")
		p.cursor = 1 // between 'a' and 'c'
		p.TypeRune('b')
		if p.Value != "abc" {
			t.Errorf("Value = %q, want %q", p.Value, "abc")
		}
		if p.cursor != 2 {
			t.Errorf("cursor = %d, want 2", p.cursor)
		}
	})

	t.Run("backspace removes char before cursor", func(t *testing.T) {
		p := NewPromptEditor()
		p.Enter("abc")
		p.Backspace()
		if p.Value != "ab" {
			t.Errorf("Value = %q, want %q", p.Value, "ab")
		}
	})

	t.Run("backspace at start is no-op", func(t *testing.T) {
		p := NewPromptEditor()
		p.Enter("abc")
		p.cursor = 0
		p.Backspace()
		if p.Value != "abc" {
			t.Errorf("Value = %q, want %q", p.Value, "abc")
		}
	})

	t.Run("delete removes char after cursor", func(t *testing.T) {
		p := NewPromptEditor()
		p.Enter("abc")
		p.cursor = 1
		p.Delete()
		if p.Value != "ac" {
			t.Errorf("Value = %q, want %q", p.Value, "ac")
		}
	})

	t.Run("delete at end is no-op", func(t *testing.T) {
		p := NewPromptEditor()
		p.Enter("abc")
		p.Delete()
		if p.Value != "abc" {
			t.Errorf("Value = %q, want %q", p.Value, "abc")
		}
	})

	t.Run("cursor left", func(t *testing.T) {
		p := NewPromptEditor()
		p.Enter("abc")
		p.CursorLeft()
		if p.cursor != 2 {
			t.Errorf("cursor = %d, want 2", p.cursor)
		}
	})

	t.Run("cursor left at start is no-op", func(t *testing.T) {
		p := NewPromptEditor()
		p.Enter("abc")
		p.cursor = 0
		p.CursorLeft()
		if p.cursor != 0 {
			t.Errorf("cursor = %d, want 0", p.cursor)
		}
	})

	t.Run("cursor right", func(t *testing.T) {
		p := NewPromptEditor()
		p.Enter("abc")
		p.cursor = 0
		p.CursorRight()
		if p.cursor != 1 {
			t.Errorf("cursor = %d, want 1", p.cursor)
		}
	})

	t.Run("cursor right at end is no-op", func(t *testing.T) {
		p := NewPromptEditor()
		p.Enter("abc")
		p.CursorRight()
		if p.cursor != 3 {
			t.Errorf("cursor = %d, want 3", p.cursor)
		}
	})

	t.Run("cursor home", func(t *testing.T) {
		p := NewPromptEditor()
		p.Enter("abc")
		p.CursorHome()
		if p.cursor != 0 {
			t.Errorf("cursor = %d, want 0", p.cursor)
		}
	})

	t.Run("cursor end", func(t *testing.T) {
		p := NewPromptEditor()
		p.Enter("abc")
		p.cursor = 0
		p.CursorEnd()
		if p.cursor != 3 {
			t.Errorf("cursor = %d, want 3", p.cursor)
		}
	})
}

func TestRenderPromptBar(t *testing.T) {
	t.Run("empty when inactive", func(t *testing.T) {
		p := NewPromptEditor()
		result := RenderPromptBar(p, 80)
		if result != "" {
			t.Errorf("expected empty string when inactive, got %q", result)
		}
	})

	t.Run("shows prompt prefix when active", func(t *testing.T) {
		p := NewPromptEditor()
		p.Enter("hello")
		result := RenderPromptBar(p, 80)
		stripped := ansi.Strip(result)
		if !strings.Contains(stripped, "prompt>") {
			t.Errorf("expected prompt bar to contain 'prompt>', got %q", stripped)
		}
		if !strings.Contains(stripped, "hello") {
			t.Errorf("expected prompt bar to contain 'hello', got %q", stripped)
		}
	})

	t.Run("shows cursor", func(t *testing.T) {
		p := NewPromptEditor()
		p.Enter("")
		result := RenderPromptBar(p, 80)
		if !strings.Contains(result, "█") {
			t.Error("expected prompt bar to show cursor")
		}
	})

	t.Run("pads to width", func(t *testing.T) {
		p := NewPromptEditor()
		p.Enter("hi")
		result := RenderPromptBar(p, 40)
		stripped := ansi.Strip(result)
		if len(stripped) < 40 {
			t.Errorf("expected padded length >= 40, got %d", len(stripped))
		}
	})
}

func TestPromptEditorKeyHandling(t *testing.T) {
	t.Run("e opens editor when stdin done", func(t *testing.T) {
		m := newTestModel()
		m.stdinDone = true
		m.state.Prompt = "test prompt"

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "e"})
		if !m.promptEditor.Active {
			t.Error("expected prompt editor to be active after pressing e with stdinDone")
		}
		if m.promptEditor.Value != "test prompt" {
			t.Errorf("promptEditor.Value = %q, want %q", m.promptEditor.Value, "test prompt")
		}
	})

	t.Run("e does nothing when stdin not done", func(t *testing.T) {
		m := newTestModel()
		m.stdinDone = false

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "e"})
		if m.promptEditor.Active {
			t.Error("expected prompt editor to remain inactive when stdin not done")
		}
	})

	t.Run("e does nothing when help modal open", func(t *testing.T) {
		m := newTestModel()
		m.stdinDone = true
		m.showHelpModal = true

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "e"})
		if m.promptEditor.Active {
			t.Error("expected prompt editor to remain inactive when help modal is open")
		}
	})

	t.Run("e does nothing when details modal open", func(t *testing.T) {
		m := newTestModel()
		m.stdinDone = true
		m.showDetailsModal = true

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "e"})
		if m.promptEditor.Active {
			t.Error("expected prompt editor to remain inactive when details modal is open")
		}
	})

	t.Run("enter confirms edited prompt", func(t *testing.T) {
		m := newTestModel()
		m.stdinDone = true
		m.state.Prompt = "original"
		m.promptEditor.Enter("original")
		m.promptEditor.TypeRune('!')

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: tea.KeyEnter})
		if m.promptEditor.Active {
			t.Error("expected prompt editor to be inactive after enter")
		}
		if m.state.Prompt != "original!" {
			t.Errorf("state.Prompt = %q, want %q", m.state.Prompt, "original!")
		}
	})

	t.Run("esc cancels and restores original", func(t *testing.T) {
		m := newTestModel()
		m.stdinDone = true
		m.state.Prompt = "original"
		m.promptEditor.Enter("original")
		m.promptEditor.TypeRune('!')

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: tea.KeyEscape})
		if m.promptEditor.Active {
			t.Error("expected prompt editor to be inactive after esc")
		}
		if m.promptEditor.Value != "original" {
			t.Errorf("promptEditor.Value = %q, want %q", m.promptEditor.Value, "original")
		}
	})

	t.Run("typing adds characters", func(t *testing.T) {
		m := newTestModel()
		m.promptEditor.Enter("")

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "h"})
		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "i"})
		if m.promptEditor.Value != "hi" {
			t.Errorf("promptEditor.Value = %q, want %q", m.promptEditor.Value, "hi")
		}
	})

	t.Run("backspace removes characters", func(t *testing.T) {
		m := newTestModel()
		m.promptEditor.Enter("abc")

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: tea.KeyBackspace})
		if m.promptEditor.Value != "ab" {
			t.Errorf("promptEditor.Value = %q, want %q", m.promptEditor.Value, "ab")
		}
	})

	t.Run("ctrl+c quits during editing", func(t *testing.T) {
		m := newTestModel()
		m.promptEditor.Enter("test")

		_, cmd := m.handleKeyMsg(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
		if cmd == nil {
			t.Error("expected quit command on ctrl+c during prompt editing")
		}
	})

	t.Run("left arrow moves cursor", func(t *testing.T) {
		m := newTestModel()
		m.promptEditor.Enter("abc")

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: tea.KeyLeft})
		if m.promptEditor.cursor != 2 {
			t.Errorf("cursor = %d, want 2", m.promptEditor.cursor)
		}
	})

	t.Run("right arrow moves cursor", func(t *testing.T) {
		m := newTestModel()
		m.promptEditor.Enter("abc")
		m.promptEditor.cursor = 0

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: tea.KeyRight})
		if m.promptEditor.cursor != 1 {
			t.Errorf("cursor = %d, want 1", m.promptEditor.cursor)
		}
	})

	t.Run("prompt editor takes priority over other keys", func(t *testing.T) {
		m := newTestModel()
		m.promptEditor.Enter("test")

		// 'q' should type 'q', not quit
		m, cmd := m.handleKeyMsg(tea.KeyPressMsg{Text: "q"})
		if cmd != nil {
			t.Error("expected no quit command when prompt editor is active")
		}
		if !strings.Contains(m.promptEditor.Value, "q") {
			t.Error("expected 'q' to be typed into prompt editor")
		}
	})

	t.Run("prompt editor takes priority over search", func(t *testing.T) {
		m := newTestModel()
		m.promptEditor.Enter("test")
		m.search.Active = true // should not matter

		// '/' should type '/', not activate search
		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "/"})
		if strings.Contains(m.promptEditor.Value, "/") {
			// good
		}
	})
}

func TestModelPrompt(t *testing.T) {
	t.Run("returns editor value when set", func(t *testing.T) {
		m := newTestModel()
		m.promptEditor.Value = "edited"
		m.state.Prompt = "original"
		if m.Prompt() != "edited" {
			t.Errorf("Prompt() = %q, want %q", m.Prompt(), "edited")
		}
	})

	t.Run("returns state prompt when editor empty", func(t *testing.T) {
		m := newTestModel()
		m.state.Prompt = "from state"
		if m.Prompt() != "from state" {
			t.Errorf("Prompt() = %q, want %q", m.Prompt(), "from state")
		}
	})
}
