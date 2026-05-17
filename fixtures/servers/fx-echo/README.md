# fx-echo

Minimal stdio MCP fixture server. Single read-only `echo` tool that returns
its input unchanged.

## Tools

| Tool | Annotations | Behaviour |
|---|---|---|
| `echo(text: string) -> text` | `readOnlyHint: true` | Returns `text` unchanged. |

## Backs

- FX-GLOBAL-01, FX-GLOBAL-02, FX-GLOBAL-03
- FX-RAW-01
- FX-CI-02 (determinism)
- FX-CI-04 (exit-code variants — base server for negative cases)

## Run

    go build && ./fx-echo

Then send JSON-RPC messages over stdin. Stdio only; no flags, no env vars.

## Notes

This is the foundation server: the build/test pattern proved out here is
replicated across the other 8 fixture servers.
