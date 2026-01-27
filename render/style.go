// Package render provides rendering utilities for terminal output.
package render

import (
	"github.com/johnnyfreeman/viewscreen/style"
)

// StyleApplier abstracts style application for testability across all renderer packages.
// This unified interface replaces the individual StyleApplier interfaces that were
// previously defined in user, system, and result packages.
//
// All text styling uses Ultraviolet for composition-safe styling. Lipgloss is only
// used for layout (padding, width) and background colors (diff highlighting).
type StyleApplier interface {
	// Text styles (Ultraviolet-based, composition-safe)
	SuccessText(text string) string
	WarningText(text string) string
	MutedText(text string) string
	ErrorText(text string) string
	ErrorBoldText(text string) string
	SuccessBoldText(text string) string

	// Output prefixes
	OutputPrefix() string
	OutputContinue() string
	Bullet() string

	// Diff-related styles
	LineNumberRender(text string) string
	LineNumberSepRender(text string) string
	DiffAddRender(text string) string
	DiffRemoveRender(text string) string
	DiffAddBg() style.Color
	DiffRemoveBg() style.Color

	// Session/header styles
	SessionHeaderRender(text string) string
	ApplyThemeBoldGradient(text string) string
	ApplySuccessGradient(text string) string
	ApplyErrorGradient(text string) string

	// Color state
	NoColor() bool
}

// DefaultStyleApplier implements StyleApplier using the actual style package.
// Use this as the default implementation in production code.
type DefaultStyleApplier struct{}

// Text styles (Ultraviolet-based, composition-safe)
func (d DefaultStyleApplier) SuccessText(text string) string     { return style.SuccessText(text) }
func (d DefaultStyleApplier) WarningText(text string) string     { return style.WarningText(text) }
func (d DefaultStyleApplier) MutedText(text string) string       { return style.MutedText(text) }
func (d DefaultStyleApplier) ErrorText(text string) string       { return style.ErrorText(text) }
func (d DefaultStyleApplier) ErrorBoldText(text string) string   { return style.ErrorBoldText(text) }
func (d DefaultStyleApplier) SuccessBoldText(text string) string { return style.SuccessBoldText(text) }

// Output prefixes
func (d DefaultStyleApplier) OutputPrefix() string   { return style.OutputPrefix }
func (d DefaultStyleApplier) OutputContinue() string { return style.OutputContinue }
func (d DefaultStyleApplier) Bullet() string         { return style.Bullet }

// Diff-related styles
func (d DefaultStyleApplier) LineNumberRender(text string) string    { return style.LineNumberText(text) }
func (d DefaultStyleApplier) LineNumberSepRender(text string) string { return style.LineNumberSepText("│") }
func (d DefaultStyleApplier) DiffAddRender(text string) string       { return style.DiffAdd.Render(text) }
func (d DefaultStyleApplier) DiffRemoveRender(text string) string    { return style.DiffRemove.Render(text) }
func (d DefaultStyleApplier) DiffAddBg() style.Color    { return style.DiffAddBg }
func (d DefaultStyleApplier) DiffRemoveBg() style.Color { return style.DiffRemoveBg }

// Session/header styles
func (d DefaultStyleApplier) SessionHeaderRender(text string) string { return style.InfoBoldText(text) }
func (d DefaultStyleApplier) ApplyThemeBoldGradient(text string) string {
	return style.ApplyThemeBoldGradient(text)
}
func (d DefaultStyleApplier) ApplySuccessGradient(text string) string {
	return style.ApplySuccessGradient(text)
}
func (d DefaultStyleApplier) ApplyErrorGradient(text string) string {
	return style.ApplyErrorGradient(text)
}

// Color state
func (d DefaultStyleApplier) NoColor() bool { return style.NoColor() }
