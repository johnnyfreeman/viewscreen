# Codex JSON Output Types

Documentation for the JSONL events emitted by `codex exec --json` (the
[Codex CLI](https://github.com/openai/codex)).

Codex's stream is structured differently from Claude Code's
[stream-json](../stream-json/index.md). Instead of message events that carry
content blocks, Codex emits a flat sequence of **envelope** events. Three of
those envelopes (`item.started`, `item.updated`, `item.completed`) wrap an
**item** — a self-contained unit of work such as an assistant message, a shell
command, or a file change. The envelope tells you *where* in the lifecycle you
are; the item tells you *what* the work is.

```
envelope (item.started / item.updated / item.completed)
   └── item (agent_message | reasoning | command_execution | ...)
```

## Envelope Types

| Type | Description |
|------|-------------|
| [thread.started](thread.md) | A new conversation thread (session) has begun |
| [turn.started](turn.md#turnstarted) | The model has begun a turn |
| [turn.completed](turn.md#turncompleted) | A turn finished; carries token usage |
| [turn.failed](turn.md#turnfailed) | A turn ended with an error |
| [item.started](item.md) | A work item began |
| [item.updated](item.md) | A work item changed (e.g. streaming output) |
| [item.completed](item.md) | A work item finished |
| [error](error.md) | A stream-level error |

## Item Types

Carried inside the `item.*` envelopes. See [item.md](item.md) for the full
reference.

| Item type | Description |
|-----------|-------------|
| `agent_message` | Assistant response text (rendered as markdown) |
| `reasoning` | Model reasoning / "thinking" summary |
| `command_execution` | A shell command and its output |
| `file_change` | File additions, updates, and deletions |
| `todo_list` | A plan / todo list with per-item completion |
| `mcp_tool_call` | An MCP tool invocation |
| `web_search` | A web search query |
| `error` | An error surfaced as a work item |

## Event Flow

```
thread.started
    |
    v
turn.started
    |
    +--> item.started --> [item.updated]* --> item.completed   (per work item)
    |        (agent_message, reasoning, command_execution, ...)
    |
    v
turn.completed (usage)   |   turn.failed (error)
```

A turn contains zero or more items. Short items (an `agent_message` or
`reasoning`) typically arrive as a single `item.completed`. Long-running items (a
`command_execution`, `file_change`, `mcp_tool_call`, or `web_search`) arrive as
an `item.started` followed by an `item.completed` for the same `id`, and may emit
`item.updated` in between.

**Important:** the same item `id` can appear in more than one envelope. A renderer
should treat the first envelope for an `id` as "create" and later envelopes as
"update", deduplicating any header it has already printed. Viewscreen's
`codex.Renderer` does this with a per-stream `id` set.

## Detecting Codex vs Claude

Viewscreen auto-detects the format per line, with no flag required. A line is
treated as Codex output when its top-level `type` is a known Codex envelope
(`thread.started`, `turn.*`, `item.*`, `error`) or, more generally, when the
`type` contains a dot (`.`) — Claude Code's stream-json types never do. See
`codex.IsEventType` and `events.Parse`.

## Common Properties

Every envelope has a `type`. The set of other fields depends on the envelope:

| Property | Type | Applies to |
|----------|------|------------|
| `type` | `string` | all envelopes |
| `thread_id` | `string` | `thread.started` |
| `usage` | `object` | `turn.completed` |
| `error` | `object` | `turn.failed` |
| `item` | `object` | `item.started`, `item.updated`, `item.completed` |
| `message` | `string` | `error` |

## Capturing a Stream

```bash
codex exec --json --sandbox read-only - <<<"your prompt" > capture.jsonl
```
