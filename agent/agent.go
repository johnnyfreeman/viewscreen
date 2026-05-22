// Package agent selects and spawns the underlying AI coding CLI (Claude Code
// or the Codex CLI) for viewscreen's prompt mode. Both spawners expose the
// same Stdout/Wait/Kill surface, so callers drive whichever agent was chosen
// through a single Process interface.
package agent

import (
	"io"

	"github.com/johnnyfreeman/viewscreen/claude"
	"github.com/johnnyfreeman/viewscreen/codex"
	"github.com/johnnyfreeman/viewscreen/config"
)

// Process is the common control surface for a spawned agent subprocess.
type Process interface {
	Stdout() io.ReadCloser
	Wait() error
	Kill() error
}

// Spawner starts an agent subprocess. When stdinReader is non-nil the prompt
// is piped via the agent's stdin (used for -p with piped input); otherwise it
// is passed as a positional argument.
type Spawner func(prompt string, stdinReader io.Reader) (Process, error)

// startClaude and startCodex are package vars so tests can stub the spawners
// without launching real subprocesses.
var (
	startClaude Spawner = func(prompt string, stdinReader io.Reader) (Process, error) {
		return claude.Start(prompt, stdinReader)
	}
	startCodex Spawner = func(prompt string, stdinReader io.Reader) (Process, error) {
		return codex.Start(prompt, stdinReader)
	}
)

// Start spawns the named agent with the given prompt. Any name other than
// config.AgentCodex falls back to Claude Code.
func Start(name, prompt string, stdinReader io.Reader) (Process, error) {
	switch name {
	case config.AgentCodex:
		return startCodex(prompt, stdinReader)
	default:
		return startClaude(prompt, stdinReader)
	}
}
