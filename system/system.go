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

// MCPServer represents a configured MCP server
type MCPServer struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Plugin represents a configured plugin
type Plugin struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Event represents a system initialization event
type Event struct {
	types.BaseEvent
	Subtype           string      `json:"subtype"`
	CWD               string      `json:"cwd"`
	Tools             []string    `json:"tools"`
	Model             string      `json:"model"`
	PermissionMode    string      `json:"permissionMode"`
	ClaudeCodeVersion string      `json:"claude_code_version"`
	Agents            []string    `json:"agents"`
	MCPServers        []MCPServer `json:"mcp_servers"`
	SlashCommands     []string    `json:"slash_commands"`
	Skills            []string    `json:"skills"`
	Plugins           []Plugin    `json:"plugins"`
	FastModeState     string      `json:"fast_mode_state"`
	APIKeySource      string      `json:"apiKeySource"`
	OutputStyle       string      `json:"output_style"`
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

// NewRenderer creates a new system Renderer with the given options
func NewRenderer(opts ...RendererOption) *Renderer {
	r := &Renderer{
		output:       os.Stdout,
		styleApplier: render.DefaultStyleApplier{},
		config:       config.Get(),
	}
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
	if r.config.IsVerbose() && len(event.MCPServers) > 0 {
		names := make([]string, len(event.MCPServers))
		for i, s := range event.MCPServers {
			names[i] = s.Name
		}
		fmt.Fprintf(out, "%s%s %s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedText("MCP Servers:"), strings.Join(names, ", "))
	}
	if r.config.IsVerbose() && len(event.Plugins) > 0 {
		names := make([]string, len(event.Plugins))
		for i, p := range event.Plugins {
			names[i] = p.Name
		}
		fmt.Fprintf(out, "%s%s %s\n", r.styleApplier.OutputContinue(), r.styleApplier.MutedText("Plugins:"), strings.Join(names, ", "))
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

