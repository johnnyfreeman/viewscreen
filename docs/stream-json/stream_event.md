# stream_event

Streaming event emitted when using `--include-partial-messages` flag. These events provide real-time updates as the assistant generates a response.

## Properties

| Property | Type | Description |
|----------|------|-------------|
| `type` | `"stream_event"` | Event type identifier |
| `event` | `object` | The streaming event object |
| `session_id` | `string` | UUID for the session |
| `parent_tool_use_id` | `string \| null` | ID of parent tool use if nested |
| `uuid` | `string` | Unique identifier for this event |

## event Object

The `event` object contains the streaming event from the Anthropic API. The `event.type` field indicates the type of streaming event.

### Event Types

| event.type | Description |
|------------|-------------|
| `message_start` | Beginning of a new message |
| `content_block_start` | Beginning of a content block |
| `content_block_delta` | Incremental update to a content block |
| `content_block_stop` | End of a content block |
| `message_delta` | Update to message metadata (stop_reason, usage) |
| `message_stop` | End of the message |

## message_start Event

Signals the start of a new assistant message.

```json
{
  "type": "stream_event",
  "event": {
    "type": "message_start",
    "message": {
      "model": "claude-opus-4-5-20251101",
      "id": "msg_01QC57b2ATRj1HDdR55GGHxK",
      "type": "message",
      "role": "assistant",
      "content": [],
      "stop_reason": null,
      "stop_sequence": null,
      "usage": {
        "input_tokens": 2,
        "cache_creation_input_tokens": 3534,
        "cache_read_input_tokens": 15643,
        "output_tokens": 1,
        "service_tier": "standard"
      }
    }
  },
  "session_id": "974e4483-930c-4663-9eef-e07806950611",
  "parent_tool_use_id": null,
  "uuid": "7ba01495-f8de-4078-b7e2-5f574bbe8efa"
}
```

## content_block_start Event

Signals the start of a content block (text or tool_use).

### Text Block Start

```json
{
  "type": "stream_event",
  "event": {
    "type": "content_block_start",
    "index": 0,
    "content_block": {
      "type": "text",
      "text": ""
    }
  },
  "session_id": "974e4483-930c-4663-9eef-e07806950611",
  "parent_tool_use_id": null,
  "uuid": "543f0fd7-8e6c-4e3a-8c13-b2a16ff5afc0"
}
```

### Tool Use Block Start

```json
{
  "type": "stream_event",
  "event": {
    "type": "content_block_start",
    "index": 0,
    "content_block": {
      "type": "tool_use",
      "id": "toolu_01N1CxR7ghUyAmUudq6yTG2U",
      "name": "Read",
      "input": {}
    }
  },
  "session_id": "fa9a0555-220c-4895-a779-0193744e703a",
  "parent_tool_use_id": null,
  "uuid": "0f3b8709-5b47-4b47-81f4-92da553c58f5"
}
```

## content_block_delta Event

Incremental updates to a content block.

### Text Delta

```json
{
  "type": "stream_event",
  "event": {
    "type": "content_block_delta",
    "index": 0,
    "delta": {
      "type": "text_delta",
      "text": "4"
    }
  },
  "session_id": "974e4483-930c-4663-9eef-e07806950611",
  "parent_tool_use_id": null,
  "uuid": "fa52dce0-38fc-4e75-a711-8175b4b17062"
}
```

### Input JSON Delta (for tool_use)

```json
{
  "type": "stream_event",
  "event": {
    "type": "content_block_delta",
    "index": 0,
    "delta": {
      "type": "input_json_delta",
      "partial_json": "{\"file"
    }
  },
  "session_id": "fa9a0555-220c-4895-a779-0193744e703a",
  "parent_tool_use_id": null,
  "uuid": "7bffddf3-5595-4fe2-a0cb-23d8a0a506f9"
}
```

## content_block_stop Event

Signals the end of a content block.

```json
{
  "type": "stream_event",
  "event": {
    "type": "content_block_stop",
    "index": 0
  },
  "session_id": "974e4483-930c-4663-9eef-e07806950611",
  "parent_tool_use_id": null,
  "uuid": "e85ce55a-98dc-4e2c-931d-ae90a32e4a48"
}
```

## message_delta Event

Updates to message metadata, including the stop reason.

```json
{
  "type": "stream_event",
  "event": {
    "type": "message_delta",
    "delta": {
      "stop_reason": "end_turn",
      "stop_sequence": null
    },
    "usage": {
      "input_tokens": 2,
      "cache_creation_input_tokens": 3534,
      "cache_read_input_tokens": 15643,
      "output_tokens": 5
    }
  },
  "session_id": "974e4483-930c-4663-9eef-e07806950611",
  "parent_tool_use_id": null,
  "uuid": "07660786-8942-4297-ae30-4ea1e0617588"
}
```

### stop_reason Values

| Value | Description |
|-------|-------------|
| `end_turn` | Assistant finished generating |
| `tool_use` | Assistant wants to use a tool |

## message_stop Event

Signals the end of the message.

```json
{
  "type": "stream_event",
  "event": {
    "type": "message_stop"
  },
  "session_id": "974e4483-930c-4663-9eef-e07806950611",
  "parent_tool_use_id": null,
  "uuid": "8a51c284-b539-422f-bd2e-48a5130e9a1f"
}
```

## Usage Notes

- Stream events are only emitted when using `--include-partial-messages` flag
- Stream events are interspersed with the normal `assistant` and `user` events
- **Important:** The `assistant` event is emitted mid-stream (after content deltas but before `content_block_stop`, `message_delta`, `message_stop`)
- Renderers should wait for `message_stop` before resetting streaming state
- Use stream events to build real-time UI updates showing text as it's generated

## Event Ordering Example

For a single turn with streaming enabled:

```
1. message_start
2. content_block_start (index=0)
3. content_block_delta (index=0) × N
4. assistant (complete message)  ← mid-stream!
5. content_block_stop (index=0)
6. message_delta
7. message_stop
```

For tool use followed by text response:

```
Turn 1 (tool_use):
  message_start → content_block_start → deltas → assistant → content_block_stop → message_delta → message_stop

user (tool_result)

Turn 2 (text):
  message_start → content_block_start → deltas → assistant → content_block_stop → message_delta → message_stop
```
