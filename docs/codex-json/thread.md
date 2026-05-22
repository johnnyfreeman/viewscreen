# thread.started

Emitted once at the very start of a session, before any turn. A "thread" is
Codex's term for a conversation; the `thread_id` identifies it for the lifetime
of the stream.

## Properties

| Property | Type | Description |
|----------|------|-------------|
| `type` | `"thread.started"` | Event type identifier |
| `thread_id` | `string` | UUID for the conversation thread |

## Example

```json
{
  "type": "thread.started",
  "thread_id": "019e4fe1-a36d-70d0-914e-1f2a230766e3"
}
```

## Rendering

Viewscreen renders this as the session header ("Codex Session") and prints the
thread id. It is also the point at which the TUI brands itself as `codex`. See
`codex.Renderer.renderThreadStarted` and `events.EventProcessor.detectAgent`.
