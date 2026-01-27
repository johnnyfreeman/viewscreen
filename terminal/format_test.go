package terminal

import "testing"

func TestDefaultWidth(t *testing.T) {
	if DefaultWidth != 80 {
		t.Errorf("DefaultWidth = %d, expected 80", DefaultWidth)
	}
}

func TestWidth(t *testing.T) {
	// Width() should return a positive value (either actual terminal width or DefaultWidth)
	w := Width()
	if w <= 0 {
		t.Errorf("Width() = %d, expected positive value", w)
	}
}
