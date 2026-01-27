package render

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/lucasb-eyer/go-colorful"
)

// maxCodeSize is the threshold above which syntax highlighting is skipped.
const maxCodeSize = 1024 * 1024

// bgFormatter is a custom chroma formatter that applies syntax highlighting
// foreground colors while forcing a specific background color per token.
// Uses Ultraviolet for proper style/content separation - this ensures that
// syntax-highlighted code with backgrounds can be safely composed with other
// styles without escape sequence conflicts.
type bgFormatter struct {
	bgColor style.Color
}

func (f bgFormatter) Format(w io.Writer, style *chroma.Style, it chroma.Iterator) error {
	// Convert background color once for all tokens
	bg := hexToRGBA(string(f.bgColor))

	for token := it(); token != chroma.EOF; token = it() {
		value := strings.TrimRight(token.Value, "\n")

		entry := style.Get(token.Type)

		// Build Ultraviolet style with background always set
		uvStyle := &uv.Style{
			Bg: bg,
		}

		if !entry.IsZero() {
			// Accumulate text attributes
			var attrs uint8
			if entry.Bold == chroma.Yes {
				attrs |= uv.AttrBold
			}
			if entry.Italic == chroma.Yes {
				attrs |= uv.AttrItalic
			}
			uvStyle.Attrs = attrs

			// Set underline via the Underline field (not as an attribute)
			if entry.Underline == chroma.Yes {
				uvStyle.Underline = uv.UnderlineSingle
			}

			// Set foreground color if specified
			if entry.Colour.IsSet() {
				uvStyle.Fg = hexToRGBA(entry.Colour.String())
			}
		}

		if _, err := fmt.Fprint(w, uvStyle.Styled(value)); err != nil {
			return err
		}
	}
	return nil
}

// hexToRGBA converts a hex color string to color.RGBA.
func hexToRGBA(hex string) color.RGBA {
	c, err := colorful.Hex(hex)
	if err != nil {
		return color.RGBA{}
	}
	return color.RGBA{
		R: uint8(c.R * 255),
		G: uint8(c.G * 255),
		B: uint8(c.B * 255),
		A: 255,
	}
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
func (c *CodeRenderer) HighlightWithBg(code, language string, bgColor style.Color) string {
	if c.shouldSkip(code) || language == "" {
		return code
	}

	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	return c.formatWith(code, lexer, bgFormatter{bgColor: bgColor})
}

// HighlightFileWithBg highlights code with a background color, detecting language from filename.
// Returns the original code if no lexer can be determined.
func (c *CodeRenderer) HighlightFileWithBg(code, filename string, bgColor style.Color) string {
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

	return c.formatWith(code, lexer, bgFormatter{bgColor: bgColor})
}
