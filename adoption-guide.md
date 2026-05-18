# MCLIP Adoption Guide for Existing Wrappers

Version: 0.2 (2026-05-18)
Status: draft, aligned with `profile-v0.md` draft 2.2.
Audience: maintainers of `mcp2cli`, `MCPorter`, `MCPShim`, `f/mcptools` (and any other MCP-to-CLI wrapper considering an `--mclip` mode).

> **Outreach status (2026-05-18):** This guide is part of the public MCLIP draft. Formal maintainer outreach has not yet begun — it is held until the reference implementation (`mclio`) is functional and the conformance harness is end-to-end exercisable. If you maintain one of the wrappers named above and want to push back on the framing or the spec itself before formal outreach reaches you, an issue on this repo or a note to `aidey@mclip.dev` (in setup) is the fastest way in.

This is not a "switch to MCLIP" pitch. It's a "here is the minimum surface area you need to add to claim MCLIP-Core conformance, and here is what each optional module adds on top." Your existing UX, daemon, codegen, REPL — all of it — stays intact. MCLIP-mode is a parallel surface.

## The core principle

Every MCLIP rule is testable against the conformance suite (`conformance-fixtures.md`). Adoption means: when invoked in MCLIP mode, your wrapper passes every MUST in MCLIP-Core (the only required level) plus every MUST in any optional module you claim. Pick the modules that match what your wrapper already does well; skip the rest.

The §15.1 claim format is:

> "<your-wrapper-name> conforms to MCLIP v0 — Core + <modules>"

The shortest valid claim is `Core`. The longest is `Core + Resources + Prompts + HTTP + Streaming + Safety + Auth + Metadata + Discovery`. Any subset is valid as long as Core is present and every claimed module's MUSTs pass.

## Two adoption shapes

### Shape A — `--mclip` flag on existing commands

Lowest friction. Add a top-level `--mclip` (or equivalent) flag that switches your wrapper into Core-conformant mode for a single invocation.

Works well when:
- Your invocation shape is close to MCLIP-Core's canonical `<binary> <server> <category> <verb>` form already.
- The work is mostly in the output renderer (envelope shape, exit codes) and credential resolution.

Examples of wrappers this likely fits without surgery: `MCPorter` (dot-notation maps cleanly), `MCPShim` (already accepts long flags).

### Shape B — `<your-binary> mclip ...` sub-command

For wrappers whose invocation shape is fundamentally different from MCLIP-Core. Add a top-level `mclip` sub-command that hosts the canonical surface as a self-contained CLI.

Works well when:
- Your wrapper's main shape is server-as-binary, session-prefix, or any pattern incompatible with `<binary> <server> <category> <verb>`.
- You want users to opt into MCLIP explicitly rather than accidentally.

Examples of wrappers this likely fits: `mcp2cli` (`work email send` → `mcp2cli mclip work tools call email-send`).

Both shapes are equally valid for claim purposes — the conformance suite tests *invocations*, not entry-point conventions.

## Module-by-module adoption cost

Rough complexity estimate per module. Numbers are calibrated to a wrapper with an existing MCP client, not a from-scratch implementation.

### MCLIP-Core (required)

**What it costs:** the §1 command shape (`<binary> <server> <category> <verb> [target]`), §2 flag mapping (long-flag-from-schema, `--input`/`--input-file` fallback), §3 stdin handling, §4 output formats (text + json; ndjson flag defined but only required if you claim Streaming), §5 JSON envelope (`{ result: ... }` / `{ error: ... }`), §6 exit codes (specific values from the §6.1 table; no generic 1 for any condition with a dedicated code), §10 pagination defaults, §12.1 stdio + §12.4–§12.5 transport errors and timeout, §13.1/§13.3/§13.4 MCLIP-specific server discovery, §14.0 safety baseline (refuse destructive in non-TTY without `--yes`), §14.4 CI-safe composition rollup (no prompts non-interactively; deterministic output; specific exit codes), §15 conformance claim format.

**Likely-most-painful items for existing wrappers:**
- Switching argument-passing style if your wrapper uses `key:value`, `--params <JSON>`, or `key:=value`. MCLIP-Core requires long-flags-from-schema with `--input` / `--input-file` fallback for ambiguous shapes only.
- Changing JSON output shape to the `{ result: ... }` envelope if your wrapper currently emits raw tool results or a custom wrapping.
- Mapping your exit codes to the specific §6.1 codes (especially if you currently use generic 1 for everything non-zero).

### MCLIP-Resources

**What it costs:** §7 verbs — `list`, `read`, `templates`, `watch`. If your wrapper already exposes resources, mostly a verb-renaming exercise. `watch` requires server `capabilities.resources.subscribe: true` and clean SIGINT handling that issues `resources/unsubscribe` before exit.

### MCLIP-Prompts

**What it costs:** §8 verbs — `list`, `get`. Trivial if your wrapper already exposes prompts. Note that prompt argument values are always strings per `[MCLIP-8-03]` (MCP prompts don't carry per-argument JSON Schema).

### MCLIP-HTTP

**What it costs:** §12.1 HTTP transport + §12.2 protocol-version header + §12.3 HTTP semantics of `--transport http`. If your wrapper already speaks HTTP, this is mostly the header rule and the no-tokens-in-query-strings rule.

### MCLIP-Streaming

**What it costs:** §9.1 (`progressToken` on outgoing requests; render progress to **stderr**), §9.2 (NDJSON format for profile-defined streaming commands — currently just `resources watch`). Most wrappers do NOT currently surface progress at all — adding it is straightforward but requires care that progress never leaks to stdout when consumers are piping JSON.

### MCLIP-Safety

**What it costs:** §14.1 confirmation prompt on stderr (default no), §14.2 `--dry-run` (validate locally; emit would-send JSON-RPC request; redact credentials per `[MCLIP-14-09]`), §14.3 trusted-server escape hatch via `safety.trustAnnotations` config. Most invasive piece is `--dry-run` — it requires local schema validation, not just request mocking.

### MCLIP-Auth

**What it costs:** §11.1 credential priority (OS keychain → `MCLIP_TOKEN_<SERVER>` env → `auth.token` config), §11.2 Bearer header transmission, §11.3 OAuth-as-extension rules, §11.9 plaintext-token warning on permissive config files. **No generic `MCLIP_TOKEN` env var; no generic `--auth-token` flag.** Most wrappers will need to rename their existing env vars and remove (or hide behind a non-conformant subcommand) any generic auth flag.

### MCLIP-Metadata

**What it costs:** §16 honouring `mclip.*` keys per the companion Extensions Track SEP. The Extensions SEP is unfiled at the time of writing; claiming this module is **premature** for any wrapper today. Wait for the SEP.

### MCLIP-Discovery

**What it costs:** §13.2 reading inherited config files from Claude Desktop / `.vscode/mcp.json` / `.cursor/mcp.json`. If your wrapper already does this (MCPorter, f/mcptools), it's a no-op once the trust rule from §13.6 is respected (same consent rule applies symmetrically to all inherited files).

## Picking a starting set

Conservative starting claim (lowest cost): **Core + Auth**. Auth requires the smallest surface change beyond Core and covers the most common real-world usage (HTTP servers behind Bearer tokens).

Most-common-wrapper-fit starting claim: **Core + Resources + Auth**. Adds resources because most wrappers already expose them and the verb-rename cost is tiny.

Maximalist starting claim: **Core + Resources + Prompts + HTTP + Safety + Auth + Discovery** (all in-scope modules except Metadata and Streaming). Streaming is light to add once you have NDJSON rendering working, so reaching maximalist is straightforward in a second pass.

## Pricing Core migration honestly by wrapper shape

The cost of adding MCLIP-Core depends heavily on how far your wrapper's existing surface is from MCLIP-Core's canonical shape. The earlier draft of this guide claimed "weekend" — that was wrong. Below is an honest estimate per wrapper shape.

| Your wrapper's shape | Likely Core-migration cost | What needs writing |
|---|---|---|
| **Shape A — close to canonical** (e.g. MCPorter's `<binary> call <server>.<tool>`, MCPShim's `--server X --tool Y`): your invocation lines up, you already emit long flags, you already have a JSON output mode. | **Days of focused work.** Renderer/exit-code remap + credential-pipeline plumbing + conformance-harness integration. | A `--mclip` flag that switches renderer + exit codes + credential resolution to spec; a small input-mode shim if your arg parser differs from long-flags-from-schema. |
| **Shape B — moderate divergence** (e.g. wrappers using `key:value` colon flags or `key:=value` typed pairs): your discovery and transport are fine, but argument parsing has to be re-emitted. | **One to two weeks.** Add a parallel parser; keep the existing one untouched; route to the parallel one when `--mclip` is set. | Above, plus a parallel schema-aware flag generator (the two-phase parser pattern from `mclio-architecture.md` is the reference design). |
| **Shape C — fundamental incompatibility** (e.g. mcp2cli's server-as-binary, mcpc's session-prefix): your top-level invocation shape cannot host `<server> <category> <verb>` without colliding with your existing UX. | **Several weeks to a small project.** A self-contained sub-command (`<binary> mclip ...`) that re-uses your wrapper's transport + discovery + auth code but with its own argv parser, renderer, exit-code table, and config-source pipeline. | Above, plus the sub-command scaffolding, its own help text, and a separate test surface for the sub-command's conformance. |

`mclio` (per `mclio-architecture.md`) gives you the design for the new pipeline pieces; don't try to invent your own from `profile-v0.md` alone.

## Step-by-step: adding MCLIP-Core

For a wrapper that already speaks MCP, has its own opinionated UX, and wants to add an `--mclip` flag (Shape A or B) or a `<binary> mclip` sub-command (Shape C):

1. **Read `profile-v0.md` once end-to-end.** It's ~800 lines; skim the appendices on first pass.
2. **Read `conformance-fixtures.md` and pick the 5-10 fixtures most likely to break against your wrapper today.** That's your TDD list. Use the harness contracts from `fixtures-spec.md` so you understand what "passing" actually measures.
3. **If Shape C, scaffold the parallel sub-command first.** Don't try to bolt MCLIP onto the existing invocation parser; the conformance failures hide in the boundary. The sub-command sees clean argv from the first character.
4. **Implement the §6.1 exit-code table and the §5.2 envelope.** Every other piece of the migration assumes correct exit codes and envelope shape. Get these right first.
5. **Implement the two-phase argv parser** (per `mclio-architecture.md` "Two-phase parser contract"). Phase 1 (global flags + path) runs before any transport / schema discovery; Phase 2 (generated tool flags) runs after `tools/list` has returned. A single-phase parser will fail FX-GLOBAL-03 and FX-COLLIDE-02.
6. **Wire up the §11 credential priority order** (keychain → per-server env → config), behind your existing transport. The conformance harness asserts on the resolved `Authorization` header; the path that gets there is yours.
7. **Add the §14.0 destructive-in-non-TTY refusal** and the `[MCLIP-14-15]` specific-exit-code discipline. These are mechanical once the renderer + exit-code table are in.
8. **Add the §14.4 CI-safe rollup checks**: `isatty(stdin)` detection, no env-var-only opt-in to non-TTY mode, no-secret-leak in error envelopes, deterministic JSON output (object-key ordering stable, no walltime in output).
9. **Wire the conformance harness** from `fixtures-spec.md` against your `--mclip` mode. Iterate until Core passes.
10. **Publish a §15.1 claim** ("`<your-wrapper>` conforms to MCLIP v0 — Core") in your README. Include the per-fixture pass/fail report the harness generated so users can see exactly what's tested.

## Where to push back on the spec

Adoption feedback is more valuable than adoption itself, especially in v0. If you find a rule that:
- Would force a breaking change to your existing surface that you don't want to make,
- Conflicts with a CLI convention your users rely on,
- Is testably ambiguous (two implementations could both pass the fixture but produce different observable behaviour),

… file an issue at `<spec repo>` referencing the rule ID. The spec is in draft (v0 revision 2.2 at the time of writing); rule changes during draft do not break stable IDs, only modify their MUSTs.

The most common useful pushback shapes are:
1. **"Your default exit code for X conflicts with our convention; here's a counter-proposal."**
2. **"Your argument-passing rule makes our existing escape-hatch unreachable; can the rule be a SHOULD instead of a MUST for case Y?"**
3. **"This module's MUST list overlaps confusingly with another module's; the boundary is unclear."**

## What MCLIP does NOT ask of you

- Vendor adoption work. MCP servers do not need to advertise anything for MCLIP-Core; if a server is plain MCP, MCLIP gives users a canonical surface against it.
- Your own UX changes. The conformant surface is parallel.
- Rewriting your transport, daemon, registry, codegen, or LLM-mediated chat features.
- Tracking the spec lockstep with every MCP revision. The profile declares its baseline MCP version explicitly per `[MCLIP-15-05]`; you claim against the baseline you support.

## Help we'd appreciate from existing wrappers

In priority order:

1. **A reading pass of `profile-v0.md` revision 2.2.** Five-line response noting "this would work" / "this would not work because X" beats a perfect-but-late SEP comment.
2. **A reading pass of `conformance-fixtures.md`.** Tell us if any fixture is impossible to satisfy without breaking your wrapper.
3. **A signal in MCP Discord `#general` (or wherever the SEP discussion ends up).** Even "we've read it; haven't decided yet" raises the SEP's sponsor-attractiveness.
4. **A draft `--mclip` mode against MCLIP-Core only.** This is what we mean by "executable reference implementation by an existing wrapper" — it's the bar that lets the SEP reach Final status.

If you respond to any of those, please CC `aidey@mclip.dev` (in setup) so feedback isn't lost in a Discord scroll.

## References

- `profile-v0.md` — the normative spec (revision 2.2 at time of writing).
- `security-model.md` — the companion trust / destructive-action / CI-safe model.
- `conformance-fixtures.md` — the per-rule fixture catalogue.
- `fixtures-spec.md` — implementation spec for the synthetic MCP fixture servers.
- `mclio-architecture.md` — the design for `mclio` (the production CLI / executable reference), for comparison with your own implementation.
