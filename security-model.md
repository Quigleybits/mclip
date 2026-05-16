# MCLIP CLI Security Model

Version: draft 1.1 (2026-05-16)
Status: draft, aligned with `profile-v0.md` draft 2.2.

## Scope

This document covers security requirements for MCLIP-conformant CLI clients. It focuses on risks introduced by exposing MCP servers through shell commands: destructive tool execution, non-interactive automation, credential handling, project-local configuration, output/log safety, and trust boundaries between the client and MCP server.

This document does not define a full OAuth UX, identity provider policy, server-side authorization model, sandbox model, or enterprise policy engine. Those are implementation or future-profile concerns.

## System Model

MCLIP has five security-relevant actors and stores:

- User or automation runner invoking the CLI with argv, stdin, environment variables, and config files.
- MCLIP client resolving configuration, credentials, MCP server capabilities, schemas, annotations, and metadata.
- MCP server receiving `tools/call`, `resources/*`, and `prompts/*` requests over stdio or Streamable HTTP.
- Local config and project files, including user config, project-local `mclip.json`, and inherited MCP config files.
- Output consumers, including humans, shell pipelines, CI logs, audit logs, and downstream scripts.

The client is a policy enforcement point. The MCP server is the source of capability discovery, but it is not trusted to decide client-side safety policy unless the user explicitly opts into trust for that server.

## Assets

- Access tokens and auth headers.
- User config and server aliases.
- Project-local config trust decisions.
- Tool inputs, especially file paths, secrets, and destructive-operation parameters.
- MCP result data, including potentially sensitive tool output.
- stdout machine output and stderr diagnostics consumed by scripts and CI.
- Audit trail of high-risk invocations.

## Trust Boundaries

### Shell And Environment To MCLIP Client

The user controls argv, stdin, env vars, and shell redirection. CI systems and scripts may invoke MCLIP without a TTY. The client must treat non-interactive execution as higher risk because prompts cannot safely happen.

Controls:

- Refuse potentially destructive tools in non-interactive mode unless `--yes` is present.
- Never block silently when stdin is required from a TTY.
- Never read generic credential env vars like `MCLIP_TOKEN`; credentials must be scoped per server alias.
- Never accept generic credentials through argv, including `--auth-token`.

### Project Files To MCLIP Client

Project-local config is attacker-controllable when a user clones or enters an untrusted repo. A malicious repo can try to shadow a trusted server alias, redirect traffic, capture credentials, or set annotation-trust flags.

Controls:

- Explicit config, environment config, and user config beat project-local config.
- Project-local config cannot override higher-priority aliases.
- Project-local entries that can receive credentials require consent stored outside the repo.
- Project-local `safety.trustAnnotations: true` requires consent stored outside the repo.
- The project-local consent rule applies symmetrically to every project-rooted file MCLIP reads, including `./mclip.json`, `./.vscode/mcp.json`, `./.cursor/mcp.json`, and any future per-project file enumerated under MCLIP-Discovery. The same threat model (attacker-controllable clone) applies; asymmetric handling would create a footgun.
- Non-interactive clients should ignore unconsented risky project-local entries rather than prompting.

### MCLIP Client To MCP Server

The MCP server receives tool inputs and credentials and returns schemas, annotations, metadata, results, and errors. The server may be buggy, compromised, hostile, or simply over-broad in its annotations.

Controls:

- Treat tools as safe only when `annotations.readOnlyHint == true`.
- Treat missing annotations and `destructiveHint == false` as potentially destructive unless `safety.trustAnnotations: true` is set for that server.
- Server-supplied `mclip.*` metadata may tighten client safety behaviour but must not relax it.
- Preserve server JSON-RPC errors and tool `isError` results without hiding their origin.

### Credential Store To MCLIP Client

Credentials can come from OS keychain, per-server env var, or config fallback. Config-stored credentials are higher risk because they are files and may be copied into repos or logs.

Controls:

- Prefer OS keychain, then `MCLIP_TOKEN_<SERVER_NAME>`, then `auth.token`.
- `auth.token` remains permitted as the last-priority source. Implementations SHOULD warn on stderr the first time a token is read from `auth.token` when the config file's POSIX mode is world- or group-readable (mode `& 0o077 != 0`), or when the file resides on a path that is not user-private. Windows clients SHOULD apply an equivalent ACL check or warn unconditionally when reading a config-stored token. The warning MUST NOT contain the token.
- Do not send access tokens in URI query strings.
- For HTTP, send tokens in the `Authorization` header only.
- If OAuth acquisition is implemented as an extension, follow MCP protected-resource metadata and resource-bound token requirements.
- Redact tokens from stderr, logs, dry-run output, and error `data`.

### MCLIP Output To Consumers

stdout is often parsed by scripts. stderr is often logged. Mixing prompts, logs, progress, or secrets into stdout breaks automation and can leak data.

Controls:

- Primary command output goes to stdout only.
- Progress, diagnostics, confirmations, and prompts go to stderr only.
- JSON errors use the envelope from `profile-v0.md` §5.2.
- `--raw` unwraps successful JSON results only; errors remain enveloped.
- Non-interactive output must be deterministic and machine-readable when `--output json` or `--output ndjson` is selected.

## Destructive Action Policy

MCLIP's default posture is deny-or-confirm for anything not explicitly read-only.

Safe tool:

- `annotations.readOnlyHint == true`.

Potentially destructive tool:

- No annotations.
- `destructiveHint == false` without `readOnlyHint == true`.
- `destructiveHint == true`.
- Any tool from a server whose annotations are not trusted and whose read-only status is ambiguous.

Required behaviour:

- Non-TTY without `--yes`: refuse and exit `64`.
- TTY with MCLIP-Safety: prompt on stderr, default no.
- `--yes`: skip prompt but do not skip validation.
- `--force`: may override client-side validation warnings but does not skip confirmation unless paired with `--yes`.
- `--dry-run`: validate locally and print the would-send JSON-RPC request, with credentials redacted and no `tools/call` sent.

## CI-Safe Behaviour

CI-safe means the same command behaves predictably without a TTY and never waits for human input.

Requirements:

- No credential prompts during any conformant invocation defined by `profile-v0.md` §1.3 and §1.4 (every `tools` / `resources` / `prompts` / `meta` verb plus every root subcommand). The narrower phrasing in earlier drafts (which named only `tools call`, `resources read`, `prompts get`) is superseded; an implementer who only blocks prompts on those three verbs leaves `list`, `describe`, `meta version`, and friends able to hang CI.
- No destructive prompt in non-TTY mode; either refuse or run only when `--yes` is present.
- No project-local consent prompt in non-TTY mode; unconsented risky entries are ignored or refused per `profile-v0.md` §13.6.
- Stable exit codes for usage, validation, auth, transport, tool error, SIGINT, and config errors. Generic exit `1` is reserved for unspecified failures only; every error condition with a code in `profile-v0.md` §6.1 MUST use the specific code.
- stdout contains only primary output.
- stderr contains diagnostics only and must redact secrets.
- JSON and NDJSON outputs remain parseable even when progress notifications occur and are deterministic across runs given identical inputs (modulo server-supplied values and the client's `progressToken`).

## Auditability

Audit logging is **OPTIONAL** for MCLIP-Core conformance. This document defines a recommended audit event schema so implementations that opt in produce comparable output; it does not require any implementation to emit, store, or expose audit events. Implementations targeting regulated environments MAY layer audit logging on top of the recommended event schema; the schema is stable so downstream tooling can rely on it.

Recommended audit event fields:

- Timestamp.
- Server alias and resolved config source class, not raw config path if sensitive.
- Transport type.
- Command category and verb.
- Tool, resource URI, or prompt name.
- Whether invocation was non-interactive.
- Whether `--yes`, `--force`, or `--dry-run` was used.
- Safety classification and reason.
- Exit code and error origin.

Do not log:

- Access tokens.
- Authorization headers.
- Raw request bodies by default.
- Full tool outputs by default.

## Resolved Security Decisions

All five open questions tracked in draft 1.0 are resolved in draft 1.1 and folded into both this document and `profile-v0.md` draft 2.2.

- **Audit logging — optional.** MCLIP-Core defines recommended audit event fields but does NOT require any implementation to emit, store, or expose them. Mandating an audit sink raises the conformance bar (sink format, rotation, retention) without script-portability payoff. Implementations targeting regulated environments MAY layer audit logging on top of the recommended event schema.
- **`auth.token` — kept, with plaintext warning.** `auth.token` remains the last-priority credential source so users on platforms without an OS keychain (Linux servers, CI without a secrets manager) are not stranded. Implementations SHOULD warn on stderr when reading a config-stored token from a file with permissive POSIX modes or non-private ACLs (see Credential Store controls above). Warnings MUST NOT contain the token.
- **Inherited project-local config files — same consent rule.** `./mclip.json`, `./.vscode/mcp.json`, `./.cursor/mcp.json`, and any future per-project file follow the identical consent rule for credential-bearing entries and `safety.trustAnnotations: true`. The threat model (attacker-controllable repo clone) is identical across files.
- **`safety.trustAnnotations` reason — SHOULD field.** The MCLIP config schema adds an optional `safety.trustAnnotationsReason: string`. When `trustAnnotations: true` is set, implementations SHOULD ensure a reason is present and SHOULD warn on stderr when it is missing. The reason is captured for human audit, not parsed by the client.
- **`--dry-run` redaction — credentials in v0, sensitive fields when the metadata extension lands.** v0 normative: credentials are redacted in `--dry-run` output (already covered by general credential controls). Forward-compatibility note: once the Extensions Track SEP defines `mclip.sensitive: true` on schema fields, MCLIP-Metadata-conformant clients MUST also redact those field values in `--dry-run` output. v0 implementations MAY pre-implement this behaviour as a non-breaking extension.

## Conformance Security Checklist

- [ ] Project-local config cannot shadow higher-priority aliases.
- [ ] Credential-bearing project-local entries require consent outside the repo.
- [ ] The consent rule applies symmetrically to `./mclip.json`, `./.vscode/mcp.json`, and `./.cursor/mcp.json`.
- [ ] `safety.trustAnnotations` cannot be enabled by unconsented project-local config.
- [ ] When `safety.trustAnnotations: true` is set, a missing `safety.trustAnnotationsReason` produces a stderr warning.
- [ ] Reading a token from `auth.token` warns when the config file has permissive POSIX modes or non-private ACLs; the warning omits the token.
- [ ] Generic `MCLIP_TOKEN` is ignored.
- [ ] Generic `--auth-token` is not accepted as credential input.
- [ ] Tokens are sent in headers, not query strings.
- [ ] Potentially destructive non-TTY calls without `--yes` fail with exit `64`.
- [ ] MCLIP-Safety prompts are stderr-only and default to no.
- [ ] `--force` alone does not skip confirmation.
- [ ] `--dry-run` does not send `tools/call`.
- [ ] `--dry-run` output redacts credentials.
- [ ] MCLIP-Metadata-conformant clients redact fields tagged `mclip.sensitive: true` in `--dry-run` output once the Extensions SEP defines the key.
- [ ] stdout/stderr separation holds in success, error, progress, and prompt cases.
- [ ] SIGINT exits `130` after best-effort cancellation or unsubscribe.
- [ ] Error envelopes preserve `origin` as `client`, `transport`, `server`, or `tool`.
- [ ] Secrets are redacted from diagnostics, dry-run output, and audit logs.

## References

- MCP 2025-11-25 Authorization: https://modelcontextprotocol.io/specification/2025-11-25/basic/authorization
- MCLIP profile draft: `profile-v0.md`
