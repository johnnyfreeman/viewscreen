package claude

import (
	"io"
	"os"
	"os/exec"
)

// Process wraps an exec.Cmd for a running claude subprocess.
type Process struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
}

// Start spawns claude with the given prompt in stream-json mode.
// If stdinReader is non-nil, the prompt is piped via claude's stdin
// (used for the -p flag with piped input). Otherwise the prompt is
// passed as a positional argument.
func Start(prompt string, stdinReader io.Reader) (*Process, error) {
	args := []string{"-p", "--output-format", "stream-json", "--verbose"}
	if stdinReader == nil {
		args = append(args, prompt)
	}

	cmd := exec.Command("claude", args...)
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
	return p.cmd.Wait()
}

// Kill terminates the subprocess.
func (p *Process) Kill() error {
	if p.cmd.Process != nil {
		return p.cmd.Process.Kill()
	}
	return nil
}
