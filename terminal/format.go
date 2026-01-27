// Package terminal provides terminal-specific utilities.
package terminal

import (
	"os"

	"golang.org/x/term"
)

// DefaultWidth is the fallback terminal width when detection fails.
const DefaultWidth = 80

// Width returns the current terminal width, or DefaultWidth if detection fails.
// This centralizes terminal width detection to avoid duplication across packages.
func Width() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	return DefaultWidth
}
