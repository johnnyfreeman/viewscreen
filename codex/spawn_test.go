package codex

import (
	"os/exec"
	"testing"
)

func TestProcessLifecycleMethodsAreNilSafe(t *testing.T) {
	var nilProcess *Process
	if err := nilProcess.Kill(); err != nil {
		t.Errorf("nil Kill returned error: %v", err)
	}
	if err := nilProcess.Wait(); err != nil {
		t.Errorf("nil Wait returned error: %v", err)
	}

	emptyProcess := &Process{}
	if err := emptyProcess.Kill(); err != nil {
		t.Errorf("empty Kill returned error: %v", err)
	}
	if err := emptyProcess.Wait(); err != nil {
		t.Errorf("empty Wait returned error: %v", err)
	}
}

func TestProcessWaitIsIdempotent(t *testing.T) {
	cmd := exec.Command("sh", "-c", "exit 0")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	proc := &Process{cmd: cmd}
	if err := proc.Wait(); err != nil {
		t.Fatalf("first Wait returned error: %v", err)
	}
	if err := proc.Wait(); err != nil {
		t.Fatalf("second Wait returned error: %v", err)
	}
}
