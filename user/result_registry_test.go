package user

import (
	"encoding/json"
	"testing"

	"github.com/johnnyfreeman/viewscreen/render"
)

func TestResultRegistry_TryRender_EmptyRegistry(t *testing.T) {
	registry := NewResultRegistry()
	ctx := &RenderContext{
		Output:         render.StringOutput(),
		OutputPrefix:   "prefix ",
		OutputContinue: "cont ",
	}

	result := registry.TryRender(ctx, json.RawMessage(`{}`))
	if result {
		t.Error("Expected empty registry to return false")
	}
}

func TestResultRegistry_TryRender_FirstMatch(t *testing.T) {
	registry := NewResultRegistry()

	calls := []string{}

	// First renderer that matches
	registry.Register(ResultRendererFunc(func(ctx *RenderContext, data json.RawMessage) bool {
		calls = append(calls, "first")
		return true
	}))

	// Second renderer should not be called
	registry.Register(ResultRendererFunc(func(ctx *RenderContext, data json.RawMessage) bool {
		calls = append(calls, "second")
		return true
	}))

	ctx := &RenderContext{
		Output:         render.StringOutput(),
		OutputPrefix:   "prefix ",
		OutputContinue: "cont ",
	}

	result := registry.TryRender(ctx, json.RawMessage(`{}`))
	if !result {
		t.Error("Expected registry to return true when renderer matches")
	}
	if len(calls) != 1 || calls[0] != "first" {
		t.Errorf("Expected only first renderer to be called, got calls: %v", calls)
	}
}

func TestResultRegistry_TryRender_FallsThrough(t *testing.T) {
	registry := NewResultRegistry()

	calls := []string{}

	// First renderer doesn't match
	registry.Register(ResultRendererFunc(func(ctx *RenderContext, data json.RawMessage) bool {
		calls = append(calls, "first")
		return false
	}))

	// Second renderer matches
	registry.Register(ResultRendererFunc(func(ctx *RenderContext, data json.RawMessage) bool {
		calls = append(calls, "second")
		return true
	}))

	ctx := &RenderContext{
		Output:         render.StringOutput(),
		OutputPrefix:   "prefix ",
		OutputContinue: "cont ",
	}

	result := registry.TryRender(ctx, json.RawMessage(`{}`))
	if !result {
		t.Error("Expected registry to return true when second renderer matches")
	}
	if len(calls) != 2 || calls[0] != "first" || calls[1] != "second" {
		t.Errorf("Expected both renderers to be called in order, got calls: %v", calls)
	}
}

func TestResultRegistry_TryRender_NoneMatch(t *testing.T) {
	registry := NewResultRegistry()

	calls := []string{}

	// Neither renderer matches
	registry.Register(ResultRendererFunc(func(ctx *RenderContext, data json.RawMessage) bool {
		calls = append(calls, "first")
		return false
	}))

	registry.Register(ResultRendererFunc(func(ctx *RenderContext, data json.RawMessage) bool {
		calls = append(calls, "second")
		return false
	}))

	ctx := &RenderContext{
		Output:         render.StringOutput(),
		OutputPrefix:   "prefix ",
		OutputContinue: "cont ",
	}

	result := registry.TryRender(ctx, json.RawMessage(`{}`))
	if result {
		t.Error("Expected registry to return false when no renderers match")
	}
	if len(calls) != 2 {
		t.Errorf("Expected both renderers to be tried, got calls: %v", calls)
	}
}

func TestResultRendererFunc_Adapter(t *testing.T) {
	called := false
	var receivedCtx *RenderContext
	var receivedData json.RawMessage

	fn := ResultRendererFunc(func(ctx *RenderContext, data json.RawMessage) bool {
		called = true
		receivedCtx = ctx
		receivedData = data
		return true
	})

	ctx := &RenderContext{
		Output:         render.StringOutput(),
		OutputPrefix:   "test",
		OutputContinue: "cont",
	}
	data := json.RawMessage(`{"key": "value"}`)

	result := fn.TryRender(ctx, data)

	if !called {
		t.Error("Expected function to be called")
	}
	if !result {
		t.Error("Expected function to return true")
	}
	if receivedCtx != ctx {
		t.Error("Expected context to be passed through")
	}
	if string(receivedData) != string(data) {
		t.Errorf("Expected data to be passed through, got %s", string(receivedData))
	}
}

func TestRenderContext_Fields(t *testing.T) {
	out := render.StringOutput()
	ctx := &RenderContext{
		Output:         out,
		OutputPrefix:   "prefix ",
		OutputContinue: "continue ",
	}

	if ctx.Output != out {
		t.Error("Expected Output to match")
	}
	if ctx.OutputPrefix != "prefix " {
		t.Errorf("Expected OutputPrefix to be 'prefix ', got %q", ctx.OutputPrefix)
	}
	if ctx.OutputContinue != "continue " {
		t.Errorf("Expected OutputContinue to be 'continue ', got %q", ctx.OutputContinue)
	}
}
