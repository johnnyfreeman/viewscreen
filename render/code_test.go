package render

import (
	"strings"
	"testing"

	"github.com/johnnyfreeman/viewscreen/style"
)

func TestCodeRenderer_NewCodeRenderer(t *testing.T) {
	t.Run("with color", func(t *testing.T) {
		cr := NewCodeRenderer(false)
		if cr == nil {
			t.Fatal("NewCodeRenderer returned nil")
		}
		if cr.noColor {
			t.Error("expected noColor to be false")
		}
		if cr.formatter == nil {
			t.Error("formatter should not be nil")
		}
		if cr.style == nil {
			t.Error("style should not be nil")
		}
	})

	t.Run("without color", func(t *testing.T) {
		cr := NewCodeRenderer(true)
		if cr == nil {
			t.Fatal("NewCodeRenderer returned nil")
		}
		if !cr.noColor {
			t.Error("expected noColor to be true")
		}
		if cr.formatter == nil {
			t.Error("formatter should not be nil even in no-color mode")
		}
	})
}

func TestCodeRenderer_Highlight(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		language string
		noColor  bool
		check    func(t *testing.T, result string)
	}{
		{
			name:     "go code with color",
			code:     `func main() { fmt.Println("hello") }`,
			language: "go",
			noColor:  false,
			check: func(t *testing.T, result string) {
				// With color enabled, result should differ from input (have ANSI codes)
				if result == `func main() { fmt.Println("hello") }` {
					t.Error("expected highlighting to modify the code")
				}
			},
		},
		{
			name:     "go code without color",
			code:     `func main() { fmt.Println("hello") }`,
			language: "go",
			noColor:  true,
			check: func(t *testing.T, result string) {
				// Without color, result should be identical to input
				if result != `func main() { fmt.Println("hello") }` {
					t.Errorf("expected unchanged code, got %q", result)
				}
			},
		},
		{
			name:     "empty language returns unchanged",
			code:     "some code",
			language: "",
			noColor:  false,
			check: func(t *testing.T, result string) {
				if result != "some code" {
					t.Errorf("expected unchanged code for empty language, got %q", result)
				}
			},
		},
		{
			name:     "empty code",
			code:     "",
			language: "go",
			noColor:  false,
			check: func(t *testing.T, result string) {
				if result != "" {
					t.Errorf("expected empty result, got %q", result)
				}
			},
		},
		{
			name:     "python code",
			code:     `def hello():\n    print("world")`,
			language: "python",
			noColor:  false,
			check: func(t *testing.T, result string) {
				// Should contain some transformation
				if result == "" {
					t.Error("result should not be empty")
				}
			},
		},
		{
			name:     "unknown language uses fallback",
			code:     "random text here",
			language: "unknownlang12345",
			noColor:  false,
			check: func(t *testing.T, result string) {
				// Should still return something (fallback lexer)
				if result == "" {
					t.Error("result should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := NewCodeRenderer(tt.noColor)
			result := cr.Highlight(tt.code, tt.language)
			tt.check(t, result)
		})
	}
}

func TestCodeRenderer_HighlightFile(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		filename string
		noColor  bool
		check    func(t *testing.T, result string)
	}{
		{
			name:     "go file with color",
			code:     `package main`,
			filename: "main.go",
			noColor:  false,
			check: func(t *testing.T, result string) {
				// Should be highlighted
				if result == "package main" {
					t.Error("expected code to be highlighted")
				}
			},
		},
		{
			name:     "go file without color",
			code:     `package main`,
			filename: "main.go",
			noColor:  true,
			check: func(t *testing.T, result string) {
				if result != "package main" {
					t.Errorf("expected unchanged code, got %q", result)
				}
			},
		},
		{
			name:     "unknown file extension",
			code:     "some random content",
			filename: "file.unknownext",
			noColor:  false,
			check: func(t *testing.T, result string) {
				// Should return unchanged since lexer can't be determined
				if result != "some random content" {
					t.Errorf("expected unchanged for unknown extension, got %q", result)
				}
			},
		},
		{
			name:     "python file",
			code:     `import sys`,
			filename: "script.py",
			noColor:  false,
			check: func(t *testing.T, result string) {
				if result == "" {
					t.Error("result should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := NewCodeRenderer(tt.noColor)
			result := cr.HighlightFile(tt.code, tt.filename)
			tt.check(t, result)
		})
	}
}

func TestCodeRenderer_HighlightDiff(t *testing.T) {
	diffContent := `--- a/file.go
+++ b/file.go
@@ -1,3 +1,3 @@
 package main
-func old() {}
+func new() {}
`

	t.Run("with color", func(t *testing.T) {
		cr := NewCodeRenderer(false)
		result := cr.HighlightDiff(diffContent)
		// Diff should be highlighted
		if result == diffContent {
			t.Error("expected diff to be highlighted")
		}
	})

	t.Run("without color", func(t *testing.T) {
		cr := NewCodeRenderer(true)
		result := cr.HighlightDiff(diffContent)
		// Should return unchanged
		if result != diffContent {
			t.Errorf("expected unchanged diff, got %q", result)
		}
	})
}

func TestCodeRenderer_HighlightWithBg(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		language string
		bgColor  style.Color
		noColor  bool
		check    func(t *testing.T, result string)
	}{
		{
			name:     "go code with green background",
			code:     `func test() {}`,
			language: "go",
			bgColor:  style.Color("#14532D"),
			noColor:  false,
			check: func(t *testing.T, result string) {
				// Result should contain original text content
				// (bgFormatter applies styling per token)
				if result == "" {
					t.Error("expected non-empty result")
				}
				// Should contain the core tokens
				if !strings.Contains(result, "func") {
					t.Error("expected result to contain 'func' token")
				}
			},
		},
		{
			name:     "no color mode returns unchanged",
			code:     `func test() {}`,
			language: "go",
			bgColor:  style.Color("#14532D"),
			noColor:  true,
			check: func(t *testing.T, result string) {
				if result != "func test() {}" {
					t.Errorf("expected unchanged code, got %q", result)
				}
			},
		},
		{
			name:     "empty language returns unchanged",
			code:     `some code`,
			language: "",
			bgColor:  style.Color("#14532D"),
			noColor:  false,
			check: func(t *testing.T, result string) {
				if result != "some code" {
					t.Errorf("expected unchanged for empty language, got %q", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := NewCodeRenderer(tt.noColor)
			result := cr.HighlightWithBg(tt.code, tt.language, tt.bgColor)
			tt.check(t, result)
		})
	}
}

func TestCodeRenderer_LargeContentSkipped(t *testing.T) {
	// Create content larger than 1MB threshold
	largeContent := strings.Repeat("x", 1024*1024+1)

	t.Run("Highlight skips large content", func(t *testing.T) {
		cr := NewCodeRenderer(false)
		result := cr.Highlight(largeContent, "go")
		if result != largeContent {
			t.Error("expected large content to be returned unchanged")
		}
	})

	t.Run("HighlightFile skips large content", func(t *testing.T) {
		cr := NewCodeRenderer(false)
		result := cr.HighlightFile(largeContent, "main.go")
		if result != largeContent {
			t.Error("expected large content to be returned unchanged")
		}
	})

	t.Run("HighlightWithBg skips large content", func(t *testing.T) {
		cr := NewCodeRenderer(false)
		result := cr.HighlightWithBg(largeContent, "go", style.Color("#000000"))
		if result != largeContent {
			t.Error("expected large content to be returned unchanged")
		}
	})
}

func TestBgFormatter(t *testing.T) {
	// Test the custom bgFormatter type
	t.Run("formats with background color", func(t *testing.T) {
		cr := NewCodeRenderer(false)
		result := cr.HighlightWithBg("test", "text", style.Color("#FF0000"))
		// Should produce some output
		if result == "" {
			t.Error("expected non-empty result")
		}
	})
}
