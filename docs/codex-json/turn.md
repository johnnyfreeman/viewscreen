# turn.started / turn.completed / turn.failed

A "turn" is one model response cycle within a thread. Each turn is bracketed by a
`turn.started` and either a `turn.completed` (with token usage) or a
`turn.failed` (with an error). The work the model does during the turn is
reported as [items](item.md) in between.

## turn.started

Marks the beginning of a turn. It carries no payload beyond its type.

### Properties

| Property | Type | Description |
|----------|------|-------------|
| `type` | `"turn.started"` | Event type identifier |

### Example

```json
{"type": "turn.started"}
```

Viewscreen renders nothing for `turn.started`, but uses it to increment the turn
counter shown in the TUI sidebar (mirroring how Claude's `assistant` events drive
the count).

## turn.completed

Marks the successful end of a turn and reports cumulative token usage.

### Properties

| Property | Type | Description |
|----------|------|-------------|
| `type` | `"turn.completed"` | Event type identifier |
| `usage` | `object` | Token accounting for the turn |

### usage Object

| Property | Type | Description |
|----------|------|-------------|
| `input_tokens` | `number` | Total input tokens |
| `cached_input_tokens` | `number` | Input tokens served from cache |
| `output_tokens` | `number` | Output tokens generated |
| `reasoning_output_tokens` | `number` | Output tokens spent on reasoning |

### Example

```json
{
  "type": "turn.completed",
  "usage": {
    "input_tokens": 11211,
    "cached_input_tokens": 4480,
    "output_tokens": 26,
    "reasoning_output_tokens": 19
  }
}
```

Viewscreen folds `usage` into the shared state: `cached_input_tokens` maps onto
the cache-read counter and `reasoning_output_tokens` onto a dedicated reasoning
counter shown in the sidebar. Codex does not report a cost, so the sidebar omits
the cost/rate lines for Codex streams (`state.State.ReportsCost`).

## turn.failed

Marks a turn that ended with an error.

### Properties

| Property | Type | Description |
|----------|------|-------------|
| `type` | `"turn.failed"` | Event type identifier |
| `error` | `object` | Failure detail |

### error Object

| Property | Type | Description |
|----------|------|-------------|
| `message` | `string` | Human-readable failure message |

### Example

```json
{
  "type": "turn.failed",
  "error": {
    "message": "model request failed: connection reset by peer"
  }
}
```

Viewscreen renders this as a styled "Turn Failed" block with the message.
