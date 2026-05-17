# fx-progress

Stdio MCP fixture server exposing one tool that emits a deterministic series
of `notifications/progress` messages before returning.

## Tools

| Tool | Annotations | Behaviour |
|---|---|---|
| `slow_count(target: integer) -> text` | `readOnlyHint: true` | Emits `target` progress notifications (`progress = 1..target`, `total = target`), then returns the text `"done <target>"`. |

Progress notifications are only emitted if the client included a
`progressToken` in the inbound `tools/call` request's `_meta`. If absent,
the tool runs silently and just returns the final text.

## Backs

- `[MCLIP-9-02]` — the rule that progress notifications go to stderr while
  the JSON result goes cleanly to stdout. `mclio` must split the
  two streams correctly so machine consumers piping stdout get only the
  tool result.
- Future FX-PROGRESS-* fixtures (none defined yet in
  `conformance-fixtures.md`).

## Determinism

There is **no** wall-clock sleep between progress emits. The fixture
produces exactly `target` notifications, in order, with monotonically
increasing `progress` values. The harness asserts on count and ordering,
not on inter-event timing.
