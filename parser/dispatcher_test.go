package parser

import (
	"bytes"
	"errors"
	"testing"
)

func TestEventDispatcher_Register(t *testing.T) {
	d := NewEventDispatcher(&bytes.Buffer{})

	called := false
	d.Register("test", func(line []byte) error {
		called = true
		return nil
	})

	handled := d.Dispatch("test", []byte("{}"))
	if !handled {
		t.Error("expected event to be handled")
	}
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestEventDispatcher_UnknownType(t *testing.T) {
	d := NewEventDispatcher(&bytes.Buffer{})

	handled := d.Dispatch("unknown", []byte("{}"))
	if handled {
		t.Error("expected unknown type to not be handled")
	}
}

func TestEventDispatcher_HandlerError(t *testing.T) {
	errOut := &bytes.Buffer{}
	d := NewEventDispatcher(errOut)

	d.Register("error_type", func(line []byte) error {
		return errors.New("test error")
	})

	handled := d.Dispatch("error_type", []byte("{}"))
	if !handled {
		t.Error("expected event to be handled even with error")
	}

	if errOut.String() != "Error parsing error_type event: test error\n" {
		t.Errorf("unexpected error output: %q", errOut.String())
	}
}

func TestEventDispatcher_MultipleHandlers(t *testing.T) {
	d := NewEventDispatcher(&bytes.Buffer{})

	var order []string
	d.Register("first", func(line []byte) error {
		order = append(order, "first")
		return nil
	})
	d.Register("second", func(line []byte) error {
		order = append(order, "second")
		return nil
	})

	d.Dispatch("second", nil)
	d.Dispatch("first", nil)

	if len(order) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(order))
	}
	if order[0] != "second" || order[1] != "first" {
		t.Errorf("unexpected call order: %v", order)
	}
}
