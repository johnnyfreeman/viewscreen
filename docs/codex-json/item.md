# item.started / item.updated / item.completed

These three envelopes wrap a single **item** — a unit of work the model performs
during a turn. The envelope's `type` marks the item's lifecycle phase; the nested
`item` object describes the work itself.

## Envelope Properties

| Property | Type | Description |
|----------|------|-------------|
| `type` | `"item.started"` \| `"item.updated"` \| `"item.completed"` | Lifecycle phase |
| `item` | `object` | The work item (see below) |

## Lifecycle

- **`item.started`** — the item began. For long-running items (shell commands,
  file changes) this carries the initial state (e.g. `status: "in_progress"`,
  `exit_code: null`).
- **`item.updated`** — the item changed while still running. Optional; not all
  item types emit it.
- **`item.completed`** — the item finished, carrying its final state (output,
  exit code, completion status).

Short items such as `agent_message` and `reasoning` usually arrive as a single
`item.completed` with no preceding `item.started`. Long-running items arrive as
`item.started` … `item.completed` sharing the same `id`.

> **Deduplication:** because an item's `id` recurs across phases, a renderer must
> avoid printing a header twice. Viewscreen prints an item's header on the first
> phase it sees for an `id`, then only appends new detail (output, exit status)
> on later phases. See `codex.Renderer.once`.

## Common Item Properties

| Property | Type | Description |
|----------|------|-------------|
| `id` | `string` | Stable identifier across the item's phases (e.g. `"item_3"`) |
| `type` | `string` | Item type (see the sections below) |
| `status` | `string` | `"in_progress"`, `"completed"`, or `"failed"` (long-running items) |

Unknown item types are rendered defensively: viewscreen keeps the raw JSON of
every item so a newly added type still shows a sensible header and any text,
message, or status it carries (`Item.UnmarshalJSON`,
`Renderer.renderUnknownItem`).

---

## agent_message

Assistant response text. Rendered as markdown.

| Property | Type | Description |
|----------|------|-------------|
| `id` | `string` | Item id |
| `type` | `"agent_message"` | Item type |
| `text` | `string` | The message body (markdown) |

```json
{
  "type": "item.completed",
  "item": {"id": "item_0", "type": "agent_message", "text": "hello"}
}
```

## reasoning

A summary of the model's reasoning ("thinking"). Rendered muted, under a
"Thinking" header.

| Property | Type | Description |
|----------|------|-------------|
| `id` | `string` | Item id |
| `type` | `"reasoning"` | Item type |
| `text` | `string` | The reasoning summary (markdown) |

```json
{
  "type": "item.completed",
  "item": {
    "id": "item_0",
    "type": "reasoning",
    "text": "**Creating a simple TODO plan**\n\nI need to respond to the user by creating a short TODO plan for adding tests..."
  }
}
```

## command_execution

A shell command and its captured output. Codex wraps commands as
`<shell> -lc <script>` (e.g. `/usr/bin/zsh -lc 'git status'`); viewscreen unwraps
the inner command for display (`codex.ShellCommand`).

| Property | Type | Description |
|----------|------|-------------|
| `id` | `string` | Item id |
| `type` | `"command_execution"` | Item type |
| `command` | `string` | The full shell invocation |
| `aggregated_output` | `string` | Combined stdout/stderr captured so far |
| `exit_code` | `number` \| `null` | Process exit code; `null` while running |
| `status` | `string` | `"in_progress"`, `"completed"`, or `"failed"` |

```json
{
  "type": "item.started",
  "item": {
    "id": "item_1",
    "type": "command_execution",
    "command": "/usr/bin/zsh -lc ls",
    "aggregated_output": "",
    "exit_code": null,
    "status": "in_progress"
  }
}
```

```json
{
  "type": "item.completed",
  "item": {
    "id": "item_1",
    "type": "command_execution",
    "command": "/usr/bin/zsh -lc ls",
    "aggregated_output": "foo.txt\n",
    "exit_code": 0,
    "status": "completed"
  }
}
```

Viewscreen prints the command header on the first phase, then on completion
prints the (line-capped) output and a failure note when `exit_code` is non-zero
or `status` is `"failed"`.

## file_change

One or more files added, updated, or deleted in a single edit.

| Property | Type | Description |
|----------|------|-------------|
| `id` | `string` | Item id |
| `type` | `"file_change"` | Item type |
| `changes` | `object[]` | The files touched |
| `status` | `string` | `"in_progress"`, `"completed"`, or `"failed"` |

### changes Array Items

| Property | Type | Description |
|----------|------|-------------|
| `path` | `string` | Absolute path of the file |
| `kind` | `string` | `"add"`, `"update"`, or `"delete"` |

```json
{
  "type": "item.completed",
  "item": {
    "id": "item_22",
    "type": "file_change",
    "changes": [
      {"path": "/home/user/project/main.go", "kind": "update"}
    ],
    "status": "completed"
  }
}
```

When a single file is touched, viewscreen labels the header with its path; when
several are touched it shows `N files` and lists each change underneath
(`codex.FileChangeSummary`).

## todo_list

A plan / todo list. Codex re-emits the whole list on each update, with each
item's `completed` flag reflecting current progress.

| Property | Type | Description |
|----------|------|-------------|
| `id` | `string` | Item id |
| `type` | `"todo_list"` | Item type |
| `items` | `object[]` | The todo entries |

### items Array Items

| Property | Type | Description |
|----------|------|-------------|
| `text` | `string` | The todo description |
| `completed` | `boolean` | Whether the entry is done |

```json
{
  "type": "item.completed",
  "item": {
    "id": "item_1",
    "type": "todo_list",
    "items": [
      {"text": "Add focused unit tests", "completed": true},
      {"text": "Run the relevant test command and fix failures", "completed": false}
    ]
  }
}
```

Codex reports only a boolean per entry (no explicit "in progress" marker), so
viewscreen maps each entry onto the shared todo model as either `completed` or
`pending`, and refreshes the sidebar task list on every update.

## mcp_tool_call

An invocation of an MCP (Model Context Protocol) tool.

| Property | Type | Description |
|----------|------|-------------|
| `id` | `string` | Item id |
| `type` | `"mcp_tool_call"` | Item type |
| `server` | `string` | MCP server name |
| `tool` | `string` | Tool name on that server |
| `status` | `string` | `"in_progress"`, `"completed"`, or `"failed"` |

```json
{
  "type": "item.completed",
  "item": {
    "id": "item_4",
    "type": "mcp_tool_call",
    "server": "github",
    "tool": "list_issues",
    "status": "completed"
  }
}
```

Viewscreen labels the header `server.tool` (falling back to whichever is present)
via `codex.MCPLabel`, and notes a failure on completion when `status` is
`"failed"`.

## web_search

A web search performed by the model.

| Property | Type | Description |
|----------|------|-------------|
| `id` | `string` | Item id |
| `type` | `"web_search"` | Item type |
| `query` | `string` | The search query |

```json
{
  "type": "item.completed",
  "item": {"id": "item_7", "type": "web_search", "query": "golang context cancellation"}
}
```

## error

An error surfaced as a work item (as opposed to the stream-level
[`error`](error.md) envelope).

| Property | Type | Description |
|----------|------|-------------|
| `id` | `string` | Item id |
| `type` | `"error"` | Item type |
| `message` | `string` | Error message (falls back to `text`) |

```json
{
  "type": "item.completed",
  "item": {"id": "item_9", "type": "error", "message": "tool call timed out"}
}
```
