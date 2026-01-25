# system

Initialization event emitted at the start of a session.

## Properties

| Property | Type | Description |
|----------|------|-------------|
| `type` | `"system"` | Event type identifier |
| `subtype` | `string` | Event subtype (e.g., `"init"`) |
| `cwd` | `string` | Current working directory |
| `session_id` | `string` | UUID for the session |
| `tools` | `string[]` | Available tool names |
| `mcp_servers` | `array` | MCP server configurations |
| `model` | `string` | Model identifier (e.g., `"claude-opus-4-5-20251101"`) |
| `permissionMode` | `string` | Permission mode (e.g., `"default"`) |
| `slash_commands` | `string[]` | Available slash commands |
| `apiKeySource` | `string` | Source of API key (e.g., `"none"`) |
| `claude_code_version` | `string` | Claude Code version (e.g., `"2.1.19"`) |
| `output_style` | `string` | Output style (e.g., `"default"`) |
| `agents` | `string[]` | Available agent names |
| `skills` | `array` | Available skills |
| `plugins` | `object[]` | Loaded plugins with `name` and `path` |
| `uuid` | `string` | Unique identifier for this event |

## Subtypes

- `init` - Session initialization

## plugins Array Items

| Property | Type | Description |
|----------|------|-------------|
| `name` | `string` | Plugin name (e.g., `"typescript-lsp"`) |
| `path` | `string` | Path to the plugin installation |

## Example

```json
{
  "type": "system",
  "subtype": "init",
  "cwd": "/home/user/project",
  "session_id": "960d3f4f-0bcb-41a8-a9b3-198e6594f9ac",
  "tools": ["Task", "Bash", "Glob", "Grep", "Read", "Edit", "Write", ...],
  "mcp_servers": [],
  "model": "claude-opus-4-5-20251101",
  "permissionMode": "default",
  "slash_commands": ["compact", "context", "cost", "init", ...],
  "apiKeySource": "none",
  "claude_code_version": "2.1.19",
  "output_style": "default",
  "agents": ["Bash", "general-purpose", "Explore", "Plan", ...],
  "skills": [],
  "plugins": [
    {"name": "typescript-lsp", "path": "/home/user/.claude/plugins/cache/..."},
    {"name": "gopls-lsp", "path": "/home/user/.claude/plugins/cache/..."}
  ],
  "uuid": "23e7d5c4-4d4b-4334-99ab-07318cb8733d"
}
```
