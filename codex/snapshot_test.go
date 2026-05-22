package codex

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileSnapshotTrackerAttachPatchesUpdate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.go")
	if err := os.WriteFile(path, []byte("package main\n\nfunc old() {}\n"), 0o600); err != nil {
		t.Fatalf("write old file: %v", err)
	}

	tracker := NewFileSnapshotTracker()
	started := &Item{
		ID:      "f1",
		Type:    ItemFileChange,
		Changes: []FileChange{{Path: path, Kind: "update"}},
	}
	tracker.Capture(started)

	if err := os.WriteFile(path, []byte("package main\n\nfunc new() {}\n"), 0o600); err != nil {
		t.Fatalf("write new file: %v", err)
	}

	completed := &Item{
		ID:      "f1",
		Type:    ItemFileChange,
		Changes: []FileChange{{Path: path, Kind: "update"}},
	}
	tracker.AttachPatches(completed)

	patch := completed.Changes[0].StructuredPatch
	if len(patch) == 0 {
		t.Fatal("StructuredPatch is empty")
	}
	got := strings.Join(patch[0].Lines, "\n")
	for _, want := range []string{" package main", "-func old() {}", "+func new() {}"} {
		if !strings.Contains(got, want) {
			t.Fatalf("patch lines = %q, want %q", got, want)
		}
	}
}

func TestFileSnapshotTrackerAttachPatchesAdd(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new.txt")

	tracker := NewFileSnapshotTracker()
	tracker.Capture(&Item{
		ID:      "f1",
		Type:    ItemFileChange,
		Changes: []FileChange{{Path: path, Kind: "add"}},
	})

	if err := os.WriteFile(path, []byte("one\ntwo\n"), 0o600); err != nil {
		t.Fatalf("write new file: %v", err)
	}

	completed := &Item{
		ID:      "f1",
		Type:    ItemFileChange,
		Changes: []FileChange{{Path: path, Kind: "add"}},
	}
	tracker.AttachPatches(completed)

	patch := completed.Changes[0].StructuredPatch
	if len(patch) != 1 || !strings.Contains(strings.Join(patch[0].Lines, "\n"), "+two") {
		t.Fatalf("StructuredPatch = %+v, want added file lines", patch)
	}
}
