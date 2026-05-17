# fx-http-error-data

Streamable-HTTP MCP fixture server that protects its MCP endpoint with the
same static Bearer-token middleware as `fx-http-auth`, but exposes a single
tool whose entire purpose is to return a JSON-RPC error containing the
inbound request headers in `error.data`.

## Tools

| Tool | Behaviour |
|---|---|
| `verbose_fail(text: string) -> error` | Always returns a JSON-RPC error. `error.code = -32000`, `error.message = "verbose fail"`, `error.data = { "received_headers": { ... }, "received_text": "..." }`. The headers include `Authorization` (and the literal Bearer token value the server saw). |

## Flags

| Flag | Default | Meaning |
|---|---|---|
| `--addr` | `127.0.0.1:0` | TCP address to bind. |
| `--token` | `test-token-abc123` | The accepted Bearer token. |
| `--path` | `/mcp` | URL path the MCP handler is mounted on. |

Bound address is printed on stdout once at startup (same format as
`fx-http-auth`):

    fx-http-error-data listening 127.0.0.1:<port>/mcp

## Backs

- **FX-AUTH-07** — no-secret-leak in `error.data`. The harness presents a
  known token, calls `verbose_fail`, then greps the client's stdout AND
  stderr for the literal token. Zero matches must be found.
- **FX-AUDIT-01** — audit-event redaction. The harness configures the
  client's audit sink to write to a temp file, calls `verbose_fail` (NOT
  dry-run), then greps the audit JSONL for the literal token. Zero matches
  must be found.

## Notes

The server uses the SDK's low-level path of returning `*jsonrpc.Error` from
a typed tool handler — the SDK's typed-handler shim detects the structured
error type and forwards it through unchanged (see
`internal/jsonrpc2.WireError`).
