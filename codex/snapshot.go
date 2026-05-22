package codex

import (
	"os"
	"strings"
)

const diffContextLines = 3

// FileSnapshotTracker captures file contents around live Codex file_change
// items and turns before/after content into structured patches. It is best
// effort: missing files, read failures, or unchanged content simply leave the
// item unchanged so transcript replay remains honest.
type FileSnapshotTracker struct {
	files map[string]map[string]fileSnapshot
}

type fileSnapshot struct {
	content string
	exists  bool
	ok      bool
}

// NewFileSnapshotTracker creates a tracker for live file changes.
func NewFileSnapshotTracker() *FileSnapshotTracker {
	return &FileSnapshotTracker{files: make(map[string]map[string]fileSnapshot)}
}

// Capture records the pre-change contents for a file_change item.
func (t *FileSnapshotTracker) Capture(item *Item) {
	if item == nil || item.ID == "" || item.Type != ItemFileChange || len(item.Changes) == 0 {
		return
	}
	if t.files == nil {
		t.files = make(map[string]map[string]fileSnapshot)
	}
	byPath := make(map[string]fileSnapshot, len(item.Changes))
	for _, change := range item.Changes {
		if change.Path == "" {
			continue
		}
		byPath[change.Path] = readFileSnapshot(change.Path)
	}
	if len(byPath) > 0 {
		t.files[item.ID] = byPath
	}
}

// AttachPatches adds structured patches to a completed file_change item when
// before/after snapshots are available.
func (t *FileSnapshotTracker) AttachPatches(item *Item) {
	if item == nil || item.ID == "" || item.Type != ItemFileChange {
		return
	}
	byPath, ok := t.files[item.ID]
	if !ok {
		return
	}
	defer delete(t.files, item.ID)

	for i := range item.Changes {
		change := &item.Changes[i]
		if len(change.StructuredPatch) > 0 || change.Path == "" {
			continue
		}
		before, ok := byPath[change.Path]
		if !ok || !before.ok {
			continue
		}
		after := readFileSnapshot(change.Path)
		if !after.ok {
			continue
		}
		if before.exists == after.exists && before.content == after.content {
			continue
		}
		change.StructuredPatch = structuredPatch(before, after)
	}
}

func readFileSnapshot(path string) fileSnapshot {
	b, err := os.ReadFile(path)
	if err == nil {
		return fileSnapshot{content: string(b), exists: true, ok: true}
	}
	if os.IsNotExist(err) {
		return fileSnapshot{exists: false, ok: true}
	}
	return fileSnapshot{}
}

func structuredPatch(before, after fileSnapshot) []PatchHunk {
	oldLines := snapshotLines(before)
	newLines := snapshotLines(after)
	ops := diffLineOps(oldLines, newLines)
	return diffHunks(ops)
}

func snapshotLines(s fileSnapshot) []string {
	if !s.exists || s.content == "" {
		return nil
	}
	content := strings.TrimSuffix(s.content, "\n")
	if content == "" {
		return nil
	}
	return strings.Split(content, "\n")
}

type lineOp struct {
	kind byte
	text string
	old  int
	new  int
}

func diffLineOps(oldLines, newLines []string) []lineOp {
	lcs := make([][]int, len(oldLines)+1)
	for i := range lcs {
		lcs[i] = make([]int, len(newLines)+1)
	}
	for i := len(oldLines) - 1; i >= 0; i-- {
		for j := len(newLines) - 1; j >= 0; j-- {
			if oldLines[i] == newLines[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}

	var ops []lineOp
	i, j := 0, 0
	for i < len(oldLines) || j < len(newLines) {
		switch {
		case i < len(oldLines) && j < len(newLines) && oldLines[i] == newLines[j]:
			ops = append(ops, lineOp{kind: ' ', text: oldLines[i], old: i + 1, new: j + 1})
			i++
			j++
		case j < len(newLines) && (i == len(oldLines) || lcs[i][j+1] > lcs[i+1][j]):
			ops = append(ops, lineOp{kind: '+', text: newLines[j], new: j + 1})
			j++
		default:
			ops = append(ops, lineOp{kind: '-', text: oldLines[i], old: i + 1})
			i++
		}
	}
	return ops
}

func diffHunks(ops []lineOp) []PatchHunk {
	changeIndexes := changedLineIndexes(ops)
	if len(changeIndexes) == 0 {
		return nil
	}

	var hunks []PatchHunk
	for i := 0; i < len(changeIndexes); {
		start := max(0, changeIndexes[i]-diffContextLines)
		end := min(len(ops), changeIndexes[i]+diffContextLines+1)
		i++
		for i < len(changeIndexes) && changeIndexes[i] <= end+diffContextLines {
			end = min(len(ops), changeIndexes[i]+diffContextLines+1)
			i++
		}
		hunks = append(hunks, patchHunk(ops[start:end]))
	}
	return hunks
}

func changedLineIndexes(ops []lineOp) []int {
	var indexes []int
	for i, op := range ops {
		if op.kind == '+' || op.kind == '-' {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func patchHunk(ops []lineOp) PatchHunk {
	h := PatchHunk{OldStart: 1, NewStart: 1}
	if len(ops) == 0 {
		return h
	}
	for _, op := range ops {
		if op.kind != '+' && h.OldLines == 0 {
			h.OldStart = op.old
		}
		if op.kind != '-' && h.NewLines == 0 {
			h.NewStart = op.new
		}
		switch op.kind {
		case '+':
			h.NewLines++
		case '-':
			h.OldLines++
		default:
			h.OldLines++
			h.NewLines++
		}
		h.Lines = append(h.Lines, string(op.kind)+op.text)
	}
	return h
}
