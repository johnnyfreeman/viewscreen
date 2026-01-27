package render

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
)

// maxCodeSize is the threshold above which syntax highlighting is skipped.
const maxCodeSize = 1024 * 1024

// bgFormatter is a custom chroma formatter that applies syntax highlighting
// foreground colors while forcing a specific background color per token.
type bgFormatter struct {
	bgColor lipgloss.Color
}

func (f bgFormatter) Format(w io.Writer, style *chroma.Style, it chroma.Iterator) error {
	for token := it(); token != chroma.EOF; token = it() {
		value := strings.TrimRight(token.Value, "\n")

		entry := style.Get(token.Type)
		s := lipgloss.NewStyle().Background(f.bgColor)

		if !entry.IsZero() {
			if entry.Bold == chroma.Yes {
				s = s.Bold(true)
			}
			if entry.Underline == chroma.Yes {
				s = s.Underline(true)
			}
			if entry.Italic == chroma.Yes {
				s = s.Italic(true)
			}
			if entry.Colour.IsSet() {
				s = s.Foreground(lipgloss.Color(entry.Colour.String()))
			}
		}

		if _, err := fmt.Fprint(w, s.Render(value)); err != nil {
			return err
		}
	}
	return nil
}

// CodeRenderer handles syntax highlighting with chroma
type CodeRenderer struct {
	formatter chroma.Formatter
	style     *chroma.Style
	noColor   bool
}

// NewCodeRenderer creates a new code renderer
func NewCodeRenderer(noColor bool) *CodeRenderer {
	var formatter chroma.Formatter
	var style *chroma.Style

	if noColor {
		formatter = formatters.NoOp
		style = styles.Fallback
	} else {
		// Use TrueColor (24-bit) formatter for richer syntax highlighting
		formatter = formatters.Get("terminal16m")
		if formatter == nil {
			formatter = formatters.TTY256
		}
		style = styles.Get("monokai")
		if style == nil {
			style = styles.Fallback
		}
	}

	return &CodeRenderer{
		formatter: formatter,
		style:     style,
		noColor:   noColor,
	}
}

// shouldSkip returns true if highlighting should be skipped for the given code.
func (c *CodeRenderer) shouldSkip(code string) bool {
	return c.noColor || len(code) > maxCodeSize
}

// formatWith tokenizes code using the lexer and formats it with the given formatter.
// Returns the original code if tokenization or formatting fails.
func (c *CodeRenderer) formatWith(code string, lexer chroma.Lexer, formatter chroma.Formatter) string {
	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, c.style, iterator); err != nil {
		return code
	}

	return buf.String()
}

// Highlight highlights code with the given language
func (c *CodeRenderer) Highlight(code, language string) string {
	if c.shouldSkip(code) || language == "" {
		return code
	}

	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	return c.formatWith(code, lexer, c.formatter)
}

// HighlightFile highlights code, detecting language from filename
func (c *CodeRenderer) HighlightFile(code, filename string) string {
	if c.shouldSkip(code) {
		return code
	}

	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		return code
	}

	return c.formatWith(code, lexer, c.formatter)
}

// HighlightDiff highlights diff/patch content
func (c *CodeRenderer) HighlightDiff(diff string) string {
	return c.Highlight(diff, "diff")
}

// HighlightWithBg highlights code with a forced background color per token.
// This allows syntax highlighting colors to show through while maintaining
// a consistent background (e.g., for diff added/removed lines).
func (c *CodeRenderer) HighlightWithBg(code, language string, bgColor lipgloss.Color) string {
	if c.shouldSkip(code) || language == "" {
		return code
	}

	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	return c.formatWith(code, lexer, bgFormatter{bgColor: bgColor})
}

// DetectLanguageFromPath returns a language hint from a file path
func DetectLanguageFromPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".jsx":
		return "jsx"
	case ".tsx":
		return "tsx"
	case ".rb":
		return "ruby"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".hpp":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".php":
		return "php"
	case ".sh", ".bash":
		return "bash"
	case ".sql":
		return "sql"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".xml":
		return "xml"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".scss":
		return "scss"
	case ".md":
		return "markdown"
	default:
		return ""
	}
}
