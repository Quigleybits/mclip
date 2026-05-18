# MCLIP Conformance Fixture Catalogue

Version: 0.2 (2026-05-16)
Status: draft, aligned with `profile-v0.md` draft 2.2 and `security-model.md` draft 1.1.

Change log:
- **0.2 (2026-05-16)**: Review-pass corrections — FX-COLLIDE-02 now exercises the `--arg-` reserved-flag prefix per [MCLIP-2-07] (previously contradicted the rule by forcing `--input`); FX-CI-01 split into text/JSON modes per §5.2; new FX-CI-01c verifies `isatty(stdin)` is the primary signal (not env vars); FX-LOCAL-04 corrected to expect exit 78 (project-local consent rule refuses load, not partial load); new FX-LOCAL-06 for the user-config-shadow case; advisory tags on FX-AUTH-01, FX-AUTH-06, FX-LOCAL-05 (SHOULD-level rules); FX-AUTH-07 split into Auth-only secret-leak and Auth+Safety dry-run (FX-AUTH-08); new FX-AUDIT-01 conditional on declared audit emission; FX-CI-03 split into JSON-normative parsing and text non-strict five-facts presence.
- **0.1 (2026-05-16)**: Initial catalogue covering draft 2.1/2.2 freeze blockers.

This document is the **catalogue**: it lists each conformance fixture, the MCLIP rule(s) it exercises, the synthetic MCP server shape needed, the client invocation, and the expected pass criteria. The fixtures **themselves** (the synthetic MCP servers + the harness) are a separate deliverable in `prd.md` §4.5–§4.6 and are tracked under `todo.md` P1 — Prototype / Conformance.

The catalogue is grouped by the draft 2.1 / 2.2 blockers the conformance suite must lock down before v0 freeze. Each fixture carries an ID of the form `FX-<theme>-NN`. Fixtures are tagged with the rule IDs they exercise so wrappers publishing per-module pass/fail can match by tag.

---

## 1. Global flag placement (§1.2)

Locks down `[MCLIP-1-02]` and `[MCLIP-1-07]`.

### FX-GLOBAL-01 — Pre-server global flags accepted

- **MCP server shape:** single tool `echo(text: string) -> text`.
- **Invocation:** `<binary> -o json --no-color demo tools call echo --text "hi"`
- **Expected:** exit 0; stdout is the §5.1 JSON success envelope; stderr empty; ANSI codes suppressed.
- **Pass criteria:** envelope shape exactly matches §5.1; no ANSI sequences in any output byte; no `--no-color` or `-o` flag leaks into the MCP `tools/call` request body.

### FX-GLOBAL-02 — Post-target global flag rejected at conformance level

- **Invocation:** `<binary> demo tools call echo --text "hi" -o json`
- **Expected:** the test harness MAY accept either outcome (the rule says implementations MAY accept post-target global flags), but the **conformance test** asserts the canonical pre-server placement form (FX-GLOBAL-01) is accepted. Implementations that only accept the canonical placement pass.
- **Pass criteria:** canonical form passes; non-canonical form is implementation-defined but MUST NOT silently treat the flag as a tool argument.

### FX-GLOBAL-03 — Generated flag before `<server>` is a usage error

- **MCP server shape:** tool with input property `text: string`.
- **Invocation:** `<binary> --text "hi" demo tools call echo`
- **Expected:** exit 64; stderr clearly names the misplaced `--text` flag; stdout empty in text mode, JSON error envelope in JSON mode.
- **Pass criteria:** exit 64; error envelope `origin: "client"`; no `tools/call` request sent.

---

## 2. Reserved root names (§1.4 + §1.8)

Locks down `[MCLIP-1-08]` and `[MCLIP-13-05]`.

### FX-RESERVED-01 — Alias `servers` is rejected at config load

- **Config:** server entry keyed `"servers"`.
- **Invocation:** any conformant invocation, e.g. `<binary> servers tools list`.
- **Expected:** exit 78 (config error); stderr names the reserved alias.
- **Pass criteria:** the client never attempts a JSON-RPC connection; the error is emitted before any transport activity.

### FX-RESERVED-02 — Alias `help` is rejected

- As FX-RESERVED-01 with alias `"help"`. Same pass criteria.

### FX-RESERVED-03 — Alias `version` is rejected

- As FX-RESERVED-01 with alias `"version"`. Same pass criteria.

### FX-RESERVED-04 — Alias starting with `-` is rejected

- Config with alias `"-evil"`. Invocation `<binary> -evil tools list`.
- **Expected:** exit 78; the leading-`-` rejection MUST happen at config load, not as a CLI-parsing error.
- **Pass criteria:** stderr message distinguishes "invalid alias" from "unknown flag".

---

## 3. Schema collision policy (§2.5 + §2.7)

Locks down the collision fallback to `--input` / `--input-file`.

### FX-COLLIDE-01 — Two properties generate the same flag

- **MCP server shape:** tool `make_widget(snake_case: string, snakeCase: string)`. Both generate `--snake-case`.
- **Invocation:** `<binary> demo tools call make_widget --snake-case "x"` (intentionally ambiguous).
- **Expected:** exit 64; stderr explains the collision and points to `--input` / `--input-file`; no `tools/call` sent.
- **Pass criteria:** the client also rejects `<binary> demo tools call make_widget --snake-case=x --snake-case=y`. The only conformant path to call this tool is `--input '{"snake_case":"x","snakeCase":"y"}'`.

### FX-COLLIDE-02 — Reserved-flag collision

- **Exercises:** `[MCLIP-2-07]`.
- **MCP server shape:** tool `dump(output: string) -> ok` whose input property `output` would generate `--output`, a reserved global flag per Appendix A.
- **Invocation A (conformant):** `<binary> demo tools call dump --arg-output "/tmp/x"`.
- **Expected A:** exit 0; the JSON-RPC `arguments` object is `{"output": "/tmp/x"}`. The bare global flag `--output` is unaffected.
- **Invocation B (must remain global):** `<binary> -o json demo tools call dump --arg-output "/tmp/x" --output json`.
- **Expected B:** exit 0; `--output json` is parsed as the global format flag (no double-set conflict); `--arg-output` carries the tool argument.
- **Invocation C (fallback path):** `<binary> demo tools call dump --input '{"output":"/tmp/x"}'`.
- **Expected C:** exit 0; identical JSON-RPC body to Invocation A. `--input` remains an additional permitted route, not a replacement for the `--arg-` form.
- **Pass criteria:** all three invocations produce the same JSON-RPC `arguments` object. A client that only accepts `--input` for this case fails Invocation A and is non-conformant.

### FX-COLLIDE-03 — Complex shape fallback

- **MCP server shape:** tool with `oneOf` schema, tuple array, and a property with a dotted name (`foo.bar`).
- **Expected:** the client emits no generated flags for these properties; `tools describe` output names each as requiring `--input` / `--input-file`.

---

## 4. `--input` / `--input-file` fallback shapes (§2.7 + §3)

Locks down conformant input source precedence.

### FX-INPUT-01 — `--input` JSON wins for ambiguous schemas

- **Invocation:** `<binary> demo tools call make_widget --input '{"snake_case":"a","snakeCase":"b"}'`.
- **Expected:** exit 0; JSON-RPC body's `arguments` is exactly that object.
- **Pass criteria:** no property-flag → JSON merging; if any property flag is also present, exit 64.

### FX-INPUT-02 — `--input-file` from disk

- **Invocation:** `<binary> demo tools call make_widget --input-file ./fixtures/widget.json`.
- **Expected:** identical to FX-INPUT-01.
- **Pass criteria:** missing file exits 66; unparseable JSON exits 65.

### FX-INPUT-03 — `--input -` reads stdin

- **Invocation:** `echo '{"snake_case":"a","snakeCase":"b"}' | <binary> demo tools call make_widget --input -`.
- **Expected:** exit 0; identical body.
- **Pass criteria:** stdin must be drained completely; no other data may be read from stdin during the same invocation.

### FX-INPUT-04 — Config / env vars cannot supply tool argument defaults

- **Setup:** config entry sets `args` / `env`; tool requires property `text`; no `--text`, no `--input`, no `--input-file`.
- **Invocation:** `<binary> demo tools call echo`.
- **Expected:** exit 64; stderr says required input `text` is missing.
- **Pass criteria:** the client MUST NOT consult any non-CLI source for tool argument defaults.

---

## 5. `--raw` semantics (§4.3)

Locks down the unwrap-on-success-only rule.

### FX-RAW-01 — Successful call unwraps in JSON mode

- **Invocation:** `<binary> -o json --raw demo tools call echo --text "hi"`.
- **Expected:** stdout is the bare `CallToolResult` JSON (no `{"result": ...}` envelope); exit 0.

### FX-RAW-02 — Tool-reported error remains enveloped

- **MCP server shape:** tool that returns `isError: true`.
- **Invocation:** `<binary> -o json --raw demo tools call fails --text "x"`.
- **Expected:** stdout is the §5.2 enveloped form (`{"result": ..., "error": {...}}`); exit 100. `--raw` MUST NOT strip the envelope.

### FX-RAW-03 — Transport error remains enveloped

- **Setup:** server alias points at unreachable endpoint.
- **Invocation:** `<binary> -o json --raw demo tools call echo --text "hi"`.
- **Expected:** stdout is the §5.2 error envelope (`{"error": {"origin": "transport", ...}}`); exit 69.

### FX-RAW-04 — `--raw` without `-o json` is a usage error

- **Invocation:** `<binary> --raw demo tools call echo --text "hi"`.
- **Expected:** exit 64; stderr explains `--raw` requires `-o json`.

---

## 6. SIGINT exit code (§6.2 + §7.4 + §9.3)

Locks down `[MCLIP-6-03]` and the `resources watch` unsubscribe path.

### FX-SIGINT-01 — Long tool call interrupted by Ctrl-C

- **MCP server shape:** tool `sleep(seconds: integer)` that just sleeps server-side.
- **Invocation:** `<binary> demo tools call sleep --seconds 30`. Send SIGINT after 1 s.
- **Expected:** exit 130; the client SHOULD have sent `notifications/cancelled` to the server.
- **Pass criteria:** exit code exactly 130; cancellation send is non-blocking (the client does not wait for ack).

### FX-SIGINT-02 — `resources watch` unsubscribes then exits 130

- **MCP server shape:** server with `capabilities.resources.subscribe: true` and a watch-target resource.
- **Invocation:** `<binary> -o ndjson demo resources watch test://changes`. Send SIGINT after one event.
- **Expected:** before exit, the client sends `resources/unsubscribe`; exit 130.
- **Pass criteria:** harness verifies the unsubscribe request was received by the server; exit 130.

### FX-SIGINT-03 — Server-initiated stream end exits 0

- **MCP server shape:** as FX-SIGINT-02, but the server emits a `notifications/resources/updated` then closes the subscription stream cleanly.
- **Invocation:** `<binary> -o ndjson demo resources watch test://changes`. No SIGINT.
- **Expected:** exit 0; no `notifications/cancelled` sent; client MUST distinguish server-ended stream from interrupted stream.

---

## 7. Project-local config trust (§13.1 + §13.6)

Locks down `[MCLIP-13-06]` and the symmetric consent rule across files.

### FX-LOCAL-01 — Project-local alias cannot shadow user-config alias

- **Setup:** user config has `demo` → server A; project-local `./mclip.json` has `demo` → server B with embedded `auth.token`.
- **Invocation:** `<binary> demo tools list` from the project directory.
- **Expected:** the request goes to server A (user-config wins); server B's `auth.token` is never read; stderr may emit one informational message about the ignored override.

### FX-LOCAL-02 — Unconsented credential-bearing project-local entry refused in non-TTY

- **Setup:** project-local `./mclip.json` has alias `demo` with `auth.token`; no user-config entry for `demo`; no consent record in user config.
- **Invocation:** `<binary> demo tools list` with `stdin` redirected from `/dev/null`.
- **Expected:** exit 78 (config error: server unknown after risky project-local entry was ignored); stderr explains why; no JSON-RPC connection attempted.

### FX-LOCAL-03 — Same rule applies to `./.vscode/mcp.json`

- **Setup:** `./.vscode/mcp.json` has alias `demo` with `auth.token`; no consent.
- **Invocation:** `<binary> demo tools list` non-interactively.
- **Expected:** exit 78; same behaviour as FX-LOCAL-02. Asymmetric handling between `./mclip.json` and the inherited files is a conformance failure.

### FX-LOCAL-04 — Unconsented project-local `safety.trustAnnotations: true` refused outright (no shadow)

- **Setup:** project-local `./mclip.json` has alias `demo` (transport stdio, no credentials) with `safety.trustAnnotations: true`. NO higher-priority alias for `demo` exists.
- **Invocation:** `<binary> demo tools call <destructive-tool> --yes` non-interactively.
- **Expected:** exit 78 — the risky project-local entry is ignored entirely per §13.1 [MCLIP-13-06]; with no other config source resolving the alias, the server is unknown. No `tools/call` is sent. Stderr names the alias and the missing consent.
- **Why the change from "load with trust flag ignored":** §13.6 forbids loading credential-bearing or `trustAnnotations`-bearing project-local entries without consent; partial-loading would still expose the user to whatever the entry's `command`/`args`/`url` were set to.

### FX-LOCAL-05 — Missing `trustAnnotationsReason` warns at first use *(advisory, SHOULD-level)*

- **Tag:** ADVISORY — exercises `[MCLIP-13-07]` which is a SHOULD, not a MUST. Implementations that elect not to warn are still conformant.
- **Setup:** user-config alias with `safety.trustAnnotations: true` but no `safety.trustAnnotationsReason` (consented; lives in user config, not project-local).
- **Invocation:** any conformant call to a destructive tool that benefits from trustAnnotations.
- **Expected:** stderr emits exactly one warning naming the alias and the missing field. The call still proceeds.
- **Pass criteria:** advisory pass — used for quality reporting, not module pass/fail.

### FX-LOCAL-06 — Project-local trust flag silently ignored when user-config alias shadows it

- **Setup:** user-config alias `demo` → server A (no trustAnnotations). Project-local `./mclip.json` has alias `demo` → server B with `safety.trustAnnotations: true` (unconsented).
- **Invocation:** `<binary> demo tools call <destructive-tool>` interactively.
- **Expected:** the request goes to server A (per §13.6: project-local cannot override higher-priority); server B's trust flag never takes effect; server B's `command` / `url` is never executed or contacted. Stderr MAY emit one informational message about the ignored override.
- **Pass criteria:** verifies the precedence rule prevents privilege escalation even when consent might be granted elsewhere for a different file.

---

## 8. Auth source ordering (§11.1)

Locks down `[MCLIP-11-01]`, `[MCLIP-11-06]`, `[MCLIP-11-07]`, `[MCLIP-11-09]`.

### FX-AUTH-01 — Keychain entry wins over env var and config *(advisory, SHOULD-level)*

- **Tag:** ADVISORY — exercises the keychain-priority part of `[MCLIP-11-01]`, which the profile marks as SHOULD (clients SHOULD support OS keychain). Clients without keychain support remain MCLIP-Auth conformant.
- **Setup:** OS keychain has `mclip/<alias>` → `K-TOKEN`; env `MCLIP_TOKEN_<ALIAS>=E-TOKEN`; config has `auth.token: "C-TOKEN"`.
- **Invocation:** any HTTP-transport `tools list`.
- **Expected (clients that support keychain):** outgoing request carries `Authorization: Bearer K-TOKEN`.
- **Expected (clients that do not):** outgoing request carries `Authorization: Bearer E-TOKEN` (env wins because keychain is unsupported). The fixture passes either case; harness records which path was taken for the implementation's quality report.

### FX-AUTH-02 — Env var wins when keychain entry is absent

- **Setup:** no keychain entry; env `MCLIP_TOKEN_<ALIAS>=E-TOKEN`; config has `auth.token: "C-TOKEN"`.
- **Expected:** `Authorization: Bearer E-TOKEN`.

### FX-AUTH-03 — Config token used when no keychain and no env var

- **Setup:** no keychain, no env var; config has `auth.token: "C-TOKEN"`.
- **Expected:** `Authorization: Bearer C-TOKEN`.

### FX-AUTH-04 — Generic `MCLIP_TOKEN` is ignored

- **Setup:** env `MCLIP_TOKEN=GENERIC` only. No per-server env. No keychain. No config token.
- **Invocation:** an HTTP-transport `tools list` against a server that requires auth.
- **Expected:** exit 77 (auth required); the `GENERIC` value MUST NOT appear in any request or in any output.

### FX-AUTH-05 — Generic `--auth-token` rejected

- **Invocation:** `<binary> --auth-token X demo tools list`.
- **Expected:** exit 64; the flag is not part of the conformant surface.

### FX-AUTH-06 — Plaintext-token warning on permissive config file *(advisory, SHOULD-level)*

- **Tag:** ADVISORY — exercises `[MCLIP-11-09]` which is a SHOULD.
- **Setup:** config file with `auth.token: "C-TOKEN"` and POSIX mode `0644` (world-readable).
- **Invocation:** `<binary> demo tools list` (HTTP transport).
- **Expected:** stderr emits exactly one warning the first time the token is read, naming the file and the permissive mode; the warning MUST NOT include the token. Subsequent calls in the same process MUST NOT re-warn.
- **Pass criteria:** advisory pass; absence of warning is acceptable but counts against quality score.

### FX-AUTH-07 — Token never appears in stdout, stderr, or transport errors *(MCLIP-Auth only)*

- **Exercises:** `[MCLIP-14-13]` (no-secret-leak) restricted to non-Safety paths.
- **Setup:** any of the above auth sources resolved.
- **Invocation:** `<binary> demo tools list` against a server that responds with verbose error `data` echoing request headers.
- **Expected:** harness greps stdout + stderr for the token literal; zero matches. The error envelope's `data` field MUST be filtered before emission.

### FX-AUTH-08 — Token redacted in `--dry-run` *(MCLIP-Auth + MCLIP-Safety)*

- **Exercises:** `[MCLIP-14-09]`. Only valid for implementations claiming both MCLIP-Auth and MCLIP-Safety.
- **Setup:** any of the above auth sources resolved.
- **Invocation:** `<binary> --dry-run demo tools call <safe-tool> --text "x"` against an HTTP-transport server.
- **Expected:** stdout is the would-send JSON-RPC request; any `Authorization` header value MUST be the literal `"<redacted>"`; any `auth.token` echoed back MUST also be redacted. Harness greps for the token literal; zero matches.

### FX-AUDIT-01 — Audit event redaction *(conditional, only when implementation declares audit emission)*

- **Tag:** CONDITIONAL — audit logging is OPTIONAL per `[MCLIP-14-08]`. Skip when the implementation does not declare audit emission via `<binary> --version` or an equivalent capability flag.
- **Setup:** as FX-AUTH-07 with the implementation's audit sink configured.
- **Expected:** the emitted audit event contains no token literal; the recommended schema fields from `security-model.md` are present.

---

## 9. Cross-cutting CI-safe composition (§14.4)

Locks down the rollup `[MCLIP-14-11]`–`[MCLIP-14-15]`.

### FX-CI-01a — Destructive tool in non-TTY pipeline, text mode

- **Invocation:** `<binary> demo tools call delete_thing --id 42 < /dev/null > out.txt 2> err.log`.
- **Expected:** exit 64; `out.txt` empty (text mode, no stdout for usage errors); `err.log` names the missing `--yes`; no `tools/call` sent; no credential leaked.

### FX-CI-01b — Destructive tool in non-TTY pipeline, JSON mode

- **Invocation:** `<binary> -o json demo tools call delete_thing --id 42 < /dev/null > out.json 2> err.log`.
- **Expected:** exit 64; `out.json` contains the §5.2 JSON error envelope `{"error": {"code": 64, "message": "...", "origin": "client", "data": {...}}}` (the harness asserts parseable JSON); `err.log` contains the human-readable message; no `tools/call` sent; no credential leaked.
- **Rationale:** §4.5 separates stdout/stderr by channel, not by exit status. JSON-mode errors MUST go on stdout per §5.2 so CI consumers can parse failures.

### FX-CI-01c — `isatty(stdin)` detection vs env var heuristic

- **Setup A:** real PTY (e.g. `script -q` on Linux, ConPTY on Windows) but with `CI=1` and `NO_TTY=1` set in env.
- **Invocation A:** `<binary> demo tools call delete_thing --id 42`. No `--yes`.
- **Expected A:** the client prompts on stderr (or proceeds per MCLIP-Safety rules) because `isatty(stdin)` is true. The env vars MUST NOT downgrade this to non-interactive mode unless the client also declares them as an additional opt-in.
- **Setup B:** redirected stdin (`</dev/null`) with no env vars set.
- **Invocation B:** as A.
- **Expected B:** exit 64 — destructive call refused in non-TTY mode per §14.0 [MCLIP-14-01], regardless of env state.
- **Pass criteria:** the harness verifies that env-variable heuristics never override true TTY state.

### FX-CI-02 — Output determinism across runs

- **Setup:** any read-only tool. Run the same invocation twice with identical inputs.
- **Expected:** the two `-o json` stdout blobs are byte-identical except for fields the harness designates as server-supplied (timestamps, server IDs) and the client's `progressToken`. Object-key ordering MUST match.

### FX-CI-03 — `--version` JSON output is parseable and contract-conformant

- **Exercises:** `[MCLIP-15-09]`.
- **Invocation:** `<binary> -o json --version`.
- **Expected:** stdout parses as JSON; the resulting object has exactly the keys `implementation`, `implementationVersion`, `mclipProfile`, `mclipDraft`, `mcpBaseline`, `modules`; each field obeys the §15.4.3 contract (canonical module enum + canonical order, `mcpBaseline` matches the profile header, `modules` includes `"core"`, no duplicates).
- **Pass criteria:** harness JSON-parses the output, validates field types, and asserts `modules` is a subset of the canonical enum in canonical order with `core` present.

### FX-CI-03b — `--version` text form contains the five required facts *(non-byte-strict)*

- **Exercises:** `[MCLIP-15-10]`.
- **Invocation:** `<binary> --version` with `--output` unset.
- **Expected:** stdout is human-readable text that contains each of the five facts (implementation name, implementation version, profile + draft, MCP baseline, claimed modules) as substrings. The exact wording, ordering, and whitespace are implementation-defined and MUST NOT be compared for byte equality.
- **Pass criteria:** the five facts can be located by substring search using the corresponding values from the JSON form (FX-CI-03). Parsers MUST NOT consume the text form for anything other than presence checks.

### FX-CI-04 — No generic exit 1 in non-interactive paths

- **Setup:** synthesise each error condition mapped to a specific exit code (64, 65, 66, 69, 70, 75, 77, 78, 100).
- **Expected:** each path returns its specific code; none returns 1.

---

## 10. Tagging convention

Each fixture in the harness output MUST report:

- The fixture ID (`FX-...`).
- The rule IDs it exercises (e.g. `[MCLIP-1-02]`, `[MCLIP-14-11]`).
- The module the fixture belongs to (`core`, `safety`, `auth`, …) so wrappers publishing per-module pass/fail can match by tag.

Harness reports SHOULD be machine-readable (JSON or JUnit XML) so wrapper-maintainer CI can publish per-module conformance badges.
