package tui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/johnnyfreeman/viewscreen/style"
)

var promptBarNewlineReplacer = strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ")

// PromptEditor holds the state for the prompt editing feature.
// When active, it captures keyboard input to let the user edit the prompt
// that will be used for the next agent invocation.
type PromptEditor struct {
	Active bool   // Whether prompt editing is active
	Value  string // Current prompt text
	cursor int    // Cursor position (byte offset)
}

// NewPromptEditor creates a new PromptEditor with default state.
func NewPromptEditor() PromptEditor {
	return PromptEditor{}
}

// Enter activates prompt editing mode, pre-populating with the given prompt.
func (p *PromptEditor) Enter(currentPrompt string) {
	p.Active = true
	p.Value = currentPrompt
	p.cursor = len(p.Value)
}

// Exit deactivates prompt editing mode, keeping the value.
func (p *PromptEditor) Exit() {
	p.Active = false
}

// Cancel deactivates prompt editing mode and restores the original value.
func (p *PromptEditor) Cancel(originalPrompt string) {
	p.Active = false
	p.Value = originalPrompt
	p.cursor = 0
}

// TypeRune inserts a rune at the cursor position.
func (p *PromptEditor) TypeRune(r rune) {
	s := string(r)
	p.Value = p.Value[:p.cursor] + s + p.Value[p.cursor:]
	p.cursor += len(s)
}

// Backspace removes the character before the cursor.
func (p *PromptEditor) Backspace() {
	if p.cursor > 0 {
		// Find the start of the previous rune
		prev := p.cursor - 1
		for prev > 0 && !isRuneStart(p.Value[prev]) {
			prev--
		}
		p.Value = p.Value[:prev] + p.Value[p.cursor:]
		p.cursor = prev
	}
}

// Delete removes the character after the cursor.
func (p *PromptEditor) Delete() {
	if p.cursor < len(p.Value) {
		// Find the end of the current rune
		next := p.cursor + 1
		for next < len(p.Value) && !isRuneStart(p.Value[next]) {
			next++
		}
		p.Value = p.Value[:p.cursor] + p.Value[next:]
	}
}

// CursorLeft moves the cursor one character to the left.
func (p *PromptEditor) CursorLeft() {
	if p.cursor > 0 {
		p.cursor--
		for p.cursor > 0 && !isRuneStart(p.Value[p.cursor]) {
			p.cursor--
		}
	}
}

// CursorRight moves the cursor one character to the right.
func (p *PromptEditor) CursorRight() {
	if p.cursor < len(p.Value) {
		p.cursor++
		for p.cursor < len(p.Value) && !isRuneStart(p.Value[p.cursor]) {
			p.cursor++
		}
	}
}

// CursorHome moves the cursor to the beginning.
func (p *PromptEditor) CursorHome() {
	p.cursor = 0
}

// CursorEnd moves the cursor to the end.
func (p *PromptEditor) CursorEnd() {
	p.cursor = len(p.Value)
}

// isRuneStart returns true if the byte is the start of a UTF-8 rune.
func isRuneStart(b byte) bool {
	return b&0xC0 != 0x80
}

// RenderPromptBar renders the prompt editor bar at the bottom of the viewport.
func RenderPromptBar(p PromptEditor, width int) string {
	if !p.Active {
		return ""
	}
	if width <= 0 {
		return ""
	}

	prefix := style.AccentText("prompt> ")
	valueWidth := width - ansi.StringWidth(prefix)
	if valueWidth <= 0 {
		return fitBarLine(prefix, width)
	}

	before := sanitizePromptBarSegment(p.Value[:p.cursor])
	after := sanitizePromptBarSegment(p.Value[p.cursor:])
	cursor := style.MutedText("█")

	return fitBarLine(prefix+renderPromptValue(before, cursor, after, valueWidth), width)
}

func sanitizePromptBarSegment(s string) string {
	return promptBarNewlineReplacer.Replace(s)
}

func renderPromptValue(before, cursor, after string, width int) string {
	if width <= 0 {
		return ""
	}

	value := before + cursor + after
	if ansi.StringWidth(value) <= width {
		return value
	}

	cursorWidth := ansi.StringWidth(cursor)
	if cursorWidth >= width {
		return ansi.Truncate(cursor, width, "")
	}

	remaining := width - cursorWidth
	beforeWidth := ansi.StringWidth(before)
	afterWidth := ansi.StringWidth(after)

	afterBudget := min(afterWidth, remaining/2)
	beforeBudget := remaining - afterBudget
	if beforeWidth < beforeBudget {
		afterBudget += beforeBudget - beforeWidth
		beforeBudget = beforeWidth
	}
	if afterWidth < afterBudget {
		beforeBudget += afterBudget - afterWidth
		afterBudget = afterWidth
	}

	return rightmostCells(before, beforeBudget) + cursor + leftmostCells(after, afterBudget)
}
