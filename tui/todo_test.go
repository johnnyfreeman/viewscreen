package tui

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/spinner"
	"github.com/johnnyfreeman/viewscreen/state"
	"github.com/johnnyfreeman/viewscreen/style"
)

func init() {
	// Initialize style with noColor for consistent test output
	style.Init(true)
}

func newTestTodoRenderer() *TodoRenderer {
	sp := spinner.New(
		spinner.WithSpinner(spinner.Spinner{Frames: []string{"⠋"}, FPS: 1}),
	)
	return NewTodoRenderer(30, sp)
}

func TestNewTodoRenderer(t *testing.T) {
	sp := spinner.New()
	r := NewTodoRenderer(30, sp)

	if r == nil {
		t.Fatal("NewTodoRenderer returned nil")
	}
	if r.width != 30 {
		t.Errorf("width = %d, want 30", r.width)
	}
}

func TestTodoRenderer_RenderItem(t *testing.T) {
	r := newTestTodoRenderer()

	t.Run("renders completed todo with checkmark", func(t *testing.T) {
		todo := state.Todo{
			Content: "Fix bug",
			Status:  "completed",
		}
		output := r.RenderItem(todo)

		if !strings.Contains(output, "✓") {
			t.Error("expected checkmark for completed todo")
		}
		if !strings.Contains(output, "Fix bug") {
			t.Error("expected todo subject in output")
		}
	})

	t.Run("completed todo uses activeForm fallback", func(t *testing.T) {
		todo := state.Todo{
			Content:    "",
			ActiveForm: "Building project",
			Status:     "completed",
		}
		output := r.RenderItem(todo)

		if !strings.Contains(output, "Building project") {
			t.Error("expected activeForm fallback for completed todo")
		}
	})

	t.Run("renders in_progress todo with spinner", func(t *testing.T) {
		todo := state.Todo{
			Content:    "Fix bug",
			ActiveForm: "Fixing bug",
			Status:     "in_progress",
		}
		output := r.RenderItem(todo)

		// Should prefer ActiveForm for in_progress
		if !strings.Contains(output, "Fixing bug") {
			t.Error("expected active form in output")
		}
	})

	t.Run("in_progress uses subject fallback", func(t *testing.T) {
		todo := state.Todo{
			Content:    "Test feature",
			ActiveForm: "",
			Status:     "in_progress",
		}
		output := r.RenderItem(todo)

		if !strings.Contains(output, "Test feature") {
			t.Error("expected subject fallback for in_progress todo")
		}
	})

	t.Run("renders pending todo with circle", func(t *testing.T) {
		todo := state.Todo{
			Content: "Review code",
			Status:  "pending",
		}
		output := r.RenderItem(todo)

		if !strings.Contains(output, "○") {
			t.Error("expected circle for pending todo")
		}
		if !strings.Contains(output, "Review code") {
			t.Error("expected todo subject in output")
		}
	})

	t.Run("pending uses activeForm fallback", func(t *testing.T) {
		todo := state.Todo{
			Content:    "",
			ActiveForm: "Waiting for review",
			Status:     "pending",
		}
		output := r.RenderItem(todo)

		if !strings.Contains(output, "Waiting for review") {
			t.Error("expected activeForm fallback for pending todo")
		}
	})

	t.Run("unknown status treated as pending", func(t *testing.T) {
		todo := state.Todo{
			Content: "Unknown task",
			Status:  "unknown",
		}
		output := r.RenderItem(todo)

		if !strings.Contains(output, "○") {
			t.Error("expected circle for unknown status todo")
		}
	})
}

func TestTodoRenderer_RenderList(t *testing.T) {
	r := newTestTodoRenderer()

	t.Run("empty todos returns empty string", func(t *testing.T) {
		output := r.RenderList(nil)
		if output != "" {
			t.Errorf("expected empty string for nil todos, got %q", output)
		}

		output = r.RenderList([]state.Todo{})
		if output != "" {
			t.Errorf("expected empty string for empty todos, got %q", output)
		}
	})

	t.Run("renders todos with header", func(t *testing.T) {
		todos := []state.Todo{
			{Content: "Task 1", Status: "completed"},
			{Content: "Task 2", Status: "in_progress"},
			{Content: "Task 3", Status: "pending"},
		}
		output := r.RenderList(todos)

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

	t.Run("single todo renders correctly", func(t *testing.T) {
		todos := []state.Todo{
			{Content: "Only task", Status: "pending"},
		}
		output := r.RenderList(todos)

		if !strings.Contains(output, "Tasks") {
			t.Error("expected 'Tasks' header in output")
		}
		if !strings.Contains(output, "Only task") {
			t.Error("expected task in output")
		}
	})
}

func TestTodoRenderer_RenderProgressBar(t *testing.T) {
	r := newTestTodoRenderer()

	t.Run("zero total returns empty string", func(t *testing.T) {
		output := r.RenderProgressBar(0, 0)
		if output != "" {
			t.Errorf("expected empty string for zero total, got %q", output)
		}
	})

	t.Run("renders filled and empty blocks", func(t *testing.T) {
		output := r.RenderProgressBar(3, 6)

		if !strings.Contains(output, "█") {
			t.Error("expected filled blocks in progress bar")
		}
		if !strings.Contains(output, "░") {
			t.Error("expected empty blocks in progress bar")
		}
		if !strings.Contains(output, "3/6") {
			t.Errorf("expected '3/6' count in output, got %q", output)
		}
	})

	t.Run("all completed", func(t *testing.T) {
		output := r.RenderProgressBar(5, 5)

		if !strings.Contains(output, "█") {
			t.Error("expected filled blocks in progress bar")
		}
		if strings.Contains(output, "░") {
			t.Error("expected no empty blocks when fully complete")
		}
		if !strings.Contains(output, "5/5") {
			t.Errorf("expected '5/5' count in output, got %q", output)
		}
	})

	t.Run("none completed", func(t *testing.T) {
		output := r.RenderProgressBar(0, 4)

		if !strings.Contains(output, "░") {
			t.Error("expected empty blocks in progress bar")
		}
		if !strings.Contains(output, "0/4") {
			t.Errorf("expected '0/4' count in output, got %q", output)
		}
	})

	t.Run("narrow width still renders", func(t *testing.T) {
		narrow := NewTodoRenderer(10, newTestSpinner())
		output := narrow.RenderProgressBar(1, 3)

		if !strings.Contains(output, "1/3") {
			t.Errorf("expected '1/3' count in narrow output, got %q", output)
		}
		if !strings.Contains(output, "█") {
			t.Error("expected at least some filled blocks in narrow bar")
		}
	})
}

func TestTodoRenderer_RenderList_WithProgressBar(t *testing.T) {
	r := newTestTodoRenderer()

	t.Run("single todo has no progress bar", func(t *testing.T) {
		todos := []state.Todo{
			{Content: "Only task", Status: "pending"},
		}
		output := r.RenderList(todos)

		if strings.Contains(output, "█") || strings.Contains(output, "░") {
			t.Error("expected no progress bar for single todo")
		}
	})

	t.Run("two or more todos show progress bar", func(t *testing.T) {
		todos := []state.Todo{
			{Content: "Task 1", Status: "completed"},
			{Content: "Task 2", Status: "pending"},
			{Content: "Task 3", Status: "pending"},
		}
		output := r.RenderList(todos)

		if !strings.Contains(output, "1/3") {
			t.Errorf("expected '1/3' progress in output, got %q", output)
		}
		if !strings.Contains(output, "█") {
			t.Error("expected filled blocks in progress bar")
		}
	})
}

func TestTodoRenderer_Truncation(t *testing.T) {
	r := NewTodoRenderer(20, newTestSpinner()) // narrow width

	todo := state.Todo{
		Content: "This is a very long task name that should be truncated",
		Status:  "pending",
	}
	output := r.RenderItem(todo)

	// The full subject should not appear
	if strings.Contains(output, "This is a very long task name that should be truncated") {
		t.Error("expected long task name to be truncated")
	}
}
