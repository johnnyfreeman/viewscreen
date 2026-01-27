package user

import (
	"encoding/json"

	"github.com/johnnyfreeman/viewscreen/render"
)

// RenderContext bundles the parameters needed for rendering tool results.
// This reduces parameter count and makes the interface cleaner.
type RenderContext struct {
	Output         *render.Output
	OutputPrefix   string
	OutputContinue string
}

// ResultRenderer defines the interface for rendering specific tool result types.
// Each implementation handles one type of tool result (edit, write, todo, etc).
type ResultRenderer interface {
	// TryRender attempts to render the tool result.
	// Returns true if this renderer handled the result, false if it should
	// be passed to the next renderer in the chain.
	TryRender(ctx *RenderContext, toolUseResult json.RawMessage) bool
}

// ResultRendererFunc is a function adapter for ResultRenderer
type ResultRendererFunc func(ctx *RenderContext, toolUseResult json.RawMessage) bool

func (f ResultRendererFunc) TryRender(ctx *RenderContext, toolUseResult json.RawMessage) bool {
	return f(ctx, toolUseResult)
}

// ResultRegistry holds result renderers in priority order
type ResultRegistry struct {
	renderers []ResultRenderer
}

// NewResultRegistry creates an empty result registry
func NewResultRegistry() *ResultRegistry {
	return &ResultRegistry{}
}

// Register adds a renderer to the registry.
// Renderers are tried in the order they are registered.
func (r *ResultRegistry) Register(renderer ResultRenderer) {
	r.renderers = append(r.renderers, renderer)
}

// TryRender attempts to render the result using registered renderers.
// Returns true if any renderer handled the result.
func (r *ResultRegistry) TryRender(ctx *RenderContext, toolUseResult json.RawMessage) bool {
	for _, renderer := range r.renderers {
		if renderer.TryRender(ctx, toolUseResult) {
			return true
		}
	}
	return false
}
