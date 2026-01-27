package result

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/types"
)

// Usage represents usage in result events
type Usage struct {
	InputTokens              int `json:"input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	OutputTokens             int `json:"output_tokens"`
}

// ModelUsage represents per-model usage
type ModelUsage struct {
	InputTokens              int     `json:"inputTokens"`
	OutputTokens             int     `json:"outputTokens"`
	CacheReadInputTokens     int     `json:"cacheReadInputTokens"`
	CacheCreationInputTokens int     `json:"cacheCreationInputTokens"`
	CostUSD                  float64 `json:"costUSD"`
	ContextWindow            int     `json:"contextWindow"`
	MaxOutputTokens          int     `json:"maxOutputTokens"`
}

// PermissionDenial represents a denied permission request
type PermissionDenial struct {
	ToolName  string          `json:"tool_name"`
	ToolUseID string          `json:"tool_use_id"`
	ToolInput json.RawMessage `json:"tool_input"`
}

// Event represents a final result event
type Event struct {
	types.BaseEvent
	Subtype           string                `json:"subtype"`
	IsError           bool                  `json:"is_error"`
	DurationMS        int                   `json:"duration_ms"`
	DurationAPIMS     int                   `json:"duration_api_ms"`
	NumTurns          int                   `json:"num_turns"`
	Result            string                `json:"result"`
	TotalCostUSD      float64               `json:"total_cost_usd"`
	Usage             Usage                 `json:"usage"`
	ModelUsage        map[string]ModelUsage `json:"modelUsage"`
	PermissionDenials []PermissionDenial    `json:"permission_denials"`
	Errors            []string              `json:"errors"`
}

// Renderer handles rendering of result events with configurable output and options
type Renderer struct {
	output       io.Writer
	config       config.Provider
	styleApplier render.StyleApplier
}

// RendererOption is a functional option for configuring a Renderer
type RendererOption func(*Renderer)

// WithOutput sets the output writer for the renderer
func WithOutput(w io.Writer) RendererOption {
	return func(r *Renderer) {
		r.output = w
	}
}

// WithConfigProvider sets a custom config provider
func WithConfigProvider(cp config.Provider) RendererOption {
	return func(r *Renderer) {
		r.config = cp
	}
}

// WithStyleApplier sets a custom style applier
func WithStyleApplier(sa render.StyleApplier) RendererOption {
	return func(r *Renderer) {
		r.styleApplier = sa
	}
}

// NewRenderer creates a new Renderer with the given options
func NewRenderer(opts ...RendererOption) *Renderer {
	r := &Renderer{
		output:       os.Stdout,
		config:       config.DefaultProvider{},
		styleApplier: render.DefaultStyleApplier{},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// renderTo writes the result event to the given output
func (r *Renderer) renderTo(out *render.Output, event Event) {
	sa := r.styleApplier
	fmt.Fprintln(out)
	if event.IsError {
		// Error header with gradient (or bold fallback for no-color mode)
		header := fmt.Sprintf("%sSession Error", sa.Bullet())
		if !sa.NoColor() {
			header = sa.ApplyErrorGradient(header)
		} else {
			header = sa.ErrorBoldText(header)
		}
		fmt.Fprintln(out, header)
		for _, err := range event.Errors {
			fmt.Fprintf(out, "%s%s\n", sa.OutputPrefix(), sa.ErrorText(err))
		}
	} else {
		// Success header with gradient (or bold fallback for no-color mode)
		header := fmt.Sprintf("%sSession Complete", sa.Bullet())
		if !sa.NoColor() {
			header = sa.ApplySuccessGradient(header)
		} else {
			header = sa.SuccessBoldText(header)
		}
		fmt.Fprintln(out, header)
	}

	fmt.Fprintf(out, "%s%s %.2fs (API: %.2fs)\n",
		sa.OutputPrefix(),
		sa.MutedText("Duration:"),
		float64(event.DurationMS)/1000, float64(event.DurationAPIMS)/1000)
	fmt.Fprintf(out, "%s%s %d\n", sa.OutputContinue(), sa.MutedText("Turns:"), event.NumTurns)
	fmt.Fprintf(out, "%s%s $%.4f\n", sa.OutputContinue(), sa.MutedText("Cost:"), event.TotalCostUSD)

	if r.config.ShowUsage() {
		fmt.Fprintf(out, "%s%s in=%d out=%d (cache: created=%d read=%d)\n",
			sa.OutputContinue(),
			sa.MutedText("Tokens:"),
			event.Usage.InputTokens, event.Usage.OutputTokens,
			event.Usage.CacheCreationInputTokens, event.Usage.CacheReadInputTokens)
	}

	if len(event.PermissionDenials) > 0 {
		fmt.Fprintf(out, "%s%s %d\n",
			sa.OutputContinue(),
			sa.WarningText("Permission Denials:"),
			len(event.PermissionDenials))
		for _, denial := range event.PermissionDenials {
			fmt.Fprintf(out, "%s  - %s (%s)\n", sa.OutputContinue(), denial.ToolName, denial.ToolUseID)
		}
	}
}

// Render outputs the result event using this renderer's configuration
func (r *Renderer) Render(event Event) {
	r.renderTo(render.WriterOutput(r.output), event)
}

// RenderToString renders the result event to a string
func (r *Renderer) RenderToString(event Event) string {
	out := render.StringOutput()
	r.renderTo(out, event)
	return out.String()
}

