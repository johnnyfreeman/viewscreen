package codex

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/testutil"
)

// newTestRenderer builds a Renderer that produces deterministic, color-free
// output suitable for substring assertions.
func newTestRenderer(t *testing.T, showUsage, verbose bool) *Renderer {
	t.Helper()
	level := 0
	if verbose {
		level = 1
	}
	return newTestRendererWithLevel(t, showUsage, level)
}

func newTestRendererWithLevel(t *testing.T, showUsage bool, level int) *Renderer {
	t.Helper()
	style.Init(true)
	t.Cleanup(func() { style.Init(false) })

	cfg := testutil.MockConfigProvider{NoColorVal: true, ShowUsageVal: showUsage, VerboseLevelVal: level}
	r := NewRenderer(
		WithConfigProvider(cfg),
		WithMarkdownRenderer(render.NewMarkdownRenderer(true, 80)),
	)
	r.SetWidth(80)
	return r
}

func itemEvent(phase string, item Item) Event {
	return Event{Type: phase, Item: &item}
}

func intPtr(n int) *int { return &n }

func TestRender_ThreadStarted(t *testing.T) {
	r := newTestRenderer(t, true, false)
	out := r.Render(Event{Type: TypeThreadStarted, ThreadID: "abc123"})
	if !strings.Contains(out, "Codex Session") {
		t.Errorf("missing header in %q", out)
	}
	if !strings.Contains(out, "abc123") {
		t.Errorf("missing thread id in %q", out)
	}
}

func TestRender_TurnStartedIsSilent(t *testing.T) {
	r := newTestRenderer(t, true, false)
	if out := r.Render(Event{Type: TypeTurnStarted}); out != "" {
		t.Errorf("turn.started should render nothing, got %q", out)
	}
}

func TestRender_AgentMessage(t *testing.T) {
	r := newTestRenderer(t, true, false)
	out := r.Render(itemEvent(TypeItemCompleted, Item{ID: "m1", Type: ItemAgentMessage, Text: "hello there"}))
	if !strings.Contains(out, "hello there") {
		t.Errorf("missing message text in %q", out)
	}
}

func TestRender_AgentMessageDedup(t *testing.T) {
	r := newTestRenderer(t, true, false)
	first := r.Render(itemEvent(TypeItemCompleted, Item{ID: "m1", Type: ItemAgentMessage, Text: "hello"}))
	second := r.Render(itemEvent(TypeItemCompleted, Item{ID: "m1", Type: ItemAgentMessage, Text: "hello"}))
	if first == "" {
		t.Fatal("first render should produce output")
	}
	if second != "" {
		t.Errorf("duplicate message should render nothing, got %q", second)
	}
}

func TestRender_Reasoning(t *testing.T) {
	r := newTestRenderer(t, true, false)
	out := r.Render(itemEvent(TypeItemCompleted, Item{ID: "r1", Type: ItemReasoning, Text: "weighing options"}))
	if !strings.Contains(out, "Thinking") {
		t.Errorf("missing Thinking header in %q", out)
	}
	if !strings.Contains(out, "weighing options") {
		t.Errorf("missing reasoning text in %q", out)
	}
}

func TestRender_CommandHeaderThenOutput(t *testing.T) {
	r := newTestRenderer(t, true, false)
	item := Item{ID: "c1", Type: ItemCommandExecution, Command: "/usr/bin/zsh -lc ls", AggregatedOutput: "foo.txt\n", ExitCode: intPtr(0), Status: "completed"}

	started := r.Render(itemEvent(TypeItemStarted, item))
	if !strings.Contains(started, "Shell") || !strings.Contains(started, "ls") {
		t.Errorf("started should show command header, got %q", started)
	}
	if strings.Contains(started, "foo.txt") {
		t.Errorf("started should not include output, got %q", started)
	}

	completed := r.Render(itemEvent(TypeItemCompleted, item))
	if strings.Contains(completed, "Shell") {
		t.Errorf("completed should not repeat the header, got %q", completed)
	}
	if !strings.Contains(completed, "1 lines") {
		t.Errorf("completed should summarize output, got %q", completed)
	}
	if strings.Contains(completed, "foo.txt") {
		t.Errorf("completed should not include default output, got %q", completed)
	}
}

func TestRender_CommandCompletedOnly(t *testing.T) {
	r := newTestRenderer(t, true, false)
	item := Item{ID: "c1", Type: ItemCommandExecution, Command: "/bin/bash -c 'echo hi'", AggregatedOutput: "hi\n", ExitCode: intPtr(0), Status: "completed"}
	out := r.Render(itemEvent(TypeItemCompleted, item))
	if !strings.Contains(out, "Shell") || !strings.Contains(out, "echo hi") {
		t.Errorf("expected header with unwrapped command, got %q", out)
	}
	if !strings.Contains(out, "1 lines") {
		t.Errorf("expected output summary, got %q", out)
	}
}

func TestRender_CommandOutputVerboseStillSummarizes(t *testing.T) {
	r := newTestRendererWithLevel(t, true, 1)
	item := Item{ID: "c1", Type: ItemCommandExecution, Command: "/bin/sh -lc cat", AggregatedOutput: "one\ntwo\n", ExitCode: intPtr(0), Status: "completed"}
	out := r.Render(itemEvent(TypeItemCompleted, item))
	if !strings.Contains(out, "2 lines") {
		t.Errorf("expected line summary at -v, got %q", out)
	}
	if strings.Contains(out, "one") || strings.Contains(out, "two") {
		t.Errorf("-v should not expand command output, got %q", out)
	}
}

func TestRender_CommandNonZeroExit(t *testing.T) {
	r := newTestRenderer(t, true, false)
	item := Item{ID: "c1", Type: ItemCommandExecution, Command: "/bin/sh -lc false", AggregatedOutput: "", ExitCode: intPtr(2), Status: "completed"}
	out := r.Render(itemEvent(TypeItemCompleted, item))
	if !strings.Contains(out, "exited with code 2") {
		t.Errorf("expected exit code note, got %q", out)
	}
}

func TestRender_CommandNoOutput(t *testing.T) {
	r := newTestRenderer(t, true, false)
	item := Item{ID: "c1", Type: ItemCommandExecution, Command: "/bin/sh -lc true", AggregatedOutput: "", ExitCode: intPtr(0), Status: "completed"}
	out := r.Render(itemEvent(TypeItemCompleted, item))
	if !strings.Contains(out, "(no output)") {
		t.Errorf("expected (no output), got %q", out)
	}
}

func TestRender_CommandOutputTruncation(t *testing.T) {
	var lines []string
	for i := range commandOutputLinesMaxVerbose + 3 {
		lines = append(lines, "line "+string(rune('a'+i)))
	}
	output := strings.Join(lines, "\n") + "\n"
	item := Item{ID: "c1", Type: ItemCommandExecution, Command: "/bin/sh -lc cat", AggregatedOutput: output, ExitCode: intPtr(0), Status: "completed"}

	t.Run("default summarizes", func(t *testing.T) {
		r := newTestRenderer(t, true, false)
		out := r.Render(itemEvent(TypeItemCompleted, item))
		if !strings.Contains(out, "13 lines") {
			t.Errorf("expected line summary, got %q", out)
		}
		if strings.Contains(out, "line a") {
			t.Errorf("default should not expand output, got %q", out)
		}
	})

	t.Run("-vv shows five lines", func(t *testing.T) {
		r := newTestRendererWithLevel(t, true, 2)
		out := r.Render(itemEvent(TypeItemCompleted, item))
		if !strings.Contains(out, "line a") || !strings.Contains(out, "line e") {
			t.Errorf("-vv should include first five lines, got %q", out)
		}
		if strings.Contains(out, "line f") {
			t.Errorf("-vv should hide the sixth line, got %q", out)
		}
		if !strings.Contains(out, "8 more lines") {
			t.Errorf("expected truncation note, got %q", out)
		}
	})

	t.Run("-vvv shows ten lines", func(t *testing.T) {
		r := newTestRendererWithLevel(t, true, 3)
		out := r.Render(itemEvent(TypeItemCompleted, item))
		if !strings.Contains(out, "line a") || !strings.Contains(out, "line j") {
			t.Errorf("-vvv should include first ten lines, got %q", out)
		}
		if strings.Contains(out, "line k") {
			t.Errorf("-vvv should hide the eleventh line, got %q", out)
		}
		if !strings.Contains(out, "3 more lines") {
			t.Errorf("expected truncation note, got %q", out)
		}
	})
}

func TestRender_CommandOutputLongLineTruncatesToWidth(t *testing.T) {
	r := newTestRendererWithLevel(t, true, 2)
	r.SetWidth(40)

	long := strings.Repeat("x", 120)
	item := Item{ID: "c1", Type: ItemCommandExecution, Command: "/bin/sh -lc cat", AggregatedOutput: long + "\n", ExitCode: intPtr(0), Status: "completed"}
	out := r.Render(itemEvent(TypeItemCompleted, item))

	if !strings.Contains(out, "more chars") {
		t.Fatalf("expected omitted-character marker, got %q", out)
	}
	if strings.Contains(out, long) {
		t.Fatalf("expected long physical line to be truncated, got %q", out)
	}
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if strings.Contains(line, "more chars") && ansi.StringWidth(line) > 40 {
			t.Errorf("truncated output line width = %d, want <= 40; line %q", ansi.StringWidth(line), line)
		}
	}
}

func TestRender_FileChangeSingle(t *testing.T) {
	r := newTestRenderer(t, true, false)
	item := Item{ID: "f1", Type: ItemFileChange, Changes: []FileChange{{Path: "/tmp/bar.txt", Kind: "add"}}, Status: "completed"}
	out := r.Render(itemEvent(TypeItemCompleted, item))
	if !strings.Contains(out, "Edit") || !strings.Contains(out, "/tmp/bar.txt") {
		t.Errorf("expected single-file header, got %q", out)
	}
	if !strings.Contains(out, "add") {
		t.Errorf("expected single-file change detail, got %q", out)
	}
}

func TestRender_FileChangeMultiple(t *testing.T) {
	r := newTestRenderer(t, true, false)
	item := Item{ID: "f1", Type: ItemFileChange, Changes: []FileChange{
		{Path: "/a.txt", Kind: "add"},
		{Path: "/b.txt", Kind: "update"},
	}, Status: "completed"}
	out := r.Render(itemEvent(TypeItemCompleted, item))
	if !strings.Contains(out, "2 files") {
		t.Errorf("expected file count summary, got %q", out)
	}
	if !strings.Contains(out, "/a.txt") || !strings.Contains(out, "/b.txt") {
		t.Errorf("expected per-file lines, got %q", out)
	}
}

func TestRender_FileChangeDedup(t *testing.T) {
	r := newTestRenderer(t, true, false)
	item := Item{ID: "f1", Type: ItemFileChange, Changes: []FileChange{{Path: "/a.txt", Kind: "add"}}}
	if out := r.Render(itemEvent(TypeItemStarted, item)); out == "" {
		t.Fatal("started should render the file change")
	}
	out := r.Render(itemEvent(TypeItemCompleted, item))
	if strings.Contains(out, "Edit") {
		t.Errorf("completed duplicate should not repeat header, got %q", out)
	}
	if !strings.Contains(out, "add") || !strings.Contains(out, "/a.txt") {
		t.Errorf("completed duplicate should render change details, got %q", out)
	}
}

func TestRender_FileChangeVeryVerboseDoesNotDumpSparseRawPayload(t *testing.T) {
	r := newTestRendererWithLevel(t, true, 3)
	event, err := ParseEvent([]byte(`{"type":"item.completed","item":{"id":"f1","type":"file_change","changes":[{"path":"/a.txt","kind":"update"}],"status":"completed"}}`))
	if err != nil {
		t.Fatalf("ParseEvent error: %v", err)
	}
	out := r.Render(event)
	if strings.Contains(out, `"type":"file_change"`) || strings.Contains(out, `"status":"completed"`) {
		t.Errorf("-vvv should not dump sparse raw payload, got %q", out)
	}
}

func TestRender_FileChangeStructuredPatch(t *testing.T) {
	r := newTestRendererWithLevel(t, true, 3)
	item := Item{
		ID:   "f1",
		Type: ItemFileChange,
		Changes: []FileChange{{
			Path: "/tmp/app.go",
			Kind: "update",
			StructuredPatch: []PatchHunk{{
				OldStart: 10,
				OldLines: 2,
				NewStart: 10,
				NewLines: 3,
				Lines:    []string{" package main", "-old()", "+new()", "+more()"},
			}},
		}},
		Status: "completed",
	}
	out := r.Render(itemEvent(TypeItemCompleted, item))
	for _, want := range []string{"10", "│", "old()", "new()", "more()"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in structured patch output, got %q", want, out)
		}
	}
}

func TestRender_TodoList(t *testing.T) {
	r := newTestRenderer(t, true, false)
	item := Item{ID: "t1", Type: ItemTodoList, Items: []TodoItem{
		{Text: "done step", Completed: true},
		{Text: "pending step", Completed: false},
	}}
	out := r.Render(itemEvent(TypeItemStarted, item))
	if !strings.Contains(out, "Update Todos") {
		t.Errorf("expected todo header, got %q", out)
	}
	if !strings.Contains(out, "✓ done step") {
		t.Errorf("expected completed marker, got %q", out)
	}
	if !strings.Contains(out, "○ pending step") {
		t.Errorf("expected pending marker, got %q", out)
	}
}

func TestRender_WebSearch(t *testing.T) {
	r := newTestRenderer(t, true, false)
	item := Item{ID: "w1", Type: ItemWebSearch, Query: "golang testing"}
	out := r.Render(itemEvent(TypeItemCompleted, item))
	if !strings.Contains(out, "Web Search") || !strings.Contains(out, "golang testing") {
		t.Errorf("expected web search header, got %q", out)
	}
}

func TestRender_MCPToolCall(t *testing.T) {
	r := newTestRenderer(t, true, false)
	item := Item{ID: "x1", Type: ItemMCPToolCall, Server: "github", Tool: "create_pr", Status: "completed"}
	out := r.Render(itemEvent(TypeItemStarted, item))
	if !strings.Contains(out, "github.create_pr") {
		t.Errorf("expected server.tool label, got %q", out)
	}
}

func TestRender_UnknownItem(t *testing.T) {
	r := newTestRenderer(t, true, false)
	item := Item{
		ID:      "x1",
		Type:    "image_generation",
		Message: "created image asset",
		Status:  "completed",
	}
	out := r.Render(itemEvent(TypeItemCompleted, item))
	if !strings.Contains(out, "Image Generation") {
		t.Errorf("expected titleized fallback header, got %q", out)
	}
	if !strings.Contains(out, "created image asset") {
		t.Errorf("expected fallback message, got %q", out)
	}
}

func TestRender_UnknownItemVerboseShowsRawPayload(t *testing.T) {
	r := newTestRenderer(t, true, true)
	item := Item{
		ID:     "x1",
		Type:   "custom_tool",
		Status: "completed",
		Raw:    []byte(`{"id":"x1","type":"custom_tool","status":"completed","detail":{"count":2}}`),
	}
	out := r.Render(itemEvent(TypeItemCompleted, item))
	if !strings.Contains(out, "Custom Tool") {
		t.Errorf("expected fallback header, got %q", out)
	}
	if !strings.Contains(out, `"count":2`) {
		t.Errorf("expected compact raw payload in verbose output, got %q", out)
	}
}

func TestRender_TurnCompletedUsage(t *testing.T) {
	t.Run("with usage shown", func(t *testing.T) {
		r := newTestRenderer(t, true, false)
		out := r.Render(Event{Type: TypeTurnCompleted, Usage: &Usage{InputTokens: 100, OutputTokens: 20, CachedInputTokens: 40, ReasoningOutputTokens: 5}})
		if !strings.Contains(out, "Turn Complete") {
			t.Errorf("expected header, got %q", out)
		}
		if !strings.Contains(out, "in=100") || !strings.Contains(out, "out=20") {
			t.Errorf("expected token counts, got %q", out)
		}
	})

	t.Run("usage hidden", func(t *testing.T) {
		r := newTestRenderer(t, false, false)
		out := r.Render(Event{Type: TypeTurnCompleted, Usage: &Usage{InputTokens: 100}})
		if strings.Contains(out, "Tokens:") {
			t.Errorf("usage should be hidden, got %q", out)
		}
	})
}

func TestRender_TurnFailed(t *testing.T) {
	r := newTestRenderer(t, true, false)
	out := r.Render(Event{Type: TypeTurnFailed, Error: &ThreadError{Message: "rate limited"}})
	if !strings.Contains(out, "Turn Failed") || !strings.Contains(out, "rate limited") {
		t.Errorf("expected failure output, got %q", out)
	}
}

func TestRender_TopLevelError(t *testing.T) {
	r := newTestRenderer(t, true, false)
	out := r.Render(Event{Type: TypeError, Message: "stream broke"})
	if !strings.Contains(out, "Error") || !strings.Contains(out, "stream broke") {
		t.Errorf("expected error output, got %q", out)
	}
}

func TestRender_NilItem(t *testing.T) {
	r := newTestRenderer(t, true, false)
	if out := r.Render(Event{Type: TypeItemCompleted}); out != "" {
		t.Errorf("nil item should render nothing, got %q", out)
	}
}

func TestShellCommand(t *testing.T) {
	cases := map[string]string{
		"/usr/bin/zsh -lc ls":            "ls",
		"/usr/bin/zsh -lc 'cat foo.txt'": "cat foo.txt",
		`/bin/bash -c "echo hi"`:         "echo hi",
		"plain command":                  "plain command",
	}
	for in, want := range cases {
		if got := ShellCommand(in); got != want {
			t.Errorf("ShellCommand(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("short", 80); got != "short" {
		t.Errorf("truncate kept-short = %q", got)
	}
	long := strings.Repeat("a", 100)
	got := truncate(long, 80)
	if len(got) != 80 || !strings.HasSuffix(got, "...") {
		t.Errorf("truncate(long) = %q (len %d)", got, len(got))
	}
	if got := truncate("a\nb", 80); got != "a b" {
		t.Errorf("truncate should flatten newlines, got %q", got)
	}
}
