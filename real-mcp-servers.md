# Real MCP Servers for MCLIP Prototype Validation

**Date:** 2026-05-16
**Status:** RATIFIED (2026-05-16). The set below is the v0 prototype validation set. See `decisions/2026-05-16-mcp-server-prototype-set.md` for the decision record.

## Ratified set

| Slot | Server | Role |
|---|---|---|
| 1 | `@modelcontextprotocol/server-everything` | Protocol-feature exerciser — covers `resources/subscribe`, `notifications/progress`, prompts, logging notifications in one place. No auth; zero-setup. |
| 2 | GitHub MCP Server | Real-world auth + destructive validator. PAT or OAuth; 60+ tools; remote Streamable HTTP. |
| 3 | Context7 MCP Server | Real-SaaS smoke test. Hosted Streamable HTTP + local stdio; API-key header auth (a credential path NOT covered by GitHub's OAuth/PAT). |
| fallback | `@modelcontextprotocol/server-filesystem` | Cheapest-setup destructive fixture if GitHub PAT setup becomes early-CI friction. Held in reserve, not primary. |

Skipped (deferred for v0): Harness MCP Server — broad coverage but adds account-provisioning friction that doesn't pay off until closer to freeze. Reintroduce if Context7's tool surface proves too narrow.

---

## Go SDK canonical import path

**`github.com/modelcontextprotocol/go-sdk`** — confirmed from the repo's `go.mod`.

- Repo: https://github.com/modelcontextprotocol/go-sdk
- Description: "The official Go SDK for Model Context Protocol servers and clients. Maintained in collaboration with Google."
- Latest release: **v1.6.0** (published 2026-05-08)
- Last commit on main: 2026-05-16

This is the correct Tier 1 dependency. In `go.mod`:

```
require github.com/modelcontextprotocol/go-sdk v1.6.0
```

---

## Server 1 — `@modelcontextprotocol/server-everything` (stdio + Streamable HTTP)

| Field | Value |
|---|---|
| Package | `@modelcontextprotocol/server-everything` |
| Repo | https://github.com/modelcontextprotocol/servers/tree/main/src/everything |
| Language | TypeScript (Node.js) |
| Transport | **stdio** (default), **Streamable HTTP**, SSE — selectable at launch |
| Auth model | None |
| Last release | 2026.1.26 (npm); last repo commit 2026-05-16 |

**Surfaces:**

- **Tools (12+):** `echo`, `trigger-long-running-operation` (progress notifications via `notifications/progress`), `get-resource-links`, `get-resource-reference`, `gzip-file-as-resource`, `get-annotated-message` (annotation support), `toggle-subscriber-updates`, `simulate-research-query` (MCP Tasks / SEP-1686), and others.
- **Resources:** Dynamic text/blob templates at `demo://resource/dynamic/{text,blob}/{index}`, static docs, session-scoped resources. Full `resources/list`, `resources/read`, `resources/subscribe`, `resources/unsubscribe` with per-session update notifications.
- **Prompts (4):** `simple-prompt`, `args-prompt`, `completable-prompt`, `resource-prompt`.
- **Streaming:** `trigger-long-running-operation` emits `notifications/progress` per step — the canonical fixture for MCLIP-Streaming.
- **Logging:** `toggle-simulated-logging` fires `notifications/message` at multiple log levels.

**Criteria hit:** stdio (#1), Streamable HTTP (#1), resources/subscribe (#3), progress streaming (#4), actively maintained (#6), stable schema (#7).

**Does NOT cover:** Auth (#2), destructive tools with annotations in the §14 sense (the annotated message tool tests the annotations data type but no tool is `destructive: true` by annotation).

**Installation:**

```
npx -y @modelcontextprotocol/server-everything stdio
npx -y @modelcontextprotocol/server-everything streamableHttp
```

**Smoke test:**

```
# stdio — list tools, then call progress tool
mclip call trigger-long-running-operation --duration 3 --steps 3 npx -y @modelcontextprotocol/server-everything
```

---

## Server 2 — GitHub MCP Server (Streamable HTTP + Bearer/OAuth + Resources)

| Field | Value |
|---|---|
| Repo | https://github.com/github/github-mcp-server |
| Language | Go |
| Transport | **Streamable HTTP** (remote at `https://api.githubcopilot.com/mcp/`); also runnable locally via stdio (Docker or binary) |
| Auth model | **OAuth** (GitHub OAuth App) **or Bearer token** (GitHub PAT via `Authorization: Bearer <token>`) |
| Last release | v1.0.4 (2026-05-11) |
| Last commit | 2026-05-15 |

**Surfaces (remote endpoint, 74 tool files in pkg/github):**

- **Tools (60+):** Issues (create, update, close — destructive), PRs (create, merge, close — destructive), file write (`create_or_update_file` — destructive), repo delete, branch delete, code scanning, Dependabot, secret scanning, Actions workflow triggers.
- **Resources:** `repository_resource.go` + `resources.go` — exposes repo file trees and file contents as MCP resources, readable via `resources/list` / `resources/read`.
- **Prompts:** `prompts.go` + `workflow_prompts.go` — workflow-level prompts documented.
- **Streaming:** Long-running Actions polling implies progress; no explicit `notifications/progress` documented but tool call roundtrip latency is real-world.

**Criteria hit:** Streamable HTTP (#1), Bearer + OAuth (#2), resources (#3), destructive tools with real-world consequence — `create_or_update_file`, `delete_file`, merge PR (#5), actively maintained at Tier 1 (#6), stable schema (toolsets versioned, breakage handled via `deprecated_tool_aliases.go`) (#7).

**Does NOT cover:** stdio (local Docker binary is stdio-ish but the interesting surface is the remote HTTP endpoint).

**Installation (remote, PAT):**

Set the header `Authorization: Bearer <GITHUB_PAT>` against `https://api.githubcopilot.com/mcp/`.

For local stdio via Docker:

```
docker run -i --rm -e GITHUB_PERSONAL_ACCESS_TOKEN=<pat> ghcr.io/github/github-mcp-server
```

**Smoke test:**

```
# List tools on the remote Streamable HTTP endpoint with Bearer auth
mclip --url https://api.githubcopilot.com/mcp/ --header "Authorization: Bearer $GITHUB_PAT" tools list

# Then call a read-only tool
mclip call get_me --url https://api.githubcopilot.com/mcp/ --header "Authorization: Bearer $GITHUB_PAT"
```

---

## Server 3 — Context7 MCP Server (hosted Streamable HTTP + API-key header auth)

| Field | Value |
|---|---|
| Repo | https://github.com/upstash/context7 |
| Language | TypeScript / hosted SaaS |
| Transport | **stdio** (`npx -y @upstash/context7-mcp`) AND **Streamable HTTP** (hosted at `https://mcp.context7.com/mcp`) |
| Auth model | API-key header (Context7 API key); OAuth supported for remote HTTP |
| Last release | 2026-05-11; ~55.4k repo stars; ~815 commits at time of research |

**Surfaces:**
- **Tools:** documentation lookup / library resolution — simple required-string-argument shape. Read-only.
- **Resources:** not primary.
- **Prompts:** not primary.

**Criteria hit:** stdio (#1), Streamable HTTP hosted (#1), API-key header auth (#2 — a different credential shape from GitHub's Bearer/OAuth, which is the value-add for the prototype), actively maintained (#6), stable schema (#7).

**Does NOT cover:** destructive tools (use GitHub for that), resources/subscribe (use server-everything for that), progress notifications.

**Installation:**

Local stdio:
```
npx -y @upstash/context7-mcp
```

Remote (hosted) HTTP — **conformant config (uses MCLIP-Auth credential path, not hardcoded headers):**
```jsonc
// mclip.json
{
  "servers": {
    "context7": {
      "transport": "http",
      "url": "https://mcp.context7.com/mcp"
      // No "headers" entry. The credential is resolved per §11.1:
      //   1. OS keychain entry "mclip" / "context7"
      //   2. Env var MCLIP_TOKEN_CONTEXT7
      //   3. auth.token field below (last priority; warns if file is world-readable)
      //
      // The resolved token is sent as Authorization: Bearer <token>
      // per §11.2 [MCLIP-11-03].
    }
  }
}
```

Provide the Context7 API key via one of:
```
# Preferred (OS keychain on macOS / libsecret on Linux / Credential Manager on Windows)
security add-generic-password -s mclip -a context7 -w '<CONTEXT7_API_KEY>'

# Or per-server env var (CI-friendly)
export MCLIP_TOKEN_CONTEXT7='<CONTEXT7_API_KEY>'
```

The hardcoded `headers.Authorization` form (`"headers": { "Authorization": "Bearer X" }`) is **non-conformant for the prototype-validation purpose** because it bypasses the §11.1 credential source ordering, the §11.2 Bearer construction, the §14.13 no-secret-leak guarantees, and the §11.9 plaintext-token warning. Use it only as a debugging escape hatch outside conformance runs.

**Smoke test:**

```
mclip context7 tools list
mclip context7 tools call resolve-library-id --library "nextjs"
```

**Why it's slot 3 instead of `server-filesystem`:** Context7's API-key-header auth path (resolved through the MCLIP-Auth credential pipeline, NOT hardcoded headers) is not covered by GitHub (Bearer/OAuth) or `server-everything` (no auth). The destructive-tool gap from dropping filesystem is acceptable because GitHub already covers destructive operations with real consequence. Filesystem remains in reserve below.

---

## Fallback — `@modelcontextprotocol/server-filesystem` (stdio + destructive tools, zero setup)

| Field | Value |
|---|---|
| Package | `@modelcontextprotocol/server-filesystem` |
| Repo | https://github.com/modelcontextprotocol/servers/tree/main/src/filesystem |
| Language | TypeScript (Node.js) |
| Transport | **stdio** |
| Auth model | None (access control via allowed-directory list in args) |
| Last release | 2026.1.14 (npm); last repo commit 2026-05-16 |

**Surfaces:**

- **Tools (9):** `read_file`, `read_multiple_files`, `write_file` (destructive), `create_directory` (destructive), `list_directory`, `move_file` (destructive), `search_files`, `get_file_info`, `list_allowed_directories`.
- **Resources:** Filesystem roots exposed as MCP resources via the Roots protocol (`roots/list`, `notifications/roots/list_changed`). Not `resources/list` in the traditional sense — it uses the client-side Roots capability — but exercises a different part of the resource protocol surface.
- **Prompts:** None.
- **Destructive tools:** `write_file`, `create_directory`, `move_file` — real filesystem mutations. These are the cleanest destructive-tool fixtures for §14 validation because the blast radius is confined to a sandboxed temp dir.

**Criteria hit:** stdio (#1), destructive tools with real effect (#5), actively maintained (#6), stable and small schema (9 tools, straightforward inputSchemas) (#7), exercises the Roots sub-protocol adjacent to resources (#3 — partial).

**Does NOT cover:** HTTP transport (#1 HTTP side), auth (#2), `notifications/progress` streaming (#4).

**Installation:**

```
npx -y @modelcontextprotocol/server-filesystem /tmp/mclip-sandbox
```

**Smoke test:**

```
# Write a file (destructive tool), then read it back
mclip call write_file --path /tmp/mclip-sandbox/test.txt --content "hello mclip" npx -y @modelcontextprotocol/server-filesystem /tmp/mclip-sandbox
mclip call read_file --path /tmp/mclip-sandbox/test.txt npx -y @modelcontextprotocol/server-filesystem /tmp/mclip-sandbox
```

---

## Coverage matrix

| Criterion | everything | github-mcp | context7 | filesystem (fallback) |
|---|:---:|:---:|:---:|:---:|
| stdio transport | yes | no (local Docker only) | yes | yes |
| Streamable HTTP transport | yes | yes | yes (hosted) | no |
| Auth-required | no | Bearer / OAuth | API-key header (also OAuth) | no |
| resources/list + read | yes | yes | not primary | partial (Roots) |
| resources/subscribe | yes | no | no | no |
| Streaming / progress notifications | yes | no | no | no |
| Destructive tools | no | yes | no | yes |
| Active maintenance | yes | yes | yes | yes |
| Stable schema | yes | yes | yes | yes |

The ratified set (everything + github-mcp + context7) collectively covers all seven criteria. The filesystem fallback adds zero-setup destructive coverage if GitHub PAT setup becomes friction.

---

## Where to start

**Start with `@modelcontextprotocol/server-everything` in stdio mode.**

Reasons:
- Zero auth setup — no PAT, no OAuth App, no Docker.
- `npx -y @modelcontextprotocol/server-everything` is a one-line cold start.
- It covers the most protocol surface in one server: tools, resources, resource subscriptions, progress notifications, prompts, logging notifications — all the things MCLIP-Core, MCLIP-Resources, and MCLIP-Streaming need to validate.
- Running it in `streamableHttp` mode (same package, `streamableHttp` arg) immediately adds the HTTP transport path without changing the server under test, so you can validate both transports against an identical tool/resource schema.
- Add the GitHub server (Server 2) in a second pass once the §11 credential-flow code path is ready, because it requires PAT wiring and the remote endpoint adds network latency.
- Add the filesystem server (Server 3) in a third pass specifically for §14 destructive-tool annotation validation — it keeps the blast radius small (sandboxed temp dir) and the schema is tiny, so conformance signal is clean.

---

## Gaps and honest caveats

- **No server here exercises `destructive: true` annotations formally** in the MCP spec sense (the annotation field in tool metadata). The `everything` server uses the `annotations` object for `priority`/`audience` fields in content, not the tool-level `destructive` annotation from the MCLIP §14 spec. You will need to either add a synthetic fixture or check whether the GitHub MCP server populates `annotations.destructive` in its tool schemas — that field was added to the spec post-2024 and server adoption is incomplete.
- **resources/subscribe on the GitHub server** is not documented; do not assume it is available until you call `initialize` and inspect the `capabilities` response.
- **The `everything` server's Streamable HTTP mode** runs on localhost by default (port 3001). For the HTTP auth path, the GitHub remote endpoint is the correct target.
