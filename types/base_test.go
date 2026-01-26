package types

import (
	"encoding/json"
	"testing"
)

func TestBaseEvent_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantErr  bool
		validate func(t *testing.T, e BaseEvent)
	}{
		{
			name: "all fields",
			json: `{
				"type": "assistant",
				"session_id": "sess_123",
				"uuid": "uuid_456",
				"parent_tool_use_id": "tool_789"
			}`,
			wantErr: false,
			validate: func(t *testing.T, e BaseEvent) {
				if e.Type != "assistant" {
					t.Errorf("Type: got %q, want %q", e.Type, "assistant")
				}
				if e.SessionID != "sess_123" {
					t.Errorf("SessionID: got %q, want %q", e.SessionID, "sess_123")
				}
				if e.UUID != "uuid_456" {
					t.Errorf("UUID: got %q, want %q", e.UUID, "uuid_456")
				}
				if e.ParentToolUseID == nil {
					t.Fatal("ParentToolUseID: expected non-nil")
				}
				if *e.ParentToolUseID != "tool_789" {
					t.Errorf("ParentToolUseID: got %q, want %q", *e.ParentToolUseID, "tool_789")
				}
			},
		},
		{
			name: "null parent_tool_use_id",
			json: `{
				"type": "system",
				"session_id": "sess_abc",
				"uuid": "uuid_def",
				"parent_tool_use_id": null
			}`,
			wantErr: false,
			validate: func(t *testing.T, e BaseEvent) {
				if e.ParentToolUseID != nil {
					t.Errorf("ParentToolUseID: expected nil, got %q", *e.ParentToolUseID)
				}
			},
		},
		{
			name: "missing parent_tool_use_id",
			json: `{
				"type": "user",
				"session_id": "sess_xyz"
			}`,
			wantErr: false,
			validate: func(t *testing.T, e BaseEvent) {
				if e.ParentToolUseID != nil {
					t.Errorf("ParentToolUseID: expected nil for missing field")
				}
				if e.UUID != "" {
					t.Errorf("UUID: expected empty for missing field, got %q", e.UUID)
				}
			},
		},
		{
			name: "empty event",
			json: `{}`,
			wantErr: false,
			validate: func(t *testing.T, e BaseEvent) {
				if e.Type != "" {
					t.Errorf("Type: expected empty, got %q", e.Type)
				}
			},
		},
		{
			name:    "invalid json",
			json:    `{"type": }`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event BaseEvent
			err := json.Unmarshal([]byte(tt.json), &event)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, event)
			}
		})
	}
}

func TestBaseEvent_JSONMarshal(t *testing.T) {
	toolID := "tool_123"
	event := BaseEvent{
		Type:            "assistant",
		SessionID:       "sess_456",
		UUID:            "uuid_789",
		ParentToolUseID: &toolID,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if result["type"] != "assistant" {
		t.Errorf("type: got %v, want %q", result["type"], "assistant")
	}
	if result["session_id"] != "sess_456" {
		t.Errorf("session_id: got %v, want %q", result["session_id"], "sess_456")
	}
	if result["uuid"] != "uuid_789" {
		t.Errorf("uuid: got %v, want %q", result["uuid"], "uuid_789")
	}
	if result["parent_tool_use_id"] != "tool_123" {
		t.Errorf("parent_tool_use_id: got %v, want %q", result["parent_tool_use_id"], "tool_123")
	}
}

func TestContentBlock_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantErr  bool
		validate func(t *testing.T, b ContentBlock)
	}{
		{
			name: "text block",
			json: `{
				"type": "text",
				"text": "Hello, world!"
			}`,
			wantErr: false,
			validate: func(t *testing.T, b ContentBlock) {
				if b.Type != "text" {
					t.Errorf("Type: got %q, want %q", b.Type, "text")
				}
				if b.Text != "Hello, world!" {
					t.Errorf("Text: got %q, want %q", b.Text, "Hello, world!")
				}
				if b.ID != "" {
					t.Errorf("ID: expected empty for text block, got %q", b.ID)
				}
				if b.Name != "" {
					t.Errorf("Name: expected empty for text block, got %q", b.Name)
				}
			},
		},
		{
			name: "tool_use block with input object",
			json: `{
				"type": "tool_use",
				"id": "tool_abc",
				"name": "Read",
				"input": {"file_path": "/test.go", "offset": 0}
			}`,
			wantErr: false,
			validate: func(t *testing.T, b ContentBlock) {
				if b.Type != "tool_use" {
					t.Errorf("Type: got %q, want %q", b.Type, "tool_use")
				}
				if b.ID != "tool_abc" {
					t.Errorf("ID: got %q, want %q", b.ID, "tool_abc")
				}
				if b.Name != "Read" {
					t.Errorf("Name: got %q, want %q", b.Name, "Read")
				}
				if b.Input == nil {
					t.Fatal("Input: expected non-nil")
				}

				var input map[string]any
				if err := json.Unmarshal(b.Input, &input); err != nil {
					t.Fatalf("failed to unmarshal input: %v", err)
				}
				if input["file_path"] != "/test.go" {
					t.Errorf("input.file_path: got %v, want %q", input["file_path"], "/test.go")
				}
			},
		},
		{
			name: "tool_use block with empty input",
			json: `{
				"type": "tool_use",
				"id": "tool_xyz",
				"name": "Bash",
				"input": {}
			}`,
			wantErr: false,
			validate: func(t *testing.T, b ContentBlock) {
				if b.Input == nil {
					t.Fatal("Input: expected non-nil even for empty object")
				}

				var input map[string]any
				if err := json.Unmarshal(b.Input, &input); err != nil {
					t.Fatalf("failed to unmarshal input: %v", err)
				}
				if len(input) != 0 {
					t.Errorf("input: expected empty map, got %v", input)
				}
			},
		},
		{
			name: "block without optional fields",
			json: `{
				"type": "text"
			}`,
			wantErr: false,
			validate: func(t *testing.T, b ContentBlock) {
				if b.Text != "" {
					t.Errorf("Text: expected empty for missing field, got %q", b.Text)
				}
				if b.Input != nil {
					t.Errorf("Input: expected nil for missing field")
				}
			},
		},
		{
			name: "text block with unicode",
			json: `{
				"type": "text",
				"text": "Hello 世界! 🚀"
			}`,
			wantErr: false,
			validate: func(t *testing.T, b ContentBlock) {
				if b.Text != "Hello 世界! 🚀" {
					t.Errorf("Text: got %q, want %q", b.Text, "Hello 世界! 🚀")
				}
			},
		},
		{
			name: "text block with newlines and special chars",
			json: `{
				"type": "text",
				"text": "Line1\nLine2\tTabbed\r\nWindows"
			}`,
			wantErr: false,
			validate: func(t *testing.T, b ContentBlock) {
				expected := "Line1\nLine2\tTabbed\r\nWindows"
				if b.Text != expected {
					t.Errorf("Text: got %q, want %q", b.Text, expected)
				}
			},
		},
		{
			name:    "invalid json",
			json:    `{"type": "text", "text": }`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var block ContentBlock
			err := json.Unmarshal([]byte(tt.json), &block)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, block)
			}
		})
	}
}

func TestContentBlock_JSONMarshal(t *testing.T) {
	tests := []struct {
		name     string
		block    ContentBlock
		validate func(t *testing.T, data []byte)
	}{
		{
			name: "text block",
			block: ContentBlock{
				Type: "text",
				Text: "Hello, world!",
			},
			validate: func(t *testing.T, data []byte) {
				var result map[string]any
				if err := json.Unmarshal(data, &result); err != nil {
					t.Fatalf("failed to unmarshal result: %v", err)
				}
				if result["type"] != "text" {
					t.Errorf("type: got %v, want %q", result["type"], "text")
				}
				if result["text"] != "Hello, world!" {
					t.Errorf("text: got %v, want %q", result["text"], "Hello, world!")
				}
				// omitempty fields should not be present
				if _, ok := result["id"]; ok {
					t.Error("id: should be omitted when empty")
				}
				if _, ok := result["name"]; ok {
					t.Error("name: should be omitted when empty")
				}
			},
		},
		{
			name: "tool_use block",
			block: ContentBlock{
				Type:  "tool_use",
				ID:    "tool_123",
				Name:  "Write",
				Input: json.RawMessage(`{"file_path":"/out.txt"}`),
			},
			validate: func(t *testing.T, data []byte) {
				var result map[string]any
				if err := json.Unmarshal(data, &result); err != nil {
					t.Fatalf("failed to unmarshal result: %v", err)
				}
				if result["type"] != "tool_use" {
					t.Errorf("type: got %v, want %q", result["type"], "tool_use")
				}
				if result["id"] != "tool_123" {
					t.Errorf("id: got %v, want %q", result["id"], "tool_123")
				}
				if result["name"] != "Write" {
					t.Errorf("name: got %v, want %q", result["name"], "Write")
				}
				// text should be omitted when empty
				if _, ok := result["text"]; ok {
					t.Error("text: should be omitted when empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.block)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}
			tt.validate(t, data)
		})
	}
}

func TestUsage_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantErr  bool
		validate func(t *testing.T, u Usage)
	}{
		{
			name: "all fields",
			json: `{
				"input_tokens": 1000,
				"cache_creation_input_tokens": 500,
				"cache_read_input_tokens": 200,
				"output_tokens": 300,
				"service_tier": "standard"
			}`,
			wantErr: false,
			validate: func(t *testing.T, u Usage) {
				if u.InputTokens != 1000 {
					t.Errorf("InputTokens: got %d, want %d", u.InputTokens, 1000)
				}
				if u.CacheCreationInputTokens != 500 {
					t.Errorf("CacheCreationInputTokens: got %d, want %d", u.CacheCreationInputTokens, 500)
				}
				if u.CacheReadInputTokens != 200 {
					t.Errorf("CacheReadInputTokens: got %d, want %d", u.CacheReadInputTokens, 200)
				}
				if u.OutputTokens != 300 {
					t.Errorf("OutputTokens: got %d, want %d", u.OutputTokens, 300)
				}
				if u.ServiceTier != "standard" {
					t.Errorf("ServiceTier: got %q, want %q", u.ServiceTier, "standard")
				}
			},
		},
		{
			name: "zero values",
			json: `{
				"input_tokens": 0,
				"output_tokens": 0
			}`,
			wantErr: false,
			validate: func(t *testing.T, u Usage) {
				if u.InputTokens != 0 {
					t.Errorf("InputTokens: got %d, want %d", u.InputTokens, 0)
				}
				if u.OutputTokens != 0 {
					t.Errorf("OutputTokens: got %d, want %d", u.OutputTokens, 0)
				}
			},
		},
		{
			name: "large token counts",
			json: `{
				"input_tokens": 999999999,
				"output_tokens": 888888888
			}`,
			wantErr: false,
			validate: func(t *testing.T, u Usage) {
				if u.InputTokens != 999999999 {
					t.Errorf("InputTokens: got %d, want %d", u.InputTokens, 999999999)
				}
				if u.OutputTokens != 888888888 {
					t.Errorf("OutputTokens: got %d, want %d", u.OutputTokens, 888888888)
				}
			},
		},
		{
			name: "empty object",
			json: `{}`,
			wantErr: false,
			validate: func(t *testing.T, u Usage) {
				if u.InputTokens != 0 {
					t.Errorf("InputTokens: expected 0 for missing field, got %d", u.InputTokens)
				}
				if u.ServiceTier != "" {
					t.Errorf("ServiceTier: expected empty for missing field, got %q", u.ServiceTier)
				}
			},
		},
		{
			name:    "invalid json - wrong type",
			json:    `{"input_tokens": "not_a_number"}`,
			wantErr: true,
		},
		{
			name:    "invalid json - syntax error",
			json:    `{"input_tokens": }`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var usage Usage
			err := json.Unmarshal([]byte(tt.json), &usage)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, usage)
			}
		})
	}
}

func TestUsage_JSONMarshal(t *testing.T) {
	usage := Usage{
		InputTokens:              1000,
		CacheCreationInputTokens: 500,
		CacheReadInputTokens:     200,
		OutputTokens:             300,
		ServiceTier:              "premium",
	}

	data, err := json.Marshal(usage)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Check all fields are present with correct values
	if result["input_tokens"] != float64(1000) {
		t.Errorf("input_tokens: got %v, want %v", result["input_tokens"], 1000)
	}
	if result["cache_creation_input_tokens"] != float64(500) {
		t.Errorf("cache_creation_input_tokens: got %v, want %v", result["cache_creation_input_tokens"], 500)
	}
	if result["cache_read_input_tokens"] != float64(200) {
		t.Errorf("cache_read_input_tokens: got %v, want %v", result["cache_read_input_tokens"], 200)
	}
	if result["output_tokens"] != float64(300) {
		t.Errorf("output_tokens: got %v, want %v", result["output_tokens"], 300)
	}
	if result["service_tier"] != "premium" {
		t.Errorf("service_tier: got %v, want %q", result["service_tier"], "premium")
	}
}
