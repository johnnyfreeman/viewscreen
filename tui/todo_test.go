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
			Subject: "Fix bug",
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
			Subject:    "",
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
			Subject:    "Fix bug",
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
			Subject:    "Test feature",
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
			Subject: "Review code",
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
			Subject:    "",
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
			Subject: "Unknown task",
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
			{Subject: "Task 1", Status: "completed"},
			{Subject: "Task 2", Status: "in_progress"},
			{Subject: "Task 3", Status: "pending"},
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
			{Subject: "Only task", Status: "pending"},
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

func TestTodoRenderer_Truncation(t *testing.T) {
	r := NewTodoRenderer(20, newTestSpinner()) // narrow width

	todo := state.Todo{
		Subject: "This is a very long task name that should be truncated",
		Status:  "pending",
	}
	output := r.RenderItem(todo)

	// The full subject should not appear
	if strings.Contains(output, "This is a very long task name that should be truncated") {
		t.Error("expected long task name to be truncated")
	}
}
