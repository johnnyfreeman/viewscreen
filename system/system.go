package system

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/render"
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
	styleApplier   render.StyleApplier
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
func WithStyleApplier(sa render.StyleApplier) RendererOption {
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
		styleApplier:   render.DefaultStyleApplier{},
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

// renderTo writes the system event to the given output
func (r *Renderer) renderTo(out *render.Output, event Event) {
	// Use gradient for session header when color is enabled
	header := fmt.Sprintf("%sSession Started", r.styleApplier.Bullet())
	if !r.styleApplier.NoColor() {
		header = r.styleApplier.ApplyThemeBoldGradient(header)
	} else {
		header = r.styleApplier.SessionHeaderRender(header)
	}
	fmt.Fprintln(out, header)
	fmt.Fprintf(out, "%s%s %s\n", r.styleApplier.OutputPrefix(), r.styleApplier.MutedRender("Model:"), event.Model)
	fmt.Fprintf(out, "%s%s %s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedRender("Version:"), event.ClaudeCodeVersion)
	fmt.Fprintf(out, "%s%s %s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedRender("CWD:"), event.CWD)
	fmt.Fprintf(out, "%s%s %d available\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedRender("Tools:"), len(event.Tools))
	if r.verboseChecker.IsVerbose() && len(event.Agents) > 0 {
		fmt.Fprintf(out, "%s%s %s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedRender("Agents:"), strings.Join(event.Agents, ", "))
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

