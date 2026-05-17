# fx-flag-collisions

Stdio MCP fixture server whose tools have input-schema shapes that intentionally
trigger MCLIP flag-name collisions in the CLI under test.

## Tools

| Tool | Property shape | What it tests |
|---|---|---|
| `make_widget` | `snake_case`, `snakeCase` (both → `--snake-case`) | FX-COLLIDE-01: collision fallback to `--input` |
| `dump` | `output` (collides with reserved global flag) | FX-COLLIDE-02: rename to `--arg-output` |
| `polyglot` | `union` (oneOf), `tuple` (prefixItems), `foo.bar` (dotted) | FX-COLLIDE-03: complex-shape fallback to `--input` |

All tools declare `readOnlyHint: true` so destructive-action logic is not
involved in collision tests.

## Backs

- FX-COLLIDE-01, FX-COLLIDE-02, FX-COLLIDE-03
- FX-INPUT-01..04 (the `--input`-fallback rule exercised by the collision tools)

## Notes

Schemas use `map[string]any` JSON Schema directly through the SDK's low-level
`Server.AddTool` path because the property names cannot be expressed via Go
struct tags (case-distinct `snake_case`/`snakeCase`, dotted `"foo.bar"`).
