package render

import (
	"strings"
	"testing"
)

func TestNewMarkdownRenderer(t *testing.T) {
	tests := []struct {
		name    string
		noColor bool
		width   int
	}{
		{"with color and default width", false, 80},
		{"without color", true, 80},
		{"narrow width", false, 40},
		{"wide width", false, 120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMarkdownRenderer(tt.noColor, tt.width)
			if mr == nil {
				t.Fatal("NewMarkdownRenderer returned nil")
			}
			if mr.noColor != tt.noColor {
				t.Errorf("noColor = %v, want %v", mr.noColor, tt.noColor)
			}
			if mr.full == nil {
				t.Error("full renderer should not be nil")
			}
			if mr.muted == nil {
				t.Error("muted renderer should not be nil")
			}
		})
	}
}

func TestMarkdownRenderer_Render(t *testing.T) {
	tests := []struct {
		name    string
		content string
		noColor bool
		check   func(t *testing.T, result string)
	}{
		{
			name:    "simple text",
			content: "Hello, world!",
			noColor: true,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "Hello, world!") {
					t.Errorf("result should contain input text, got %q", result)
				}
			},
		},
		{
			name:    "heading",
			content: "# Title",
			noColor: true,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "Title") {
					t.Errorf("result should contain heading text, got %q", result)
				}
			},
		},
		{
			name:    "code block",
			content: "```go\nfunc main() {}\n```",
			noColor: true,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "func main()") {
					t.Errorf("result should contain code, got %q", result)
				}
			},
		},
		{
			name:    "bullet list",
			content: "- item 1\n- item 2\n- item 3",
			noColor: true,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "item 1") {
					t.Errorf("result should contain list items, got %q", result)
				}
			},
		},
		{
			name:    "empty content",
			content: "",
			noColor: true,
			check: func(t *testing.T, result string) {
				// Empty content should still return a newline
				if result != "\n" {
					t.Errorf("expected single newline for empty content, got %q", result)
				}
			},
		},
		{
			name:    "with color renders differently",
			content: "**bold text**",
			noColor: false,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "bold text") {
					t.Errorf("result should contain bold text, got %q", result)
				}
			},
		},
		{
			name:    "inline code",
			content: "Use `fmt.Println()` to print",
			noColor: true,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "fmt.Println()") {
					t.Errorf("result should contain inline code, got %q", result)
				}
			},
		},
		{
			name:    "link",
			content: "Check [this link](https://example.com)",
			noColor: true,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "this link") {
					t.Errorf("result should contain link text, got %q", result)
				}
			},
		},
		{
			name:    "result ends with newline",
			content: "Some content",
			noColor: true,
			check: func(t *testing.T, result string) {
				if !strings.HasSuffix(result, "\n") {
					t.Errorf("result should end with newline, got %q", result)
				}
			},
		},
		{
			name:    "result has exactly one trailing newline",
			content: "Content\n\n\n",
			noColor: true,
			check: func(t *testing.T, result string) {
				if strings.HasSuffix(result, "\n\n") {
					t.Errorf("result should have exactly one trailing newline, got %q", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMarkdownRenderer(tt.noColor, 80)
			result := mr.Render(tt.content)
			tt.check(t, result)
		})
	}
}

func TestMarkdownRenderer_RenderMuted(t *testing.T) {
	tests := []struct {
		name    string
		content string
		noColor bool
		check   func(t *testing.T, result string)
	}{
		{
			name:    "simple text muted",
			content: "Thinking...",
			noColor: true,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "Thinking...") {
					t.Errorf("result should contain input text, got %q", result)
				}
			},
		},
		{
			name:    "muted heading",
			content: "# Thinking",
			noColor: true,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "Thinking") {
					t.Errorf("result should contain heading text, got %q", result)
				}
			},
		},
		{
			name:    "muted code block",
			content: "```\nthinking about code\n```",
			noColor: true,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "thinking about code") {
					t.Errorf("result should contain code, got %q", result)
				}
			},
		},
		{
			name:    "empty muted content",
			content: "",
			noColor: true,
			check: func(t *testing.T, result string) {
				// Empty content should still return a newline
				if result != "\n" {
					t.Errorf("expected single newline for empty content, got %q", result)
				}
			},
		},
		{
			name:    "muted with color",
			content: "Some muted text",
			noColor: false,
			check: func(t *testing.T, result string) {
				// With ANSI codes, words may be split, so check for individual words
				if !strings.Contains(result, "muted") {
					t.Errorf("result should contain 'muted', got %q", result)
				}
			},
		},
		{
			name:    "muted result ends with newline",
			content: "Muted content",
			noColor: true,
			check: func(t *testing.T, result string) {
				if !strings.HasSuffix(result, "\n") {
					t.Errorf("result should end with newline, got %q", result)
				}
			},
		},
		{
			name:    "muted result has exactly one trailing newline",
			content: "Content\n\n\n",
			noColor: true,
			check: func(t *testing.T, result string) {
				if strings.HasSuffix(result, "\n\n") {
					t.Errorf("result should have exactly one trailing newline, got %q", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := NewMarkdownRenderer(tt.noColor, 80)
			result := mr.RenderMuted(tt.content)
			tt.check(t, result)
		})
	}
}

func TestMarkdownRenderer_NilRenderers(t *testing.T) {
	// Test that methods handle nil renderers gracefully
	t.Run("Render with nil full renderer", func(t *testing.T) {
		mr := &MarkdownRenderer{
			full:  nil,
			muted: nil,
		}
		result := mr.Render("test content")
		if result != "test content" {
			t.Errorf("expected unchanged content when full renderer is nil, got %q", result)
		}
	})

	t.Run("RenderMuted with nil muted renderer", func(t *testing.T) {
		mr := &MarkdownRenderer{
			full:  nil,
			muted: nil,
		}
		result := mr.RenderMuted("test content")
		if result != "test content" {
			t.Errorf("expected unchanged content when muted renderer is nil, got %q", result)
		}
	})
}

func TestMarkdownRenderer_NoColorSameRenderers(t *testing.T) {
	// In no-color mode, full and muted should use the same renderer
	mr := NewMarkdownRenderer(true, 80)

	fullResult := mr.Render("**test**")
	mutedResult := mr.RenderMuted("**test**")

	// Both should produce the same output in no-color mode
	if fullResult != mutedResult {
		t.Errorf("in no-color mode, Render and RenderMuted should produce same output\nRender: %q\nRenderMuted: %q", fullResult, mutedResult)
	}
}

func TestMarkdownRenderer_WidthRespected(t *testing.T) {
	// Test that narrow width wraps long lines
	longLine := "This is a very long line that should be wrapped when using a narrow terminal width setting in the markdown renderer"

	narrowMR := NewMarkdownRenderer(true, 40)
	narrowResult := narrowMR.Render(longLine)

	wideMR := NewMarkdownRenderer(true, 200)
	wideResult := wideMR.Render(longLine)

	// Narrow result should have more lines due to wrapping
	narrowLines := strings.Count(narrowResult, "\n")
	wideLines := strings.Count(wideResult, "\n")

	if narrowLines <= wideLines {
		t.Errorf("narrow width should produce more lines than wide width\nnarrow lines: %d, wide lines: %d", narrowLines, wideLines)
	}
}

func TestGetMutedStyle(t *testing.T) {
	// Test that getMutedStyle returns a valid style config
	style := getMutedStyle()

	// Check that the style has the expected muted color
	expectedMuted := "#71717A"
	if style.Document.Color == nil || *style.Document.Color != expectedMuted {
		t.Errorf("Document color should be %q", expectedMuted)
	}

	// Check heading is bold
	if style.Heading.Bold == nil || !*style.Heading.Bold {
		t.Error("Heading should be bold")
	}

	// Check that code block has subtle color
	expectedSubtle := "#52525B"
	if style.Code.Color == nil || *style.Code.Color != expectedSubtle {
		t.Errorf("Code color should be %q", expectedSubtle)
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("stringPtr", func(t *testing.T) {
		s := "test"
		ptr := stringPtr(s)
		if ptr == nil {
			t.Fatal("stringPtr returned nil")
		}
		if *ptr != s {
			t.Errorf("*stringPtr(%q) = %q, want %q", s, *ptr, s)
		}
	})

	t.Run("boolPtr", func(t *testing.T) {
		b := true
		ptr := boolPtr(b)
		if ptr == nil {
			t.Fatal("boolPtr returned nil")
		}
		if *ptr != b {
			t.Errorf("*boolPtr(%v) = %v, want %v", b, *ptr, b)
		}
	})

	t.Run("uintPtr", func(t *testing.T) {
		u := uint(42)
		ptr := uintPtr(u)
		if ptr == nil {
			t.Fatal("uintPtr returned nil")
		}
		if *ptr != u {
			t.Errorf("*uintPtr(%d) = %d, want %d", u, *ptr, u)
		}
	})
}
