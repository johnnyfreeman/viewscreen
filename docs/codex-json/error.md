# error

A stream-level error envelope. This is distinct from the [`error` *item*](item.md#error)
(which appears inside an `item.*` envelope): a top-level `error` reports a problem
with the stream or session itself rather than a single unit of work.

## Properties

| Property | Type | Description |
|----------|------|-------------|
| `type` | `"error"` | Event type identifier |
| `message` | `string` | Human-readable error message |

## Example

```json
{
  "type": "error",
  "message": "stream closed unexpectedly"
}
```

Viewscreen renders this as a styled "Error" block with the message
(`codex.Renderer.renderError`).

## Note on detection

Unlike the other Codex envelopes, `error` has no dot in its `type`, so it is
matched explicitly. `web_search`/`reasoning`/etc. are *item* types and never
appear as a top-level `type`. See `codex.IsEventType`.
