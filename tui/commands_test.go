package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestTerminalModeResetSequence(t *testing.T) {
	seq := terminalModeResetSequence()

	for _, want := range []string{
		ansi.ResetModeBracketedPaste,
		ansi.ResetModeMouseX10,
		ansi.ResetModeMouseNormal,
		ansi.ResetModeMouseHighlight,
		ansi.ResetModeMouseButtonEvent,
		ansi.ResetModeMouseAnyEvent,
		ansi.ResetModeMouseExtUtf8,
		ansi.ResetModeMouseExtSgr,
		ansi.ResetModeMouseExtUrxvt,
		ansi.ResetModeMouseExtSgrPixel,
		ansi.ResetModeFocusEvent,
	} {
		if !strings.Contains(seq, want) {
			t.Fatalf("terminal reset sequence missing %q in %q", want, seq)
		}
	}
}
