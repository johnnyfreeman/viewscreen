# result

Final result event emitted at the end of a session.

## Properties

| Property | Type | Description |
|----------|------|-------------|
| `type` | `"result"` | Event type identifier |
| `subtype` | `string` | Result subtype (e.g., `"success"`) |
| `is_error` | `boolean` | Whether the session ended with an error |
| `duration_ms` | `number` | Total duration in milliseconds |
| `duration_api_ms` | `number` | API call duration in milliseconds |
| `num_turns` | `number` | Number of conversation turns |
| `result` | `string` | Final text result |
| `session_id` | `string` | UUID for the session |
| `total_cost_usd` | `number` | Total cost in USD |
| `usage` | `object` | Aggregate token usage |
| `modelUsage` | `object` | Per-model usage breakdown |
| `permission_denials` | `array` | List of denied permission requests |
| `errors` | `string[]` | Error messages (only present for `error_during_execution` subtype) |
| `uuid` | `string` | Unique identifier for this event |

## Subtypes

- `success` - Session completed successfully
- `error_during_execution` - Error occurred during execution (e.g., invalid session ID, missing configuration)

## usage Object

| Property | Type | Description |
|----------|------|-------------|
| `input_tokens` | `number` | Total input tokens |
| `cache_creation_input_tokens` | `number` | Tokens for cache creation |
| `cache_read_input_tokens` | `number` | Tokens read from cache |
| `output_tokens` | `number` | Total output tokens |
| `server_tool_use` | `object` | Server-side tool usage |
| `service_tier` | `string` | Service tier |
| `cache_creation` | `object` | Cache creation breakdown by tier |

### cache_creation Object

| Property | Type | Description |
|----------|------|-------------|
| `ephemeral_5m_input_tokens` | `number` | Tokens in 5-minute ephemeral cache |
| `ephemeral_1h_input_tokens` | `number` | Tokens in 1-hour ephemeral cache |

### server_tool_use Object

| Property | Type | Description |
|----------|------|-------------|
| `web_search_requests` | `number` | Number of web searches |
| `web_fetch_requests` | `number` | Number of web fetches |

## modelUsage Object

Keyed by model ID (e.g., `"claude-opus-4-5-20251101"`):

| Property | Type | Description |
|----------|------|-------------|
| `inputTokens` | `number` | Input tokens for this model |
| `outputTokens` | `number` | Output tokens for this model |
| `cacheReadInputTokens` | `number` | Cache read tokens |
| `cacheCreationInputTokens` | `number` | Cache creation tokens |
| `webSearchRequests` | `number` | Web search count |
| `costUSD` | `number` | Cost for this model |
| `contextWindow` | `number` | Context window size |
| `maxOutputTokens` | `number` | Max output tokens |

## permission_denials Array Items

| Property | Type | Description |
|----------|------|-------------|
| `tool_name` | `string` | Name of the tool |
| `tool_use_id` | `string` | Tool use ID |
| `tool_input` | `object` | Input that was denied |

## Example (Success)

```json
{
  "type": "result",
  "subtype": "success",
  "is_error": false,
  "duration_ms": 2303,
  "duration_api_ms": 2290,
  "num_turns": 1,
  "result": "4",
  "session_id": "960d3f4f-0bcb-41a8-a9b3-198e6594f9ac",
  "total_cost_usd": 0.030087749999999996,
  "usage": {
    "input_tokens": 2,
    "cache_creation_input_tokens": 3541,
    "cache_read_input_tokens": 15643,
    "output_tokens": 5,
    "server_tool_use": {"web_search_requests": 0, "web_fetch_requests": 0}
  },
  "modelUsage": {
    "claude-opus-4-5-20251101": {
      "inputTokens": 2,
      "outputTokens": 5,
      "costUSD": 0.030087749999999996,
      "contextWindow": 200000,
      "maxOutputTokens": 64000
    }
  },
  "permission_denials": []
}
```

## Example (With Permission Denials)

```json
{
  "type": "result",
  "subtype": "success",
  "is_error": false,
  "num_turns": 3,
  "permission_denials": [
    {
      "tool_name": "Write",
      "tool_use_id": "toolu_01Ua2ufAQ3Yzo3YvaAzKo53Z",
      "tool_input": {
        "file_path": "/home/user/test.txt",
        "content": "hello world\n"
      }
    }
  ]
}
```

## Example (Error During Execution)

Emitted when an error occurs during session execution (e.g., invalid resume session ID, configuration errors).

```json
{
  "type": "result",
  "subtype": "error_during_execution",
  "duration_ms": 0,
  "duration_api_ms": 0,
  "is_error": true,
  "num_turns": 0,
  "session_id": "701a5ae9-7860-41b6-b092-48be21901dc3",
  "total_cost_usd": 0,
  "usage": {
    "input_tokens": 0,
    "cache_creation_input_tokens": 0,
    "cache_read_input_tokens": 0,
    "output_tokens": 0,
    "server_tool_use": {"web_search_requests": 0, "web_fetch_requests": 0},
    "service_tier": "standard"
  },
  "modelUsage": {},
  "permission_denials": [],
  "errors": [
    "Error: --resume requires a valid session ID when used with --print..."
  ],
  "uuid": "1824789d-bd70-49ab-afd9-c4afd0f38a0a"
}
```
