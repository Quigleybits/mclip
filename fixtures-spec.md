# MCLIP Conformance Fixture Servers — Implementation Spec

Version: 0.1 (2026-05-16)
Status: implementation spec, aligned with `conformance-fixtures.md` v0.2 and `profile-v0.md` draft 2.2.

This document is the **implementation spec** for the synthetic MCP servers that back the conformance fixture catalogue. Each server is a minimal, deterministic MCP server whose tools / resources / prompts exist solely to exercise specific MCLIP rules. The fixtures themselves (FX-* IDs) live in `conformance-fixtures.md`; this document maps each fixture to the server(s) it requires.

## Goals and non-goals

**Goals:**
- Every FX-* in `conformance-fixtures.md` runs against one or more of the servers spec'd here.
- Synthetic servers are deterministic: identical inputs produce identical outputs, so `[MCLIP-14-14]` (output determinism) can be tested.
- Servers are minimal: each one exists to exercise its declared rule set, nothing more. No "useful" tools beyond what fixtures need.
- Servers are easy to run in CI: stdio servers spawn as subprocess, HTTP servers run on ephemeral ports.

**Non-goals:**
- These are not example servers anyone should publish. They are conformance test infrastructure.
- They are not security-hardened beyond what fixtures need.
- They do not need pretty error messages — the fixtures assert on shape, not prose.

## Language and SDK

- **Primary language: Go.** Same language as the reference CLI (per user decision). Uses the official Tier 1 Go MCP SDK.
- **TypeScript permitted for servers that need Streamable HTTP + a maintained MCP HTTP server library.** The Go SDK's HTTP server surface is the canonical path; TS is the fallback if HTTP server primitives are awkward in Go at spec time. (Reference CLI architecture doc tracks this dependency choice.)
- All servers compiled / runnable from `fixtures/servers/<server-name>/`. Each server has a `README.md` that names the fixtures it backs.

## Server catalogue

Nine servers cover the fixture catalogue. Server IDs map 1:1 with directory names under `fixtures/servers/`.

### `fx-echo` — basic stdio server

- **Transport:** stdio only.
- **Auth:** none.
- **Tools:**
  - `echo(text: string) -> text` — returns input unchanged. `annotations.readOnlyHint: true`.
- **Backs:** FX-GLOBAL-01, FX-GLOBAL-02, FX-GLOBAL-03, FX-RAW-01, FX-CI-02 (determinism), FX-CI-04 (exit-code variants — used as a base for negative cases).

### `fx-flag-collisions` — schema collision generator

- **Transport:** stdio only.
- **Auth:** none.
- **Tools:**
  - `make_widget(snake_case: string, snakeCase: string) -> text` — both properties produce `--snake-case`. Tests FX-COLLIDE-01 (collision fallback to `--input`).
  - `dump(output: string) -> text` — property `output` collides with reserved global flag, so must become `--arg-output`. Tests FX-COLLIDE-02.
  - `polyglot(union: <oneOf>, tuple: <tuple>, "foo.bar": string) -> text` — exercises FX-COLLIDE-03 (complex shape fallback).
- All tools have `annotations.readOnlyHint: true` to keep destructive logic out of collision tests.

### `fx-destructive` — destructive-action exercises

- **Transport:** stdio.
- **Auth:** none.
- **Tools:**
  - `delete_thing(id: integer) -> { ok: bool }` — NO annotations (so it's potentially destructive per §14.0 baseline). Backs FX-CI-01a, FX-CI-01b.
  - `fails(text: string) -> CallToolResult{isError: true}` — always returns `isError: true`. Backs FX-RAW-02.
  - `sleep(seconds: integer) -> ok` — server-side sleep; backs FX-SIGINT-01. The handler MUST be cancellable so `notifications/cancelled` from client actually stops it (otherwise the test can't tell whether the client sent cancel).
  - `safe_read() -> text` — `annotations.readOnlyHint: true`. Used as a positive control alongside the destructive tools.

### `fx-resources-watch` — resource subscription server

- **Transport:** stdio.
- **Auth:** none.
- **Server capability:** `capabilities.resources.subscribe: true`.
- **Resources:**
  - `test://changes` — a fake resource the server emits `notifications/resources/updated` for, on a deterministic schedule.
- **Behaviour:** the server is configured via two CLI flags at spawn time (NOT via control tools — the CLI under test owns the only stdio MCP connection, so a control tool reachable through MCP would create a second-client requirement the harness can't satisfy):
  - `--event-interval=<duration>` — emit `notifications/resources/updated` for `test://changes` at this fixed cadence after the first `resources/subscribe` arrives. Default `200ms`.
  - `--event-count=<int>` — total number of events to emit. After the Nth event:
    - If `N >= 0` (default `5`): emit the N events, then close the subscription stream cleanly (server-initiated end of stream). Drives FX-SIGINT-03 (expected client exit 0).
    - If `N == -1`: emit forever until the client unsubscribes or disconnects. Drives FX-SIGINT-02 (harness SIGINTs the client after one event; expected unsubscribe + exit 130).
- **Tools:** none. The fixture deliberately has no tools so the harness cannot confuse "tool roundtrip" with "subscription event" behaviour.
- **Backs:** FX-SIGINT-02 (spawn with `--event-count=-1`), FX-SIGINT-03 (spawn with `--event-count=5 --event-interval=100ms`).
- **Determinism note:** the schedule is wall-clock-driven but the event payload is deterministic (`{"uri":"test://changes","seq":N}` where N is the 1-based sequence number); the harness asserts on payload equality and event count, not on inter-event timing.

### `fx-http-auth` — HTTP transport + Bearer auth

- **Transport:** Streamable HTTP.
- **Auth:** Bearer token check. Server expects `Authorization: Bearer <some-known-token>` on every request after `initialize`; returns 401 otherwise.
- **Tools:**
  - `whoami() -> text` — returns the Bearer-token suffix the server saw (last 4 chars) so the harness can assert on which credential source resolved.
  - `read_only(text: string) -> text` — read-only, used to exercise authed conformant invocations.
- **Server capability:** none beyond Core tools.
- **Backs:** FX-AUTH-01, FX-AUTH-02, FX-AUTH-03, FX-AUTH-04, FX-AUTH-05, FX-AUTH-06, FX-AUTH-07, FX-AUTH-08.

### `fx-http-error-data` — HTTP server that echoes request headers in error data

- **Transport:** Streamable HTTP.
- **Auth:** Bearer (same as fx-http-auth).
- **Tools:**
  - `verbose_fail(text: string) -> { error with data: { received_headers: {...} } }` — server returns a JSON-RPC error whose `data` field includes the request headers it saw. Used to test that the client filters credentials from error `data` before emission.
- **Backs:** FX-AUTH-07 (no-leak assertion).

### `fx-pagination` — list/paginate exerciser

- **Transport:** stdio.
- **Auth:** none.
- **Tools:** 50 generated read-only tools named `t01` … `t50`. The fixture server returns them paginated with `nextCursor` opaque strings. Page size is server-decided (10) so auto-paginate has to loop.
- **Backs:** FX-CI-02 (determinism over multi-page list), `tools list --limit` paths.

### `fx-prompts` — prompts surface

- **Transport:** stdio.
- **Auth:** none.
- **Prompts:**
  - `greet({ name: string, lang?: string }) -> message[]` — minimal prompt to verify §8 (Prompts) handling. Argument values MUST be treated as strings per `[MCLIP-8-03]`.
- **Backs:** any future FX-PROMPTS-* (out of scope for the draft 2.1 fixture pass, included here as scaffolding for MCLIP-Prompts module conformance).

### `fx-progress` — progress notifications

- **Transport:** stdio.
- **Auth:** none.
- **Tools:**
  - `slow_count(target: integer) -> text` — over the course of the call, emits `notifications/progress` updates with monotonically increasing `progress`. Final response is the text "done <target>".
- **Backs:** any future FX-PROGRESS-* and the `[MCLIP-9-02]` rule that progress goes to stderr while the JSON result goes cleanly to stdout.

## Server-implementation conventions

These conventions are normative for the fixture servers themselves so the harness can assert on stable behaviour.

1. **Determinism.** Every server response — except request IDs the client controls — MUST be a pure function of the inputs. No timestamps, no random IDs, no environment lookups. The pagination cursor MUST be deterministic (e.g. `"cursor:N"` where N is the page number).
2. **Schema declaration.** Each tool's `inputSchema` MUST be JSON Schema draft 2020-12 (the version MCP 2025-11-25 targets). Use only the JSON-Schema features the spec covers; the conformance suite uses what the servers expose to validate the client's schema → flag generator.
3. **Annotations.** Tools that should be treated as safe MUST declare `annotations.readOnlyHint: true`. Tools that should be treated as destructive MUST omit annotations entirely (so the rule "no annotations → potentially destructive" is exercised) OR set `annotations.destructiveHint: true`. Do NOT set `destructiveHint: false` on a destructive tool in fixtures — that's a server lie, useful only for adversarial fixtures we haven't spec'd yet.
4. **Capabilities.** Each server's `initialize` response MUST declare exactly the capabilities the fixtures rely on, no more. `fx-resources-watch` declares `resources.subscribe: true`; nothing else does.
5. **Error responses.** When asserting on error envelopes, the server's JSON-RPC error MUST include the expected `code` and a stable `message`. The harness compares on `code` and a substring of `message`.
6. **No state across invocations.** Each fresh stdio process MUST behave identically. HTTP servers MAY hold session state per connection but MUST NOT carry it across connections.
7. **Logging.** Servers MUST NOT log to stdout (would corrupt the stdio MCP transport). Stderr is permitted for debug logging; the harness ignores server stderr unless explicitly capturing it.

## Harness contract

The conformance harness (separate deliverable in P1) wraps the fixture servers and the reference CLI together. Per-fixture, the harness:

1. Starts the named server (subprocess for stdio; HTTP listener on an ephemeral port for HTTP servers).
2. Resolves the server's address into a synthetic `<binary> <server> ...` invocation — either by writing a temporary `mclip.json` or by passing `--config` to point at a pre-baked config under `fixtures/configs/`.
3. Invokes the reference CLI with the fixture's command line.
4. Captures stdout, stderr, exit code, AND (for HTTP fixtures) the raw HTTP requests the server received.
5. Compares against the fixture's expected outputs. Comparison rules:
   - Exit code: exact.
   - stdout (JSON / NDJSON modes): JSON-deep-equal modulo fields the harness declares as non-deterministic (server-supplied IDs, progressToken).
   - stdout (text mode): substring match for the rules that allow it; exact match for byte-strict assertions.
   - stderr: substring match for human-readable warnings; never byte-strict (locale, colour codes).
   - Captured HTTP requests: exact match on `Authorization` header, exact match on URL path, JSON-deep-equal on request body.

The harness emits machine-readable results (JUnit XML or JSON) so a wrapper maintainer can publish per-module pass/fail badges per the `prd.md` §4 deliverable.

### Per-fixture harness contracts for client-side-only assertions

A handful of fixtures cannot be validated by server-observable behaviour alone. The harness needs an explicit contract for each.

**FX-AUTH-08 — `--dry-run` redaction (MCLIP-Auth + MCLIP-Safety).** The dry-run path MUST NOT issue `tools/call`, so the server never sees the request. Harness contract:

1. Set up token sources: write `MCLIP_TOKEN_FXHTTPAUTH=test-token-abc123` in the test env; OR write `auth.token: "test-token-abc123"` to the test config with a known POSIX mode; OR (if the platform supports it) preload an OS-keychain entry under service `mclip` account `fxhttpauth`.
2. Invoke `<binary> --dry-run -o json fxhttpauth tools call read_only --text "x"`.
3. Assert exit 0.
4. Parse stdout as JSON. Assert it represents the would-send JSON-RPC request (method, params shape).
5. **Grep stdout for the literal `test-token-abc123`.** Zero matches MUST be found. Any `Authorization` header value MUST equal the literal string `"<redacted>"`. Any `auth.token` echoed back MUST equal `"<redacted>"`.
6. Verify the fixture server (`fx-http-auth`) received zero HTTP requests during the test (no `tools/call`, no `initialize`).
7. Repeat steps 2–6 for each token source variant to confirm redaction works regardless of which credential path resolved the token.

**FX-AUDIT-01 — audit-event redaction (conditional on declared audit emission).** Audit logging is OPTIONAL per `[MCLIP-14-08]`. Skip the fixture entirely when the implementation does not declare audit emission. Detection rule:

1. Invoke `<binary> -o json --version`. Parse the JSON; if the implementation exposes a (currently optional) `auditSink` field (proposed for v0.3 of the `--version` schema, not yet normative), or if the implementation's documentation declares audit emission, run the rest of the fixture. Otherwise mark as SKIPPED.

When running:

1. Configure the audit sink to write to a known temp file (`MCLIP_AUDIT_PATH=$temp/audit.jsonl` or equivalent implementation-defined env / flag).
2. Set up a credential per FX-AUTH-08 step 1.
3. Invoke `<binary> fxhttpauth tools call read_only --text "x"` (normal call, NOT dry-run).
4. Assert exit 0.
5. Read `$temp/audit.jsonl`. Parse each line as JSON. **Grep for the token literal across the whole file.** Zero matches MUST be found.
6. Assert the recommended audit-event fields from `security-model.md` are present on at least one event line.

**FX-AUTH-07 — no-secret-leak in error data (MCLIP-Auth).** Uses `fx-http-error-data`. Harness contract:

1. Set up token per FX-AUTH-08 step 1.
2. Invoke `<binary> -o json fxhttperrordata tools call verbose_fail --text "x"`. The server intentionally returns a JSON-RPC error whose `data` field echoes the request headers it received (including `Authorization`).
3. Assert exit per the error envelope (most likely 1 or a server-mapped code).
4. Parse stdout as JSON. Assert the error envelope shape (`origin: "server"`).
5. **Grep stdout AND stderr for the token literal.** Zero matches MUST be found. The client MUST filter `Authorization` and any value matching the resolved token from the `error.data` field per `[MCLIP-14-13]`.

These three contracts are normative for the conformance harness. A harness implementation that lists FX-AUTH-08 / FX-AUDIT-01 / FX-AUTH-07 as "covered" without the grep assertions described here is not actually testing them.

## Build / CI

- Each fixture server has its own `go.mod` (or `package.json` for the TS HTTP one if needed); they DO NOT share a build with the reference CLI.
- The harness depends on compiled fixture-server binaries; it does NOT recompile them per test run.
- CI workflow: build all fixture servers → build reference CLI → run harness → publish results.

## Coverage map (fixture → server)

| Fixture ID | Server(s) |
|---|---|
| FX-GLOBAL-01, -02, -03 | `fx-echo` |
| FX-RESERVED-01..04 | (no server needed — config-load tests) |
| FX-COLLIDE-01, -02, -03 | `fx-flag-collisions` |
| FX-INPUT-01..04 | `fx-flag-collisions` |
| FX-RAW-01 | `fx-echo` |
| FX-RAW-02 | `fx-destructive` (uses `fails`) |
| FX-RAW-03 | (no server — alias points at unreachable endpoint) |
| FX-RAW-04 | `fx-echo` (usage-error test) |
| FX-SIGINT-01 | `fx-destructive` (uses `sleep`) |
| FX-SIGINT-02, -03 | `fx-resources-watch` |
| FX-LOCAL-01..06 | (no server needed — config-trust tests) |
| FX-AUTH-01..06 | `fx-http-auth` |
| FX-AUTH-07 | `fx-http-error-data` |
| FX-AUTH-08 | `fx-http-auth` |
| FX-AUDIT-01 | `fx-http-error-data` |
| FX-CI-01a, -01b, -01c | `fx-destructive` |
| FX-CI-02 | `fx-pagination` (multi-page list) |
| FX-CI-03, -03b | (no server — `--version` is local-only) |
| FX-CI-04 | combination across servers per error condition |

## Open items the next implementer will hit

These are flagged here so the implementer doesn't rediscover them on day 1:

1. **OS keychain in CI.** FX-AUTH-01 expects keychain priority over env var, but CI runners (GitHub Actions, etc.) typically have no keychain. The harness MUST detect "keychain unavailable" and run the advisory-tagged variant (env-wins) instead of marking failure.
2. **Windows ACL check for FX-AUTH-06.** POSIX-mode check (`mode & 0o077 != 0`) doesn't translate to Windows. The Go implementation's reference path is documented at `[MCLIP-11-09]` — Windows MAY warn unconditionally. The fixture harness needs platform-conditional logic to assert the right warning shape.
3. **PTY for FX-CI-01c.** Running a process under a real PTY in tests requires platform-specific code (`creack/pty` on Unix, ConPTY on Windows). The harness MUST support both or the fixture is reduced to non-PTY only.
4. **Pagination cursor opacity.** `fx-pagination` uses `"cursor:N"` as a debugging convenience but the harness MUST NOT parse this — it's the *server's* cursor; the client's only contract is "opaque". Document the cursor format in the server's README so reviewers don't think it's a client-side convention.
