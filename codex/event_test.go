package codex

import "testing"

func TestIsEventType(t *testing.T) {
	codexTypes := []string{
		TypeThreadStarted, TypeTurnStarted, TypeTurnCompleted, TypeTurnFailed,
		TypeItemStarted, TypeItemUpdated, TypeItemCompleted, TypeError,
	}
	for _, tt := range codexTypes {
		if !IsEventType(tt) {
			t.Errorf("IsEventType(%q) = false, want true", tt)
		}
	}

	// Claude Code event types must not be mistaken for codex events.
	claudeTypes := []string{"system", "assistant", "user", "stream_event", "result", ""}
	for _, tt := range claudeTypes {
		if IsEventType(tt) {
			t.Errorf("IsEventType(%q) = true, want false", tt)
		}
	}
}

func TestParseEvent_ThreadStarted(t *testing.T) {
	event, err := ParseEvent([]byte(`{"type":"thread.started","thread_id":"abc123"}`))
	if err != nil {
		t.Fatalf("ParseEvent error: %v", err)
	}
	if event.Type != TypeThreadStarted {
		t.Errorf("Type = %q, want %q", event.Type, TypeThreadStarted)
	}
	if event.ThreadID != "abc123" {
		t.Errorf("ThreadID = %q, want abc123", event.ThreadID)
	}
}

func TestParseEvent_TurnCompletedUsage(t *testing.T) {
	line := `{"type":"turn.completed","usage":{"input_tokens":100,"cached_input_tokens":40,"output_tokens":12,"reasoning_output_tokens":3}}`
	event, err := ParseEvent([]byte(line))
	if err != nil {
		t.Fatalf("ParseEvent error: %v", err)
	}
	if event.Usage == nil {
		t.Fatal("Usage is nil")
	}
	if event.Usage.InputTokens != 100 || event.Usage.CachedInputTokens != 40 ||
		event.Usage.OutputTokens != 12 || event.Usage.ReasoningOutputTokens != 3 {
		t.Errorf("unexpected usage: %+v", *event.Usage)
	}
}

func TestParseEvent_TurnFailed(t *testing.T) {
	event, err := ParseEvent([]byte(`{"type":"turn.failed","error":{"message":"boom"}}`))
	if err != nil {
		t.Fatalf("ParseEvent error: %v", err)
	}
	if event.Error == nil || event.Error.Message != "boom" {
		t.Errorf("unexpected error: %+v", event.Error)
	}
}

func TestParseEvent_CommandExecutionItem(t *testing.T) {
	line := `{"type":"item.completed","item":{"id":"item_1","type":"command_execution","command":"/usr/bin/zsh -lc ls","aggregated_output":"foo.txt\n","exit_code":0,"status":"completed"}}`
	event, err := ParseEvent([]byte(line))
	if err != nil {
		t.Fatalf("ParseEvent error: %v", err)
	}
	if event.Item == nil {
		t.Fatal("Item is nil")
	}
	if event.Item.Type != ItemCommandExecution {
		t.Errorf("item type = %q, want %q", event.Item.Type, ItemCommandExecution)
	}
	if event.Item.Command != "/usr/bin/zsh -lc ls" {
		t.Errorf("command = %q", event.Item.Command)
	}
	if event.Item.ExitCode == nil || *event.Item.ExitCode != 0 {
		t.Errorf("exit code = %v, want 0", event.Item.ExitCode)
	}
}

func TestParseEvent_FileChangeItem(t *testing.T) {
	line := `{"type":"item.completed","item":{"id":"i","type":"file_change","changes":[{"path":"/a.txt","kind":"add"},{"path":"/b.txt","kind":"update"}],"status":"completed"}}`
	event, err := ParseEvent([]byte(line))
	if err != nil {
		t.Fatalf("ParseEvent error: %v", err)
	}
	if got := len(event.Item.Changes); got != 2 {
		t.Fatalf("changes len = %d, want 2", got)
	}
	if event.Item.Changes[0].Path != "/a.txt" || event.Item.Changes[0].Kind != "add" {
		t.Errorf("unexpected change[0]: %+v", event.Item.Changes[0])
	}
}

func TestParseEvent_TodoListItem(t *testing.T) {
	line := `{"type":"item.started","item":{"id":"i","type":"todo_list","items":[{"text":"step one","completed":true},{"text":"step two","completed":false}]}}`
	event, err := ParseEvent([]byte(line))
	if err != nil {
		t.Fatalf("ParseEvent error: %v", err)
	}
	if got := len(event.Item.Items); got != 2 {
		t.Fatalf("items len = %d, want 2", got)
	}
	if !event.Item.Items[0].Completed || event.Item.Items[1].Completed {
		t.Errorf("unexpected completion flags: %+v", event.Item.Items)
	}
}

func TestParseEvent_Invalid(t *testing.T) {
	if _, err := ParseEvent([]byte(`{not json`)); err == nil {
		t.Error("expected error for invalid JSON")
	}
}
