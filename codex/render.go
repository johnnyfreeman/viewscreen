package codex

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/johnnyfreeman/viewscreen/config"
	"github.com/johnnyfreeman/viewscreen/render"
	"github.com/johnnyfreeman/viewscreen/style"
	"github.com/johnnyfreeman/viewscreen/terminal"
	"github.com/johnnyfreeman/viewscreen/textutil"
)

const (
	// commandOutputLinesVeryVerbose mirrors Claude read-output expansion at -vv.
	commandOutputLinesVeryVerbose = 5

	// commandOutputLinesMaxVerbose mirrors Claude read-output expansion at -vvv.
	commandOutputLinesMaxVerbose = 10
)

// argWidth is the maximum width of a header argument (command/path) before it
// is truncated. Matches the tools header renderer.
const argWidth = 80

// Renderer turns codex events into styled terminal output. It is stateful: a
// single Renderer instance must be used for the lifetime of one stream so that
// it can deduplicate items that codex reports more than once (an item.started
// followed by an item.completed for the same id).
type Renderer struct {
	md         *render.MarkdownRenderer
	config     config.Provider
	headerSeen map[string]bool
	width      int
}

// RendererOption configures a Renderer.
type RendererOption func(*Renderer)

// WithConfigProvider sets a custom config provider.
func WithConfigProvider(cp config.Provider) RendererOption {
	return func(r *Renderer) {
		r.config = cp
	}
}

// WithMarkdownRenderer sets a custom markdown renderer (used in tests).
func WithMarkdownRenderer(md *render.MarkdownRenderer) RendererOption {
	return func(r *Renderer) {
		r.md = md
	}
}

// NewRenderer creates a Renderer with default dependencies.
func NewRenderer(opts ...RendererOption) *Renderer {
	cfg := config.Get()
	r := &Renderer{
		config:     cfg,
		md:         render.NewMarkdownRenderer(cfg.NoColor(), terminal.Width()),
		headerSeen: make(map[string]bool),
		width:      terminal.Width(),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// SetWidth updates the word-wrap width of markdown and command output.
func (r *Renderer) SetWidth(width int) {
	r.md.SetWidth(width)
	if width > 0 {
		r.width = width
	}
}

// Render renders a single codex event to a string. It returns "" for events
// that produce no output (e.g. turn.started, or a duplicate item).
func (r *Renderer) Render(event Event) string {
	switch event.Type {
	case TypeThreadStarted:
		return r.renderThreadStarted(event)
	case TypeTurnCompleted:
		return r.renderTurnCompleted(event)
	case TypeTurnFailed:
		return r.renderTurnFailed(event)
	case TypeError:
		return r.renderError(event.Message)
	case TypeItemStarted:
		return r.renderItem("started", event.Item)
	case TypeItemUpdated:
		return r.renderItem("updated", event.Item)
	case TypeItemCompleted:
		return r.renderItem("completed", event.Item)
	default: // turn.started and any unknown envelope
		return ""
	}
}

// once reports whether id is being seen for the first time, marking it seen.
func (r *Renderer) once(id string) bool {
	if id == "" {
		return true // unidentified items always render
	}
	if r.headerSeen[id] {
		return false
	}
	r.headerSeen[id] = true
	return true
}

func (r *Renderer) renderThreadStarted(event Event) string {
	out := render.StringOutput()
	fmt.Fprintln(out, style.BulletHeader("Codex Session"))
	if event.ThreadID != "" {
		fmt.Fprintf(out, "%s%s %s\n", style.OutputPrefix, style.MutedText("Thread:"), event.ThreadID)
	}
	fmt.Fprintln(out)
	return out.String()
}

func (r *Renderer) renderTurnCompleted(event Event) string {
	out := render.StringOutput()
	fmt.Fprintln(out)
	fmt.Fprintln(out, style.BulletSuccessHeader("Turn Complete"))
	if event.Usage != nil && r.config.ShowUsage() {
		u := event.Usage
		fmt.Fprintf(out, "%s%s in=%d out=%d (cached=%d reasoning=%d)\n",
			style.OutputPrefix, style.MutedText("Tokens:"),
			u.InputTokens, u.OutputTokens, u.CachedInputTokens, u.ReasoningOutputTokens)
	}
	return out.String()
}

func (r *Renderer) renderTurnFailed(event Event) string {
	msg := ""
	if event.Error != nil {
		msg = event.Error.Message
	}
	out := render.StringOutput()
	fmt.Fprintln(out)
	fmt.Fprintln(out, style.BulletErrorHeader("Turn Failed"))
	if msg != "" {
		fmt.Fprintf(out, "%s%s\n", style.OutputPrefix, style.ErrorText(msg))
	}
	return out.String()
}

func (r *Renderer) renderError(message string) string {
	out := render.StringOutput()
	fmt.Fprintln(out, style.BulletErrorHeader("Error"))
	if message != "" {
		fmt.Fprintf(out, "%s%s\n", style.OutputPrefix, style.ErrorText(message))
	}
	return out.String()
}

func (r *Renderer) renderItem(phase string, item *Item) string {
	if item == nil {
		return ""
	}
	switch item.Type {
	case ItemAgentMessage:
		if item.Text == "" || !r.once(item.ID) {
			return ""
		}
		return r.md.Render(item.Text)
	case ItemReasoning:
		if item.Text == "" || !r.once(item.ID) {
			return ""
		}
		return r.renderReasoning(item.Text)
	case ItemCommandExecution:
		return r.renderCommand(phase, item)
	case ItemFileChange:
		if !r.once(item.ID) {
			return ""
		}
		return r.renderFileChange(item)
	case ItemTodoList:
		if !r.once(item.ID) {
			return ""
		}
		return r.renderTodoList(item)
	case ItemMCPToolCall:
		return r.renderMCPToolCall(phase, item)
	case ItemWebSearch:
		if !r.once(item.ID) {
			return ""
		}
		return r.renderWebSearch(item)
	case ItemError:
		if !r.once(item.ID) {
			return ""
		}
		return r.renderError(firstNonEmpty(item.Message, item.Text))
	default:
		return r.renderUnknownItem(phase, item)
	}
}

func (r *Renderer) renderReasoning(text string) string {
	out := render.StringOutput()
	fmt.Fprintln(out, style.MutedText(style.Bullet+" Thinking"))
	fmt.Fprint(out, r.md.RenderMuted(text))
	return out.String()
}

// renderCommand renders a shell command. The header (the command itself) is
// printed the first time the item is seen; the output and exit status are
// printed when the item completes.
func (r *Renderer) renderCommand(phase string, item *Item) string {
	out := render.StringOutput()
	if r.once(item.ID) {
		fmt.Fprint(out, header("Shell", ShellCommand(item.Command)))
	}
	if phase == "completed" {
		r.writeCommandOutput(out, item)
	}
	return out.String()
}

func (r *Renderer) writeCommandOutput(out *render.Output, item *Item) {
	pw := textutil.NewPrefixedWriter(out, style.OutputPrefix, style.OutputContinue)
	body := strings.TrimRight(item.AggregatedOutput, "\n")
	if body != "" {
		for _, line := range r.commandOutputLines(strings.Split(body, "\n")) {
			pw.WriteLine(r.truncateCommandOutputLine(line))
		}
	}
	if item.ExitCode != nil && *item.ExitCode != 0 {
		pw.WriteLine(style.ErrorText(fmt.Sprintf("exited with code %d", *item.ExitCode)))
	} else if item.Status == "failed" {
		pw.WriteLine(style.ErrorText("failed"))
	} else if body == "" {
		pw.WriteLine(style.MutedText("(no output)"))
	}
}

func (r *Renderer) renderFileChange(item *Item) string {
	out := render.StringOutput()
	fmt.Fprint(out, header("Edit", FileChangeSummary(item.Changes)))
	if len(item.Changes) > 1 {
		pw := textutil.NewPrefixedWriter(out, style.OutputPrefix, style.OutputContinue)
		for _, c := range item.Changes {
			pw.WriteLinef("%s %s", changeKindLabel(c.Kind), c.Path)
		}
	}
	return out.String()
}

func (r *Renderer) renderTodoList(item *Item) string {
	out := render.StringOutput()
	fmt.Fprint(out, header("Update Todos", ""))
	pw := textutil.NewPrefixedWriter(out, style.OutputPrefix, style.OutputContinue)
	for _, todo := range item.Items {
		if todo.Completed {
			pw.WriteLinef("%s %s", style.SuccessText("✓"), style.MutedText(todo.Text))
		} else {
			pw.WriteLinef("%s %s", style.MutedText("○"), todo.Text)
		}
	}
	return out.String()
}

func (r *Renderer) renderMCPToolCall(phase string, item *Item) string {
	out := render.StringOutput()
	if r.once(item.ID) {
		fmt.Fprint(out, header(MCPLabel(item), ""))
	}
	if phase == "completed" && item.Status == "failed" {
		pw := textutil.NewPrefixedWriter(out, style.OutputPrefix, style.OutputContinue)
		pw.WriteLine(style.ErrorText("failed"))
	}
	return out.String()
}

func (r *Renderer) renderWebSearch(item *Item) string {
	return header("Web Search", item.Query)
}

func (r *Renderer) renderUnknownItem(phase string, item *Item) string {
	out := render.StringOutput()
	arg := firstNonEmpty(item.Command, item.Query, item.Status)
	if r.once(item.ID) {
		fmt.Fprint(out, header(unknownItemLabel(item.Type), arg))
	}
	if phase == "completed" || phase == "updated" {
		r.writeUnknownItemDetails(out, item, arg)
	}
	return out.String()
}

func (r *Renderer) writeUnknownItemDetails(out *render.Output, item *Item, headerArg string) {
	pw := textutil.NewPrefixedWriter(out, style.OutputPrefix, style.OutputContinue)
	detail := firstNonEmpty(item.Message, item.Text)
	if detail != "" && detail != headerArg {
		for _, line := range strings.Split(strings.TrimRight(detail, "\n"), "\n") {
			pw.WriteLine(line)
		}
	}
	if item.Status == "failed" {
		pw.WriteLine(style.ErrorText("failed"))
	}
	if r.config.IsVerbose() && len(item.Raw) > 0 {
		if compact := compactRawItem(item.Raw); compact != "" {
			pw.WriteLine(style.MutedText(compact))
		}
	}
}

// commandOutputLines applies the same read-output expansion policy as Claude:
// default and -v show a summary, -vv shows 5 lines, and -vvv shows 10.
func (r *Renderer) commandOutputLines(lines []string) []string {
	level := r.config.GetVerboseLevel()
	maxLines := 0
	switch {
	case level >= 3:
		maxLines = commandOutputLinesMaxVerbose
	case level >= 2:
		maxLines = commandOutputLinesVeryVerbose
	}

	if maxLines == 0 {
		return []string{style.MutedText(fmt.Sprintf("%d lines", len(lines)))}
	}
	if len(lines) <= maxLines {
		return lines
	}

	hidden := len(lines) - maxLines
	result := append([]string(nil), lines[:maxLines]...)
	return append(result, style.MutedText(textutil.TruncationIndicator(hidden)))
}

func (r *Renderer) truncateCommandOutputLine(line string) string {
	limit := max(20, r.width-ansi.StringWidth(style.OutputPrefix))
	if ansi.StringWidth(line) <= limit {
		return line
	}

	originalWidth := ansi.StringWidth(line)
	omitted := originalWidth - limit
	var tail string
	for {
		tail = style.MutedText(fmt.Sprintf("… (+%d more chars)", omitted))
		next := originalWidth - max(0, limit-ansi.StringWidth(tail))
		if next == omitted {
			break
		}
		omitted = next
	}
	return ansi.Truncate(line, limit, tail)
}

// header renders a tool-style header line: a gradient bullet + label, followed
// by an optional muted argument. It mirrors the layout of tools.HeaderRenderer
// so codex output is visually consistent with the Claude renderers.
func header(label, arg string) string {
	var b strings.Builder
	b.WriteString(style.ApplyThemeBoldGradient(style.Bullet + " " + label))
	if arg != "" {
		b.WriteString(" " + style.MutedText(truncate(arg, argWidth)))
	}
	b.WriteString("\n")
	return b.String()
}

// ShellCommand extracts the inner command from codex's shell invocation,
// which wraps commands as "<shell> -lc <script>". It is exported so callers
// that need the human-readable command (e.g. the live spinner label) share the
// same unwrapping the renderer uses for command headers.
func ShellCommand(cmd string) string {
	for _, sep := range []string{" -lc ", " -c "} {
		if i := strings.Index(cmd, sep); i != -1 {
			return unquote(strings.TrimSpace(cmd[i+len(sep):]))
		}
	}
	return cmd
}

func unquote(s string) string {
	if len(s) >= 2 {
		first, last := s[0], s[len(s)-1]
		if (first == '\'' && last == '\'') || (first == '"' && last == '"') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// FileChangeSummary returns the human-readable argument for a file_change item:
// the single path when only one file is touched, or "N files" otherwise. It is
// exported so the live spinner can label an in-flight file change the same way
// the inline header does (mirroring ShellCommand and MCPLabel).
func FileChangeSummary(changes []FileChange) string {
	switch len(changes) {
	case 0:
		return ""
	case 1:
		return changes[0].Path
	default:
		return fmt.Sprintf("%d files", len(changes))
	}
}

func changeKindLabel(kind string) string {
	switch kind {
	case "add":
		return style.SuccessText("add")
	case "delete":
		return style.ErrorText("delete")
	default:
		return style.WarningText("update")
	}
}

// MCPLabel returns a "server.tool" label for an mcp_tool_call item, falling
// back to whichever of server/tool is present. It is exported so the live
// spinner can label an in-flight MCP call the same way the header does.
func MCPLabel(item *Item) string {
	switch {
	case item.Server != "" && item.Tool != "":
		return item.Server + "." + item.Tool
	case item.Tool != "":
		return item.Tool
	case item.Server != "":
		return item.Server
	default:
		return "MCP Tool"
	}
}

func unknownItemLabel(itemType string) string {
	if itemType == "" {
		return "Codex Item"
	}
	parts := strings.Fields(strings.NewReplacer("_", " ", "-", " ").Replace(itemType))
	for i, part := range parts {
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func compactRawItem(raw json.RawMessage) string {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
