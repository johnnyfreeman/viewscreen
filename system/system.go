package system

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/types"
)

// Event represents a system initialization event
type Event struct {
	types.BaseEvent
	Subtype           string   `json:"subtype"`
	CWD               string   `json:"cwd"`
	Tools             []string `json:"tools"`
	Model             string   `json:"model"`
	PermissionMode    string   `json:"permissionMode"`
	ClaudeCodeVersion string   `json:"claude_code_version"`
	Agents            []string `json:"agents"`
}

// StyleApplier abstracts style application for testability
type StyleApplier interface {
	NoColor() bool
	ApplyThemeBoldGradient(text string) string
	SessionHeaderRender(text string) string
	MutedRender(text string) string
	Bullet() string
	OutputPrefix() string
	OutputContinue() string
}

// DefaultStyleApplier uses the actual style package
type DefaultStyleApplier struct{}

func (d DefaultStyleApplier) NoColor() bool                         { return style.NoColor() }
func (d DefaultStyleApplier) ApplyThemeBoldGradient(text string) string { return style.ApplyThemeBoldGradient(text) }
func (d DefaultStyleApplier) SessionHeaderRender(text string) string    { return style.SessionHeader.Render(text) }
func (d DefaultStyleApplier) MutedRender(text string) string            { return style.Muted.Render(text) }
func (d DefaultStyleApplier) Bullet() string                            { return style.Bullet }
func (d DefaultStyleApplier) OutputPrefix() string                      { return style.OutputPrefix }
func (d DefaultStyleApplier) OutputContinue() string                    { return style.OutputContinue }

// VerboseChecker abstracts verbose flag checking for testability
type VerboseChecker interface {
	IsVerbose() bool
}

// DefaultVerboseChecker uses the actual config package
type DefaultVerboseChecker struct{}

func (d DefaultVerboseChecker) IsVerbose() bool { return config.Verbose }

// Renderer handles rendering system events
type Renderer struct {
	output         io.Writer
	styleApplier   StyleApplier
	verboseChecker VerboseChecker
}

// RendererOption is a functional option for configuring a Renderer
type RendererOption func(*Renderer)

// WithOutput sets a custom output writer
func WithOutput(w io.Writer) RendererOption {
	return func(r *Renderer) {
		r.output = w
	}
}

// WithStyleApplier sets a custom style applier
func WithStyleApplier(sa StyleApplier) RendererOption {
	return func(r *Renderer) {
		r.styleApplier = sa
	}
}

// WithVerboseChecker sets a custom verbose checker
func WithVerboseChecker(vc VerboseChecker) RendererOption {
	return func(r *Renderer) {
		r.verboseChecker = vc
	}
}

// NewRenderer creates a new system Renderer with default dependencies
func NewRenderer() *Renderer {
	return &Renderer{
		output:         os.Stdout,
		styleApplier:   DefaultStyleApplier{},
		verboseChecker: DefaultVerboseChecker{},
	}
}

// NewRendererWithOptions creates a new system Renderer with custom options
func NewRendererWithOptions(opts ...RendererOption) *Renderer {
	r := NewRenderer()
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Render outputs the system event
func (r *Renderer) Render(event Event) {
	// Use gradient for session header when color is enabled
	header := fmt.Sprintf("%sSession Started", r.styleApplier.Bullet())
	if !r.styleApplier.NoColor() {
		header = r.styleApplier.ApplyThemeBoldGradient(header)
	} else {
		header = r.styleApplier.SessionHeaderRender(header)
	}
	fmt.Fprintln(r.output, header)
	fmt.Fprintf(r.output, "%s%s %s\n", r.styleApplier.OutputPrefix(), r.styleApplier.MutedRender("Model:"), event.Model)
	fmt.Fprintf(r.output, "%s%s %s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedRender("Version:"), event.ClaudeCodeVersion)
	fmt.Fprintf(r.output, "%s%s %s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedRender("CWD:"), event.CWD)
	fmt.Fprintf(r.output, "%s%s %d available\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedRender("Tools:"), len(event.Tools))
	if r.verboseChecker.IsVerbose() && len(event.Agents) > 0 {
		fmt.Fprintf(r.output, "%s%s %s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedRender("Agents:"), strings.Join(event.Agents, ", "))
	}
	fmt.Fprintln(r.output)
}

// Package-level renderer for backward compatibility
var defaultRenderer *Renderer

func getDefaultRenderer() *Renderer {
	if defaultRenderer == nil {
		defaultRenderer = NewRenderer()
	}
	return defaultRenderer
}

// Render is a package-level convenience function for backward compatibility
func Render(event Event) {
	getDefaultRenderer().Render(event)
}

// RenderToString renders the system event to a string
func (r *Renderer) RenderToString(event Event) string {
	var sb strings.Builder
	// Use gradient for session header when color is enabled
	header := fmt.Sprintf("%sSession Started", r.styleApplier.Bullet())
	if !r.styleApplier.NoColor() {
		header = r.styleApplier.ApplyThemeBoldGradient(header)
	} else {
		header = r.styleApplier.SessionHeaderRender(header)
	}
	sb.WriteString(header + "\n")
	sb.WriteString(fmt.Sprintf("%s%s %s\n", r.styleApplier.OutputPrefix(), r.styleApplier.MutedRender("Model:"), event.Model))
	sb.WriteString(fmt.Sprintf("%s%s %s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedRender("Version:"), event.ClaudeCodeVersion))
	sb.WriteString(fmt.Sprintf("%s%s %s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedRender("CWD:"), event.CWD))
	sb.WriteString(fmt.Sprintf("%s%s %d available\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedRender("Tools:"), len(event.Tools)))
	if r.verboseChecker.IsVerbose() && len(event.Agents) > 0 {
		sb.WriteString(fmt.Sprintf("%s%s %s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedRender("Agents:"), strings.Join(event.Agents, ", ")))
	}
	sb.WriteString("\n")
	return sb.String()
}

// RenderToString is a package-level convenience function for backward compatibility
func RenderToString(event Event) string {
	return getDefaultRenderer().RenderToString(event)
}
