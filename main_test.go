package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/johnnyfreeman/viewscreen/parser"
)

func TestNewRunner_Defaults(t *testing.T) {
	r := NewRunner()

	if r.errOutput == nil {
		t.Error("expected errOutput to be set")
	}
	if r.parserFactory == nil {
		t.Error("expected parserFactory to be set")
	}
	if r.exitFunc == nil {
		t.Error("expected exitFunc to be set")
	}
	if r.parseFlags == nil {
		t.Error("expected parseFlags to be set")
	}
}

func TestNewRunner_WithOptions(t *testing.T) {
	t.Run("WithErrOutput", func(t *testing.T) {
		buf := &bytes.Buffer{}
		r := NewRunner(WithErrOutput(buf))

		if r.errOutput != buf {
			t.Error("expected custom errOutput to be set")
		}
	})

	t.Run("WithParserFactory", func(t *testing.T) {
		factoryCalled := false
		factory := func() *parser.Parser {
			factoryCalled = true
			return parser.NewParserWithOptions(
				parser.WithInput(strings.NewReader("")),
			)
		}

		r := NewRunner(
			WithParserFactory(factory),
			WithParseFlags(func() {}),
		)

		r.Run()

		if !factoryCalled {
			t.Error("expected custom parserFactory to be called")
		}
	})

	t.Run("WithExitFunc", func(t *testing.T) {
		exitCalled := false
		exitCode := 0
		exitFunc := func(code int) {
			exitCalled = true
			exitCode = code
		}

		// Create a parser that returns an error
		factory := func() *parser.Parser {
			return parser.NewParserWithOptions(
				parser.WithInput(&errorReader{}),
			)
		}

		errBuf := &bytes.Buffer{}
		r := NewRunner(
			WithErrOutput(errBuf),
			WithParserFactory(factory),
			WithExitFunc(exitFunc),
			WithParseFlags(func() {}),
		)

		r.Run()

		if !exitCalled {
			t.Error("expected exit function to be called")
		}
		if exitCode != 1 {
			t.Errorf("expected exit code 1, got %d", exitCode)
		}
	})

	t.Run("WithParseFlags", func(t *testing.T) {
		flagsCalled := false
		parseFlags := func() {
			flagsCalled = true
		}

		r := NewRunner(
			WithParserFactory(func() *parser.Parser {
				return parser.NewParserWithOptions(
					parser.WithInput(strings.NewReader("")),
				)
			}),
			WithParseFlags(parseFlags),
		)

		r.Run()

		if !flagsCalled {
			t.Error("expected parseFlags to be called")
		}
	})
}

// errorReader always returns an error when reading
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}

func TestRunner_Run_Success(t *testing.T) {
	flagsCalled := false
	factoryCalled := false

	r := NewRunner(
		WithParseFlags(func() {
			flagsCalled = true
		}),
		WithParserFactory(func() *parser.Parser {
			factoryCalled = true
			return parser.NewParserWithOptions(
				parser.WithInput(strings.NewReader("")),
			)
		}),
	)

	r.Run()

	if !flagsCalled {
		t.Error("expected parseFlags to be called")
	}
	if !factoryCalled {
		t.Error("expected parserFactory to be called")
	}
}

func TestRunner_Run_Error(t *testing.T) {
	errBuf := &bytes.Buffer{}
	exitCalled := false
	exitCode := -1

	r := NewRunner(
		WithErrOutput(errBuf),
		WithParseFlags(func() {}),
		WithParserFactory(func() *parser.Parser {
			return parser.NewParserWithOptions(
				parser.WithInput(&errorReader{}),
			)
		}),
		WithExitFunc(func(code int) {
			exitCalled = true
			exitCode = code
		}),
	)

	r.Run()

	if !exitCalled {
		t.Error("expected exit function to be called on error")
	}
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(errBuf.String(), "error reading input") {
		t.Errorf("expected error message in stderr, got: %s", errBuf.String())
	}
}

func TestRunner_Run_ValidEvents(t *testing.T) {
	// Test with a valid system event
	input := `{"type":"system","subtype":"init","cwd":"/test","model":"test-model","claude_code_version":"1.0.0","tools":[]}`

	r := NewRunner(
		WithParseFlags(func() {}),
		WithParserFactory(func() *parser.Parser {
			return parser.NewParserWithOptions(
				parser.WithInput(strings.NewReader(input)),
			)
		}),
	)

	// Should not panic or call exit
	r.Run()
}

func TestRunner_Run_MultipleEvents(t *testing.T) {
	// Test with multiple valid events
	events := []string{
		`{"type":"system","subtype":"init","cwd":"/test","model":"test-model","claude_code_version":"1.0.0","tools":[]}`,
		`{"type":"result","subtype":"success","is_error":false,"duration_ms":100,"result":"done"}`,
	}
	input := strings.Join(events, "\n")

	r := NewRunner(
		WithParseFlags(func() {}),
		WithParserFactory(func() *parser.Parser {
			return parser.NewParserWithOptions(
				parser.WithInput(strings.NewReader(input)),
			)
		}),
	)

	// Should not panic or call exit
	r.Run()
}

func TestWithErrOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	opt := WithErrOutput(buf)

	r := &Runner{}
	opt(r)

	if r.errOutput != buf {
		t.Error("expected errOutput to be set by option")
	}
}

func TestWithParserFactory(t *testing.T) {
	called := false
	factory := func() *parser.Parser {
		called = true
		return nil
	}
	opt := WithParserFactory(factory)

	r := &Runner{}
	opt(r)

	if r.parserFactory == nil {
		t.Error("expected parserFactory to be set by option")
	}

	r.parserFactory()
	if !called {
		t.Error("expected factory to be callable")
	}
}

func TestWithExitFunc(t *testing.T) {
	called := false
	capturedCode := -1
	exitFn := func(code int) {
		called = true
		capturedCode = code
	}
	opt := WithExitFunc(exitFn)

	r := &Runner{}
	opt(r)

	if r.exitFunc == nil {
		t.Error("expected exitFunc to be set by option")
	}

	r.exitFunc(42)
	if !called {
		t.Error("expected exit function to be callable")
	}
	if capturedCode != 42 {
		t.Errorf("expected exit code 42, got %d", capturedCode)
	}
}

func TestWithParseFlags(t *testing.T) {
	called := false
	parseFn := func() {
		called = true
	}
	opt := WithParseFlags(parseFn)

	r := &Runner{}
	opt(r)

	if r.parseFlags == nil {
		t.Error("expected parseFlags to be set by option")
	}

	r.parseFlags()
	if !called {
		t.Error("expected parseFlags to be callable")
	}
}
