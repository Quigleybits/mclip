# fx-pagination

Stdio MCP fixture server exposing 50 read-only tools, `t01` … `t50`, to
exercise `tools/list` auto-pagination in the CLI under test.

## Tools

50 tools, all with the same shape:

| Tool name | InputSchema | Output |
|---|---|---|
| `t01` … `t50` | `{ "type": "object", "properties": {}, "additionalProperties": false }` | Text content equal to the tool's own name. |

All 50 declare `readOnlyHint: true`.

## Pagination

`ServerOptions.PageSize = 10`, so `tools/list` returns 10 tools at a time.
A client doing a full enumeration follows the opaque `nextCursor` field four
times (page 1: t01..t10, page 2: t11..t20, …, page 5: t41..t50).

## Cursor opacity

The `fixtures-spec.md` originally sketched a debug-friendly `"cursor:N"`
cursor shape; the Go SDK's actual cursor encoding is a base64-wrapped
gob-encoded `pageToken` struct. The opacity is stronger this way — the
harness MUST treat the cursor as opaque and round-trip the bytes verbatim.
Do not parse, decode, or assert on cursor contents.

## Backs

- FX-CI-02 — multi-page determinism. The same `tools/list` walk must
  produce identical pages on repeat runs.
- General `tools list --limit N` paths in `mclio`.
