package render

import (
	"bytes"
	"testing"
)

func TestWriterOutput(t *testing.T) {
	var buf bytes.Buffer
	out := WriterOutput(&buf)

	out.WriteString("hello ")
	out.Write([]byte("world"))

	if got := buf.String(); got != "hello world" {
		t.Errorf("WriterOutput: got %q, want %q", got, "hello world")
	}

	// String() should return empty for WriterOutput
	if got := out.String(); got != "" {
		t.Errorf("WriterOutput.String(): got %q, want %q", got, "")
	}
}

func TestStringOutput(t *testing.T) {
	out := StringOutput()

	out.WriteString("hello ")
	out.Write([]byte("world"))

	if got := out.String(); got != "hello world" {
		t.Errorf("StringOutput.String(): got %q, want %q", got, "hello world")
	}
}
