# Stream JSON Output Types

Documentation for JSON object types emitted by `claude -p --output-format stream-json`.

## Event Types

| Type | Description |
|------|-------------|
| [system](system.md) | Session initialization event with configuration |
| [assistant](assistant.md) | Assistant message (text response or tool use request) |
| [user](user.md) | Tool result returned to the assistant |
| [result](result.md) | Final session result with usage and cost |
| [stream_event](stream_event.md) | Real-time streaming events (requires `--include-partial-messages`) |

## Event Flow

```
session start
    |
    v
[result] (subtype: error_during_execution) -- if startup error
    |
    v
[system] (subtype: init)
    |
    v
[stream_event]* --> [assistant] --> [stream_event]* (text or tool_use)
    |
    +---> [user] (tool_result) --> [stream_event]* --> [assistant] --> ...
    |
    v
[result] (subtype: success)

* stream_event only with --include-partial-messages flag
```

**Important:** When streaming is enabled (`--include-partial-messages`), the `assistant` event
is emitted **mid-stream** — after the content deltas but before `content_block_stop`,
`message_delta`, and `message_stop`. Renderers must not reset streaming state when
processing the `assistant` event; instead, wait for `message_stop` to finalize the message.

### Detailed Stream Event Order (per turn)

```
message_start
content_block_start (index=0)
content_block_delta (index=0) × N
[assistant]  ← emitted here, mid-stream
content_block_stop (index=0)
message_delta
message_stop
```

## Common Properties

All events share these properties:

| Property | Type | Description |
|----------|------|-------------|
| `type` | `string` | Event type identifier |
| `session_id` | `string` | UUID for the session |
| `uuid` | `string` | Unique identifier for this event |
