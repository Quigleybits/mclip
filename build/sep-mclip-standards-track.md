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

Working files in this repo remain split for editing efficiency (profile-v0.md evolves separately from the SEP cover material, with its own change log and version stamps). The merge is a one-time mechanical step at SEP-file submission and is owned by the SEP author, not by reviewers. A short script under `scripts/build-sep.sh` (TBD as part of the SEP submission preparation) will perform the concatenation deterministically and re-run if either source changes during review.

This addresses the structural concern that a SEP draft pointing at companion files cannot be evaluated as a SEP submission — at submission, the artefact will be one self-contained file. The working split form is internal repo convenience only.

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

The reference CLI is being developed in Go using the official Tier 1 Go MCP SDK (`github.com/modelcontextprotocol/go-sdk` v1.6.0). Architecture: `mclio-architecture.md`. Status at time of SEP filing: architectural spec complete; coding begins after sponsor signal per `problem-statement.md` execution sequence.

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



**Profile specification, draft v0 — revision 2.2**
Author: Aidan Quigley
Date: 2026-05-16
Status: **Draft** — pre-SEP, pre-implementation
Baseline MCP version: `2025-11-25` ([spec](https://modelcontextprotocol.io/specification/2025-11-25))

---

## 0. Conventions

### 0.1 Normative language

The key words **MUST**, **MUST NOT**, **REQUIRED**, **SHALL**, **SHALL NOT**, **SHOULD**, **SHOULD NOT**, **RECOMMENDED**, **MAY**, and **OPTIONAL** in this document are to be interpreted as described in BCP 14 [RFC 2119] [RFC 8174] when, and only when, they appear in all capitals.

### 0.2 Conformance markers

Every normative rule in this document carries a conformance marker of the form `[MCLIP-§-NN]` where `§` is the section number and `NN` is the rule number within that section. The conformance test suite (separate deliverable) targets these markers by ID.

### 0.3 Scope

This profile applies to **CLI implementations** that translate MCP server capabilities into command-line invocations. It does not modify the MCP specification itself.

MCLIP v0 is structured as a required **MCLIP-Core** level plus eight independently-claimable optional modules — see §0.7 for the conformance-level table and §15 for the claim format.

The companion CLI-metadata extension (§16) reserves the `mclip.*` namespace under MCP's `_meta` field. Specific keys, types, and semantics are published as a separate **Extensions Track SEP** rather than inline in this profile. MCLIP-Core MUST function correctly against vanilla MCP servers that publish no `mclip.*` metadata; a client claims MCLIP-Metadata only when it implements the Extensions SEP's keys.

### 0.4 Terminology

- **MCP server**: a process implementing the Model Context Protocol per the 2025-11-25 spec.
- **MCLIP client**: a CLI implementation conforming to this profile. Synonymous with **wrapper** at the profile level; "wrapper" emphasises the multi-implementation ecosystem (`mcp2cli`, `MCPorter`, `MCPShim`, `f/mcptools`).
- **Wrapper**: a tool that translates MCP into a CLI surface. MCLIP defines the rules wrappers MUST follow to produce a uniform surface.
- **Capability area**: one of `tools`, `resources`, `prompts` (per MCP spec).
- **Tool result**: the `CallToolResult` returned by `tools/call`, containing `content[]`, optionally `structuredContent`, and `isError`.
- **Canonical surface**: the command shape, flags, output schema, exit codes, and error envelope that every MCLIP-conformant client MUST produce for the same MCP server. The output of this profile applied to an MCP server.
- **Conformance**: an implementation's adherence to every MUST rule in MCLIP-Core plus every MUST rule in each claimed module. See §15.
- **Conformance level**: a named subset of this profile claimable independently. MCLIP-Core is required; the eight modules in §0.7 are optional.
- **Module**: a named optional conformance level (e.g. MCLIP-Resources, MCLIP-Safety). Each module is self-contained: its rules activate only when the module is claimed.
- **Extension**: an additive specification outside this profile. The CLI-metadata extension (§16) is the only extension referenced by v0; it lives in a companion Extensions Track SEP and MUST NOT be required for Core conformance.
- **Fixture**: a synthetic MCP server or tool definition used by the conformance test suite (separate deliverable) to exercise specific rules. The fixture corpus covers the audit's known fragmentation points plus the canonical scenarios per module.

### 0.5 MCP as source of truth

`[MCLIP-0-01]` MCLIP clients MUST derive their command surface from MCP server discovery methods (`tools/list`, `resources/list`, `prompts/list`, their associated schemas, `annotations`, and the server's advertised `capabilities`). They MUST NOT require vendor-specific command definitions for Core conformance (§0.7).

> Rationale: the foundational promise of MCLIP is "if a company adopts MCP, a user or agent automatically knows how to interact with it via CLI." That promise only holds if discovery is the sole input to the command surface. Hardcoded per-vendor command tables would break it.

### 0.6 Profile scope: mechanics, not semantics

This profile standardises the **mechanics** of CLI access over MCP — command shape, flag mapping, output envelope, exit codes, and error model. It does **not** standardise **domain semantics** — tool names, verbs, or argument names across services.

In practice this means:

```
mclip linear tools call create-comment --issue-id ENG-123
mclip github tools call create-issue --title "..."
```

will follow the same grammar, flag-mapping rules, output shape, and error semantics. It does **not** mean every service exposes the same semantic verbs (e.g. a common `create-issue` or `delete-user` across vendors). MCP itself does not standardise domain tool names, so MCLIP cannot honestly promise that. A cross-vendor semantic-verb layer is out of scope for v0 and would require a separate profile built on top of MCLIP.

### 0.7 Conformance levels

MCLIP v0 defines one required level and eight optional modules. An implementation claims **MCLIP-Core** if it passes every MUST in the Core sections. It claims a module by additionally passing every MUST in that module's section(s). Modules are independent; claiming MCLIP-Resources does not require claiming MCLIP-Prompts. See §15 for the canonical claim format.

| Level | Required? | Covers |
|---|---|---|
| **MCLIP-Core** | Yes | §1 (Command shape — `tools` and `meta` verbs only), §2 (Argument and flag conventions), §3 (Stdin handling), §4 (Output formats — `text` and `json`; `ndjson` flag is defined but only required when MCLIP-Streaming is claimed), §5 (JSON envelope), §6 (Exit codes), §10 (Pagination), §12.1 stdio item + §12.4–§12.5 (Transport errors and timeout), §13.1, §13.3, §13.4 (MCLIP-specific server discovery), §14.0 (Minimum safety baseline), §14.4 (Non-interactive / CI-safe composition rollup), §15 (Conformance). |
| **MCLIP-Resources** | No | §7 (Resources verbs: `list`, `read`, `templates`, `watch`). |
| **MCLIP-Prompts** | No | §8 (Prompts verbs: `list`, `get`). |
| **MCLIP-HTTP** | No | Streamable HTTP transport in §12: HTTP item of §12.1, §12.2 (protocol-version header), and HTTP semantics of §12.3 (`--transport http`). |
| **MCLIP-Streaming** | No | §9.1 (Progress notifications) and §9.2 (NDJSON event output for profile-defined streaming commands such as `resources watch`). Incremental `tools/call` result streaming is deferred until MCP provides a deterministic signal for it. |
| **MCLIP-Safety** | No | §14.1–§14.3 (Confirmation prompt UX, `--dry-run`, `--force` semantics, `mclip.confirm_message` support, trusted-server escape hatch). Builds on the Core safety baseline (§14.0). |
| **MCLIP-Auth** | No | §11 (Credential resolution, env-var conventions, Bearer header transmission). |
| **MCLIP-Metadata** | No | §16 (Honouring `mclip.*` keys per the companion Extensions Track SEP). |
| **MCLIP-Discovery** | No | §13.2 (Inherited config files from Claude Desktop, `.vscode/mcp.json`, `.cursor/mcp.json`). |

`[MCLIP-0-02]` A MCLIP-Core-conformant client MUST function correctly against vanilla MCP servers that publish no `mclip.*` metadata. Modules MAY extend behaviour within their own scope; they MUST NOT relax safety defaults (per §14.0 and §16.3).

Task-augmented execution (`--follow`, `--detach`, the `meta tasks *` subcommands previously sketched in §9.3) is **deferred to MCLIP v1** pending stabilisation of MCP's `tasks/*` API. Revision 2 of v0 removes these from the v0 surface entirely.

---

## 1. Command shape

### 1.1 Root binary name

`[MCLIP-1-01]` The reference implementation MUST be invoked as `mclip`. Implementations that are not the reference (e.g. existing wrappers adding an "MCLIP mode") MAY use a different binary name but MUST accept the canonical command shape defined in §1.2 under their own binary or under an `mclip` subcommand or alias.

### 1.2 Canonical invocation shape

`[MCLIP-1-02]` Commands MUST follow the shape:

```
<binary> [global-flags...] <server> <category> <verb> [target] [command-flags...]
<binary> [global-flags...] <root-subcommand> ...
```

Where:
- `[global-flags]` are profile-defined client flags that affect parsing, config, transport, output, or diagnostics before the server command is resolved: `--config`, `--output` / `-o`, `--raw`, `--no-color`, `--verbose` / `-v`, `--quiet` / `-q`, `--timeout`, and `--transport`
- `<server>` is the server alias (per §13 discovery rules)
- `<category>` is one of `tools`, `resources`, `prompts`, or `meta`
- `<verb>` is a category-specific verb (§1.3)
- `[target]` is a tool name, URI, or prompt name as required by the verb
- `[command-flags]` are command-specific profile flags plus generated tool or prompt argument flags; generated flags MUST appear after the command target when a target exists

`[MCLIP-1-07]` Conformant clients MUST accept the global-flag placement shown above. They MAY additionally accept global flags after the command target, but conformance tests MUST use the canonical pre-server placement. Tool-generated flags MUST NOT be accepted before `<server>`, because they cannot be interpreted before the selected tool schema is known.

### 1.3 Reserved category verbs

`[MCLIP-1-03]` Conformant clients MUST implement the following verbs. The Module column identifies which conformance level requires the verb. Verbs belonging to a module that is not claimed MUST NOT be advertised as MCLIP-conformant by that implementation.

| Category | Verb | Target | Module | Description |
|---|---|---|---|---|
| `tools` | `list` | — | Core | Enumerate available tools |
| `tools` | `call` | tool name | Core | Invoke a tool with the given inputs |
| `tools` | `describe` | tool name | Core | Print the tool's full schema and annotations |
| `meta` | `ping` | — | Core | Verify server reachability |
| `meta` | `info` | — | Core | Print server identity and capabilities |
| `meta` | `version` | — | Core | Print the server's reported protocol version |
| `resources` | `list` | — | Resources | Enumerate available resources |
| `resources` | `read` | URI | Resources | Read a resource's contents |
| `resources` | `templates` | — | Resources | List resource URI templates |
| `resources` | `watch` | URI | Resources | Subscribe to update notifications (§7.4) |
| `prompts` | `list` | — | Prompts | Enumerate available prompts |
| `prompts` | `get` | prompt name | Prompts | Retrieve a prompt with arguments |

### 1.4 Root-level subcommands

`[MCLIP-1-04]` Conformant clients MUST implement these root-level subcommands (not bound to any server):

| Command | Purpose |
|---|---|
| `<binary> servers list` | Enumerate known servers from the config |
| `<binary> servers add <name> ...` | Add a server to the user config |
| `<binary> servers remove <name>` | Remove a server from the user config |
| `<binary> --help` / `-h` | Print top-level help and exit 0 |
| `<binary> --version` | Print the implementation version, MCLIP profile version, MCP baseline, and claimed modules per §15.4.3, exit 0 |

`[MCLIP-1-08]` Server aliases MUST NOT be any reserved root command name: `servers`, `help`, or `version`. Server aliases MUST NOT begin with `-`. Implementations MAY reject additional aliases that are not portable as shell arguments, but they MUST reject at least these names to avoid ambiguity with root-level commands and flags.

### 1.5 Help

`[MCLIP-1-05]` Every subcommand MUST accept `--help` and `-h`. Passing either MUST cause the client to print help text to stdout and exit 0. This includes invalid subcommands paired with `-h`.

> Rationale: clig.dev "Display help when passed `-h` or `--help` flags. Ignore any other flags and arguments that are passed."

### 1.6 Verb-noun vs noun-verb

`[MCLIP-1-06]` MCLIP uses **category-then-verb** ordering (`tools list`, `tools call`, `resources read`). Implementations MUST NOT define equivalent verb-then-category top-level aliases (e.g. `mclip list tools`) at the conformance-tested surface, to ensure script portability. Users MAY add their own shell aliases.

> Rationale: Modern noun-verb shape (docker, gcloud, aws, gh) lets the three MCP capability areas namespace their verbs symmetrically. The audit (`wrapper-audit.md`) found five distinct invocation shapes across existing wrappers; locking one is the highest-leverage portability fix.

---

## 2. Argument and flag conventions

### 2.1 Long options

`[MCLIP-2-01]` Conformant clients MUST accept both `--flag value` and `--flag=value` forms for any flag that takes a value (POSIX-style and GNU-style). They MUST treat the two as equivalent.

> Reference: POSIX.1-2017 XBD §12.2 + GNU `getopt_long` convention.

### 2.2 Short options

`[MCLIP-2-02]` Where this profile defines a short option (e.g. `-h`, `-o`, `-y`, `-q`, `-v`), conformant clients MUST accept it as a single-letter alias for the long form.

`[MCLIP-2-03]` Implementations MAY accept POSIX-style short option combination (`-abc` ≡ `-a -b -c`) but this profile does not require it. Tests targeting MCLIP MUST NOT rely on combination.

### 2.3 End-of-options separator

`[MCLIP-2-04]` Conformant clients MUST honour `--` as the end-of-options separator (POSIX Guideline 10). Arguments after `--` MUST be treated as positional, even if they begin with `-`.

### 2.4 Stdin sentinel

`[MCLIP-2-05]` Conformant clients MUST accept the single character `-` as a positional argument or flag value to mean "read from stdin", per POSIX Guideline 13. This applies wherever the schema accepts a string, file path, or input blob.

### 2.5 Schema property → flag name mapping

`[MCLIP-2-06]` For each property in a tool's `inputSchema`, MCLIP clients MUST generate a flag name by converting the property name to **lower-kebab-case**:

- `snake_case` → `--snake-case`
- `camelCase` → `--camel-case`
- `PascalCase` → `--pascal-case`
- `already-kebab` → `--already-kebab`

`[MCLIP-2-07]` Generated flag names MUST NOT collide with reserved flag names (Appendix A). If a tool's input property would collide, the client MUST prefix the generated flag with `--arg-` (e.g. `output` → `--arg-output`).

`[MCLIP-2-13]` If two or more schema properties produce the same generated flag name after lower-kebab-case conversion (for example `fooBar`, `foo_bar`, and `foo-bar` all producing `--foo-bar`), the client MUST NOT choose one property implicitly. All colliding properties MUST be supplied through `--input` or `--input-file`; an invocation using the ambiguous generated flag MUST exit `64` and name the colliding properties.

### 2.6 Type mapping

`[MCLIP-2-08]` JSON-Schema types map to CLI inputs as follows:

| JSON-Schema | CLI form | Example |
|---|---|---|
| `string` | `--flag <value>` | `--name "alice"` |
| `integer` / `number` | `--flag <value>` | `--count 5` |
| `boolean` | `--flag` / `--no-flag` | `--verbose` / `--no-verbose` |
| `string` w/ `enum` | `--flag <enum-value>` | `--colour red` (client validates) |
| `array` of primitives | repeated `--flag <item>` | `--tag bug --tag urgent` |
| `array` of objects | use `--input` (§2.7) | — |
| `object` (1 level, primitive leaves only) | dotted flags | `--metadata.role admin` |
| `object` (2+ levels) | use `--input` (§2.7) | — |
| `null` | `--flag null` literal | — |

`[MCLIP-2-09]` Conformant clients MUST accept repeated-flag form for array properties. They MUST NOT mandate comma-separated form. They MAY additionally accept comma-separated form when the array's `items.type` is `string` and no string value contains a comma.

> Rationale: audit found four incompatible array encodings. Repeated flags are the only form safe for arbitrary values. POSIX Guideline 8 permits but does not mandate comma-separation.

`[MCLIP-2-14]` For JSON Schema constructs that do not have an unambiguous flag mapping in this profile (`oneOf`, `anyOf`, `allOf`, `not`, `const`, `additionalProperties`, `patternProperties`, tuple-style arrays, property names containing whitespace, and property names containing literal `.`), clients MUST accept `--input` and `--input-file` as the portable representation. Implementations MAY expose additional convenience flags for these shapes, but those flags are outside Core conformance and MUST NOT be required by the conformance suite.

`[MCLIP-2-15]` A nullable schema of the form `type: ["T", "null"]` maps as type `T` for flag generation. The exact flag value `null` represents JSON null only when the schema permits null. When a string schema must distinguish the string value `"null"` from JSON null, callers MUST use `--input` or `--input-file`.

### 2.7 Complex input fallback

`[MCLIP-2-10]` Conformant clients MUST accept `--input <json>` and `--input-file <path>` flags on `tools call`. These provide the entire tool input as a JSON object, bypassing flag-level mapping. `--input` and `--input-file` are mutually exclusive. Neither flag may be mixed with individual property flags; mixing MUST exit `64` (usage error).

`[MCLIP-2-16]` Tool input values MUST come only from explicit invocation input: property flags, `--input`, or `--input-file`. Config files and environment variables MUST NOT supply tool argument defaults in the conformant surface. Within property flags, repeated array flags append in encounter order; for non-array properties repeated flags use the last occurrence; for boolean `--foo` / `--no-foo`, the last occurrence wins.

### 2.8 Boolean negation

`[MCLIP-2-11]` For any boolean flag `--foo` generated from a schema property or defined in this profile, the form `--no-foo` MUST set the value to `false`. When neither `--foo` nor `--no-foo` is given for a generated schema property, the client MUST omit that property from the locally constructed input object. Servers MAY apply their own defaults; Core clients MUST NOT synthesize schema `default` values into tool input.

### 2.9 Required vs optional

`[MCLIP-2-12]` Properties listed in the tool input schema's `required` array MUST be supplied via flag, `--input`, or `--input-file`. Missing required input MUST cause exit `64` with a clear error message naming each missing property.

---

## 3. Stdin handling

### 3.1 Stdin sentinel for flag values

`[MCLIP-3-01]` When a flag value is exactly `-`, the client MUST read the value from stdin. Reading is greedy: the entire stdin contents (up to EOF) is treated as the value.

### 3.2 `--input -`

`[MCLIP-3-02]` `--input -` MUST read a complete JSON object from stdin and treat it as the tool input.

### 3.3 TTY guard

`[MCLIP-3-03]` If the client requires stdin but stdin is a TTY, the client MUST NOT block silently. It MUST print a one-line indication to stderr (e.g. `mclip: reading input from stdin (Ctrl-D to end)`) before blocking.

> Reference: clig.dev "If your command is expecting to have something piped to it and stdin is an interactive terminal, display help immediately and quit." MCLIP softens this to a prompt-line because some interactive workflows are legitimate.

---

## 4. Output formats

### 4.1 Format selection

`[MCLIP-4-01]` The flag `--output <format>` (short: `-o <format>`) selects output format. Valid values are `text`, `json`, and `ndjson`. The default is:

- `text` when stdout is a TTY and the command is non-streaming
- `json` when stdout is not a TTY and the command is non-streaming
- `ndjson` for streaming commands (§9) regardless of TTY

`[MCLIP-4-02]` Conformant clients MUST accept all three values. Other format aliases (`yaml`, `table`, etc.) MAY be supported but MUST NOT be required.

### 4.2 Text format

`[MCLIP-4-03]` `text` output is intended for human readers. The exact rendering is not normative, but conformant clients MUST:

- Print primary content to stdout
- Print progress, status, and log messages to stderr
- Render `result.content[]` text parts in a readable form
- Honour `NO_COLOR` env var and `--no-color` flag (§A)
- Detect TTY and disable colour automatically when not a TTY

### 4.3 JSON format

`[MCLIP-4-04]` `json` output is intended for machine consumption. See §5.

`[MCLIP-4-07]` The flag `--raw` is an opt-in modifier for `--output json`. On successful commands, `--raw -o json` MUST print the MCP result object directly instead of the MCLIP success envelope. Errors MUST still use the standard error envelope (§5.2), including tool-reported errors (`result.isError == true`), so failure handling remains portable. `--raw` with `text` or `ndjson` output MUST exit `64`.

### 4.4 NDJSON format

`[MCLIP-4-05]` `ndjson` output emits one complete JSON value per line, separated by `\n`. Each line MUST be independently parseable. Used for streaming (§9).

### 4.5 stdout vs stderr separation

`[MCLIP-4-06]` Conformant clients MUST send primary command output (the result envelope or its text rendering) to **stdout** only. They MUST send progress, log, diagnostic, and confirmation-prompt output to **stderr** only.

> Reference: clig.dev + POSIX convention. This is non-negotiable for shell pipeability.

---

## 5. JSON output envelope

### 5.1 Success envelope

`[MCLIP-5-01]` On successful execution, JSON output MUST be a single JSON object with the shape:

```json
{
  "result": <MCP result, verbatim>
}
```

For `tools call`, `<MCP result>` is the `CallToolResult` per MCP spec, including `content[]`, optionally `structuredContent`, and `isError`. For `resources read`, it is the `ReadResourceResult`. For `prompts get`, it is the `GetPromptResult`. For `*/list` commands, it is the list response including `nextCursor` if present.

`[MCLIP-5-02]` Clients MUST NOT alter, reorder, or strip fields from the MCP result. The `result` key is the only added wrapper.

### 5.2 Error envelope

`[MCLIP-5-03]` On error, JSON output MUST be a single JSON object with the shape:

```json
{
  "error": {
    "code": <integer>,
    "message": <string, human-readable>,
    "origin": <"server"|"transport"|"client"|"tool">,
    "data": <freeform JSON, optional>
  }
}
```

`[MCLIP-5-04]` `origin` MUST be one of:
- `"server"` — the MCP server returned a JSON-RPC error
- `"transport"` — the transport layer failed (connection refused, timeout)
- `"client"` — the CLI itself errored (config invalid, usage error, schema validation)
- `"tool"` — the tool executed but returned `isError: true` in its result

`[MCLIP-5-05]` When `origin == "server"`, the `code` and `data` MUST be the JSON-RPC error code and `data` field verbatim from the server response.

`[MCLIP-5-06]` When `origin == "tool"`, the envelope MUST be:

```json
{
  "result": <CallToolResult with isError: true>,
  "error": {
    "code": 100,
    "message": "Tool reported error",
    "origin": "tool"
  }
}
```

Both `result` and `error` keys are present so consumers can read the tool's own diagnostics from `result.content[]`.

### 5.3 No protocol noise

`[MCLIP-5-07]` Clients MUST NOT include `jsonrpc`, `id`, or other JSON-RPC protocol-level fields in CLI output. The envelope is for the CLI consumer, not for the wire.

---

## 6. Exit codes

### 6.1 Canonical codes

`[MCLIP-6-01]` Conformant clients MUST exit with one of the following codes:

| Code | Meaning | Trigger |
|---|---|---|
| `0` | Success | Normal completion |
| `1` | Generic error | Unspecified failure not covered below |
| `64` | Usage error | Bad CLI args, mixing `--input` with property flags, missing required input |
| `65` | Data error | `--input` JSON failed to parse, schema validation failed |
| `66` | Input file missing | `--input-file` path does not exist or is unreadable |
| `69` | Server unavailable | Cannot connect to MCP server; transport error |
| `70` | Internal client error | Bug in MCLIP client; should be reported |
| `75` | Temporary failure | Retryable error (rate limit, transient I/O) |
| `77` | Permission denied | Authentication or authorization failure |
| `78` | Config error | mclip.json malformed, server name unknown |
| `100` | Tool reported error | MCP call succeeded but `result.isError == true` |
| `130` | Interrupted | SIGINT received and acted upon |

`[MCLIP-6-02]` Code values `0` and `1` follow universal CLI convention. Codes 64–78 are derived from BSD `sysexits.h` and remain the most precedented category-error codes in published practice despite their formal deprecation. Code `100` is MCLIP-specific. Code `130` follows the common shell convention of `128 + SIGINT`.

### 6.2 SIGINT

`[MCLIP-6-03]` On SIGINT (Ctrl-C), the client MUST attempt to send `notifications/cancelled` to the MCP server per MCP spec §6.1 (cancellation is fire-and-forget). The client MUST NOT block on acknowledgement. After sending (or skipping if no in-flight request), the client MUST exit `130`.

### 6.3 SIGPIPE

`[MCLIP-6-04]` On SIGPIPE (downstream pipe closed), clients SHOULD exit silently with code `0` if the closure occurred during normal output flush, or `141` (128 + SIGPIPE) if the producer state was indeterminate. This matches GNU coreutils convention.

---

## 7. Resources

### 7.1 List

`[MCLIP-7-01]` `<binary> <server> resources list` MUST issue `resources/list` and return the response. It MUST support `--cursor <opaque>`, `--limit <N>`, and `--no-paginate` per §10.

### 7.2 Read

`[MCLIP-7-02]` `<binary> <server> resources read <uri>` MUST issue `resources/read { uri }` and return the response.

`[MCLIP-7-03]` In `text` output mode, if the response contents are text and there is exactly one content item, the client SHOULD print the text directly to stdout without the JSON envelope. JSON mode is unaffected by this rendering choice.

### 7.3 Templates

`[MCLIP-7-04]` `<binary> <server> resources templates` MUST issue `resources/templates/list` and return the response.

### 7.4 Watch

`[MCLIP-7-05]` `<binary> <server> resources watch <uri>` MUST issue `resources/subscribe { uri }` and then stream `notifications/resources/updated` events to stdout in `ndjson` format until interrupted. Each stdout line MUST be a JSON object with `type: "resource.updated"`, `uri`, and `notification` fields, where `notification` contains the server notification payload verbatim. SIGINT MUST trigger `resources/unsubscribe` and exit `130`. A server-initiated normal end of stream exits `0`.

> Note: requires server-side `capabilities.resources.subscribe: true`. Clients MUST check this capability before attempting to subscribe; otherwise exit `69` with a clear error.

---

## 8. Prompts

### 8.1 List

`[MCLIP-8-01]` `<binary> <server> prompts list` MUST issue `prompts/list` and return the response with pagination support per §10.

### 8.2 Get

`[MCLIP-8-02]` `<binary> <server> prompts get <name> [--arg-name value...]` MUST issue `prompts/get { name, arguments }` where `arguments` is constructed from `--<arg-name> <value>` flags (kebab-case-mapped from the prompt's declared argument names).

`[MCLIP-8-03]` Prompts do not carry per-argument JSON-Schema in the MCP spec — arguments are simple `{ name, description?, required? }` triples. MCLIP clients MUST treat all prompt argument values as strings.

---

## 9. Streaming

This section is MCLIP-Streaming module. Core clients MAY drop progress notifications silently; they MUST NOT print progress or partial results to stdout under any circumstances.

### 9.1 Progress notifications

`[MCLIP-9-01]` Clients claiming MCLIP-Streaming MUST include a `progressToken` in `_meta` on every outgoing request that could plausibly produce a progress notification (i.e. `tools/call`, `resources/read`, `prompts/get`). The token format is implementation-defined but MUST be unique per outgoing request.

`[MCLIP-9-02]` On receiving `notifications/progress` with a matching `progressToken`, MCLIP-Streaming clients MUST render progress to **stderr**. The rendering is non-normative but SHOULD include the `progress`, `total` (if present), and `message` (if present) fields.

`[MCLIP-9-03]` Progress output MUST NOT mix with stdout. Pipe consumers reading JSON or NDJSON from stdout MUST be unaffected by progress.

### 9.2 NDJSON streaming output

`[MCLIP-9-04]` For profile-defined streaming commands, MCLIP-Streaming clients MUST use NDJSON format (§4.4) on stdout. Each line MUST be an independently-parseable JSON object with a stable `type` discriminator. The only v0 profile-defined streaming command is `resources watch` (§7.4).

`[MCLIP-9-05]` Incremental `tools/call` result streaming is deferred from v0. Clients MUST NOT claim MCLIP-Streaming conformance merely because they expose implementation-specific streaming tool output. Such output MAY exist as an extension, but it MUST NOT change the behaviour of conformant non-streaming `tools call` invocations.

### 9.3 Cancellation

`[MCLIP-9-07]` Cancellation semantics are defined entirely in §6.2 (SIGINT triggers `notifications/cancelled` and exit `130`). Clients MUST NOT assume the server will stop processing before they exit; the cancellation is fire-and-forget. This rule applies to MCLIP-Core; MCLIP-Streaming adds no cancellation rules beyond Core.

> ID note: `[MCLIP-9-06]` remains unused after v0 draft 2 removed task-augmented streaming rules. Rule IDs are not renumbered to keep in-flight references stable.

> Task-augmented cancellation (`tasks/cancel` per MCP spec) is deferred to MCLIP v1 along with the rest of task support.

---

## 10. Pagination

### 10.1 Default behaviour

`[MCLIP-10-01]` For `*/list` commands, the default behaviour MUST be to auto-paginate: the client iterates until `nextCursor` is absent and concatenates results before returning.

### 10.2 Explicit cursor

`[MCLIP-10-02]` `--cursor <opaque>` causes the client to issue only one paginated request with the given cursor and return only that page (plus its `nextCursor`).

### 10.3 Limit

`[MCLIP-10-03]` `--limit <N>` caps the total number of items returned across auto-pagination. The client MUST stop iterating once `N` items have been collected, even if more pages exist.

### 10.4 No paginate

`[MCLIP-10-04]` `--no-paginate` causes the client to issue exactly one paginated request and return its result without iteration.

### 10.5 Cursor opacity

`[MCLIP-10-05]` Clients MUST NOT parse, modify, or persist cursors across sessions. Cursors are opaque to clients per MCP spec.

---

## 11. Auth & session behaviour

### 11.1 Token sources (priority order)

`[MCLIP-11-01]` When a server requires authentication, conformant clients MUST resolve credentials from these sources, in this priority order (first match wins):

1. OS-keychain entry under service name `mclip` and account `<server-name>` (clients SHOULD support)
2. Env var `MCLIP_TOKEN_<SERVER_NAME>` where `<SERVER_NAME>` is the server alias upper-cased with `-` replaced by `_`
3. `auth.token` field in the server's entry in the config file (§13)

`[MCLIP-11-06]` MCLIP-Auth clients MUST NOT expose a generic `--auth-token <value>` flag on the conformant invocation surface. Tokens passed as command-line arguments leak through shell history and process inspection on common platforms. Implementations MAY provide dedicated auth-management commands, but those commands are outside v0 conformance.

`[MCLIP-11-07]` MCLIP-Auth clients MUST NOT read a generic `MCLIP_TOKEN` environment variable. Tokens are scoped to one server alias through `MCLIP_TOKEN_<SERVER_NAME>` to avoid accidental wrong-server delivery.

`[MCLIP-11-09]` When MCLIP-Auth clients read a token from the `auth.token` config field, they SHOULD warn on stderr the first time per process if the config file's POSIX mode satisfies `mode & 0o077 != 0` (world- or group-readable) or — on Windows — if an equivalent ACL check shows access beyond the current user. Implementations that cannot perform the ACL check SHOULD warn unconditionally when reading a config-stored token. The warning MUST NOT contain the token value.

`[MCLIP-11-02]` Clients MUST NOT prompt for credentials on stdin during any conformant invocation defined by this profile (every verb under §1.3 across `tools`, `resources`, `prompts`, and `meta`, plus every root subcommand under §1.4). Interactive credential prompts MUST be confined to a dedicated `<binary> servers auth <name>` subcommand (RESERVED for v1).

### 11.2 Auth header transmission

`[MCLIP-11-03]` For Streamable HTTP transport, the resolved token MUST be sent as `Authorization: Bearer <token>` unless the server config specifies a different scheme via `auth.scheme` (deferred to v1; v0 clients MUST default to Bearer). Clients MUST NOT include access tokens in URI query strings.

### 11.3 OAuth flows

`[MCLIP-11-04]` Full OAuth 2.1 + PKCE flow standardisation is **out of scope for v0**. MCLIP-Auth v0 specifies how a pre-provisioned token is selected and transmitted; it does not require clients to obtain or refresh tokens.

`[MCLIP-11-08]` If an implementation provides HTTP OAuth acquisition or refresh as an extension, it MUST follow the MCP 2025-11-25 authorization requirements for protected-resource metadata discovery and resource-bound access tokens. Provider-specific OAuth helpers MAY live behind the `<binary> servers auth <name>` subcommand without breaking conformance, but they MUST NOT change the behaviour of conformant invocations that use pre-provisioned credentials.

### 11.4 Session reuse

`[MCLIP-11-05]` Clients MAY reuse connections to the same server across invocations via a background daemon or session cache. Whether and how is non-normative for v0. The CLI behaviour (output, exit codes) MUST be identical whether or not session reuse occurs.

---

## 12. Transport

### 12.1 Supported transports

`[MCLIP-12-01]` Conformant clients MUST support both transports defined by MCP 2025-11-25:

- **stdio**: client spawns the server as a subprocess
- **Streamable HTTP**: client connects to a URL endpoint

`[MCLIP-12-02]` Clients MUST NOT support the deprecated HTTP+SSE transport from MCP 2024-11-05 unless they also explicitly support the 2024-11-05 spec version end-to-end.

### 12.2 Protocol version header

`[MCLIP-12-03]` For Streamable HTTP, clients MUST send `MCP-Protocol-Version: 2025-11-25` (or whichever MCP spec version they target) on all requests after initialisation, per MCP spec.

### 12.3 Transport selection

`[MCLIP-12-04]` Transport is determined by the server config entry (§13). The `--transport <stdio|http>` flag MAY be used to override but MUST NOT be required when the config is unambiguous.

### 12.4 Transport errors

`[MCLIP-12-05]` Connection failures (TCP refused, subprocess spawn failed, DNS failure, TLS error) MUST exit `69` with `origin: "transport"` in the error envelope.

### 12.5 Timeout

`[MCLIP-12-06]` Clients MUST honour `--timeout <seconds>` for individual requests. The default timeout for non-streaming requests is implementation-defined but MUST be finite. Streaming requests (§9) MUST NOT time out on idle output but MAY time out on initial connection.

---

## 13. Server discovery

### 13.1 Config sources (priority order)

`[MCLIP-13-01]` Conformant clients MUST resolve server configuration from these sources in priority order (earlier wins for duplicate server names or overlapping keys):

1. Explicit `--config <path>` flag
2. `$MCLIP_CONFIG` environment variable (path to a config file)
3. User config: `$XDG_CONFIG_HOME/mclip/config.json` (or `~/.config/mclip/config.json` if `XDG_CONFIG_HOME` is unset)
4. User-home fallback: `~/.mclip.json`
5. Project-local: `./mclip.json` (current working directory)

`[MCLIP-13-06]` Project-local configuration MUST NOT override an explicit, environment, user, or user-home server entry with the same alias. A project-local server entry that can receive credentials or that sets `safety.trustAnnotations: true` MUST require explicit user consent stored outside the project directory before use. The consent rule applies symmetrically to every project-rooted file enumerated by §13.1 source 5 and §13.2 (currently `./mclip.json`, `./.vscode/mcp.json`, `./.cursor/mcp.json`, plus any future per-project file added under MCLIP-Discovery). Implementations MAY satisfy the consent requirement with a first-run interactive prompt OR a user-config allowlist. In non-interactive mode (stdin is not a TTY), clients MUST NOT prompt for consent; they MUST ignore unconsented risky project-local entries and continue without them (or exit `78` if no other config source resolves the requested server alias).

### 13.2 Inherited config files

`[MCLIP-13-02]` Clients SHOULD additionally discover and consume entries from these de-facto MCP config files (read-only; merged with lower priority than MCLIP-specific files above):

- `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS Claude Desktop)
- `%APPDATA%\Claude\claude_desktop_config.json` (Windows Claude Desktop)
- `~/.config/Claude/claude_desktop_config.json` (Linux Claude Desktop)
- `./.vscode/mcp.json`
- `./.cursor/mcp.json`

Inheriting entries from these files lowers adoption friction for users with existing MCP setups.

### 13.3 Config schema

`[MCLIP-13-03]` The canonical MCLIP config schema is:

> Canonical schema URL: `https://mclip.dev/schemas/config/v0.json`. `mclip.dev` is project-controlled as of 2026-05-16 (see `naming-check.md`). GitHub remains the source repository for the specification and implementation work; `mclip.dev` is the stable public front door and schema namespace, and MAY redirect to repository-hosted source artifacts.

```json
{
  "$schema": "https://mclip.dev/schemas/config/v0.json",
  "servers": {
    "<server-name>": {
      "transport": "stdio" | "http",
      "command": "...",       // stdio only
      "args": ["..."],        // stdio only
      "env": { "...": "..." },// stdio only
      "url": "https://...",   // http only
      "headers": { "...": "..." }, // http only, optional
      "auth": {
        "token": "...",       // optional; prefer OS keychain or per-server env var
        "scheme": "Bearer"    // optional; defaults to Bearer
      },
      "safety": {
        "trustAnnotations": false,           // optional; defaults to false
        "trustAnnotationsReason": "..."      // optional; human-readable audit string
      }
    }
  }
}
```

`[MCLIP-13-04]` Clients MUST accept the existing `mcpServers` key (Claude Desktop convention) as a synonym for `servers`. When both are present, `servers` takes precedence and the client SHOULD print a warning to stderr.

`[MCLIP-13-07]` When a server entry sets `safety.trustAnnotations: true`, conformant clients SHOULD ensure `safety.trustAnnotationsReason` is a non-empty string and SHOULD warn on stderr at first use when the field is missing or empty. The reason is captured for human audit and MUST NOT influence client parsing or runtime behaviour.

### 13.4 Server name resolution

`[MCLIP-13-05]` The positional `<server>` argument in command invocations refers to a key in the merged config. Unknown server names, aliases beginning with `-`, and aliases colliding with reserved root command names (§1.4) MUST exit `78` (config error).

---

## 14. Destructive-action safeguards & non-interactive composition

This section defines three layers. §14.0 is the **Core safety baseline** required by MCLIP-Core. §14.1–§14.3 define **MCLIP-Safety**, required only when that module is claimed. §14.4 is the **non-interactive (CI-safe) composition rollup** and is required by MCLIP-Core; it cross-references rules that live in other sections.

### 14.0 Core safety baseline (MCLIP-Core)

`[MCLIP-14-00]` Core clients MUST treat a tool as **safe** only when its `annotations.readOnlyHint == true` is explicitly declared by the server. Any other tool — including one with `annotations.destructiveHint == false` and one with no annotations at all — MUST be treated as **potentially destructive**.

`[MCLIP-14-01]` For a potentially-destructive tool invoked via `tools call`, if `--yes` (or `-y`) is NOT given AND stdin is NOT a TTY, Core clients MUST refuse to run and exit `64` with a clear stderr message naming the missing safeguard.

`[MCLIP-14-02]` When stdin IS a TTY, a Core-only client (one that does not claim MCLIP-Safety) MAY run potentially-destructive tools without further safeguards. The interactive prompt UX is defined by MCLIP-Safety (§14.1) and is REQUIRED only when MCLIP-Safety is claimed.

`[MCLIP-14-08]` Audit logging is OPTIONAL for every MCLIP conformance level. The `security-model.md` companion document defines a recommended audit-event schema; implementations MAY emit events conforming to it but are not required to. Implementations that emit audit events MUST redact secrets per `security-model.md` Auditability controls.

> Rationale: MCP spec warns *"Clients MUST consider tool annotations to be untrusted unless they come from trusted servers."* Treating `readOnlyHint == true` as the sole positive safe signal is the asymmetric safe choice — a hostile server lying that way exposes itself directly to scrutiny. Honouring `destructiveHint == false` as a skip-confirmation signal would let any server bypass safeguards by simply negating or omitting the destructive hint.

### 14.1 Confirmation prompt (MCLIP-Safety)

`[MCLIP-14-03]` MCLIP-Safety clients MUST prompt on stderr (`Proceed with <server> tools call <name>? [y/N]:`) for any potentially-destructive tool when stdin is a TTY and `--yes` was not given. The default response is `no` (empty input or any non-`y`/`yes` response refuses). The prompt MAY be customised via `mclip.confirm_message` (defined in the companion Extensions Track SEP — see §16).

`[MCLIP-14-04]` MCLIP-Safety clients MUST accept `--yes` / `-y` to skip the prompt non-interactively. Behaviour with `--yes` is identical whether stdin is a TTY or not.

`[MCLIP-14-05]` MCLIP-Safety clients MUST accept `--force` to override client-side validation warnings beyond the standard prompt. `--force` alone MUST NOT skip the confirmation prompt; `--force --yes` MUST.

### 14.2 Dry-run (MCLIP-Safety)

`[MCLIP-14-06]` MCLIP-Safety clients MUST accept `--dry-run` on `tools call`. When set, the client MUST NOT issue `tools/call`. Instead it MUST validate the input against `inputSchema` locally and print to stdout the JSON-RPC request that *would* have been sent. Exit `0` on validation pass, `65` on validation failure.

`[MCLIP-14-09]` `--dry-run` output MUST redact credential values: tokens resolved from any source defined in §11.1, the `auth.token` field if present in the would-send body, and any `Authorization` header value. Redaction MUST replace the value with the literal string `"<redacted>"` so the output remains valid JSON.

`[MCLIP-14-10]` Forward-compatibility: once the companion Extensions Track SEP defines `mclip.sensitive: true` (per §16), MCLIP-Metadata-conformant clients MUST additionally redact the values of input fields whose schema carries that key in `--dry-run` output, using the same `"<redacted>"` sentinel. v0 MCLIP-Metadata claimants MAY pre-implement this redaction as a non-breaking extension; non-Metadata clients MUST NOT apply heuristic field-name redaction in v0, to keep `--dry-run` output deterministic across implementations.

> Note: MCP v2025-11-25 does not standardise a server-side dry-run flag. Future revisions of this profile may add server-side dry-run support if the MCP spec gains it.

### 14.3 Trusted-server escape hatch (MCLIP-Safety)

`[MCLIP-14-07]` MCLIP-Safety clients MAY honour a per-server "trust annotations" opt-in declared as `safety.trustAnnotations: true` in the server config (§13.3). When this flag is set for a server, the client MAY honour `annotations.destructiveHint == false` as a skip-confirmation signal for that server only. Default is untrusted for every server. This flag MUST be treated as security-sensitive config and MUST obey the project-local consent rule in §13.1.

> Rationale: advanced users who have vetted a specific server's annotations may opt out of confirm-by-default for that server. The escape hatch must be explicit per-server config, never inferred from server behaviour.

### 14.4 Non-interactive (CI-safe) composition (MCLIP-Core)

This subsection consolidates the non-interactive guarantees that every MCLIP-Core client MUST provide. Every rule here is normative; rules introduced in earlier sections are cross-referenced rather than restated. The conformance test suite SHOULD exercise these rules together because CI-safety is an emergent property of correct §4.5 + §6 + §11 + §13 + §14.0 + §14.1 composition.

`[MCLIP-14-11]` Detection of non-interactive mode MUST be based on `isatty(stdin)` returning false. Clients MUST NOT use environment variables (e.g. `CI`, `NO_TTY`) as the primary signal, because those are inconsistent across CI providers. Clients MAY honour such variables as additional opt-ins to non-interactive behaviour but MUST NOT use them to override a true TTY into pseudo-non-interactive mode.

`[MCLIP-14-12]` In non-interactive mode, conformant clients MUST NOT prompt on stdin for any reason. This subsumes credential prompts (per §11.1 [MCLIP-11-02]), destructive-action prompts (per §14.0 [MCLIP-14-01]), project-local consent prompts (per §13.1 [MCLIP-13-06]), and any implementation-defined prompt outside this profile. A would-be prompt MUST instead be resolved by: pre-supplied flags (`--yes`), pre-stored consent, or refusal with the appropriate non-zero exit code from §6.1.

`[MCLIP-14-13]` Conformant clients MUST NOT write secrets (any value resolved as a credential per §11.1, any `Authorization` header value, any field redacted per §14.2 [MCLIP-14-09]–[MCLIP-14-10]) to stdout or stderr. This applies to verbose logging, progress notifications, error envelopes, and dry-run output. The `data` field of the error envelope (§5.2) MUST be filtered before emission.

`[MCLIP-14-14]` `--output json` and `--output ndjson` output MUST be deterministic across runs given identical inputs and identical MCP server responses, modulo (a) values supplied by the MCP server itself (timestamps, server-generated IDs), (b) request IDs the client uses for JSON-RPC correlation, and (c) the `progressToken` per §9.1. In particular: object key ordering, whitespace, and pagination iteration order MUST be stable so machine consumers can diff or hash output reliably.

`[MCLIP-14-15]` Exit codes in non-interactive mode MUST be one of the specific values in §6.1; generic exit `1` MUST NOT be returned for any condition that has a dedicated code in the §6.1 table.

> Composition test: a Core-conformant client invoked as `<binary> <server> tools call <destructive-tool>` in a non-TTY pipeline with no `--yes` MUST: emit no stdout, emit a clear stderr message naming the missing safeguard, redact every credential and `Authorization` value the message references, and exit `64`. This single composition tests rules from §4.5, §6.1, §11.1, §14.0, and §14.4.

---

## 15. Conformance

### 15.1 Conformance claim format

`[MCLIP-15-01]` An implementation claims MCLIP v0 conformance with a statement of the form:

> "<implementation> conforms to MCLIP v0 — Core + <module list>"

where `<module list>` is a comma-separated list of zero or more module names drawn from §0.7. The shortest valid claim is "Core"; the fullest is "Core + Resources + Prompts + HTTP + Streaming + Safety + Auth + Metadata + Discovery".

`[MCLIP-15-02]` A claim is valid only if the implementation passes every MUST in the Core sections (per §0.7 table) AND every MUST in each claimed module's sections. An implementation MAY claim Core alone. A claim of any module without Core is invalid and MUST NOT be advertised as MCLIP-conformant.

### 15.2 Conformance test suite

The MCLIP v0 conformance test suite (separate deliverable) is organised by module. Each test is tagged with the rule ID(s) it exercises and the module it belongs to. Implementations MAY submit selective test runs and publish per-module pass/fail results.

### 15.3 Forwards compatibility

`[MCLIP-15-03]` Implementations MAY exceed this profile (add flags, alternate forms, extensions) provided no extension changes the behaviour of conformant invocations defined here. An extension that causes a conformant invocation to produce different output, different exit code, or different stderr/stdout split is a conformance break.

### 15.4 Versioning and MCP compatibility

#### 15.4.1 Profile version identifier

`[MCLIP-15-04]` This profile is versioned `v0`. Draft revisions are numbered (`draft v0 — revision 2.2` etc.) until v0 is frozen; subsequent freezes will increment to `v1`, `v2`, etc. Each revision specifies which previous rules are deprecated, modified, or unchanged (see Change log).

#### 15.4.2 Baseline MCP version declaration

`[MCLIP-15-05]` Every revision of this profile MUST declare exactly one **baseline MCP version** in the document header (top of `profile-v0.md`). v0 draft 2.2 declares `2025-11-25`. The baseline is the MCP spec revision against which every normative rule in this profile is evaluated. A conformant client MAY transact with servers advertising newer MCP versions only if it does not relax any rule in this profile.

`[MCLIP-15-06]` When a new MCP spec revision is published, this profile does NOT automatically advance its baseline. A subsequent profile revision MUST explicitly bump the baseline declaration, list every MCP change item that affects MCLIP rules in §15.5, and either:
- Re-affirm each MCLIP rule unchanged with a short "verified against MCP <new-date>" note, OR
- Modify or add MCLIP rules to track the MCP change, with conformance markers.

`[MCLIP-15-07]` MCLIP draft revisions during a single v0 / v1 cycle MUST track the baseline MCP version conservatively: a draft MAY advance the baseline only when (a) the new MCP revision is final (not a working draft), and (b) the profile-side change list is complete. Drafts MUST NOT silently change behaviour because the MCP spec changed; every behavioural delta in this profile MUST appear in §15.5.

#### 15.4.3 `<binary> --version` output format

`[MCLIP-15-08]` Conformant clients MUST implement `<binary> --version`. Output MUST exit 0 and go to stdout. Two forms are defined; the **JSON form is the normative machine-readable interface** and the **text form is for humans**.

`[MCLIP-15-09]` **JSON form (NORMATIVE for parsing).** When `--output json` is set, the client MUST emit exactly this object:

```json
{
  "implementation": "<impl-name>",
  "implementationVersion": "<impl-version>",
  "mclipProfile": "v0",
  "mclipDraft": "<draft-rev>" | null,
  "mcpBaseline": "<yyyy-mm-dd>",
  "modules": ["core", "<claimed-module>", "..."]
}
```

Field contract:
- `implementation` — non-empty UTF-8 string. No leading/trailing whitespace. SHOULD be lower-kebab-case but parsers MUST NOT depend on case.
- `implementationVersion` — non-empty UTF-8 string. The profile does NOT constrain it to SemVer because Python, Go, Rust, Node, and Bun implementations use different version conventions; parsers MUST treat this field as opaque.
- `mclipProfile` — fixed literal `"v0"` for this profile; future profiles will use `"v1"`, `"v2"`, etc.
- `mclipDraft` — for draft revisions, the literal revision string from the profile header (e.g. `"2.2"`); once the profile freezes, this field MUST be `null` or absent.
- `mcpBaseline` — the ISO-8601 date string from this profile's header declaration; MUST NOT be the MCP version the running server is advertising.
- `modules` — JSON array of lower-kebab-case strings drawn from the canonical enum: `"core"`, `"resources"`, `"prompts"`, `"http"`, `"streaming"`, `"safety"`, `"auth"`, `"metadata"`, `"discovery"`. The array MUST include `"core"` for any client claiming any MCLIP-Core-or-higher conformance. Order MUST be the enum order above. Duplicate entries are a conformance break. Listing a module the implementation does not pass conformance tests for is a conformance break.

`[MCLIP-15-10]` **Text form (human-readable; non-parseable).** When `--output` is unset or `--output text`, the client MUST emit human-readable output containing — at minimum — the same five facts (implementation name, implementation version, profile + draft, MCP baseline, claimed modules). The exact text shape is implementation-defined and is NOT conformance-tested for byte equality. The recommended shape is:

```
<impl-name> <impl-version> (mclip v0 draft <draft-rev>; mcp baseline <yyyy-mm-dd>)
modules: core, <claimed-module>, ...
```

Parsers MUST NOT consume the text form; tooling that needs to compare versions or check module conformance MUST use `--output json`.

> Rationale: parsers across ecosystems cannot reliably regex an implementation name that may contain spaces or punctuation; the only portable contract is the JSON form. Constraining text would either force English phrasing (which excludes localised CLIs) or restrict implementer freedom for no machine-readable payoff. The five-facts requirement keeps the human form informative.

#### 15.4.4 Module + module-version interplay

`[MCLIP-15-11]` Modules introduced after MCLIP v0 (in v1, v2, etc.) MUST be claimable independently of newer Core revisions when that is technically possible. Conversely, a Core revision MUST NOT silently re-require a previously-optional module. If a future Core revision absorbs a module, the change MUST be listed in that revision's §15.5 and old conformance claims naming the absorbed module MUST remain valid (treated as Core conformance going forward).

### 15.5 Notable changes since v0 draft 1

The following items changed during the draft 2 series (drafts 2, 2.1, 2.2):

- §9.5, §9.6 (task-augmented execution: `--follow`, `--detach`, `meta tasks *`) — DELETED. Task support deferred to MCLIP v1.
- §14 (Destructive-action safeguards) — REWRITTEN. New rule §14.0 makes `readOnlyHint == true` the sole positive safe signal; `destructiveHint == false` no longer skips confirmation. Confirmation prompt, `--dry-run`, `--force`, and trusted-server escape hatch moved into MCLIP-Safety module.
- §16 (CLI-metadata extension) — REWRITTEN. Key definitions removed from this profile; namespace reserved with key semantics deferred to a companion **Extensions Track SEP**.
- §13.2 (inherited config files) — UNCHANGED in content, but reclassified as MCLIP-Discovery module rather than Core requirement.
- §11 (Auth) — UNCHANGED in content, reclassified as MCLIP-Auth module rather than Core.
- §12 (Transport) — split: stdio is Core, Streamable HTTP is MCLIP-HTTP module.

Rule IDs are stable except where rules were deleted (no renumbering of surviving rules to keep test-suite references intact).

---

## 16. MCLIP-Metadata module — CLI-hint extension

### 16.1 Status and home

This module reserves the `mclip.*` namespace under MCP's `_meta` field for CLI-hint metadata. **The specific keys, their types, and their semantics are defined in a companion Extensions Track SEP** filed separately under the MCP Extensions process (per [SEP 2133 — Extensions](https://modelcontextprotocol.io/seps/2133-extensions)). This profile reserves the namespace and the trust posture; the Extensions SEP carries the key catalogue.

`[MCLIP-16-01]` MCLIP-Core clients MUST function correctly against vanilla MCP servers that publish no `mclip.*` metadata. A client claims MCLIP-Metadata only when it implements every MUST rule for `mclip.*` keys in the companion Extensions SEP.

> Pre-SEP status (v0 draft 2.1): the Extensions SEP is unfiled. Until it is filed and accepted, claiming MCLIP-Metadata is premature. Implementations that wish to experiment with `mclip.*` keys SHOULD do so behind an opt-in flag and SHOULD NOT advertise MCLIP-Metadata conformance.

### 16.2 Namespace reservation

`[MCLIP-16-02]` All MCLIP-defined metadata keys live under the `mclip.` prefix in `_meta`. Other namespaces (e.g. `vendor.foo`) MUST be ignored by MCLIP clients for the purpose of CLI surface generation. Clients MAY pass non-`mclip.*` `_meta` keys through to consumers (e.g. in `tools describe` JSON output) but MUST NOT let them influence command shape, flag mapping, or safety decisions.

### 16.3 No privilege grant

`[MCLIP-16-03]` Server-supplied `mclip.*` metadata MUST NOT relax safety defaults established in §14.0. Specifically: a server that sets `mclip.destructive: false` (a key the Extensions SEP will define) on a tool that is potentially destructive per §14.0 MUST still be treated as potentially destructive by the client. The metadata extension can tighten safety, never loosen it.

> Rationale: MCP spec warns "Clients MUST consider tool annotations to be untrusted unless they come from trusted servers." Applying the same posture to MCLIP metadata prevents a hostile server from disabling confirmations. The trusted-server escape hatch (§14.3) is the only mechanism for opting into a less-strict posture, and it is per-server config, never server-supplied.

### 16.4 Forward compatibility

`[MCLIP-16-04]` Clients MUST treat unknown keys under the `mclip.` prefix as forwards-compatible: ignore them silently rather than error. This allows future versions of the Extensions SEP to introduce new keys without breaking existing clients.

---

## Appendix A — Reserved flag reference

These flag names are reserved by MCLIP v0. Tool input schemas whose properties would generate these flag names MUST be remapped per §2.5.

| Flag | Short | Section | Description |
|---|---|---|---|
| `--help` | `-h` | §1.5 | Print help, exit 0 |
| `--version` | — | §1.4 | Print version, exit 0 |
| `--output` | `-o` | §4.1 | Output format: text / json / ndjson |
| `--raw` | — | §4.3 | Return bare MCP result for successful `--output json` commands |
| `--no-color` | — | §4.2 | Disable ANSI color |
| `--verbose` | `-v` | — | Increase stderr verbosity |
| `--quiet` | `-q` | — | Suppress non-essential stderr |
| `--input` | — | §2.7 | JSON tool input |
| `--input-file` | — | §2.7 | Path to JSON tool input |
| `--server` | — | — | Reserved; v0 uses positional `<server>` only |
| `--config` | — | §13 | Override config file path |
| `--transport` | — | §12.3 | stdio / http override |
| `--timeout` | — | §12.5 | Request timeout in seconds |
| `--cursor` | — | §10.2 | Pagination cursor |
| `--limit` | — | §10.3 | Max items |
| `--no-paginate` | — | §10.4 | Single-page fetch |
| `--dry-run` | — | §14.2 | Validate without invoking (MCLIP-Safety) |
| `--yes` | `-y` | §14.1 | Skip confirmation (MCLIP-Safety; non-TTY-baseline in §14.0) |
| `--force` | — | §14.1 | Override safety checks (MCLIP-Safety) |
| `--auth-token` | — | §11.1 | Reserved; MUST NOT be used for generic credential passing in v0 |

---

## Appendix B — Exit code reference

| Code | Symbol (informal) | Section |
|---|---|---|
| 0 | OK | §6.1 |
| 1 | GENERIC_ERROR | §6.1 |
| 64 | USAGE | §6.1 |
| 65 | DATA_ERROR | §6.1 |
| 66 | NO_INPUT_FILE | §6.1 |
| 69 | SERVER_UNAVAILABLE | §6.1, §12.5 |
| 70 | INTERNAL_ERROR | §6.1 |
| 75 | TEMP_FAIL | §6.1 |
| 77 | NOPERM | §6.1 |
| 78 | CONFIG_ERROR | §6.1, §13.4 |
| 100 | TOOL_REPORTED_ERROR | §6.1, §5.2 |
| 130 | INTERRUPTED | §6.2 |
| 141 | SIGPIPE | §6.3 |

---

## Appendix C — Worked examples

### C.1 Calling a tool

```
$ mclip linear tools call create_comment \
    --issue-id ENG-123 \
    --body "Looks good"

# stdout (text mode):
Created comment c_abc123 on ENG-123.

# stderr: (empty on success)
# exit: 0
```

JSON mode of the same call:

```
$ mclip -o json linear tools call create_comment --issue-id ENG-123 --body "Looks good"

{
  "result": {
    "content": [
      { "type": "text", "text": "Created comment c_abc123 on ENG-123." }
    ],
    "structuredContent": { "id": "c_abc123", "issue": "ENG-123" },
    "isError": false
  }
}

# exit: 0
```

### C.2 Destructive tool, non-interactive

```
$ mclip linear tools call delete_issue --issue-id ENG-123
mclip: refusing to run destructive tool without --yes (non-interactive).
# stderr above; stdout empty
# exit: 64
```

### C.3 Destructive tool, interactive

```
$ mclip linear tools call delete_issue --issue-id ENG-123
Proceed with linear tools call delete_issue? [y/N]: y
Deleted issue ENG-123.
# exit: 0
```

### C.4 Resource watch (NDJSON)

```
$ mclip -o ndjson filesystem resources watch file:///tmp/state.json
{"type":"resource.updated","uri":"file:///tmp/state.json","notification":{"uri":"file:///tmp/state.json"}}
{"type":"resource.updated","uri":"file:///tmp/state.json","notification":{"uri":"file:///tmp/state.json"}}
# Ctrl-C sends resources/unsubscribe
# exit: 130
```

### C.5 Server unavailable

```
$ mclip linear tools list
mclip: cannot connect to server "linear" (transport: stdio): exec: "linear-mcp": executable file not found in $PATH
# stderr above
# stdout (json mode):
{
  "error": {
    "code": 69,
    "message": "Cannot connect to MCP server",
    "origin": "transport",
    "data": { "server": "linear", "underlying": "exec failed" }
  }
}
# exit: 69
```

### C.6 Tool reported error

```
$ mclip -o json linear tools call create_comment --issue-id NONEXISTENT --body "..."

{
  "result": {
    "content": [{ "type": "text", "text": "Issue NONEXISTENT not found." }],
    "isError": true
  },
  "error": {
    "code": 100,
    "message": "Tool reported error",
    "origin": "tool"
  }
}
# exit: 100
```

### C.7 Listing with pagination

```
$ mclip linear tools list --limit 50
# Auto-paginates until 50 items collected OR nextCursor is absent
# stdout: text rendering of tool list (50 entries)
# exit: 0
```

---

## References

### Normative (cited by MUST/SHOULD rules above)

- [RFC 2119] Bradner, S., "Key words for use in RFCs to Indicate Requirement Levels"
- [RFC 8174] Leiba, B., "Ambiguity of Uppercase vs Lowercase in RFC 2119 Key Words"
- [MCP 2025-11-25] Model Context Protocol Specification, [modelcontextprotocol.io/specification/2025-11-25](https://modelcontextprotocol.io/specification/2025-11-25)
- [MCP 2025-11-25 Authorization] [modelcontextprotocol.io/specification/2025-11-25/basic/authorization](https://modelcontextprotocol.io/specification/2025-11-25/basic/authorization)
- [JSON-RPC 2.0] [jsonrpc.org/specification](https://www.jsonrpc.org/specification)
- [JSON Schema 2020-12] [json-schema.org/draft/2020-12](https://json-schema.org/draft/2020-12)
- [POSIX 12.2] IEEE 1003.1-2017, "Utility Syntax Guidelines", [pubs.opengroup.org](https://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap12.html)
- [GNU CLI] GNU Coding Standards, "Command-Line Interfaces", [gnu.org/prep/standards](https://www.gnu.org/prep/standards/html_node/Command_002dLine-Interfaces.html)
- [clig.dev] Command Line Interface Guidelines, [clig.dev](https://clig.dev)
- [sysexits] BSD `sysexits.h`, [man.freebsd.org](https://man.freebsd.org/cgi/man.cgi?query=sysexits&sektion=3)

### Process and governance (non-normative)

Governance, SEP workflow, working-group, and maintainer references are consolidated in `prd.md` §13. This profile cites them only via that section, so the link set has a single source of truth and stays in sync across all repository documents.

---

## Change log

- **v0 draft 2.2 (2026-05-16)**: Resolved the five open security questions from `security-model.md` draft 1.0 and folded the decisions into both documents. Added a Core CI-safe composition rollup (§14.4) and broadened the credential-prompt prohibition. Expanded §15.4 into a full version + MCP-compatibility policy with normative `<binary> --version` output formats. New rules: `[MCLIP-11-02]` (broadened) covers every conformant invocation, not only `tools call` / `resources read` / `prompts get`; `[MCLIP-11-09]` plaintext-token warning when `auth.token` is read from a world/group-readable config file (or Windows ACL equivalent); `[MCLIP-13-06]` (broadened) so the project-local consent rule applies symmetrically to `./mclip.json`, `./.vscode/mcp.json`, `./.cursor/mcp.json`, and future per-project files, AND forbids consent prompts in non-interactive mode; `[MCLIP-13-07]` SHOULD have `safety.trustAnnotationsReason` whenever `trustAnnotations: true`; `[MCLIP-14-08]` audit logging is OPTIONAL for every conformance level; `[MCLIP-14-09]` `--dry-run` MUST redact credential values; `[MCLIP-14-10]` MCLIP-Metadata clients redact `mclip.sensitive: true` fields in `--dry-run` once the Extensions SEP defines that key; `[MCLIP-14-11]`–`[MCLIP-14-15]` consolidate the non-interactive (CI-safe) baseline as a Core requirement (TTY detection, no-prompt rollup, no-secret-leak, output determinism, specific-exit-code); `[MCLIP-15-05]`–`[MCLIP-15-11]` version + MCP-compatibility policy (baseline declaration, draft tracking discipline, `--version` text + JSON formats, module-list contract, Core-absorbs-module migration rule).
- **v0 draft 2.1 (2026-05-16)**: Tightened pre-freeze blockers: global flag placement and reserved root aliases; schema collision fallback to `--input`; `--raw` JSON success semantics; SIGINT exit code `130`; NDJSON v0 narrowed to profile-defined streaming commands; auth narrowed to keychain / per-server env / config token sources with no generic `MCLIP_TOKEN` or generic `--auth-token`; explicit config precedence changed so user/explicit config beats project-local config; `safety.trustAnnotations` config encoding defined with project-local consent requirements.
- **v0 draft 2 (2026-05-15)**: Restructured into **MCLIP-Core + 8 optional modules** (Resources, Prompts, HTTP, Streaming, Safety, Auth, Metadata, Discovery) — see §0.7 conformance table and §15.1 claim format. Substantive changes: (1) §14 rewritten — only `annotations.readOnlyHint == true` is a positive safe signal, `destructiveHint == false` no longer skips confirmation by default; (2) §16 reserves the `mclip.*` namespace and defers key definitions to a companion **Extensions Track SEP**; (3) §9.5–§9.6 (task-augmented execution) removed entirely, with all task support deferred to MCLIP v1; (4) §11 Auth reclassified as MCLIP-Auth module; §12 Transport split (stdio Core, Streamable HTTP → MCLIP-HTTP module); §13.2 inherited config files → MCLIP-Discovery module. Rule IDs preserved except for deletions in §9.
- **v0 draft 1 (2026-05-15)**: Initial draft. Status: pre-SEP, pre-implementation. Awaiting Transport WG discussion and reference-implementation prototype before formal SEP filing.

## References

- Profile (full normative text): `profile-v0.md` (draft 2.2)
- Security model: `security-model.md` (draft 1.1)
- Conformance fixture catalogue: `conformance-fixtures.md` (v0.2)
- Fixture-server implementation spec: `fixtures-spec.md`
- Reference-CLI architecture: `mclio-architecture.md`
- Wrapper audit (motivation evidence): `wrapper-audit.md`
- Maintainer-facing comparison: `wrapper-comparison.md`
- Adoption guide: `adoption-guide.md`
- Companion Extensions Track SEP: `sep-extensions-mclip-metadata.md`
- Governance, SEP workflow, working-group, and maintainer references: `prd.md` §13
