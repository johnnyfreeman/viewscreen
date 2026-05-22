package codex

import (
	"io"
	"os"
	"os/exec"
	"sync"
)

// Process wraps an exec.Cmd for a running codex subprocess. It mirrors the
// claude package's Process so the TUI and legacy paths can drive either agent
// through the same Stdout/Wait/Kill interface.
type Process struct {
	cmd      *exec.Cmd
	stdout   io.ReadCloser
	waitOnce sync.Once
	waitErr  error
}

// Start spawns codex with the given prompt in JSONL streaming mode
// ("codex exec --json"). If stdinReader is non-nil, the prompt is read from
// codex's stdin (the "-" prompt argument) — used for the -p flag with piped
// input. Otherwise the prompt is passed as a positional argument.
func Start(prompt string, stdinReader io.Reader) (*Process, error) {
	args := []string{"exec", "--json"}
	if stdinReader != nil {
		args = append(args, "-")
	} else {
		args = append(args, prompt)
	}

	cmd := exec.Command("codex", args...)
	if stdinReader != nil {
		cmd.Stdin = stdinReader
	}
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &Process{cmd: cmd, stdout: stdout}, nil
}

// Stdout returns the stdout pipe of the subprocess.
func (p *Process) Stdout() io.ReadCloser {
	return p.stdout
}

// Wait waits for the subprocess to exit.
func (p *Process) Wait() error {
	if p == nil || p.cmd == nil {
		return nil
	}
	p.waitOnce.Do(func() {
		p.waitErr = p.cmd.Wait()
	})
	return p.waitErr
}

// Kill terminates the subprocess.
func (p *Process) Kill() error {
	if p == nil || p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	return p.cmd.Process.Kill()
}
