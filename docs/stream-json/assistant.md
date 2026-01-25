# assistant

Message event from the assistant (Claude).

## Properties

| Property | Type | Description |
|----------|------|-------------|
| `type` | `"assistant"` | Event type identifier |
| `message` | `object` | The assistant message object |
| `parent_tool_use_id` | `string \| null` | ID of parent tool use if nested |
| `session_id` | `string` | UUID for the session |
| `error` | `string` | Error type (present for API errors, e.g., `"unknown"`) |
| `uuid` | `string` | Unique identifier for this event |

## message Object

| Property | Type | Description |
|----------|------|-------------|
| `model` | `string` | Model identifier |
| `id` | `string` | Message ID (e.g., `"msg_01F89..."`) |
| `type` | `"message"` | Always `"message"` |
| `role` | `"assistant"` | Always `"assistant"` |
| `content` | `array` | Content blocks (text or tool_use) |
| `stop_reason` | `string \| null` | Reason for stopping |
| `stop_sequence` | `string \| null` | Stop sequence if applicable |
| `usage` | `object` | Token usage information |
| `context_management` | `object \| null` | Context management info |

## content Array Items

### Text Content Block

| Property | Type | Description |
|----------|------|-------------|
| `type` | `"text"` | Content type |
| `text` | `string` | The text content |

### Tool Use Content Block

| Property | Type | Description |
|----------|------|-------------|
| `type` | `"tool_use"` | Content type |
| `id` | `string` | Tool use ID (e.g., `"toolu_01Pg6..."`) |
| `name` | `string` | Tool name (e.g., `"Bash"`, `"Read"`) |
| `input` | `object` | Tool input parameters |

## usage Object

| Property | Type | Description |
|----------|------|-------------|
| `input_tokens` | `number` | Input tokens used |
| `cache_creation_input_tokens` | `number` | Tokens for cache creation |
| `cache_read_input_tokens` | `number` | Tokens read from cache |
| `cache_creation` | `object` | Cache creation breakdown by tier |
| `output_tokens` | `number` | Output tokens used |
| `service_tier` | `string` | Service tier (e.g., `"standard"`) |

### cache_creation Object

| Property | Type | Description |
|----------|------|-------------|
| `ephemeral_5m_input_tokens` | `number` | Tokens in 5-minute ephemeral cache |
| `ephemeral_1h_input_tokens` | `number` | Tokens in 1-hour ephemeral cache |

## Example (Text Response)

```json
{
  "type": "assistant",
  "message": {
    "model": "claude-opus-4-5-20251101",
    "id": "msg_01F89mJEggpRxDzDL9AHMDsF",
    "type": "message",
    "role": "assistant",
    "content": [{"type": "text", "text": "4"}],
    "stop_reason": null,
    "usage": {"input_tokens": 2, "output_tokens": 1}
  },
  "session_id": "960d3f4f-0bcb-41a8-a9b3-198e6594f9ac",
  "uuid": "3ebf771d-6ae5-412f-a9f0-2de7710c7bba"
}
```

## Example (Tool Use)

```json
{
  "type": "assistant",
  "message": {
    "content": [{
      "type": "tool_use",
      "id": "toolu_01Pg6fQD3jhd3igkCRUUiFax",
      "name": "Bash",
      "input": {"command": "ls -la", "description": "List files"}
    }]
  },
  "session_id": "227f43f6-e238-496b-ae57-acf7057ed19f"
}
```

## Example (API Error)

When an API error occurs (e.g., invalid model), the assistant event includes an `error` property and the error message in the content.

```json
{
  "type": "assistant",
  "message": {
    "id": "62cc41b7-9000-4038-ad59-8d7664cc893a",
    "model": "<synthetic>",
    "role": "assistant",
    "stop_reason": "stop_sequence",
    "type": "message",
    "content": [{
      "type": "text",
      "text": "API Error: 404 {\"type\":\"error\",\"error\":{\"type\":\"not_found_error\",\"message\":\"model: nonexistent-model\"}}"
    }],
    "usage": {"input_tokens": 0, "output_tokens": 0}
  },
  "session_id": "70257673-32c9-45bb-8219-1f38497fc477",
  "error": "unknown",
  "uuid": "763f420f-beee-4b7a-a0d1-0ffeb4f6a1e8"
}
```
