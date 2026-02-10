package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestSearchEnterExit(t *testing.T) {
	s := NewSearch()

	if s.Active {
		t.Error("expected search to start inactive")
	}

	s.Enter()
	if !s.Active {
		t.Error("expected search to be active after Enter")
	}
	if s.Query != "" {
		t.Error("expected empty query after Enter")
	}

	s.TypeRune('h')
	s.TypeRune('i')
	if s.Query != "hi" {
		t.Errorf("query = %q, want %q", s.Query, "hi")
	}

	s.Exit()
	if s.Active {
		t.Error("expected search to be inactive after Exit")
	}
	if s.Query != "hi" {
		t.Error("expected query to be preserved after Exit")
	}

	s.Clear()
	if s.Query != "" {
		t.Error("expected query to be cleared after Clear")
	}
	if s.HasQuery() {
		t.Error("expected HasQuery to be false after Clear")
	}
}

func TestSearchBackspace(t *testing.T) {
	s := NewSearch()
	s.Enter()
	s.TypeRune('a')
	s.TypeRune('b')
	s.TypeRune('c')

	s.Backspace()
	if s.Query != "ab" {
		t.Errorf("query = %q, want %q", s.Query, "ab")
	}

	s.Backspace()
	s.Backspace()
	if s.Query != "" {
		t.Errorf("query = %q, want empty", s.Query)
	}

	// Backspace on empty should not panic
	s.Backspace()
	if s.Query != "" {
		t.Errorf("query = %q, want empty after extra backspace", s.Query)
	}
}

func TestSearchUpdateMatches(t *testing.T) {
	t.Run("basic matching", func(t *testing.T) {
		s := NewSearch()
		s.Enter()
		s.Query = "hello"

		content := "line one\nhello world\nline three\nhello again"
		s.UpdateMatches(content)

		if s.MatchCount() != 2 {
			t.Errorf("MatchCount() = %d, want 2", s.MatchCount())
		}
		if s.CurrentLine() != 1 {
			t.Errorf("CurrentLine() = %d, want 1", s.CurrentLine())
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		s := NewSearch()
		s.Enter()
		s.Query = "Hello"

		content := "HELLO world\nhello there\nHeLLo"
		s.UpdateMatches(content)

		if s.MatchCount() != 3 {
			t.Errorf("MatchCount() = %d, want 3", s.MatchCount())
		}
	})

	t.Run("no matches", func(t *testing.T) {
		s := NewSearch()
		s.Enter()
		s.Query = "xyz"

		content := "line one\nline two"
		s.UpdateMatches(content)

		if s.MatchCount() != 0 {
			t.Errorf("MatchCount() = %d, want 0", s.MatchCount())
		}
		if s.CurrentLine() != -1 {
			t.Errorf("CurrentLine() = %d, want -1", s.CurrentLine())
		}
	})

	t.Run("empty query", func(t *testing.T) {
		s := NewSearch()
		s.Enter()
		s.Query = ""

		content := "hello world"
		s.UpdateMatches(content)

		if s.MatchCount() != 0 {
			t.Errorf("MatchCount() = %d, want 0 for empty query", s.MatchCount())
		}
	})

	t.Run("strips ANSI sequences before matching", func(t *testing.T) {
		s := NewSearch()
		s.Enter()
		s.Query = "hello"

		// Simulate ANSI-styled content
		content := "\x1b[31mhello\x1b[0m world\nplain line"
		s.UpdateMatches(content)

		if s.MatchCount() != 1 {
			t.Errorf("MatchCount() = %d, want 1 (should match through ANSI)", s.MatchCount())
		}
	})
}

func TestSearchNavigation(t *testing.T) {
	s := NewSearch()
	s.Enter()
	s.Query = "match"

	content := "match one\nno\nmatch two\nno\nmatch three"
	s.UpdateMatches(content)

	if s.MatchCount() != 3 {
		t.Fatalf("MatchCount() = %d, want 3", s.MatchCount())
	}

	// Should start at first match
	if s.CurrentMatchIndex() != 1 {
		t.Errorf("CurrentMatchIndex() = %d, want 1", s.CurrentMatchIndex())
	}
	if s.CurrentLine() != 0 {
		t.Errorf("CurrentLine() = %d, want 0", s.CurrentLine())
	}

	// Next match
	s.NextMatch()
	if s.CurrentMatchIndex() != 2 {
		t.Errorf("after NextMatch: CurrentMatchIndex() = %d, want 2", s.CurrentMatchIndex())
	}
	if s.CurrentLine() != 2 {
		t.Errorf("after NextMatch: CurrentLine() = %d, want 2", s.CurrentLine())
	}

	// Next again
	s.NextMatch()
	if s.CurrentMatchIndex() != 3 {
		t.Errorf("after 2nd NextMatch: CurrentMatchIndex() = %d, want 3", s.CurrentMatchIndex())
	}

	// Wrap around
	s.NextMatch()
	if s.CurrentMatchIndex() != 1 {
		t.Errorf("after wrap NextMatch: CurrentMatchIndex() = %d, want 1", s.CurrentMatchIndex())
	}

	// Prev match wraps backward
	s.PrevMatch()
	if s.CurrentMatchIndex() != 3 {
		t.Errorf("after PrevMatch wrap: CurrentMatchIndex() = %d, want 3", s.CurrentMatchIndex())
	}

	// Prev again
	s.PrevMatch()
	if s.CurrentMatchIndex() != 2 {
		t.Errorf("after 2nd PrevMatch: CurrentMatchIndex() = %d, want 2", s.CurrentMatchIndex())
	}
}

func TestSearchNavigationNoMatches(t *testing.T) {
	s := NewSearch()
	s.Enter()
	s.Query = "nope"

	content := "line one\nline two"
	s.UpdateMatches(content)

	// Should not panic
	s.NextMatch()
	s.PrevMatch()

	if s.CurrentLine() != -1 {
		t.Errorf("CurrentLine() = %d, want -1 with no matches", s.CurrentLine())
	}
}

func TestRenderSearchBar(t *testing.T) {
	t.Run("no search active and no query", func(t *testing.T) {
		s := NewSearch()
		bar := RenderSearchBar(s, 80)
		if bar != "" {
			t.Errorf("expected empty string, got %q", bar)
		}
	})

	t.Run("active search with query", func(t *testing.T) {
		s := NewSearch()
		s.Enter()
		s.Query = "test"

		bar := RenderSearchBar(s, 80)
		plain := ansi.Strip(bar)

		if !strings.Contains(plain, "/") {
			t.Error("expected search bar to contain /")
		}
		if !strings.Contains(plain, "test") {
			t.Error("expected search bar to contain query 'test'")
		}
	})

	t.Run("shows match count", func(t *testing.T) {
		s := NewSearch()
		s.Enter()
		s.Query = "hello"
		s.UpdateMatches("hello world\nhello again\nno match")

		bar := RenderSearchBar(s, 80)
		plain := ansi.Strip(bar)

		if !strings.Contains(plain, "1/2") {
			t.Errorf("expected bar to show '1/2', got %q", plain)
		}
	})

	t.Run("shows no matches", func(t *testing.T) {
		s := NewSearch()
		s.Enter()
		s.Query = "xyz"
		s.UpdateMatches("hello world")

		bar := RenderSearchBar(s, 80)
		plain := ansi.Strip(bar)

		if !strings.Contains(plain, "no matches") {
			t.Errorf("expected bar to show 'no matches', got %q", plain)
		}
	})

	t.Run("inactive with query shows muted", func(t *testing.T) {
		s := NewSearch()
		s.Enter()
		s.Query = "test"
		s.Exit()

		bar := RenderSearchBar(s, 80)
		if bar == "" {
			t.Error("expected non-empty bar when query exists after exit")
		}
	})
}

func TestItoa(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{100, "100"},
		{-5, "-5"},
	}

	for _, tc := range tests {
		got := itoa(tc.input)
		if got != tc.want {
			t.Errorf("itoa(%d) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestHandleKeyMsgSearch(t *testing.T) {
	t.Run("/ enters search mode", func(t *testing.T) {
		m := newTestModel()
		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "/"})
		if !m.search.Active {
			t.Error("expected search to be active after /")
		}
	})

	t.Run("/ does not enter search when modal open", func(t *testing.T) {
		m := newTestModel()
		m.showHelpModal = true
		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "/"})
		if m.search.Active {
			t.Error("expected search to remain inactive when help modal is open")
		}
	})

	t.Run("typing in search mode adds to query", func(t *testing.T) {
		m := newTestModel()
		m.search.Enter()
		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "h"})
		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "i"})
		if m.search.Query != "hi" {
			t.Errorf("search query = %q, want %q", m.search.Query, "hi")
		}
	})

	t.Run("enter confirms search", func(t *testing.T) {
		m := newTestModel()
		m.search.Enter()
		m.search.Query = "test"
		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: tea.KeyEnter})
		if m.search.Active {
			t.Error("expected search input to be inactive after enter")
		}
		if m.search.Query != "test" {
			t.Error("expected query to be preserved after enter")
		}
	})

	t.Run("esc in search mode clears search", func(t *testing.T) {
		m := newTestModel()
		m.search.Enter()
		m.search.Query = "test"
		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: tea.KeyEscape})
		if m.search.Active {
			t.Error("expected search to be inactive after esc")
		}
		if m.search.Query != "" {
			t.Error("expected query to be cleared after esc in active search")
		}
	})

	t.Run("esc in normal mode clears inactive query", func(t *testing.T) {
		m := newTestModel()
		m.search.Query = "test"
		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: tea.KeyEscape})
		if m.search.Query != "" {
			t.Error("expected query to be cleared after esc")
		}
	})

	t.Run("n and N navigate matches", func(t *testing.T) {
		m := newTestModel()
		m.content.WriteString("match one\nno\nmatch two\nno\nmatch three")
		m.viewport.SetContent(m.content.String())
		m.search.Query = "match"
		m.search.UpdateMatches(m.content.String())

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "n"})
		if m.search.CurrentMatchIndex() != 2 {
			t.Errorf("after n: CurrentMatchIndex() = %d, want 2", m.search.CurrentMatchIndex())
		}

		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Text: "N"})
		if m.search.CurrentMatchIndex() != 1 {
			t.Errorf("after N: CurrentMatchIndex() = %d, want 1", m.search.CurrentMatchIndex())
		}
	})

	t.Run("backspace in search mode removes character", func(t *testing.T) {
		m := newTestModel()
		m.search.Enter()
		m.search.Query = "abc"
		m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: tea.KeyBackspace})
		if m.search.Query != "ab" {
			t.Errorf("search query = %q, want %q", m.search.Query, "ab")
		}
	})

	t.Run("ctrl+c quits from search mode", func(t *testing.T) {
		m := newTestModel()
		m.search.Enter()
		_, cmd := m.handleKeyMsg(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
		if cmd == nil {
			t.Error("expected quit command on ctrl+c in search mode")
		}
	})

	t.Run("q does not quit in search mode", func(t *testing.T) {
		m := newTestModel()
		m.search.Enter()
		m, cmd := m.handleKeyMsg(tea.KeyPressMsg{Text: "q"})
		if cmd != nil {
			t.Error("expected no quit command when typing q in search mode")
		}
		if m.search.Query != "q" {
			t.Errorf("search query = %q, want %q", m.search.Query, "q")
		}
	})
}

func TestSearchViewportHeight(t *testing.T) {
	t.Run("viewport height reduced when search active", func(t *testing.T) {
		m := NewModel()
		m.ready = false
		m = m.handleWindowSizeMsg(tea.WindowSizeMsg{Width: 120, Height: 50})

		heightWithout := m.viewport.Height()

		// Activate search
		m.search.Enter()
		m = m.handleWindowSizeMsg(tea.WindowSizeMsg{Width: 120, Height: 50})
		heightWith := m.viewport.Height()

		if heightWith != heightWithout-1 {
			t.Errorf("viewport height with search = %d, want %d (one less than %d)", heightWith, heightWithout-1, heightWithout)
		}
	})
}
