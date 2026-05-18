# MCLIP

**M**CP **C**ommand-**L**ine **I**nterface **P**rofile — a CLI conformance profile over the Model Context Protocol.

> If a service exposes an MCP server, any MCLIP-conformant client produces the same scriptable surface for it: the same command shape, the same flags, the same JSON envelope, the same exit codes.

Live site: <https://mclip.dev>

## What this repo is

A standards-documentation project for MCLIP — a profile that defines how MCP-to-CLI client implementations expose MCP server capabilities as command-line invocations. MCLIP is not a new protocol and does not require vendor-side work; it is a client-side conformance layer that sits on top of MCP.

The MCP-to-CLI translation space already exists and is crowded (`mcp2cli`, `MCPorter`, `MCPShim`, `f/mcptools`, `FastMCP generate-cli`, Apify `mcpc`, IBM `mcp-cli`, `developit/mcp-cmd`). Each picks its own flag conventions, command shapes, output formats, error structures, and resource/prompt surfaces. A script written for one of them does not run against another. MCLIP standardises the translation itself.

## Specification set

Read in this order to orient:

| # | File | What it covers |
|---|------|----------------|
| 1 | [`prd.md`](prd.md) | Product requirements, scope, deliverables, success criteria, governance routing. |
| 2 | [`profile-v0.md`](profile-v0.md) | Normative profile specification — Core plus eight independently-claimable modules. Every rule tagged `[MCLIP-§-NN]` for conformance tracking. |
| 3 | [`security-model.md`](security-model.md) | Trust boundaries, destructive-action policy, CI-safe behaviour, credential handling, auditability. |
| 4 | [`conformance-fixtures.md`](conformance-fixtures.md) | 30+ executable conformance fixtures with expected exit codes and output shapes. |
| 5 | [`adoption-guide.md`](adoption-guide.md) | For existing MCP-to-CLI wrapper maintainers — minimum surface for MCLIP-Core conformance and the per-module cost. |
| 6 | [`sep-standards-mclip-profile.md`](sep-standards-mclip-profile.md) | Standards Track SEP draft (wraps the profile for filing under `modelcontextprotocol/seps`). |
| 7 | [`sep-extensions-mclip-metadata.md`](sep-extensions-mclip-metadata.md) | Extensions Track SEP draft (optional `mclip.*` CLI-hint metadata keys, per SEP-2133). |

Supporting research and analysis:

- [`product-brief.md`](product-brief.md) — one-pager summary.
- [`use-cases.md`](use-cases.md) — concrete MCLIP use cases.
- [`wrapper-audit.md`](wrapper-audit.md) — quantitative audit of the eight existing MCP-to-CLI wrappers.
- [`wrapper-comparison.md`](wrapper-comparison.md) — maintainer-facing one-page comparison.
- [`governance-recommendation.md`](governance-recommendation.md) — routing the SEP through the MCP governance process.
- [`naming-check.md`](naming-check.md) — trademark / package-registry / domain collision audit for "MCLIP".
- [`real-mcp-servers.md`](real-mcp-servers.md) — the ratified prototype-validation real-server set.

Implementation specs:

- [`mclio-architecture.md`](mclio-architecture.md) — architecture spec for `mclio`, the production CLI that doubles as the standard's executable reference.
- [`fixtures-spec.md`](fixtures-spec.md) — implementation spec for the 9 synthetic MCP fixture servers.

Decision records (append-only): [`decisions/`](decisions/).

## Implementation

- [`fixtures/`](fixtures/) — 9 synthetic MCP servers + verify harness, Go, pinned to `github.com/modelcontextprotocol/go-sdk` v1.6.0.
- [`mclio`](https://github.com/Quigleybits/mclio) — production CLI (separate repo). Greenfield Go binary; doubles as the standard's executable reference and the recommended daily-driver MCP CLI.
- [`site/`](site/) — Astro source for <https://mclip.dev>.

## Canonical identity

- **Public front door:** <https://mclip.dev>
- **Schema namespace:** <https://mclip.dev/schemas/config/v0.json> (Draft 2020-12 JSON Schema; tracks `profile-v0.md` §13.3)
- **Source repository:** this repo (specs, fixtures, decisions).

## Status

Pre-SEP. The seven core specification documents are written and internally consistent. The 9 fixture servers and the verify harness are built. The remaining build queue is the `mclio` binary, the full conformance harness against it, and a companion consumer app. External coordination — Discord post, draft SEP PR, wrapper-maintainer outreach — is on hold until the build queue is complete and exercised end-to-end.

## Licence

Apache-2.0. See [`LICENSE`](LICENSE).
