package content

import (
	"encoding/json"
	"testing"
)

func TestExtractText(t *testing.T) {
	tests := []struct {
		name string
		raw  json.RawMessage
		want string
	}{
		{
			name: "empty input",
			raw:  nil,
			want: "",
		},
		{
			name: "empty array",
			raw:  json.RawMessage(`[]`),
			want: "",
		},
		{
			name: "simple string",
			raw:  json.RawMessage(`"hello world"`),
			want: "hello world",
		},
		{
			name: "string with newlines",
			raw:  json.RawMessage(`"line1\nline2\nline3"`),
			want: "line1\nline2\nline3",
		},
		{
			name: "single text block",
			raw:  json.RawMessage(`[{"type": "text", "text": "hello"}]`),
			want: "hello",
		},
		{
			name: "multiple text blocks",
			raw: json.RawMessage(`[
				{"type": "text", "text": "hello"},
				{"type": "text", "text": "world"}
			]`),
			want: "hello\nworld",
		},
		{
			name: "mixed block types",
			raw: json.RawMessage(`[
				{"type": "text", "text": "first"},
				{"type": "image", "data": "..."},
				{"type": "text", "text": "second"}
			]`),
			want: "first\nsecond",
		},
		{
			name: "skip empty text blocks",
			raw: json.RawMessage(`[
				{"type": "text", "text": ""},
				{"type": "text", "text": "content"}
			]`),
			want: "content",
		},
		{
			name: "fallback for invalid JSON",
			raw:  json.RawMessage(`not valid json`),
			want: "not valid json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractText(tt.raw)
			if got != tt.want {
				t.Errorf("ExtractText() = %q, want %q", got, tt.want)
			}
		})
	}
}
