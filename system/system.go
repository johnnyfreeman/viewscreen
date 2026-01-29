package system

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/render"
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


// Renderer handles rendering system events
type Renderer struct {
	output       io.Writer
	styleApplier render.StyleApplier
	config       config.Provider
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
func WithStyleApplier(sa render.StyleApplier) RendererOption {
	return func(r *Renderer) {
		r.styleApplier = sa
	}
}

// WithConfigProvider sets a custom config provider
func WithConfigProvider(cp config.Provider) RendererOption {
	return func(r *Renderer) {
		r.config = cp
	}
}

// NewRenderer creates a new system Renderer with default dependencies
func NewRenderer() *Renderer {
	return &Renderer{
		output:       os.Stdout,
		styleApplier: render.DefaultStyleApplier{},
		config:       config.DefaultProvider{},
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

// renderTo writes the system event to the given output
func (r *Renderer) renderTo(out *render.Output, event Event) {
	fmt.Fprintln(out, style.BulletHeader("Session Started"))
	fmt.Fprintf(out, "%s%s %s\n", r.styleApplier.OutputPrefix(), r.styleApplier.MutedText("Model:"), event.Model)
	fmt.Fprintf(out, "%s%s %s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedText("Version:"), event.ClaudeCodeVersion)
	fmt.Fprintf(out, "%s%s %s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedText("CWD:"), event.CWD)
	fmt.Fprintf(out, "%s%s %d available\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedText("Tools:"), len(event.Tools))
	if r.config.IsVerbose() && len(event.Agents) > 0 {
		fmt.Fprintf(out, "%s%s %s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedText("Agents:"), strings.Join(event.Agents, ", "))
	}
	fmt.Fprintln(out)
}

// Render outputs the system event
func (r *Renderer) Render(event Event) {
	r.renderTo(render.WriterOutput(r.output), event)
}

// RenderToString renders the system event to a string
func (r *Renderer) RenderToString(event Event) string {
	out := render.StringOutput()
	r.renderTo(out, event)
	return out.String()
}

