# viewscreen

A terminal renderer for Claude Code's streaming JSON output. Parses JSONL events from stdin and renders them with markdown formatting, syntax highlighting, and styled output.

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

### Flags

- `-v` - Verbose output (show more details)
- `-no-color` - Disable colored output
- `-usage` - Show token usage in result (default: true)

## Event Types

viewscreen handles the following event types:

- `system` - System messages and configuration
- `assistant` - Assistant responses
- `user` - User input
- `stream_event` - Streaming content deltas
- `result` - Final results with token usage

## License

MIT
