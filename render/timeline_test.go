package render

import (
	"strings"
	"testing"

	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/timeline"
)

func TestTimelineRendererRenderEntries(t *testing.T) {
	style.Init(true)
	r := NewTimelineRenderer()

	got := r.RenderEntries([]timeline.Entry{
		{Kind: "assistant", Body: "hello\n"},
		{Kind: "tool", Lines: []string{"one", "two"}},
	})

	if got != "hello\none\ntwo\n" {
		t.Fatalf("RenderEntries() = %q", got)
	}
}

func TestTimelineRendererRenderActivity(t *testing.T) {
	style.Init(true)
	r := NewTimelineRenderer()

	got := r.RenderActivity(timeline.Activity{Name: "Shell", Input: "go test ./..."}, "...")

	if !strings.Contains(got, "... Shell") {
		t.Fatalf("activity header missing name: %q", got)
	}
	if !strings.Contains(got, "go test ./...") {
		t.Fatalf("activity header missing input: %q", got)
	}
}
