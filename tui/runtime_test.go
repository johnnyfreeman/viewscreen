package tui

import (
	"io"
	"strings"
	"testing"
)

func TestStreamInputReader(t *testing.T) {
	t.Run("uses piped stdin as stream input", func(t *testing.T) {
		input := strings.NewReader("stream json\n")

		got, err := io.ReadAll(streamInputReader(input, false))
		if err != nil {
			t.Fatalf("ReadAll returned error: %v", err)
		}
		if string(got) != "stream json\n" {
			t.Errorf("streamInputReader() = %q, want piped input", string(got))
		}
	})

	t.Run("does not read terminal stdin as stream input", func(t *testing.T) {
		terminalInput := strings.NewReader("q\n")

		got, err := io.ReadAll(streamInputReader(terminalInput, true))
		if err != nil {
			t.Fatalf("ReadAll returned error: %v", err)
		}
		if string(got) != "" {
			t.Errorf("streamInputReader() = %q, want empty stream for terminal stdin", string(got))
		}
	})

	t.Run("nil stdin is treated as empty stream", func(t *testing.T) {
		got, err := io.ReadAll(streamInputReader(nil, false))
		if err != nil {
			t.Fatalf("ReadAll returned error: %v", err)
		}
		if string(got) != "" {
			t.Errorf("streamInputReader(nil) = %q, want empty stream", string(got))
		}
	})
}
