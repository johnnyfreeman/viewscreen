package user

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/testutil"
)

func TestRenderer_Render_TaskListResult(t *testing.T) {
	var buf bytes.Buffer
	r := NewRenderer(
		WithOutput(&buf),
		WithConfigProvider(testutil.MockConfigProvider{NoColorVal: true}),
		WithStyleApplier(testutil.MockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	toolUseResult := json.RawMessage(`{"tasks":[
		{"id":"1","subject":"Design schema","status":"completed","blockedBy":[]},
		{"id":"2","subject":"Build API","status":"in_progress","blockedBy":[]},
		{"id":"3","subject":"Write tests","status":"pending","blockedBy":[]}
	]}`)

	event := Event{
		Message:       Message{Role: "user", Content: []ToolResultContent{}},
		ToolUseResult: toolUseResult,
	}

	r.Render(event)
	output := buf.String()

	if !strings.Contains(output, "[SUCCESS:✓]") || !strings.Contains(output, "[MUTED:Design schema]") {
		t.Errorf("expected completed task with checkmark and muted content, got: %q", output)
	}
	if !strings.Contains(output, "[WARNING:→]") || !strings.Contains(output, "Build API") {
		t.Errorf("expected in_progress task with arrow, got: %q", output)
	}
	if !strings.Contains(output, "[MUTED:○]") || !strings.Contains(output, "[MUTED:Write tests]") {
		t.Errorf("expected pending task with muted circle, got: %q", output)
	}
}

func TestRenderer_Render_TaskListResult_Empty(t *testing.T) {
	var buf bytes.Buffer
	r := NewRenderer(
		WithOutput(&buf),
		WithConfigProvider(testutil.MockConfigProvider{NoColorVal: true}),
		WithStyleApplier(testutil.MockStyleApplier{}),
		WithCodeHighlighter(mockCodeHighlighter{}),
	)

	event := Event{
		Message:       Message{Role: "user", Content: []ToolResultContent{}},
		ToolUseResult: json.RawMessage(`{"tasks":[]}`),
	}

	r.Render(event)

	// An empty task list is not handled by TaskListRenderer; it must not panic
	// and must not emit status indicators.
	if out := buf.String(); strings.Contains(out, "✓") || strings.Contains(out, "○") {
		t.Errorf("expected no status indicators for empty task list, got: %q", out)
	}
}

func TestTaskCreateRenderer(t *testing.T) {
	r := NewTaskCreateRenderer(testutil.MockStyleApplier{})

	t.Run("renders created task id", func(t *testing.T) {
		out := render.StringOutput()
		ctx := &RenderContext{Output: out, ToolName: "TaskCreate"}
		if !r.TryRender(ctx, json.RawMessage(`{"task":{"id":"7","subject":"Build API"}}`)) {
			t.Fatal("expected TryRender to handle TaskCreate result")
		}
		if !strings.Contains(out.String(), "created task #7") {
			t.Errorf("output = %q, want 'created task #7'", out.String())
		}
	})

	t.Run("ignores result from a different tool", func(t *testing.T) {
		out := render.StringOutput()
		ctx := &RenderContext{Output: out, ToolName: "TaskGet"}
		if r.TryRender(ctx, json.RawMessage(`{"task":{"id":"7","subject":"Build API"}}`)) {
			t.Error("TaskCreateRenderer must not claim a TaskGet result")
		}
	})
}

func TestTaskUpdateRenderer(t *testing.T) {
	r := NewTaskUpdateRenderer(testutil.MockStyleApplier{})

	t.Run("renders status change", func(t *testing.T) {
		out := render.StringOutput()
		ctx := &RenderContext{Output: out, ToolName: "TaskUpdate"}
		ok := r.TryRender(ctx, json.RawMessage(
			`{"success":true,"taskId":"1","updatedFields":["status"],"statusChange":{"from":"pending","to":"in_progress"}}`))
		if !ok {
			t.Fatal("expected TryRender to handle TaskUpdate result")
		}
		if got := out.String(); !strings.Contains(got, "pending") || !strings.Contains(got, "in_progress") || !strings.Contains(got, "→") {
			t.Errorf("output = %q, want 'pending → in_progress'", got)
		}
	})

	t.Run("falls back to updated fields when no status change", func(t *testing.T) {
		out := render.StringOutput()
		ctx := &RenderContext{Output: out, ToolName: "TaskUpdate"}
		if !r.TryRender(ctx, json.RawMessage(`{"success":true,"taskId":"1","updatedFields":["subject"]}`)) {
			t.Fatal("expected TryRender to handle TaskUpdate result")
		}
		if !strings.Contains(out.String(), "updated subject") {
			t.Errorf("output = %q, want 'updated subject'", out.String())
		}
	})
}
