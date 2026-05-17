# fx-destructive

Stdio MCP fixture server for destructive and error-producing flows.

## Tools

| Tool | Annotations | Behaviour |
|---|---|---|
| `delete_thing(id: integer) -> { ok: bool }` | **none** (intentional) | Always returns `{ok: true}`. |
| `fails(text: string) -> error` | none | Always returns `isError: true` with a fixture-failure message. |
| `sleep(seconds: integer) -> text` | none | Sleeps server-side for `seconds`; cancellable via `notifications/cancelled`. |
| `safe_read() -> text` | `readOnlyHint: true` | Returns the literal string `"safe"`. |

## Backs

- FX-CI-01a, FX-CI-01b — destructive-action gating via `delete_thing`'s
  missing annotations (the `[MCLIP-14-0]` rule).
- FX-RAW-02 — `fails` exercises `CallToolResult.isError` pass-through.
- FX-SIGINT-01 — `sleep` must cancel cleanly when the client sends
  `notifications/cancelled`. The Go SDK propagates that into the handler's
  `context.Context`; the handler `select`s on `ctx.Done()` so the server
  exits the call promptly rather than letting `sleep` run to completion.

## Notes

`delete_thing` deliberately carries no `Annotations` so that the `nil`
serializes to "annotations field absent" — the conformance test for
"no annotations → potentially destructive" depends on the absence, not
on `destructiveHint: true`.
