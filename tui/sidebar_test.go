package tui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"
	"github.com/johnnyfreeman/viewscreen/state"
	"github.com/johnnyfreeman/viewscreen/style"
)

func init() {
	// Initialize style with noColor for consistent test output
	// Without this, gradient functions produce ANSI codes that make
	// substring checks fail (e.g., "█ █" becomes "\x1b[...m█\x1b[m\x1b[...m \x1b[m...")
	style.Init(true)
}

func newTestSpinner() spinner.Model {
	// Use v2 functional options API
	return spinner.New(
		spinner.WithSpinner(spinner.Spinner{Frames: []string{"⠋"}, FPS: 1}),
	)
}

func TestNewSidebarRenderer(t *testing.T) {
	styles := NewSidebarStyles()
	sp := newTestSpinner()
	r := NewSidebarRenderer(styles, sp)

	if r == nil {
		t.Fatal("NewSidebarRenderer returned nil")
	}
	if r.width != sidebarWidth {
		t.Errorf("width = %d, want %d", r.width, sidebarWidth)
	}
}

func TestSidebarRenderer_RenderLogo(t *testing.T) {
	r := NewSidebarRenderer(NewSidebarStyles(), newTestSpinner())

	output := r.RenderLogo()

	// Should contain the decoration dots
	if !strings.Contains(output, "·") {
		t.Error("expected decoration dots in logo")
	}

	// Should contain "claude" text
	if !strings.Contains(output, "claude") {
		t.Error("expected 'claude' text in logo")
	}

	// Should contain logo lines
	for _, line := range logoLines {
		if !strings.Contains(output, line) {
			t.Errorf("expected logo line %q in output", line)
		}
	}
}

func TestSidebarRenderer_RenderPrompt(t *testing.T) {
	r := NewSidebarRenderer(NewSidebarStyles(), newTestSpinner())

	t.Run("empty prompt returns empty string", func(t *testing.T) {
		output := r.RenderPrompt("")
		if output != "" {
			t.Errorf("expected empty string for empty prompt, got %q", output)
		}
	})

	t.Run("renders prompt with quotes", func(t *testing.T) {
		output := r.RenderPrompt("Hello world")
		if !strings.Contains(output, "\"Hello world\"") {
			t.Errorf("expected quoted prompt in output, got %q", output)
		}
	})

	t.Run("wraps long prompts", func(t *testing.T) {
		longPrompt := strings.Repeat("word ", 50)
		output := r.RenderPrompt(longPrompt)

		// Should contain the prompt text
		if !strings.Contains(output, "word") {
			t.Errorf("expected prompt text in output, got %q", output)
		}
	})
}

func TestSidebarRenderer_RenderLabelValue(t *testing.T) {
	r := NewSidebarRenderer(NewSidebarStyles(), newTestSpinner())

	output := r.RenderLabelValue("Label", "Value")

	if !strings.Contains(output, "Label") {
		t.Errorf("expected 'Label' in output, got %q", output)
	}
	if !strings.Contains(output, "Value") {
		t.Errorf("expected 'Value' in output, got %q", output)
	}
}

func TestSidebarRenderer_RenderSessionInfo(t *testing.T) {
	r := NewSidebarRenderer(NewSidebarStyles(), newTestSpinner())

	t.Run("renders model, turns, and cost", func(t *testing.T) {
		output := r.RenderSessionInfo("claude-3-opus", 5, 0.1234)

		if !strings.Contains(output, "Model") {
			t.Error("expected 'Model' label in output")
		}
		if !strings.Contains(output, "claude-3-opus") {
			t.Error("expected model name in output")
		}
		if !strings.Contains(output, "Turns") {
			t.Error("expected 'Turns' label in output")
		}
		if !strings.Contains(output, "5") {
			t.Error("expected turn count in output")
		}
		if !strings.Contains(output, "Cost") {
			t.Error("expected 'Cost' label in output")
		}
		if !strings.Contains(output, "$0.1234") {
			t.Error("expected cost in output")
		}
	})

	t.Run("truncates long model names", func(t *testing.T) {
		longModel := strings.Repeat("a", 50)
		output := r.RenderSessionInfo(longModel, 1, 0)

		// Should not contain the full model name
		if strings.Contains(output, longModel) {
			t.Error("expected long model name to be truncated")
		}
		// Should contain ellipsis
		if !strings.Contains(output, "...") {
			t.Error("expected ellipsis for truncated model name")
		}
	})
}

func TestSidebarRenderer_RenderCurrentTool(t *testing.T) {
	r := NewSidebarRenderer(NewSidebarStyles(), newTestSpinner())

	t.Run("empty tool name returns empty string", func(t *testing.T) {
		output := r.RenderCurrentTool("", "")
		if output != "" {
			t.Errorf("expected empty string for empty tool name, got %q", output)
		}
	})

	t.Run("renders tool with header", func(t *testing.T) {
		output := r.RenderCurrentTool("Read", "")

		if !strings.Contains(output, "Running") {
			t.Error("expected 'Running' header in output")
		}
		if !strings.Contains(output, "Read") {
			t.Error("expected tool name in output")
		}
	})

	t.Run("includes short input", func(t *testing.T) {
		output := r.RenderCurrentTool("Read", "file.txt")

		if !strings.Contains(output, "Read") {
			t.Error("expected tool name in output")
		}
		if !strings.Contains(output, "file.txt") {
			t.Error("expected short input in output")
		}
	})

	t.Run("excludes long input", func(t *testing.T) {
		longInput := strings.Repeat("x", 30)
		output := r.RenderCurrentTool("Read", longInput)

		if !strings.Contains(output, "Read") {
			t.Error("expected tool name in output")
		}
		// Long input should NOT be included
		if strings.Contains(output, longInput) {
			t.Error("expected long input to be excluded")
		}
	})
}

func TestSidebarRenderer_RenderTodo(t *testing.T) {
	r := NewSidebarRenderer(NewSidebarStyles(), newTestSpinner())

	t.Run("renders completed todo", func(t *testing.T) {
		todo := state.Todo{
			Subject: "Fix bug",
			Status:  "completed",
		}
		output := r.RenderTodo(todo)

		if !strings.Contains(output, "✓") {
			t.Error("expected checkmark for completed todo")
		}
		if !strings.Contains(output, "Fix bug") {
			t.Error("expected todo subject in output")
		}
	})

	t.Run("renders in_progress todo with active form", func(t *testing.T) {
		todo := state.Todo{
			Subject:    "Fix bug",
			ActiveForm: "Fixing bug",
			Status:     "in_progress",
		}
		output := r.RenderTodo(todo)

		// Should prefer ActiveForm for in_progress
		if !strings.Contains(output, "Fixing bug") {
			t.Error("expected active form in output")
		}
	})

	t.Run("renders pending todo", func(t *testing.T) {
		todo := state.Todo{
			Subject: "Review code",
			Status:  "pending",
		}
		output := r.RenderTodo(todo)

		if !strings.Contains(output, "○") {
			t.Error("expected circle for pending todo")
		}
		if !strings.Contains(output, "Review code") {
			t.Error("expected todo subject in output")
		}
	})

	t.Run("uses activeForm fallback for completed", func(t *testing.T) {
		todo := state.Todo{
			Subject:    "",
			ActiveForm: "Building project",
			Status:     "completed",
		}
		output := r.RenderTodo(todo)

		if !strings.Contains(output, "Building project") {
			t.Error("expected activeForm fallback for completed todo")
		}
	})

	t.Run("uses subject fallback for in_progress", func(t *testing.T) {
		todo := state.Todo{
			Subject:    "Test feature",
			ActiveForm: "",
			Status:     "in_progress",
		}
		output := r.RenderTodo(todo)

		if !strings.Contains(output, "Test feature") {
			t.Error("expected subject fallback for in_progress todo")
		}
	})
}

func TestSidebarRenderer_RenderTodos(t *testing.T) {
	r := NewSidebarRenderer(NewSidebarStyles(), newTestSpinner())

	t.Run("empty todos returns empty string", func(t *testing.T) {
		output := r.RenderTodos(nil)
		if output != "" {
			t.Errorf("expected empty string for nil todos, got %q", output)
		}

		output = r.RenderTodos([]state.Todo{})
		if output != "" {
			t.Errorf("expected empty string for empty todos, got %q", output)
		}
	})

	t.Run("renders todos with header", func(t *testing.T) {
		todos := []state.Todo{
			{Subject: "Task 1", Status: "completed"},
			{Subject: "Task 2", Status: "in_progress"},
			{Subject: "Task 3", Status: "pending"},
		}
		output := r.RenderTodos(todos)

		if !strings.Contains(output, "Tasks") {
			t.Error("expected 'Tasks' header in output")
		}
		if !strings.Contains(output, "Task 1") {
			t.Error("expected Task 1 in output")
		}
		if !strings.Contains(output, "Task 2") {
			t.Error("expected Task 2 in output")
		}
		if !strings.Contains(output, "Task 3") {
			t.Error("expected Task 3 in output")
		}
	})
}

func TestSidebarRenderer_Render(t *testing.T) {
	r := NewSidebarRenderer(NewSidebarStyles(), newTestSpinner())

	t.Run("renders complete sidebar", func(t *testing.T) {
		s := state.NewState()
		s.Model = "claude-3-opus"
		s.TurnCount = 5
		s.TotalCost = 0.1234
		s.Prompt = "Hello"
		s.Todos = []state.Todo{
			{Subject: "Task 1", Status: "completed"},
		}

		output := r.Render(s, 40)

		// Check all sections are present
		if !strings.Contains(output, "claude") {
			t.Error("expected logo in output")
		}
		if !strings.Contains(output, "Hello") {
			t.Error("expected prompt in output")
		}
		if !strings.Contains(output, "claude-3-opus") {
			t.Error("expected model in output")
		}
		if !strings.Contains(output, "5") {
			t.Error("expected turn count in output")
		}
		if !strings.Contains(output, "Task 1") {
			t.Error("expected todo in output")
		}
	})

	t.Run("renders tool in progress", func(t *testing.T) {
		s := state.NewState()
		s.ToolInProgress = true
		s.CurrentTool = "Read"
		s.CurrentToolInput = "test.go"

		output := r.Render(s, 40)

		if !strings.Contains(output, "Running") {
			t.Error("expected Running header in output")
		}
		if !strings.Contains(output, "Read") {
			t.Error("expected tool name in output")
		}
	})

	t.Run("omits tool section when not in progress", func(t *testing.T) {
		s := state.NewState()
		s.ToolInProgress = false
		s.CurrentTool = "Read" // Set but not in progress

		output := r.Render(s, 40)

		// Running header should not appear
		if strings.Contains(output, "Running") {
			t.Error("expected no Running header when tool not in progress")
		}
	})
}

func TestRenderSidebar(t *testing.T) {
	// Test backward compatibility function
	s := state.NewState()
	s.Model = "test-model"
	s.TurnCount = 1
	s.TotalCost = 0.01

	styles := NewSidebarStyles()
	sp := newTestSpinner()

	output := RenderSidebar(s, sp, 40, styles)

	if output == "" {
		t.Error("expected non-empty output from RenderSidebar")
	}
	if !strings.Contains(output, "test-model") {
		t.Error("expected model name in output")
	}
}

func TestNewSidebarStyles(t *testing.T) {
	styles := NewSidebarStyles()

	// Verify Container style is initialized (the only lipgloss style remaining)
	// All text styling is now handled by Ultraviolet functions in style/uvstyle.go
	if styles.Container.GetWidth() != sidebarWidth {
		t.Errorf("Container width = %d, want %d", styles.Container.GetWidth(), sidebarWidth)
	}
}

func TestNewHeaderStyles(t *testing.T) {
	styles := NewHeaderStyles()

	// Verify Modal style has border
	if styles.Modal.GetBorderStyle() != lipgloss.RoundedBorder() {
		t.Error("expected modal to have rounded border")
	}
}

func TestRenderHeader(t *testing.T) {
	t.Run("renders single line with all info", func(t *testing.T) {
		s := state.NewState()
		s.Model = "test-model"
		s.TurnCount = 3
		s.TotalCost = 0.05

		output := RenderHeader(s, 100)

		if output == "" {
			t.Error("expected non-empty output from RenderHeader")
		}
		if !strings.Contains(output, "VIEWSCREEN") {
			t.Error("expected VIEWSCREEN title in output")
		}
		if !strings.Contains(output, "test-model") {
			t.Error("expected model name in output")
		}
		if !strings.Contains(output, "3") {
			t.Error("expected turn count in output")
		}
		if !strings.Contains(output, "$0.05") {
			t.Error("expected cost in output")
		}
		if !strings.Contains(output, "[d]") {
			t.Error("expected key hint [d] in output")
		}
		if !strings.Contains(output, "─") {
			t.Error("expected decoration in output")
		}
	})

	t.Run("truncates long model names", func(t *testing.T) {
		s := state.NewState()
		s.Model = "very-long-model-name-that-exceeds-limit"
		s.TurnCount = 1
		s.TotalCost = 0

		output := RenderHeader(s, 100)

		// Should not contain the full model name
		if strings.Contains(output, "very-long-model-name-that-exceeds-limit") {
			t.Error("expected long model name to be truncated")
		}
		// Should contain truncation indicator
		if !strings.Contains(output, "..") {
			t.Error("expected truncation indicator for long model name")
		}
	})
}

func TestRenderDetailsModal(t *testing.T) {
	s := state.NewState()
	s.Model = "claude-opus"
	s.TurnCount = 5
	s.TotalCost = 0.1234
	s.Prompt = "Hello world"
	s.ToolInProgress = true
	s.CurrentTool = "Read"
	s.Todos = []state.Todo{
		{Subject: "Task 1", Status: "completed"},
	}

	styles := NewHeaderStyles()
	sp := newTestSpinner()

	output := RenderDetailsModal(s, sp, 100, 40, styles)

	// Check all sections are present
	if !strings.Contains(output, "claude") {
		t.Error("expected logo in output")
	}
	if !strings.Contains(output, "Hello world") {
		t.Error("expected prompt in output")
	}
	if !strings.Contains(output, "claude-opus") {
		t.Error("expected model in output")
	}
	if !strings.Contains(output, "Running") {
		t.Error("expected Running header in output")
	}
	if !strings.Contains(output, "Task 1") {
		t.Error("expected todo in output")
	}
	if !strings.Contains(output, "Press d or Esc to close") {
		t.Error("expected close hint in output")
	}
}

func TestLayoutModeConstants(t *testing.T) {
	// Verify layout mode constants are defined correctly
	if LayoutSidebar != 0 {
		t.Error("expected LayoutSidebar to be 0")
	}
	if LayoutHeader != 1 {
		t.Error("expected LayoutHeader to be 1")
	}
}

func TestLayoutBreakpoint(t *testing.T) {
	if breakpointWidth != 80 {
		t.Errorf("breakpointWidth = %d, want 80", breakpointWidth)
	}
	if headerHeight != 1 {
		t.Errorf("headerHeight = %d, want 1", headerHeight)
	}
}
