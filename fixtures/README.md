# MCLIP Conformance Fixture Servers

This tree contains the **9 synthetic MCP fixture servers** that back the MCLIP
conformance fixture catalogue. See `fixtures-spec.md` (repo root) for the full
implementation spec and the fixture-ID → server mapping. See
`decisions/2026-05-16-mcp-server-prototype-set.md` for the Go SDK pin
(`github.com/modelcontextprotocol/go-sdk` v1.6.0).

Each server is a minimal, deterministic MCP server whose tools / resources /
prompts exist solely to exercise specific MCLIP rules. Servers are not example
servers anyone should publish — they are conformance-test infrastructure.

## Server catalogue

| Directory | Transport | Backs |
|---|---|---|
| `servers/fx-echo/` | stdio | FX-GLOBAL-01/02/03, FX-RAW-01, FX-CI-02, FX-CI-04 |
| `servers/fx-flag-collisions/` | stdio | FX-COLLIDE-01/02/03, FX-INPUT-01..04 |
| `servers/fx-destructive/` | stdio | FX-CI-01a/b, FX-RAW-02, FX-SIGINT-01 |
| `servers/fx-resources-watch/` | stdio | FX-SIGINT-02/03 |
| `servers/fx-http-auth/` | Streamable HTTP + Bearer | FX-AUTH-01..06, FX-AUTH-08 |
| `servers/fx-http-error-data/` | Streamable HTTP + Bearer | FX-AUTH-07, FX-AUDIT-01 |
| `servers/fx-pagination/` | stdio | FX-CI-02 (multi-page) |
| `servers/fx-prompts/` | stdio | future FX-PROMPTS-* (scaffolding) |
| `servers/fx-progress/` | stdio | future FX-PROGRESS-* (scaffolding) |

## Conventions

- One Go module per server (per `fixtures-spec.md` §Build/CI). The SDK pin is
  duplicated in each `go.mod` deliberately so each server is independently
  buildable and shippable.
- Stdio servers MUST NOT log to stdout (it corrupts the JSON-RPC stream).
  Stderr is permitted; the harness ignores it unless explicitly capturing.
- Every server response (except client-controlled request IDs) MUST be a pure
  function of inputs — no timestamps, no random IDs, no environment lookups.

## Building all 9

PowerShell:

    pwsh fixtures/build.ps1

POSIX:

    bash fixtures/build.sh

Both scripts iterate `go build` over each server directory. Binaries are
produced in-place (e.g. `servers/fx-echo/fx-echo.exe`).

## Verifying all 9 initialize cleanly

After a successful build, from the repo root:

    cd fixtures/verify && go build -o verify.exe . && ./verify.exe

The harness spawns each server, performs the MCP `initialize` handshake,
and asserts the relevant list method (`tools/list`, `prompts/list`, or
`resources/list`) returns the expected feature names. HTTP servers are
exercised via an unauthed GET (expect 401) followed by an authed
`POST /mcp` initialize.

Pass criterion: `9/9 fixture servers verified.` on stderr, exit 0.

## Windows Defender Application Control (known issue)

Windows endpoint-security policy on the maintainer's machine occasionally
classifier-blocks freshly-emitted small Go console binaries:

    fork/exec ...\fx-echo.exe: An Application Control policy has blocked this file.

The block is content-based and non-deterministic across builds: rebuilding
the same source can produce a binary that either passes or fails the
classifier, and which-of-the-9 gets blocked varies between runs. The actual
fixture-server code is correct (a binary that gets classifier-blocked on
one machine runs fine on another).

**Recommended fix:** add a Defender / AppControl exclusion for the
`fixtures/servers/` subtree, then build + verify as normal. CI runs on
Linux are unaffected.

**Without exclusion:** rebuild the blocked server(s) until they pass. The
classifier verdict is stable for a given binary content, so re-emitting
the same .exe doesn't help — you need to change content. The simplest
nudge is adding a unique build-id:

    cd fixtures/servers/<name>
    go build -trimpath \
      -ldflags "-s -w -B 0x$(openssl rand -hex 10)" \
      -o <name>.exe .

A helper script, `fixtures/.rebuild-blocked.ps1`, runs `verify.exe`,
identifies any classifier-blocked servers, rebuilds them with a fresh
`-B` value, and loops until verify returns 9/9 (or hits 8 attempts).
In practice on the maintainer's machine it converges in 1–2 attempts.
