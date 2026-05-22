package agent

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/johnnyfreeman/viewscreen/config"
)

type stubProcess struct{}

func (stubProcess) Stdout() io.ReadCloser { return io.NopCloser(strings.NewReader("")) }
func (stubProcess) Wait() error           { return nil }
func (stubProcess) Kill() error           { return nil }

// stubSpawners replaces the package spawners for the duration of a test and
// records which one ran with what arguments.
func stubSpawners(t *testing.T) (claudeCalls, codexCalls *[]string) {
	t.Helper()
	origClaude, origCodex := startClaude, startCodex
	t.Cleanup(func() {
		startClaude, startCodex = origClaude, origCodex
	})

	var claudePrompts, codexPrompts []string
	startClaude = func(prompt string, _ io.Reader) (Process, error) {
		claudePrompts = append(claudePrompts, prompt)
		return stubProcess{}, nil
	}
	startCodex = func(prompt string, _ io.Reader) (Process, error) {
		codexPrompts = append(codexPrompts, prompt)
		return stubProcess{}, nil
	}
	return &claudePrompts, &codexPrompts
}

func TestStartDispatch(t *testing.T) {
	cases := []struct {
		name       string
		agent      string
		wantClaude int
		wantCodex  int
	}{
		{"codex", config.AgentCodex, 0, 1},
		{"claude", config.AgentClaude, 1, 0},
		{"empty falls back to claude", "", 1, 0},
		{"unknown falls back to claude", "gemini", 1, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			claudeCalls, codexCalls := stubSpawners(t)

			proc, err := Start(tc.agent, "do the thing", nil)
			if err != nil {
				t.Fatalf("Start returned error: %v", err)
			}
			if proc == nil {
				t.Fatal("Start returned nil process")
			}
			if got := len(*claudeCalls); got != tc.wantClaude {
				t.Errorf("claude spawns = %d, want %d", got, tc.wantClaude)
			}
			if got := len(*codexCalls); got != tc.wantCodex {
				t.Errorf("codex spawns = %d, want %d", got, tc.wantCodex)
			}
		})
	}
}

func TestStartPassesPrompt(t *testing.T) {
	_, codexCalls := stubSpawners(t)

	if _, err := Start(config.AgentCodex, "render this", nil); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if len(*codexCalls) != 1 || (*codexCalls)[0] != "render this" {
		t.Errorf("codex prompts = %v, want [\"render this\"]", *codexCalls)
	}
}

func TestStartPropagatesError(t *testing.T) {
	origCodex := startCodex
	t.Cleanup(func() { startCodex = origCodex })

	wantErr := errors.New("boom")
	startCodex = func(string, io.Reader) (Process, error) { return nil, wantErr }

	if _, err := Start(config.AgentCodex, "x", nil); !errors.Is(err, wantErr) {
		t.Errorf("Start error = %v, want %v", err, wantErr)
	}
}
