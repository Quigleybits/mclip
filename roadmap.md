# MCLIP Roadmap

Candidate additions for future profile revisions. Non-normative. Promotion requires a profile rule (`[MCLIP-§-NN]`), a conformance fixture, and reference behaviour in `mclio`.

## v1 candidates

### `meta ping` — health-check verb

MCP defines a `ping` utility at the JSON-RPC level. MCLIP v0 reserves the `meta` category verb (§1.3) but does not project ping into the CLI. A future rule could specify:

- Invocation: `<binary> <server> meta ping`
- Behaviour: issue MCP `ping`; exit `0` on response, exit per §6.1 transport codes on failure
- Output: JSON envelope on `-o json`; minimal text on default `-o text`
- Likely Core (trivial cost) — pending wrapper-maintainer feedback on whether `meta` verbs belong in the required level.

### `--log-level` — project MCP server logging

MCP's `logging` server capability supports `debug`/`info`/`warning`/`error` levels via `logging/setLevel`. MCLIP v0 exposes only client-side `--verbose`/`--quiet` for stderr noise; it does not project server-emitted log messages. A future **MCLIP-Logging** module could define:

- `--log-level <level>` flag — issues `logging/setLevel` after initialise.
- NDJSON log surface on stderr, matching the §9 progress conventions.
- Interaction rules with `--verbose`/`--quiet` (probably independent axes: one gates server messages, the other gates client diagnostics).

## Implementation-only (out of profile scope)

These do not affect script portability across conformant wrappers. If pursued they belong in `mclio-architecture.md`, not the profile:

- Shell completions (bash / zsh / fish / PowerShell). Per-binary polish; conformant tools may or may not ship them without affecting conformance claims.
