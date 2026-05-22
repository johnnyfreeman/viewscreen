# viewscreen

A terminal renderer for AI coding agents' streaming JSON output. Parses JSONL events from stdin and renders them with markdown formatting, syntax highlighting, and styled output. It understands both [Claude Code](https://claude.com/claude-code) (`--output-format stream-json`) and the [Codex CLI](https://github.com/openai/codex) (`codex exec --json`) event streams, auto-detecting the format per line. The TUI brands itself for whichever agent produced the stream — Codex streams show "codex" branding, Claude streams show "claude" — detected automatically (and seeded from `-agent` in prompt mode).

## Installation

```bash
go install github.com/johnnyfreeman/viewscreen@latest
```

Or build from source:

```bash
git clone https://github.com/johnnyfreeman/viewscreen.git
cd viewscreen
go build
```

## Usage

Pipe Claude Code's JSON output to viewscreen:

```bash
claude --output-format stream-json | viewscreen
```

Or pipe Codex CLI's JSON output to viewscreen:

```bash
codex exec --json "your prompt" | viewscreen
```

### Launching an agent directly

Instead of piping, viewscreen can spawn the agent for you. Pass a prompt as an
argument and viewscreen runs the agent in streaming-JSON mode and renders its
output live (in the TUI when stdout is a terminal):

```bash
# Spawn Claude Code (default)
viewscreen "explain this codebase"

# Spawn the Codex CLI
viewscreen -agent codex "explain this codebase"
```

With `-p`, the prompt is read from stdin instead of an argument:

```bash
echo "explain this codebase" | viewscreen -agent codex -p
```

### Flags

- `-v` - Verbose output (show more details)
- `-no-color` - Disable colored output
- `-usage` - Show token usage in result (default: true)
- `-p` - Treat stdin as a prompt (not a JSON stream)
- `-agent` - Agent to spawn in prompt mode: `claude` (default) or `codex`

## Event Types

### Claude Code (stream-json)

See [docs/stream-json](docs/stream-json/index.md) for the full reference.

- `system` - System messages and configuration
- `assistant` - Assistant responses
- `user` - User input
- `stream_event` - Streaming content deltas
- `result` - Final results with token usage

### Codex CLI (`codex exec --json`)

See [docs/codex-json](docs/codex-json/index.md) for the full reference.

- `thread.started` / `turn.started` / `turn.completed` / `turn.failed` - Session and turn envelopes
- `item.started` / `item.updated` / `item.completed` - Work items, including:
  - `agent_message` - Assistant responses (rendered as markdown)
  - `reasoning` - Model reasoning summaries
  - `command_execution` - Shell commands and their output
  - `file_change` - File additions, updates, and deletions
  - `todo_list` - Plan / todo updates
  - `mcp_tool_call` - MCP tool invocations
  - `web_search` - Web searches
- `error` - Stream errors

## License

MIT
