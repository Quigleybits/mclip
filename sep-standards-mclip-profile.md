# SEP — MCLIP: MCP Command-Line Interface Profile

SEP: TBD — to be assigned by Core Maintainers on PR open
Title: MCLIP — MCP Command-Line Interface Profile (v0)
Author: Aidan Quigley (`@<github-handle-tbd>`, aidanjohnquigley@gmail.com)
Status: Pre-Draft — pending PR open and sponsor signal; will move to Draft upon sponsorship per [MCP SEP guidelines](https://modelcontextprotocol.io/community/sep-guidelines)
Type: Standards Track
Created: 2026-05-16
Requires: MCP 2025-11-25
Companion: SEP TBD — MCLIP CLI Metadata Extension (Extensions Track)
PR: TBD — added on PR open
Sponsor: TBD — primary candidate `@pja-ant` per `sponsor-candidates.md`; secondary `@pcarleton`; tertiary `@kurtisvg`

## Abstract

This SEP defines **MCLIP** — a conformance profile that specifies how MCP-to-CLI client implementations (`mcp2cli`, `MCPorter`, `MCPShim`, `f/mcptools`, and similar) expose MCP server capabilities as command-line invocations. The profile standardises the *mechanics* of CLI access — command shape, JSON-Schema → flag mapping, output envelope, exit codes, destructive-action safeguards, server discovery — so a script written against one conformant client runs unchanged against any other conformant client targeting the same MCP server.

MCLIP does not modify the MCP wire protocol. It does not require vendor-side work. It is a client-side conformance layer, structured as **MCLIP-Core** (required for any conformance claim) plus eight independently-claimable optional modules (`Resources`, `Prompts`, `HTTP`, `Streaming`, `Safety`, `Auth`, `Metadata`, `Discovery`).

The complete normative profile is in `profile-v0.md` (draft 2.2 at time of filing). This SEP introduces the profile, motivates the design, and documents the rationale behind specific rule decisions.

## Filing form

A SEP filed under `modelcontextprotocol/modelcontextprotocol/seps/` is a single self-contained file. To produce that file from this repository, concatenate in this order:

1. The Preamble through "Open Issues" sections of this file (sep-standards-mclip-profile.md) — the **front matter**.
2. The full §0–§16 normative content of `profile-v0.md`, plus its Appendices A–C — the **normative body**.
3. The "References" section of this file (sep-standards-mclip-profile.md) — the **back matter**.

This produces one document with the SEP preamble, the full normative spec, the rationale and security analysis, and a consolidated reference list — exactly what reviewers expect from a single-file SEP submission.

Working files in this repo remain split for editing efficiency (profile-v0.md evolves separately from the SEP cover material, with its own change log and version stamps). The merge is a one-time mechanical step at SEP-file submission and is owned by the SEP author, not by reviewers. The script `scripts/build-sep.sh` (already implemented) performs the concatenation deterministically and writes `build/sep-mclip-standards-track.md`; re-run it whenever either source changes during review.

This addresses the structural concern that a SEP draft pointing at companion files cannot be evaluated as a SEP submission — at submission, the artefact is `build/sep-mclip-standards-track.md`, a single self-contained file (~1130 lines as of draft 2.2). The working split form is internal repo convenience only.

## Motivation

MCP's promise to users is *"if a service has adopted MCP, you automatically know how to interact with it."* In an LLM-mediated chat context, that promise holds: clients translate user intent into `tools/call` requests transparently. In a scripting / CLI context, it doesn't.

An audit of the eight currently-published MCP-to-CLI wrappers (`wrapper-audit.md`, 2026-05-15) found five concrete portability failures across the wrapper set: argument-passing style, JSON output flag + shape, invocation pattern, array encoding, and server reference. A bash script written against one wrapper silently fails or errors when run against another. Today, the MCP CLI promise holds *within* each wrapper, not *across* them.

The wrapper space has organic convergence on only two narrow points: `--json` as the plurality choice for the JSON output flag (4 of 8 wrappers), and consumption of the de-facto `~/.vscode/mcp.json` configuration format (2 of 8 wrappers, independently). Nothing else. There is no evidence of deliberate cross-wrapper coordination.

This SEP fixes that with a profile any wrapper can adopt as an `--mclip` mode (or equivalent), without replacing its existing UX.

## Specification

The full normative specification is `profile-v0.md`. This section summarises what is in scope, what each conformance level guarantees, and how rule IDs are organised so reviewers can navigate the spec efficiently.

### Conformance levels

| Level | Status | Covers |
|---|---|---|
| **MCLIP-Core** | Required | Command shape (§1), flag conventions (§2), stdin handling (§3), output formats text + json (§4), JSON envelope (§5), exit codes (§6), pagination (§10), stdio transport (§12 partial), MCLIP-specific server discovery (§13), Core safety baseline (§14.0), CI-safe composition (§14.4), conformance machinery (§15). |
| **MCLIP-Resources** | Optional | `resources` verbs: `list`, `read`, `templates`, `watch` (§7). |
| **MCLIP-Prompts** | Optional | `prompts` verbs: `list`, `get` (§8). |
| **MCLIP-HTTP** | Optional | Streamable HTTP transport (§12 HTTP item, §12.2, §12.3 HTTP semantics). |
| **MCLIP-Streaming** | Optional | Progress notifications + NDJSON streaming output (§9). |
| **MCLIP-Safety** | Optional | Confirmation prompt UX, `--dry-run`, `--force`, trusted-server escape hatch (§14.1–§14.3). |
| **MCLIP-Auth** | Optional | Credential resolution (keychain → per-server env → config), Bearer transmission (§11). |
| **MCLIP-Metadata** | Optional | `mclip.*` `_meta` keys per the companion Extensions Track SEP (§16). |
| **MCLIP-Discovery** | Optional | Inherited config files from Claude Desktop / `.vscode/mcp.json` / `.cursor/mcp.json` (§13.2). |

A claim takes the form *"X conforms to MCLIP v0 — Core + <modules>"*. Core alone is a valid claim; any module without Core is invalid.

### High-leverage rules

The following rules address the five fragmentation points from `wrapper-audit.md` directly:

| Audit failure | Resolving rule(s) |
|---|---|
| Argument-passing style (4 incompatible styles) | `[MCLIP-2-06]` long-flag-from-schema lower-kebab-case + `[MCLIP-2-07]` `--arg-` prefix for reserved-flag collisions + `[MCLIP-2-10]` `--input`/`--input-file` fallback |
| JSON output flag + shape | `[MCLIP-5-01]` `{ "result": ... }` envelope + `[MCLIP-5-02]` MCP result verbatim + `[MCLIP-5-03]` `{ "error": ... }` shape with `origin` discriminator |
| Invocation pattern (5 incompatible shapes) | `[MCLIP-1-02]` canonical shape `<binary> [global-flags] <server> <category> <verb> [target] [command-flags]` + `[MCLIP-1-07]` global-flag-placement rule + `[MCLIP-1-06]` category-then-verb ordering |
| Array encoding (4 incompatible forms) | `[MCLIP-2-09]` repeated-flag form MUST be accepted; comma-separated MAY be accepted but not mandated |
| Server reference (no shared convention) | `[MCLIP-13-01]` config-source priority + `[MCLIP-13-04]` `mcpServers`/`servers` synonym + `[MCLIP-13-02]` inherited-config support (MCLIP-Discovery) |

### Conformance fixture catalogue

The conformance test suite (`conformance-fixtures.md`, ~30 fixtures grouped by 8 themes) targets each rule ID by tag. Implementations submit per-module pass/fail results; the harness emits JUnit XML / JSON so wrapper maintainers can publish per-module conformance badges. The fixture servers themselves are spec'd in `fixtures-spec.md`.

This SEP fulfils the SEP-2484 conformance-test-requirement for Standards Track SEPs: every MUST in the profile maps to one or more fixtures, and the fixture set is buildable before the SEP reaches Final status.

### Version and MCP-compatibility policy

Per `[MCLIP-15-05]` through `[MCLIP-15-11]`, this profile declares an explicit baseline MCP version (`2025-11-25` for v0) and does NOT auto-advance with MCP spec revisions. A new MCP spec revision triggers a profile revision that explicitly enumerates which MCLIP rules change. The `--version` JSON output (`[MCLIP-15-09]`) is the normative machine-readable interface for runtime version checks.

## Rationale

### Why a profile rather than a new protocol or a new wrapper?

Profiles have direct precedent: A2DP/HFP on Bluetooth, OIDC on OAuth, WebAuthn on CTAP, FHIR profiles on FHIR. Each is "if you expose X capability over Y surface, these are the rules." MCLIP is structurally identical: "if you expose MCP capability over a command line, these are the rules." A new protocol would require vendor adoption (a higher bar). A new wrapper would be wrapper #9, competing for users with eight existing tools — the very fragmentation this SEP is solving.

### Why MCLIP-Core + 8 optional modules, not one monolithic claim?

A wrapper that already speaks HTTP, exposes resources, and ignores prompts shouldn't be barred from claiming useful MCLIP conformance because it doesn't do prompts. Modular conformance lets each wrapper claim what it credibly supports and incrementally add modules. `[MCLIP-15-02]` makes Core required so a claim is always anchored to a shared baseline.

### Why specify the JSON envelope shape, not just the flag name?

`wrapper-audit.md` found that 4 of 8 wrappers agree on `--json` as the flag, but disagree on the resulting *shape* (`mcp2cli` wraps in a custom `app_id`/`summary` envelope; `mcpc` enforces MCP-spec schemas; others emit raw tool results). A consumer parsing one wrapper's JSON output cannot use the same parser against another. Specifying both the flag and the shape is the only way to make the "JSON output is portable" claim true.

### Why `readOnlyHint == true` as the sole positive safe signal?

MCP spec explicitly warns clients to treat tool annotations as untrusted unless they come from trusted servers. The asymmetric design — only `readOnlyHint == true` is positive, `destructiveHint == false` is not — means a hostile server cannot bypass safety by lying about destructiveness. The trusted-server escape hatch (`safety.trustAnnotations` per `[MCLIP-14-07]`) is the user-controlled mechanism for opting into relaxed defaults; it is per-server, never server-supplied.

### Why a separate Extensions Track SEP for the `mclip.*` metadata keys?

Per SEP-2133, Extensions Track is the canonical home for additive vocabularies that don't modify the wire protocol. Decoupling the metadata keys lets the main profile (this SEP) advance through review without being held up by key-by-key bikeshed. The main profile reserves the namespace (§16); the Extensions SEP carries the catalogue. Both can advance independently.

### Why direct individual-contributor filing instead of a WG?

The Transports WG charter explicitly excludes application-layer profiles. No tooling/hosting WG exists. SEP-986 (tool name format conformance) and SEP-1730 (SDK tiering) are direct structural precedents: Standards Track SEPs specifying implementor-side conformance without wire-protocol changes, filed by individual contributors. See `governance-recommendation.md` for the full analysis.

## Backwards Compatibility

### Effect on MCP servers

**None.** This profile is consumed entirely by CLI clients. MCP servers continue to expose `tools/list`, `resources/list`, `prompts/list`, schemas, annotations, and capabilities exactly as the MCP spec defines. Vendors who opt into the companion Extensions Track SEP for `mclip.*` metadata get a better CLI UX; vendors who don't still get a fully functional canonical surface.

### Effect on existing MCP-to-CLI wrappers

**Opt-in.** A wrapper that doesn't adopt MCLIP continues to operate exactly as it does today. A wrapper that adopts MCLIP adds a parallel `--mclip` mode (or equivalent sub-command) per `adoption-guide.md`; its existing surface is untouched.

There is a one-time decision cost for adopting wrappers, priced honestly in `adoption-guide.md` by wrapper shape: days of work for wrappers whose invocation shape is close to canonical; one to two weeks for moderate-divergence wrappers; several weeks for wrappers with fundamental shape incompatibility (server-as-binary, session-prefix). The wrapper's existing UX continues to work unchanged after adoption.

### Effect on MCP-protocol future revisions

This profile pins to MCP 2025-11-25 as its declared baseline (`[MCLIP-15-05]`). Future MCP spec revisions do not auto-advance the profile baseline. When MCP publishes a new revision, a follow-on MCLIP revision explicitly tracks which rules change. Wrappers that stay on the older MCP baseline remain conformant to the older MCLIP profile.

## Reference Implementation

The reference CLI is being developed in Go using the official Tier 1 Go MCP SDK (`github.com/modelcontextprotocol/go-sdk` v1.6.0). Architecture: `reference-cli-architecture.md`. Status at time of SEP filing: architectural spec complete; coding begins after sponsor signal per `problem-statement.md` execution sequence.

The reference implementation:
- Implements MCLIP-Core plus all eight optional modules (with MCLIP-Metadata behind a build tag pending the companion Extensions SEP).
- Targets reproducible, statically-linked binaries for Linux / macOS / Windows.
- Passes the entire conformance fixture catalogue (`conformance-fixtures.md`) against the synthetic fixture servers spec'd in `fixtures-spec.md`.
- Validates against three real MCP servers (`real-mcp-servers.md`): `@modelcontextprotocol/server-everything` (protocol-feature exerciser), GitHub MCP Server (auth + destructive validation), Context7 MCP Server (real-SaaS smoke test with API-key header auth via the MCLIP-Auth credential path).

Per MCP SEP guidelines, Final status requires a working reference implementation. The implementation timeline is tied to the SEP review cadence: scaffolding begins on sponsor acceptance; Core + Auth completes before Accepted status; full module coverage before Final.

## Security Considerations

A dedicated companion document — `security-model.md` (draft 1.1) — covers MCLIP's full trust boundaries, destructive-action policy, non-interactive (CI-safe) behaviour, credential handling, project-local config trust, output redaction, and auditability. Highlights:

- **Project-local config consent rule** (`[MCLIP-13-06]`): credential-bearing or `safety.trustAnnotations`-bearing entries in any project-rooted file (including inherited `./.vscode/mcp.json` / `./.cursor/mcp.json`) require explicit user consent stored outside the project directory. Non-interactive clients MUST refuse rather than prompt.
- **No generic credential surface** (`[MCLIP-11-06]`, `[MCLIP-11-07]`): no generic `--auth-token` flag; no generic `MCLIP_TOKEN` env var; credentials are scoped per-server alias.
- **CI-safe rollup** (§14.4): `isatty(stdin)` is the only signal for non-interactive mode (env-var-only opt-in MUST NOT override true TTY state); no prompts in non-interactive mode for any reason; no secrets in stdout/stderr; deterministic JSON/NDJSON output; specific exit codes (generic `1` reserved for unspecified failures only).
- **Tightening-only metadata semantics** (`[MCLIP-16-03]` + companion Extensions SEP): server-supplied `mclip.*` metadata MAY raise safety bars (`mclip.destructive: true` adds the destructive treatment) but MUST NOT lower them.

The full conformance security checklist is in `security-model.md`.

## Open Issues

- **Naming.** The MCLIP name has been sanity-checked against trademarks, software ecosystems, and known ML-community usage (`naming-check.md`). Verdict: safe to proceed; OpenAI's CLIP collision is initials-only with strong audience separation. No blockers.
- **Wrapper-maintainer reading-pass.** The `prd.md` §6 "at least one wrapper-maintainer publicly engaged" success criterion is unmet at time of filing. Outreach is in flight per `wrapper-outreach-plan.md`; reading-pass evidence will be added to this SEP's review thread as it comes in.
- **Sponsor.** Three Core Maintainer candidates identified in `sponsor-candidates.md`: `@pja-ant` (primary), `@pcarleton` (secondary, conformance angle), `@kurtisvg` (tertiary, transport-adjacent). Tagged on the draft PR per the problem-statement execution sequence.

## References

- Profile (full normative text): `profile-v0.md` (draft 2.2)
- Security model: `security-model.md` (draft 1.1)
- Conformance fixture catalogue: `conformance-fixtures.md` (v0.2)
- Fixture-server implementation spec: `fixtures-spec.md`
- Reference-CLI architecture: `reference-cli-architecture.md`
- Wrapper audit (motivation evidence): `wrapper-audit.md`
- Maintainer-facing comparison: `wrapper-comparison.md`
- Adoption guide: `adoption-guide.md`
- Companion Extensions Track SEP: `sep-extensions-mclip-metadata.md`
- Governance, SEP workflow, working-group, and maintainer references: `prd.md` §13
