package result

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/style"
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
	output    io.Writer
	showUsage func() bool
	noColor   func() bool
}

// RendererOption is a functional option for configuring a Renderer
type RendererOption func(*Renderer)

// WithOutput sets the output writer for the renderer
func WithOutput(w io.Writer) RendererOption {
	return func(r *Renderer) {
		r.output = w
	}
}

// WithShowUsage sets the function to check if usage should be shown
func WithShowUsage(fn func() bool) RendererOption {
	return func(r *Renderer) {
		r.showUsage = fn
	}
}

// WithNoColor sets the function to check if color is disabled
func WithNoColor(fn func() bool) RendererOption {
	return func(r *Renderer) {
		r.noColor = fn
	}
}

// NewRenderer creates a new Renderer with the given options
func NewRenderer(opts ...RendererOption) *Renderer {
	r := &Renderer{
		output:    os.Stdout,
		showUsage: func() bool { return config.ShowUsage },
		noColor:   style.NoColor,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Render outputs the result event using this renderer's configuration
func (r *Renderer) Render(event Event) {
	fmt.Fprintln(r.output)
	if event.IsError {
		// Error header with gradient
		header := fmt.Sprintf("%sSession Error", style.Bullet)
		if !r.noColor() {
			header = style.ApplyErrorGradient(header)
		} else {
			header = style.Error.Bold(true).Render(header)
		}
		fmt.Fprintln(r.output, header)
		for _, err := range event.Errors {
			fmt.Fprintf(r.output, "%s%s\n", style.OutputPrefix, style.Error.Render(err))
		}
	} else {
		// Success header with gradient
		header := fmt.Sprintf("%sSession Complete", style.Bullet)
		if !r.noColor() {
			header = style.ApplySuccessGradient(header)
		} else {
			header = style.Success.Bold(true).Render(header)
		}
		fmt.Fprintln(r.output, header)
	}

	fmt.Fprintf(r.output, "%s%s %.2fs (API: %.2fs)\n",
		style.OutputPrefix,
		style.Muted.Render("Duration:"),
		float64(event.DurationMS)/1000, float64(event.DurationAPIMS)/1000)
	fmt.Fprintf(r.output, "%s%s %d\n", style.OutputContinue, style.Muted.Render("Turns:"), event.NumTurns)
	fmt.Fprintf(r.output, "%s%s $%.4f\n", style.OutputContinue, style.Muted.Render("Cost:"), event.TotalCostUSD)

	if r.showUsage() {
		fmt.Fprintf(r.output, "%s%s in=%d out=%d (cache: created=%d read=%d)\n",
			style.OutputContinue,
			style.Muted.Render("Tokens:"),
			event.Usage.InputTokens, event.Usage.OutputTokens,
			event.Usage.CacheCreationInputTokens, event.Usage.CacheReadInputTokens)
	}

	if len(event.PermissionDenials) > 0 {
		fmt.Fprintf(r.output, "%s%s %d\n",
			style.OutputContinue,
			style.Warning.Render("Permission Denials:"),
			len(event.PermissionDenials))
		for _, denial := range event.PermissionDenials {
			fmt.Fprintf(r.output, "%s  - %s (%s)\n", style.OutputContinue, denial.ToolName, denial.ToolUseID)
		}
	}
}

// defaultRenderer is the default renderer used by the Render function
var defaultRenderer = NewRenderer()

// Render outputs the result event to the terminal using the default renderer
func Render(event Event) {
	defaultRenderer.Render(event)
}
