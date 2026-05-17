# fx-http-auth

Streamable-HTTP MCP fixture server protected by a single static Bearer token.

## Auth

Every request to the MCP endpoint requires
`Authorization: Bearer <token>`. The auth middleware
(`auth.RequireBearerToken`) returns `401 Unauthorized` for missing or wrong
tokens.

The fixtures-spec phrase "every request after `initialize`" is read here as
"every request, including the initialize handshake". Gating the initialize
handshake itself would require body-parsing middleware that adds nothing the
conformance suite actually exercises; the harness presents credentials from
the start.

## Tools

| Tool | Annotations | Behaviour |
|---|---|---|
| `whoami() -> text` | `readOnlyHint: true` | Returns the last 4 chars of the Bearer token the server received on the request. Lets the harness assert which credential source resolved (env var, config file, OS keychain). |
| `read_only(text: string) -> text` | `readOnlyHint: true` | Echoes input. Used to exercise authed conformant invocations. |

## Flags

| Flag | Default | Meaning |
|---|---|---|
| `--addr` | `127.0.0.1:0` | TCP address to bind. `:0` picks an ephemeral port. |
| `--token` | `test-token-abc123` | The one Bearer token the server will accept. |
| `--path` | `/mcp` | URL path the MCP handler is mounted on. |

On startup the server prints the bound address + path on **stdout** as a
single deterministic line:

    fx-http-auth listening 127.0.0.1:<port>/mcp

The harness reads that line to learn the ephemeral port, then issues MCP
requests to that URL.

## Backs

- FX-AUTH-01 through FX-AUTH-06 — Bearer presence, source priority, redaction.
- FX-AUTH-08 — dry-run redaction. Note: dry-run does NOT issue `tools/call`,
  so this server receives zero HTTP requests during the FX-AUTH-08 test; the
  presence assertion is on the client's stdout, not on server state.
