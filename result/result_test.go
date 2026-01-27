package result

import (
	"bytes"
	"strings"
	"testing"

	"github.com/johnnyfreeman/viewscreen/testutil"
)

func TestNewRenderer(t *testing.T) {
	r := NewRenderer()

	if r == nil {
		t.Fatal("NewRenderer returned nil")
	}

	if r.output == nil {
		t.Error("expected output to be non-nil")
	}

	if r.config == nil {
		t.Error("expected config to be non-nil")
	}

	if r.styleApplier == nil {
		t.Error("expected styleApplier to be non-nil")
	}
}

func TestNewRendererWithOptions(t *testing.T) {
	t.Run("with custom output", func(t *testing.T) {
		buf := &bytes.Buffer{}
		r := NewRenderer(WithOutput(buf))

		if r.output != buf {
			t.Error("expected custom output writer")
		}
	})

	t.Run("with custom config provider", func(t *testing.T) {
		cp := testutil.MockConfigProvider{ShowUsageVal: true}
		r := NewRenderer(WithConfigProvider(cp))

		if !r.config.ShowUsage() {
			t.Error("expected showUsage to return true")
		}
	})

	t.Run("with custom styleApplier", func(t *testing.T) {
		sa := testutil.MockStyleApplier{NoColorVal: true}
		r := NewRenderer(WithStyleApplier(sa))

		if r.styleApplier == nil {
			t.Error("expected custom styleApplier to be set")
		}
	})

	t.Run("with multiple options", func(t *testing.T) {
		buf := &bytes.Buffer{}
		cp := testutil.MockConfigProvider{ShowUsageVal: true}

		r := NewRenderer(
			WithOutput(buf),
			WithConfigProvider(cp),
			WithStyleApplier(testutil.MockStyleApplier{NoColorVal: true}),
		)

		if r.output != buf {
			t.Error("expected custom output")
		}

		if !r.config.ShowUsage() {
			t.Error("expected showUsage to return true")
		}

		if r.styleApplier == nil {
			t.Error("expected styleApplier to be set")
		}
	})
}

func TestRenderer_Render_Success(t *testing.T) {
	buf := &bytes.Buffer{}
	r := NewRenderer(
		WithOutput(buf),
		WithConfigProvider(testutil.MockConfigProvider{ShowUsageVal: false}),
		WithStyleApplier(testutil.MockStyleApplier{NoColorVal: true}),
	)

	event := Event{
		IsError:       false,
		DurationMS:    5000,
		DurationAPIMS: 3000,
		NumTurns:      10,
		TotalCostUSD:  0.1234,
	}

	r.Render(event)

	output := buf.String()

	// Check header
	if !strings.Contains(output, "Session Complete") {
		t.Error("expected 'Session Complete' in output")
	}

	// Check duration formatting
	if !strings.Contains(output, "5.00s") {
		t.Error("expected duration '5.00s' in output")
	}
	if !strings.Contains(output, "3.00s") {
		t.Error("expected API duration '3.00s' in output")
	}

	// Check turns
	if !strings.Contains(output, "10") {
		t.Error("expected turns '10' in output")
	}

	// Check cost formatting
	if !strings.Contains(output, "$0.1234") {
		t.Error("expected cost '$0.1234' in output")
	}
}

func TestRenderer_Render_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	r := NewRenderer(
		WithOutput(buf),
		WithConfigProvider(testutil.MockConfigProvider{ShowUsageVal: false}),
		WithStyleApplier(testutil.MockStyleApplier{NoColorVal: true}),
	)

	event := Event{
		IsError:       true,
		DurationMS:    2000,
		DurationAPIMS: 1500,
		NumTurns:      3,
		TotalCostUSD:  0.05,
		Errors:        []string{"Something went wrong", "Another error"},
	}

	r.Render(event)

	output := buf.String()

	// Check error header
	if !strings.Contains(output, "Session Error") {
		t.Error("expected 'Session Error' in output")
	}

	// Check errors are listed
	if !strings.Contains(output, "Something went wrong") {
		t.Error("expected first error message in output")
	}
	if !strings.Contains(output, "Another error") {
		t.Error("expected second error message in output")
	}
}

func TestRenderer_Render_WithUsage(t *testing.T) {
	buf := &bytes.Buffer{}
	r := NewRenderer(
		WithOutput(buf),
		WithConfigProvider(testutil.MockConfigProvider{ShowUsageVal: true}),
		WithStyleApplier(testutil.MockStyleApplier{NoColorVal: true}),
	)

	event := Event{
		IsError:       false,
		DurationMS:    1000,
		DurationAPIMS: 800,
		NumTurns:      1,
		TotalCostUSD:  0.01,
		Usage: Usage{
			InputTokens:              1000,
			OutputTokens:             500,
			CacheCreationInputTokens: 100,
			CacheReadInputTokens:     200,
		},
	}

	r.Render(event)

	output := buf.String()

	// Check usage info
	if !strings.Contains(output, "in=1000") {
		t.Errorf("expected input tokens 'in=1000' in output, got %s", output)
	}
	if !strings.Contains(output, "out=500") {
		t.Errorf("expected output tokens 'out=500' in output, got %s", output)
	}
	if !strings.Contains(output, "created=100") {
		t.Errorf("expected cache created 'created=100' in output, got %s", output)
	}
	if !strings.Contains(output, "read=200") {
		t.Errorf("expected cache read 'read=200' in output, got %s", output)
	}
}

func TestRenderer_Render_WithoutUsage(t *testing.T) {
	buf := &bytes.Buffer{}
	r := NewRenderer(
		WithOutput(buf),
		WithConfigProvider(testutil.MockConfigProvider{ShowUsageVal: false}),
		WithStyleApplier(testutil.MockStyleApplier{NoColorVal: true}),
	)

	event := Event{
		IsError:       false,
		DurationMS:    1000,
		DurationAPIMS: 800,
		NumTurns:      1,
		TotalCostUSD:  0.01,
		Usage: Usage{
			InputTokens:  1000,
			OutputTokens: 500,
		},
	}

	r.Render(event)

	output := buf.String()

	// Should NOT contain token counts
	if strings.Contains(output, "Tokens:") {
		t.Error("expected no token info when showUsage is false")
	}
}

func TestRenderer_Render_WithPermissionDenials(t *testing.T) {
	buf := &bytes.Buffer{}
	r := NewRenderer(
		WithOutput(buf),
		WithConfigProvider(testutil.MockConfigProvider{ShowUsageVal: false}),
		WithStyleApplier(testutil.MockStyleApplier{NoColorVal: true}),
	)

	event := Event{
		IsError:       false,
		DurationMS:    1000,
		DurationAPIMS: 800,
		NumTurns:      1,
		TotalCostUSD:  0.01,
		PermissionDenials: []PermissionDenial{
			{
				ToolName:  "Bash",
				ToolUseID: "tool_123",
			},
			{
				ToolName:  "Write",
				ToolUseID: "tool_456",
			},
		},
	}

	r.Render(event)

	output := buf.String()

	// Check permission denials count
	if !strings.Contains(output, "Permission Denials:") {
		t.Error("expected 'Permission Denials:' in output")
	}
	if !strings.Contains(output, "2") {
		t.Error("expected permission denial count '2' in output")
	}

	// Check individual denials
	if !strings.Contains(output, "Bash") {
		t.Error("expected 'Bash' tool name in output")
	}
	if !strings.Contains(output, "tool_123") {
		t.Error("expected 'tool_123' tool use ID in output")
	}
	if !strings.Contains(output, "Write") {
		t.Error("expected 'Write' tool name in output")
	}
	if !strings.Contains(output, "tool_456") {
		t.Error("expected 'tool_456' tool use ID in output")
	}
}

func TestRenderer_Render_NoPermissionDenials(t *testing.T) {
	buf := &bytes.Buffer{}
	r := NewRenderer(
		WithOutput(buf),
		WithConfigProvider(testutil.MockConfigProvider{ShowUsageVal: false}),
		WithStyleApplier(testutil.MockStyleApplier{NoColorVal: true}),
	)

	event := Event{
		IsError:           false,
		DurationMS:        1000,
		DurationAPIMS:     800,
		NumTurns:          1,
		TotalCostUSD:      0.01,
		PermissionDenials: []PermissionDenial{}, // Empty
	}

	r.Render(event)

	output := buf.String()

	// Should NOT contain permission denials section
	if strings.Contains(output, "Permission Denials:") {
		t.Error("expected no permission denials section when empty")
	}
}

func TestRenderer_Render_ZeroValues(t *testing.T) {
	buf := &bytes.Buffer{}
	r := NewRenderer(
		WithOutput(buf),
		WithConfigProvider(testutil.MockConfigProvider{ShowUsageVal: true}),
		WithStyleApplier(testutil.MockStyleApplier{NoColorVal: true}),
	)

	event := Event{
		IsError:       false,
		DurationMS:    0,
		DurationAPIMS: 0,
		NumTurns:      0,
		TotalCostUSD:  0,
		Usage: Usage{
			InputTokens:              0,
			OutputTokens:             0,
			CacheCreationInputTokens: 0,
			CacheReadInputTokens:     0,
		},
	}

	r.Render(event)

	output := buf.String()

	// Should still render with zero values
	if !strings.Contains(output, "Session Complete") {
		t.Error("expected 'Session Complete' even with zero values")
	}
	if !strings.Contains(output, "0.00s") {
		t.Error("expected '0.00s' for zero duration")
	}
	if !strings.Contains(output, "$0.0000") {
		t.Error("expected '$0.0000' for zero cost")
	}
}

func TestRenderer_Render_LargeCost(t *testing.T) {
	buf := &bytes.Buffer{}
	r := NewRenderer(
		WithOutput(buf),
		WithConfigProvider(testutil.MockConfigProvider{ShowUsageVal: false}),
		WithStyleApplier(testutil.MockStyleApplier{NoColorVal: true}),
	)

	event := Event{
		IsError:       false,
		DurationMS:    100000,
		DurationAPIMS: 90000,
		NumTurns:      100,
		TotalCostUSD:  12.3456,
	}

	r.Render(event)

	output := buf.String()

	// Check large cost formatting
	if !strings.Contains(output, "$12.3456") {
		t.Errorf("expected cost '$12.3456' in output, got %s", output)
	}

	// Check large duration
	if !strings.Contains(output, "100.00s") {
		t.Errorf("expected duration '100.00s' in output, got %s", output)
	}
}

func TestUsage_JSONFields(t *testing.T) {
	// Test that Usage struct has correct JSON tags
	u := Usage{
		InputTokens:              100,
		CacheCreationInputTokens: 200,
		CacheReadInputTokens:     300,
		OutputTokens:             400,
	}

	if u.InputTokens != 100 {
		t.Errorf("InputTokens: got %d, want 100", u.InputTokens)
	}
	if u.CacheCreationInputTokens != 200 {
		t.Errorf("CacheCreationInputTokens: got %d, want 200", u.CacheCreationInputTokens)
	}
	if u.CacheReadInputTokens != 300 {
		t.Errorf("CacheReadInputTokens: got %d, want 300", u.CacheReadInputTokens)
	}
	if u.OutputTokens != 400 {
		t.Errorf("OutputTokens: got %d, want 400", u.OutputTokens)
	}
}

func TestModelUsage_Fields(t *testing.T) {
	// Test ModelUsage struct
	mu := ModelUsage{
		InputTokens:              100,
		OutputTokens:             200,
		CacheReadInputTokens:     50,
		CacheCreationInputTokens: 25,
		CostUSD:                  0.05,
		ContextWindow:            100000,
		MaxOutputTokens:          4096,
	}

	if mu.InputTokens != 100 {
		t.Errorf("InputTokens: got %d, want 100", mu.InputTokens)
	}
	if mu.CostUSD != 0.05 {
		t.Errorf("CostUSD: got %f, want 0.05", mu.CostUSD)
	}
	if mu.ContextWindow != 100000 {
		t.Errorf("ContextWindow: got %d, want 100000", mu.ContextWindow)
	}
}

func TestPermissionDenial_Fields(t *testing.T) {
	pd := PermissionDenial{
		ToolName:  "Write",
		ToolUseID: "use_123",
		ToolInput: []byte(`{"file_path": "/test.txt"}`),
	}

	if pd.ToolName != "Write" {
		t.Errorf("ToolName: got %q, want 'Write'", pd.ToolName)
	}
	if pd.ToolUseID != "use_123" {
		t.Errorf("ToolUseID: got %q, want 'use_123'", pd.ToolUseID)
	}
}

func TestEvent_Fields(t *testing.T) {
	event := Event{
		Subtype:       "test_subtype",
		IsError:       true,
		DurationMS:    5000,
		DurationAPIMS: 4000,
		NumTurns:      5,
		Result:        "test result",
		TotalCostUSD:  0.25,
		Usage: Usage{
			InputTokens:  1000,
			OutputTokens: 500,
		},
		ModelUsage: map[string]ModelUsage{
			"claude-3": {InputTokens: 1000, OutputTokens: 500},
		},
		PermissionDenials: []PermissionDenial{
			{ToolName: "Bash", ToolUseID: "id1"},
		},
		Errors: []string{"error1", "error2"},
	}

	if event.Subtype != "test_subtype" {
		t.Errorf("Subtype: got %q, want 'test_subtype'", event.Subtype)
	}
	if !event.IsError {
		t.Error("IsError: got false, want true")
	}
	if event.DurationMS != 5000 {
		t.Errorf("DurationMS: got %d, want 5000", event.DurationMS)
	}
	if event.NumTurns != 5 {
		t.Errorf("NumTurns: got %d, want 5", event.NumTurns)
	}
	if event.Result != "test result" {
		t.Errorf("Result: got %q, want 'test result'", event.Result)
	}
	if len(event.ModelUsage) != 1 {
		t.Errorf("ModelUsage: got %d entries, want 1", len(event.ModelUsage))
	}
	if len(event.PermissionDenials) != 1 {
		t.Errorf("PermissionDenials: got %d, want 1", len(event.PermissionDenials))
	}
	if len(event.Errors) != 2 {
		t.Errorf("Errors: got %d, want 2", len(event.Errors))
	}
}

func TestRenderer_Render_WithColor(t *testing.T) {
	buf := &bytes.Buffer{}
	r := NewRenderer(
		WithOutput(buf),
		WithConfigProvider(testutil.MockConfigProvider{ShowUsageVal: false}),
		WithStyleApplier(testutil.MockStyleApplier{NoColorVal: false}), // Color enabled
	)

	event := Event{
		IsError:       false,
		DurationMS:    1000,
		DurationAPIMS: 800,
		NumTurns:      1,
		TotalCostUSD:  0.01,
	}

	r.Render(event)

	output := buf.String()

	// Should contain Session Complete (with or without ANSI codes)
	if !strings.Contains(output, "Session Complete") {
		t.Error("expected 'Session Complete' in output")
	}
}

func TestRenderer_Render_ErrorWithColor(t *testing.T) {
	buf := &bytes.Buffer{}
	r := NewRenderer(
		WithOutput(buf),
		WithConfigProvider(testutil.MockConfigProvider{ShowUsageVal: false}),
		WithStyleApplier(testutil.MockStyleApplier{NoColorVal: false}), // Color enabled
	)

	event := Event{
		IsError:       true,
		DurationMS:    1000,
		DurationAPIMS: 800,
		NumTurns:      1,
		TotalCostUSD:  0.01,
		Errors:        []string{"Test error"},
	}

	r.Render(event)

	output := buf.String()

	// Should contain Session Error (with or without ANSI codes)
	if !strings.Contains(output, "Session Error") {
		t.Error("expected 'Session Error' in output")
	}
	if !strings.Contains(output, "Test error") {
		t.Error("expected error message in output")
	}
}
