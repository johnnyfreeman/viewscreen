# user

Message event representing tool results returned to the assistant.

## Properties

| Property | Type | Description |
|----------|------|-------------|
| `type` | `"user"` | Event type identifier |
| `message` | `object` | The user message object |
| `parent_tool_use_id` | `string \| null` | ID of parent tool use if nested |
| `session_id` | `string` | UUID for the session |
| `uuid` | `string` | Unique identifier for this event |
| `tool_use_result` | `string \| object` | Tool result summary (varies by tool) |

## message Object

| Property | Type | Description |
|----------|------|-------------|
| `role` | `"user"` | Always `"user"` |
| `content` | `array` | Content blocks (tool_result) |

## content Array Items (tool_result)

| Property | Type | Description |
|----------|------|-------------|
| `type` | `"tool_result"` | Content type |
| `tool_use_id` | `string` | ID of the tool use this responds to |
| `content` | `string` | Tool output content |
| `is_error` | `boolean` | Whether the tool returned an error |

## tool_use_result Variants

The `tool_use_result` field varies by tool type:

### Bash Tool Result

| Property | Type | Description |
|----------|------|-------------|
| `stdout` | `string` | Standard output |
| `stderr` | `string` | Standard error |
| `interrupted` | `boolean` | Whether command was interrupted |
| `isImage` | `boolean` | Whether output is an image |

### Grep Tool Result

| Property | Type | Description |
|----------|------|-------------|
| `mode` | `string` | Search mode (e.g., `"content"`) |
| `numFiles` | `number` | Number of files matched |
| `filenames` | `string[]` | Matched filenames |
| `content` | `string` | Matched content |
| `numLines` | `number` | Number of lines matched |

### Error Result

When `is_error` is `true`, result is a string like `"Error: File does not exist."`

## Example (Bash Success)

```json
{
  "type": "user",
  "message": {
    "role": "user",
    "content": [{
      "tool_use_id": "toolu_01Pg6fQD3jhd3igkCRUUiFax",
      "type": "tool_result",
      "content": "total 4\ndrwxr-xr-x. 1 user user 40 Jan 25 11:48 .\n...",
      "is_error": false
    }]
  },
  "session_id": "227f43f6-e238-496b-ae57-acf7057ed19f",
  "tool_use_result": {
    "stdout": "total 4\n...",
    "stderr": "",
    "interrupted": false,
    "isImage": false
  }
}
```

## Example (Tool Error)

```json
{
  "type": "user",
  "message": {
    "role": "user",
    "content": [{
      "type": "tool_result",
      "content": "<tool_use_error>File does not exist.</tool_use_error>",
      "is_error": true,
      "tool_use_id": "toolu_01HTNEz4X7A6kj2hcpDiWAok"
    }]
  },
  "tool_use_result": "Error: File does not exist."
}
```

## Example (Permission Denied)

```json
{
  "type": "user",
  "message": {
    "role": "user",
    "content": [{
      "type": "tool_result",
      "content": "Claude requested permissions to write to /path/file.txt, but you haven't granted it yet.",
      "is_error": true,
      "tool_use_id": "toolu_01Ua2ufAQ3Yzo3YvaAzKo53Z"
    }]
  },
  "tool_use_result": "Error: Claude requested permissions to write..."
}
```
