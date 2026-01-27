// Package render provides rendering utilities for terminal output.
package render

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/johnnyfreeman/viewscreen/style"
)

// StyleApplier abstracts style application for testability across all renderer packages.
// This unified interface replaces the individual StyleApplier interfaces that were
// previously defined in user, system, and result packages.
type StyleApplier interface {
	// Text styles (lipgloss-based, for simple cases)
	ErrorRender(text string) string
	MutedRender(text string) string
	SuccessRender(text string) string
	WarningRender(text string) string

	// Ultraviolet-based text styles (for composition-safe styling)
	// Use these when styled text might be embedded in other styled content.
	UVSuccessText(text string) string
	UVWarningText(text string) string
	UVMutedText(text string) string
	UVErrorText(text string) string
	UVErrorBoldText(text string) string
	UVSuccessBoldText(text string) string

	// Output prefixes
	OutputPrefix() string
	OutputContinue() string
	Bullet() string

	// Diff-related styles
	LineNumberRender(text string) string
	LineNumberSepRender(text string) string
	DiffAddRender(text string) string
	DiffRemoveRender(text string) string
	DiffAddBg() lipgloss.Color
	DiffRemoveBg() lipgloss.Color

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

// Text styles (lipgloss-based)
func (d DefaultStyleApplier) ErrorRender(text string) string   { return style.Error.Render(text) }
func (d DefaultStyleApplier) MutedRender(text string) string   { return style.Muted.Render(text) }
func (d DefaultStyleApplier) SuccessRender(text string) string { return style.Success.Render(text) }
func (d DefaultStyleApplier) WarningRender(text string) string { return style.Warning.Render(text) }

// Ultraviolet-based text styles (composition-safe)
func (d DefaultStyleApplier) UVSuccessText(text string) string   { return style.SuccessText(text) }
func (d DefaultStyleApplier) UVWarningText(text string) string   { return style.WarningText(text) }
func (d DefaultStyleApplier) UVMutedText(text string) string     { return style.MutedText(text) }
func (d DefaultStyleApplier) UVErrorText(text string) string     { return style.ErrorText(text) }
func (d DefaultStyleApplier) UVErrorBoldText(text string) string { return style.ErrorBoldText(text) }
func (d DefaultStyleApplier) UVSuccessBoldText(text string) string {
	return style.SuccessBoldText(text)
}

// Output prefixes
func (d DefaultStyleApplier) OutputPrefix() string   { return style.OutputPrefix }
func (d DefaultStyleApplier) OutputContinue() string { return style.OutputContinue }
func (d DefaultStyleApplier) Bullet() string         { return style.Bullet }

// Diff-related styles
func (d DefaultStyleApplier) LineNumberRender(text string) string    { return style.LineNumber.Render(text) }
func (d DefaultStyleApplier) LineNumberSepRender(text string) string { return style.LineNumberSep.Render("│") }
func (d DefaultStyleApplier) DiffAddRender(text string) string       { return style.DiffAdd.Render(text) }
func (d DefaultStyleApplier) DiffRemoveRender(text string) string    { return style.DiffRemove.Render(text) }
func (d DefaultStyleApplier) DiffAddBg() lipgloss.Color              { return style.DiffAddBg }
func (d DefaultStyleApplier) DiffRemoveBg() lipgloss.Color           { return style.DiffRemoveBg }

// Session/header styles
func (d DefaultStyleApplier) SessionHeaderRender(text string) string { return style.SessionHeader.Render(text) }
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
