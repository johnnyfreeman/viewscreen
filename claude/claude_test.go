package claude

import "testing"

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
