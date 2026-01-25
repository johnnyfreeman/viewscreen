package result

import (
	"encoding/json"
	"fmt"

	"github.com/jfreeman/viewscreen/config"
	"github.com/jfreeman/viewscreen/style"
	"github.com/jfreeman/viewscreen/types"
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

// Render outputs the result event to the terminal
func Render(event Event) {
	fmt.Println()
	if event.IsError {
		// Error header with gradient
		header := fmt.Sprintf("%sSession Error", style.Bullet)
		if !style.NoColor() {
			header = style.ApplyErrorGradient(header)
		} else {
			header = style.Error.Bold(true).Render(header)
		}
		fmt.Println(header)
		for _, err := range event.Errors {
			fmt.Printf("%s%s\n", style.OutputPrefix, style.Error.Render(err))
		}
	} else {
		// Success header with gradient
		header := fmt.Sprintf("%sSession Complete", style.Bullet)
		if !style.NoColor() {
			header = style.ApplySuccessGradient(header)
		} else {
			header = style.Success.Bold(true).Render(header)
		}
		fmt.Println(header)
	}

	fmt.Printf("%s%s %.2fs (API: %.2fs)\n",
		style.OutputPrefix,
		style.Muted.Render("Duration:"),
		float64(event.DurationMS)/1000, float64(event.DurationAPIMS)/1000)
	fmt.Printf("%s%s %d\n", style.OutputContinue, style.Muted.Render("Turns:"), event.NumTurns)
	fmt.Printf("%s%s $%.4f\n", style.OutputContinue, style.Muted.Render("Cost:"), event.TotalCostUSD)

	if config.ShowUsage {
		fmt.Printf("%s%s in=%d out=%d (cache: created=%d read=%d)\n",
			style.OutputContinue,
			style.Muted.Render("Tokens:"),
			event.Usage.InputTokens, event.Usage.OutputTokens,
			event.Usage.CacheCreationInputTokens, event.Usage.CacheReadInputTokens)
	}

	if len(event.PermissionDenials) > 0 {
		fmt.Printf("%s%s %d\n",
			style.OutputContinue,
			style.Warning.Render("Permission Denials:"),
			len(event.PermissionDenials))
		for _, denial := range event.PermissionDenials {
			fmt.Printf("%s  - %s (%s)\n", style.OutputContinue, denial.ToolName, denial.ToolUseID)
		}
	}
}
