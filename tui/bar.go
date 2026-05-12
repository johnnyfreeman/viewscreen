package tui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// fitBarLine returns a single terminal row that is exactly width cells wide.
func fitBarLine(line string, width int) string {
	if width <= 0 {
		return ""
	}
	if ansi.StringWidth(line) > width {
		line = ansi.Truncate(line, width, "")
	}
	if visibleLen := ansi.StringWidth(line); visibleLen < width {
		line += strings.Repeat(" ", width-visibleLen)
	}
	return line
}

func leftmostCells(s string, width int) string {
	if width <= 0 {
		return ""
	}
	return ansi.Cut(s, 0, width)
}

func rightmostCells(s string, width int) string {
	if width <= 0 {
		return ""
	}
	visibleLen := ansi.StringWidth(s)
	if visibleLen <= width {
		return s
	}
	return ansi.Cut(s, visibleLen-width, visibleLen)
}
