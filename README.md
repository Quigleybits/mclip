# MCLIP

**M**CP **C**ommand-**L**ine **I**nterface **P**rofile — a CLI conformance profile over the Model Context Protocol.

> If a service exposes an MCP server, any MCLIP-conformant client produces the same scriptable surface for it: the same command shape, the same flags, the same JSON envelope, the same exit codes.

MCP does not define a CLI itself. MCLIP defines a canonical CLI projection from MCP's protocol objects — a deterministic rendering layer that turns any MCP server into a uniform shell surface.

A tool's `inputSchema` becomes long flags. A `tools/call` result becomes a JSON envelope on stdout. An MCP error becomes a specific exit code.

Live site: <https://mclip.dev>

## What this repo is

A standards-documentation project for MCLIP — a profile that defines how MCP-to-CLI client implementations expose MCP server capabilities as command-line invocations. MCLIP is not a new protocol and does not require vendor-side work; it is a client-side conformance layer that sits on top of MCP.

The MCP-to-CLI translation space already exists and is crowded (`mcp2cli`, `MCPorter`, `MCPShim`, `f/mcptools`, `FastMCP generate-cli`, Apify `mcpc`, IBM `mcp-cli`, `developit/mcp-cmd`). Each picks its own flag conventions, command shapes, output formats, error structures, and resource/prompt surfaces. A script written for one of them does not run against another. MCLIP standardises the translation itself.

Concretely — the same MCP tool, across wrappers:

```bash
# f/mcptools
mcp call create_issue --params '{"title":"Bug","body":"..."}' \
  npx -y @modelcontextprotocol/server-github

# Apify mcpc
mcpc @github tools-call create_issue title:="Bug" body:="..."

# Under MCLIP, every conformant wrapper:
<binary> github tools call create_issue --title "Bug" --body "..."
```

Different command verbs, different argument styles, different server references, different output shapes. MCLIP-Core ([profile-v0.md §1.2, §2.5, §5](profile-v0.md)) pins all four.

## How MCP maps to the CLI

MCLIP doesn't try to turn every MCP message into a shell command. The projection draws three lines:

| MCP element | MCLIP treatment |
|---|---|
| **Tools, resources, prompts** | User-facing commands. Tool input schemas project to long flags; results project to a JSON envelope on stdout. Resources expose `list`/`read`/`templates`/`watch`; prompts expose `list`/`get`. |
| **Sampling, elicitation, roots** | Runtime mechanics, not standalone commands. Surface as flags, interactive prompts (with non-interactive refusal per §14.12), or config — not invocations. |
| **Initialize, transport, pagination, auth, progress, cancellation, diagnostics** | CLI infrastructure: hidden handshake, `--transport`, `--timeout`, `--cursor`/`--limit`/`--no-paginate`, env-var credentials (§11), NDJSON progress on stderr (MCLIP-Streaming), `SIGINT` cancellation (§6.2), `--verbose`/`--quiet`. |

MCP **primitives** become commands; MCP **schemas** become input contracts; MCP **responses** become a JSON envelope; MCP **lifecycle and interaction mechanics** become CLI runtime behaviour. The nine conformance modules in [`profile-v0.md`](profile-v0.md) §0.7 are organised around this split.

## Specification set

Read in this order to orient:

| # | File | What it covers |
|---|------|----------------|
| 1 | [`profile-v0.md`](profile-v0.md) | Normative profile specification — Core plus eight independently-claimable modules. Every rule tagged `[MCLIP-§-NN]` for conformance tracking. |
| 2 | [`security-model.md`](security-model.md) | Trust boundaries, destructive-action policy, CI-safe behaviour, credential handling, auditability. |
| 3 | [`conformance-fixtures.md`](conformance-fixtures.md) | 30+ executable conformance fixtures with expected exit codes and output shapes. |
| 4 | [`adoption-guide.md`](adoption-guide.md) | For existing MCP-to-CLI wrapper maintainers — minimum surface for MCLIP-Core conformance and the per-module cost. |
| 5 | [`sep-standards-mclip-profile.md`](sep-standards-mclip-profile.md) | Standards Track SEP draft (wraps the profile for filing under `modelcontextprotocol/seps`). |
| 6 | [`sep-extensions-mclip-metadata.md`](sep-extensions-mclip-metadata.md) | Extensions Track SEP draft (optional `mclip.*` CLI-hint metadata keys, per SEP-2133). |

Supporting docs:

- [`product-brief.md`](product-brief.md) — one-pager summary.
- [`use-cases.md`](use-cases.md) — concrete MCLIP use cases.
- [`real-mcp-servers.md`](real-mcp-servers.md) — the prototype-validation real-server set.

Implementation specs:

- [`mclio-architecture.md`](mclio-architecture.md) — architecture spec for `mclio`, the production CLI that doubles as the standard's executable reference.
- [`fixtures-spec.md`](fixtures-spec.md) — implementation spec for the 9 synthetic MCP fixture servers.

## Implementation

- [`fixtures/`](fixtures/) — 9 synthetic MCP servers + verify harness, Go, pinned to `github.com/modelcontextprotocol/go-sdk` v1.6.0.
- **`mclio`** — production CLI (separate repo, in development). Greenfield Go binary; doubles as the standard's executable reference and the recommended daily-driver MCP CLI.
- [`site/`](site/) — Astro source for <https://mclip.dev>.

## Canonical identity

- **Public front door:** <https://mclip.dev>
- **Schema namespace:** <https://mclip.dev/schemas/config/v0.json> (Draft 2020-12 JSON Schema; tracks `profile-v0.md` §13.3)
- **Source repository:** this repo (specs, fixtures, decisions).

## Status

Pre-SEP. The core specification set is drafted and self-consistent. The 9 fixture servers and the verify harness are built. The remaining build queue is the `mclio` binary, the full conformance harness against it, and a companion consumer app.

## Licence

Apache-2.0. See [`LICENSE`](LICENSE).
