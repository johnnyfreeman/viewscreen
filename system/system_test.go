package system

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/style"
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

	if r.styleApplier == nil {
		t.Error("expected styleApplier to be non-nil")
	}

	if r.config == nil {
		t.Error("expected config to be non-nil")
	}
}

func TestNewRendererWithOptions(t *testing.T) {
	t.Run("with custom output", func(t *testing.T) {
		buf := &bytes.Buffer{}
		r := NewRendererWithOptions(WithOutput(buf))

		if r.output != buf {
			t.Error("expected custom output writer")
		}
	})

	t.Run("with custom style applier", func(t *testing.T) {
		mock := &testutil.TrackingStyleApplier{}
		r := NewRendererWithOptions(WithStyleApplier(mock))

		if r.styleApplier != mock {
			t.Error("expected custom style applier")
		}
	})

	t.Run("with custom config provider", func(t *testing.T) {
		mock := testutil.MockConfigProvider{VerboseVal: true}
		r := NewRendererWithOptions(WithConfigProvider(mock))

		if r.config != mock {
			t.Error("expected custom config provider")
		}
	})

	t.Run("with multiple options", func(t *testing.T) {
		buf := &bytes.Buffer{}
		styleMock := &testutil.TrackingStyleApplier{}
		configMock := testutil.MockConfigProvider{}

		r := NewRendererWithOptions(
			WithOutput(buf),
			WithStyleApplier(styleMock),
			WithConfigProvider(configMock),
		)

		if r.output != buf {
			t.Error("expected custom output writer")
		}
		if r.styleApplier != styleMock {
			t.Error("expected custom style applier")
		}
		if r.config != configMock {
			t.Error("expected custom config provider")
		}
	})
}

func TestRenderer_Render_BasicEvent(t *testing.T) {
	style.Init(false) // Enable color
	output := &bytes.Buffer{}
	styleMock := testutil.MockStyleApplier{NoColorVal: false}
	configMock := testutil.MockConfigProvider{VerboseVal: false}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithStyleApplier(styleMock),
		WithConfigProvider(configMock),
	)

	event := Event{
		Model:             "claude-3-opus",
		ClaudeCodeVersion: "1.0.0",
		CWD:               "/home/user/project",
		Tools:             []string{"Read", "Write", "Bash"},
	}

	r.Render(event)

	result := output.String()
	stripped := testutil.StripANSI(result)

	// Check header content (gradient applied, so strip ANSI to verify text)
	if !strings.Contains(stripped, "Session Started") {
		t.Errorf("expected 'Session Started' in output, got: %s", stripped)
	}

	// Check model line (uses mock, so expect mock output)
	if !strings.Contains(result, "[MUTED:Model:] claude-3-opus") {
		t.Errorf("expected model line in output, got: %s", result)
	}

	// Check version line
	if !strings.Contains(result, "[MUTED:Version:] 1.0.0") {
		t.Error("expected version line in output")
	}

	// Check CWD line
	if !strings.Contains(result, "[MUTED:CWD:] /home/user/project") {
		t.Error("expected CWD line in output")
	}

	// Check tools count
	if !strings.Contains(result, "[MUTED:Tools:] 3 available") {
		t.Error("expected tools count in output")
	}
}

func TestRenderer_Render_NoColorMode(t *testing.T) {
	style.Init(true) // Disable color
	defer style.Init(false) // Restore for other tests

	output := &bytes.Buffer{}
	styleMock := testutil.MockStyleApplier{NoColorVal: true}
	configMock := testutil.MockConfigProvider{VerboseVal: false}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithStyleApplier(styleMock),
		WithConfigProvider(configMock),
	)

	event := Event{
		Model:             "claude-3-sonnet",
		ClaudeCodeVersion: "2.0.0",
		CWD:               "/tmp",
		Tools:             []string{"Read"},
	}

	r.Render(event)

	result := output.String()

	// In no-color mode, header should be plain text (bold only, no color codes)
	if !strings.Contains(result, "Session Started") {
		t.Errorf("expected 'Session Started' in output, got: %s", result)
	}

	// Should not contain color escape sequences (only bold \x1b[1m is allowed)
	if strings.Contains(result, "\x1b[38") {
		t.Error("no-color mode should not have color codes")
	}
}

func TestRenderer_Render_VerboseWithAgents(t *testing.T) {
	style.Init(true) // No color for predictable output
	defer style.Init(false)

	output := &bytes.Buffer{}
	styleMock := testutil.MockStyleApplier{NoColorVal: true}
	configMock := testutil.MockConfigProvider{VerboseVal: true}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithStyleApplier(styleMock),
		WithConfigProvider(configMock),
	)

	event := Event{
		Model:             "claude-3-opus",
		ClaudeCodeVersion: "1.0.0",
		CWD:               "/home/user",
		Tools:             []string{"Read", "Write"},
		Agents:            []string{"coder", "reviewer", "tester"},
	}

	r.Render(event)

	result := output.String()

	// Check agents line is present when verbose and agents exist
	if !strings.Contains(result, "[MUTED:Agents:] coder, reviewer, tester") {
		t.Errorf("expected agents line in verbose mode, got: %s", result)
	}
}

func TestRenderer_Render_VerboseWithoutAgents(t *testing.T) {
	output := &bytes.Buffer{}
	styleMock := testutil.MockStyleApplier{NoColorVal: true}
	configMock := testutil.MockConfigProvider{VerboseVal: true}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithStyleApplier(styleMock),
		WithConfigProvider(configMock),
	)

	event := Event{
		Model:             "claude-3-opus",
		ClaudeCodeVersion: "1.0.0",
		CWD:               "/home/user",
		Tools:             []string{"Read"},
		Agents:            []string{}, // Empty agents
	}

	r.Render(event)

	result := output.String()

	// Agents line should not appear when agents list is empty
	if strings.Contains(result, "Agents:") {
		t.Error("agents line should not appear when agents list is empty")
	}
}

func TestRenderer_Render_NonVerboseWithAgents(t *testing.T) {
	output := &bytes.Buffer{}
	styleMock := testutil.MockStyleApplier{NoColorVal: true}
	configMock := testutil.MockConfigProvider{VerboseVal: false}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithStyleApplier(styleMock),
		WithConfigProvider(configMock),
	)

	event := Event{
		Model:             "claude-3-opus",
		ClaudeCodeVersion: "1.0.0",
		CWD:               "/home/user",
		Tools:             []string{"Read"},
		Agents:            []string{"coder", "reviewer"}, // Agents present but verbose is false
	}

	r.Render(event)

	result := output.String()

	// Agents line should not appear when not in verbose mode
	if strings.Contains(result, "Agents:") {
		t.Error("agents line should not appear when not in verbose mode")
	}
}

func TestRenderer_Render_EmptyTools(t *testing.T) {
	output := &bytes.Buffer{}
	styleMock := testutil.MockStyleApplier{NoColorVal: true}
	configMock := testutil.MockConfigProvider{VerboseVal: false}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithStyleApplier(styleMock),
		WithConfigProvider(configMock),
	)

	event := Event{
		Model:             "claude-3-opus",
		ClaudeCodeVersion: "1.0.0",
		CWD:               "/home/user",
		Tools:             []string{},
	}

	r.Render(event)

	result := output.String()

	// Should show 0 tools
	if !strings.Contains(result, "[MUTED:Tools:] 0 available") {
		t.Errorf("expected 0 tools in output, got: %s", result)
	}
}

func TestRenderer_Render_OutputFormat(t *testing.T) {
	output := &bytes.Buffer{}
	styleMock := testutil.MockStyleApplier{NoColorVal: true}
	configMock := testutil.MockConfigProvider{VerboseVal: false}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithStyleApplier(styleMock),
		WithConfigProvider(configMock),
	)

	event := Event{
		Model:             "model",
		ClaudeCodeVersion: "version",
		CWD:               "cwd",
		Tools:             []string{"tool1", "tool2"},
	}

	r.Render(event)

	result := output.String()
	lines := strings.Split(strings.TrimRight(result, "\n"), "\n")

	// Should have 5 lines: header, model, version, cwd, tools (trailing newline is trimmed)
	if len(lines) != 5 {
		t.Errorf("expected 5 lines, got %d: %v", len(lines), lines)
	}

	// Check proper prefixes
	if !strings.HasPrefix(lines[1], "  ⎿  ") {
		t.Errorf("second line should start with OutputPrefix, got: %q", lines[1])
	}

	for i := 2; i <= 4; i++ {
		if !strings.HasPrefix(lines[i], "     ") {
			t.Errorf("line %d should start with OutputContinue, got: %q", i+1, lines[i])
		}
	}
}

func TestEvent_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantErr  bool
		validate func(t *testing.T, e Event)
	}{
		{
			name: "basic system event",
			json: `{
				"type": "system",
				"subtype": "init",
				"model": "claude-3-opus",
				"claude_code_version": "1.0.0",
				"cwd": "/home/user/project",
				"tools": ["Read", "Write", "Bash"],
				"permissionMode": "default",
				"agents": ["coder"]
			}`,
			wantErr: false,
			validate: func(t *testing.T, e Event) {
				if e.Type != "system" {
					t.Errorf("expected Type 'system', got %q", e.Type)
				}
				if e.Subtype != "init" {
					t.Errorf("expected Subtype 'init', got %q", e.Subtype)
				}
				if e.Model != "claude-3-opus" {
					t.Errorf("expected Model 'claude-3-opus', got %q", e.Model)
				}
				if e.ClaudeCodeVersion != "1.0.0" {
					t.Errorf("expected ClaudeCodeVersion '1.0.0', got %q", e.ClaudeCodeVersion)
				}
				if e.CWD != "/home/user/project" {
					t.Errorf("expected CWD '/home/user/project', got %q", e.CWD)
				}
				if len(e.Tools) != 3 {
					t.Errorf("expected 3 tools, got %d", len(e.Tools))
				}
				if e.PermissionMode != "default" {
					t.Errorf("expected PermissionMode 'default', got %q", e.PermissionMode)
				}
				if len(e.Agents) != 1 || e.Agents[0] != "coder" {
					t.Errorf("expected agents ['coder'], got %v", e.Agents)
				}
			},
		},
		{
			name: "event with empty tools and agents",
			json: `{
				"type": "system",
				"model": "claude-3-sonnet",
				"claude_code_version": "2.0.0",
				"cwd": "/tmp",
				"tools": [],
				"agents": []
			}`,
			wantErr: false,
			validate: func(t *testing.T, e Event) {
				if len(e.Tools) != 0 {
					t.Errorf("expected empty tools, got %d", len(e.Tools))
				}
				if len(e.Agents) != 0 {
					t.Errorf("expected empty agents, got %d", len(e.Agents))
				}
			},
		},
		{
			name: "event with many tools",
			json: `{
				"type": "system",
				"model": "claude-3-opus",
				"claude_code_version": "1.0.0",
				"cwd": "/",
				"tools": ["Read", "Write", "Edit", "Bash", "Glob", "Grep", "WebFetch", "Task"]
			}`,
			wantErr: false,
			validate: func(t *testing.T, e Event) {
				if len(e.Tools) != 8 {
					t.Errorf("expected 8 tools, got %d", len(e.Tools))
				}
			},
		},
		{
			name: "event with multiple agents",
			json: `{
				"type": "system",
				"model": "claude-3-opus",
				"claude_code_version": "1.0.0",
				"cwd": "/",
				"tools": [],
				"agents": ["coder", "reviewer", "tester", "documenter"]
			}`,
			wantErr: false,
			validate: func(t *testing.T, e Event) {
				if len(e.Agents) != 4 {
					t.Errorf("expected 4 agents, got %d", len(e.Agents))
				}
				expected := []string{"coder", "reviewer", "tester", "documenter"}
				for i, agent := range e.Agents {
					if agent != expected[i] {
						t.Errorf("agent %d: expected %q, got %q", i, expected[i], agent)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event Event
			err := json.Unmarshal([]byte(tt.json), &event)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, event)
			}
		})
	}
}


func TestDefaultStyleApplier(t *testing.T) {
	dsa := render.DefaultStyleApplier{}

	// Test that the methods don't panic and return expected types
	_ = dsa.NoColor()
	_ = dsa.ApplyThemeBoldGradient("test")
	_ = dsa.SessionHeaderRender("test")
	_ = dsa.MutedText("test")

	if dsa.OutputPrefix() != "  ⎿  " {
		t.Errorf("expected output prefix '  ⎿  ', got %q", dsa.OutputPrefix())
	}

	if dsa.OutputContinue() != "     " {
		t.Errorf("expected output continue '     ', got %q", dsa.OutputContinue())
	}
}

func TestDefaultConfigProvider(t *testing.T) {
	dp := config.DefaultProvider{}

	// Test that the methods don't panic
	// The actual values depend on config globals which we can't easily control here
	_ = dp.IsVerbose()
	_ = dp.NoColor()
	_ = dp.ShowUsage()
}

func TestRenderer_Render_MutedTextStyleCalls(t *testing.T) {
	output := &bytes.Buffer{}
	styleMock := &testutil.TrackingStyleApplier{MockStyleApplier: testutil.MockStyleApplier{NoColorVal: true}}
	configMock := testutil.MockConfigProvider{VerboseVal: true}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithStyleApplier(styleMock),
		WithConfigProvider(configMock),
	)

	event := Event{
		Model:             "claude-3-opus",
		ClaudeCodeVersion: "1.0.0",
		CWD:               "/home/user",
		Tools:             []string{"Read"},
		Agents:            []string{"coder"},
	}

	r.Render(event)

	// Should have called MutedText for: Model, Version, CWD, Tools, Agents
	expectedMutedTextCalls := []string{"Model:", "Version:", "CWD:", "Tools:", "Agents:"}
	if len(styleMock.MutedTextCalls) != len(expectedMutedTextCalls) {
		t.Errorf("expected %d MutedText calls, got %d: %v", len(expectedMutedTextCalls), len(styleMock.MutedTextCalls), styleMock.MutedTextCalls)
	}

	for i, expected := range expectedMutedTextCalls {
		if i < len(styleMock.MutedTextCalls) && styleMock.MutedTextCalls[i] != expected {
			t.Errorf("MutedText call %d: expected %q, got %q", i, expected, styleMock.MutedTextCalls[i])
		}
	}
}

func TestRenderer_Render_SpecialCharacters(t *testing.T) {
	output := &bytes.Buffer{}
	styleMock := testutil.MockStyleApplier{NoColorVal: true}
	configMock := testutil.MockConfigProvider{VerboseVal: false}

	r := NewRendererWithOptions(
		WithOutput(output),
		WithStyleApplier(styleMock),
		WithConfigProvider(configMock),
	)

	event := Event{
		Model:             "claude-3-opus-20240229",
		ClaudeCodeVersion: "1.0.0-beta+build.123",
		CWD:               "/home/user/my project/with spaces",
		Tools:             []string{"Read", "Write"},
	}

	r.Render(event)

	result := output.String()

	// Verify special characters are preserved
	if !strings.Contains(result, "claude-3-opus-20240229") {
		t.Error("model with dashes should be preserved")
	}
	if !strings.Contains(result, "1.0.0-beta+build.123") {
		t.Error("semver with build metadata should be preserved")
	}
	if !strings.Contains(result, "/home/user/my project/with spaces") {
		t.Error("path with spaces should be preserved")
	}
}
