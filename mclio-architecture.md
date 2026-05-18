# mclio — Architecture Spec

Version: 0.3 (2026-05-18)
Status: implementation spec, aligned with `profile-v0.md` draft 2.2.

Implementation language: **Go**, using the official Tier 1 Go MCP SDK. Portable installation, fast startup, CI ergonomics, and executable conformance take priority over SDK ecosystem size. TypeScript remains in scope for fixtures and examples, not the `mclio` binary.

This document is the **architecture spec** for the next implementation session. It is a coder's brief, not a tutorial — it names the design decisions a Go implementer needs to make on day one, with rationale.

## Design intent

`mclio` is the production-grade MCP→CLI tool that doubles as **the executable reference for the MCLIP standard**. Its job is two-fold: be the tool people install for daily use, AND be the unambiguous tie-breaker when implementers ask "what does the spec mean here?" Boring, correct, fast, small. It does not compete on features with `mcp2cli`, `MCPorter`, or `MCPShim`; its differentiator is *faithful and polished*, not *feature-rich*. Three implementation principles flow from that:

1. **Spec-faithfulness over ergonomics.** When a spec rule and an ergonomic shortcut conflict, the spec wins. No "we know users want X" overrides.
2. **Thin abstractions.** The CLI is a pipeline: parse args → resolve config → resolve credentials → connect transport → dispatch verb → render output. Each step is one file, one ~100-line function. New behaviour is added by extending the table, not refactoring the pipeline.
3. **Failure modes are deterministic.** Every error path maps to a §6.1 exit code and a §5.2 error envelope. There is no `panic` reachable from user input.

## Package layout

```
github.com/Quigleybits/mclio/
├── cmd/mclio/                  // main entrypoint; tiny — just wires up the pipeline
│   └── main.go
├── internal/
│   ├── argv/                   // global-flag detection + canonical-shape parser (§1.2)
│   ├── config/                 // config sources (§13.1), inheritance (§13.2), schema validation (§13.3)
│   ├── credentials/            // keychain / env / config token resolution (§11.1)
│   ├── envelope/               // §5 JSON envelope construction (success + error)
│   ├── exits/                  // §6.1 exit-code table + helpers
│   ├── render/                 // text / json / ndjson rendering (§4); stdout/stderr discipline (§4.5)
│   ├── safety/                 // §14.0 baseline + §14.1-§14.3 MCLIP-Safety module
│   ├── transport/              // stdio + Streamable HTTP wrappers around the SDK
│   ├── version/                // --version output, §15.4 contract
│   └── verbs/                  // tools/resources/prompts/meta verb handlers
├── go.mod
├── go.sum
└── README.md
```

Rules of thumb:
- Anything that imports the MCP SDK lives under `internal/transport` or `internal/verbs`. The rest of the code talks to a small `transport.Client` interface so swapping SDKs or running unit tests doesn't require live transports.
- Every file under `internal/` exports as few names as possible. The CLI's public surface is the binary, not the package API.

## Dependency choices

The MCP SDK does the heavy lifting (JSON-RPC framing, transport handling, server initialisation). The dependencies named here cover gaps the SDK does not fill.

| Concern | Choice | Reason |
|---|---|---|
| MCP protocol | `github.com/modelcontextprotocol/go-sdk` (Tier 1, pending exact import path verified by research agent) | Canonical reference; tracks spec revisions; first-party. |
| JSON Schema validation | `github.com/santhosh-tekuri/jsonschema/v6` | Best-maintained 2020-12-supporting Go validator. Used to validate tool input against `inputSchema` for `--dry-run` and pre-call validation. |
| Argument parsing | **Custom two-phase parser**, not Cobra/CLI library | See "Two-phase parser contract" below for the required design. A single-phase parser cannot correctly handle generated flags because their schemas are only known after transport connect + `tools/list`. ~300 lines of hand-written parsing wins over wrestling a generic library into the two-phase shape. |
| OS keychain | `github.com/zalando/go-keyring` | Cross-platform (Keychain on macOS, libsecret on Linux, Credential Manager on Windows). Permits graceful absence: returns `ErrNotFound` rather than panicking when the keychain is unavailable. |
| TTY detection | `golang.org/x/term` for `IsTerminal(stdin.Fd())` | Stdlib-adjacent, no transitive deps, works on every supported platform. Required for `[MCLIP-14-11]`. |
| Windows ACL check | `golang.org/x/sys/windows` | Used by `[MCLIP-11-09]` plaintext-token warning on Windows. The POSIX `mode & 0o077 != 0` check is straightforward via `os.Stat`. |
| Pretty errors / colour | **None.** Plain ANSI codes via constants in `internal/render`. `--no-color` and `NO_COLOR` env var disable. | Adding a colour library buys very little and adds dependency surface. |
| Testing | Standard `testing` package; `testscript` for end-to-end CLI tests | The conformance harness (separate deliverable) is the integration story. Unit tests use plain `testing`. |

**Versions:** pin everything in `go.mod`; do not depend on `latest`. The `mclio` binary MUST be reproducible.

## Two-phase parser contract

The §1.2 canonical shape is `<binary> [global-flags...] <server> <category> <verb> [target] [command-flags...]`. The `command-flags` set is **schema-derived per tool/prompt** and cannot be enumerated until transport is up and `tools/list` (or `prompts/list` etc.) has returned. The argv parser therefore runs in two phases, with config + transport + discovery between them.

### Phase 1 — pre-discovery parse (in `internal/argv`)

Run before `config.Resolve`. Inputs: `os.Args`. Outputs: `ParsedShape{ Globals, Server, Category, Verb, Target, RawCommandTail []string }`.

Phase 1 MUST:
- Recognise every reserved global flag from Appendix A of `profile-v0.md` (`--config`, `--output`/`-o`, `--raw`, `--no-color`, `--verbose`/`-v`, `--quiet`/`-q`, `--timeout`, `--transport`).
- Reject any flag-shaped argument (`--something`) that appears BEFORE `<server>` and is not a reserved global flag. This catches generated-flag-too-early cases like FX-GLOBAL-03 — exit 64.
- Resolve `<server>`, `<category>`, `<verb>`, and `[target]` positionally per §1.2.
- Validate `<category>` ∈ {`tools`, `resources`, `prompts`, `meta`}; otherwise exit 64.
- Reject server aliases that begin with `-` or match reserved root names (`servers`, `help`, `version`) per `[MCLIP-1-08]`.
- Stop at `<verb>` (or `[target]`, when the verb requires one) and collect everything afterwards as `RawCommandTail` WITHOUT interpreting it. Phase 1 MUST NOT validate, type-check, or shape-check any token in `RawCommandTail` — those tokens are generated flags whose schemas are not known yet.
- Honour `--help`/`-h` at any position and short-circuit to help output (§1.5).

Phase 1 MUST NOT:
- Consult any config file (that's `config.Resolve`).
- Connect any transport.
- Read or resolve any credential.

### Inter-phase work

After Phase 1:
1. `config.Resolve(ParsedShape.Globals.ConfigPath, ParsedShape.Server)` → server entry, transport spec, auth source list.
2. `credentials.Get(server entry)` → resolved credential per §11.1 (keychain → env → config).
3. `transport.Connect(server entry)` → live `mcp.Client`.
4. `verbs.Discover(client, ParsedShape.Category, ParsedShape.Verb, ParsedShape.Target)` → for `tools call`, fetch the target tool's `inputSchema`; for `prompts get`, fetch the prompt's `arguments` triple list; for `tools list` / `resources list` / etc., no discovery needed.

### Phase 2 — schema-aware parse (in `internal/argv` plus `internal/verbs`)

Inputs: `ParsedShape.RawCommandTail` + the discovered schema (or nil for verbs without per-call schemas). Outputs: the parsed argument set ready for the verb dispatcher.

Phase 2 MUST:
- Generate the flag set per §2.5 lower-kebab-case rule. For reserved-flag collisions, prefix with `--arg-` per `[MCLIP-2-07]`. For property-name collisions (two schema properties mapping to the same generated flag), refuse generated flags and require `--input` / `--input-file` per `[MCLIP-2-13]`.
- Type-check per §2.6: integer/number/boolean/string/enum validation; repeated flags for array properties; dotted flags for shallow primitive objects; reject `oneOf`/`anyOf`/tuple/dotted-name/whitespace-name shapes from generated flags and route them through `--input`/`--input-file` per `[MCLIP-2-14]`.
- Enforce `[MCLIP-2-10]`: `--input` and `--input-file` are mutually exclusive AND neither can be mixed with individual property flags. Mixing → exit 64.
- Enforce `[MCLIP-2-12]` required-property presence; missing → exit 64 with each missing name listed.
- Enforce `[MCLIP-2-16]`: tool input MUST come only from explicit invocation input. Config and env vars MUST NOT supply tool argument defaults.

Phase 2 MUST NOT:
- Reach back into config or env for any tool argument default.
- Allow a generated flag whose name collides with a reserved global flag to be parsed as either the global or the tool argument by accident — the `--arg-` prefix is the only conformant path.

### Why this matters

Day-1 implementers who write a single-phase parser end up either (a) failing FX-GLOBAL-03 (rejecting too late, after some configs are loaded) or (b) failing FX-COLLIDE-02 (accepting bare `--output` as a tool arg because the global pre-pass was missing). The phase boundary is the architectural defence against both failure modes; spelling it out here is the only way the custom-parser choice is defensible against "just use Cobra".

## Pipeline contract

The CLI's runtime is six stages. Each is a pure function except where noted; failures surface as a `mclip.Error` with a §6.1 exit code and a §5.2 envelope.

```
                       +-------------------+
stdin  -->  arg parse  | argv.Parse        |  --> error 64 if syntax invalid
                       +---------+---------+
                                 |
                                 v
                       +-------------------+
                       | config.Resolve    |  --> error 78 if config malformed
                       +---------+---------+
                                 |
                                 v
                       +-------------------+
                       | credentials.Get   |  --> error 77 if auth required but missing
                       +---------+---------+
                                 |
                                 v
                       +-------------------+
                       | transport.Connect |  --> error 69 if transport fails
                       +---------+---------+
                                 |
                                 v
                       +-------------------+
                       | verbs.Dispatch    |  --> error 100 if tool isError, 64/65 if input invalid
                       +---------+---------+
                                 |
                                 v
                       +-------------------+
                       | render.Emit       |  --> stdout/stderr per §4.5
                       +-------------------+
                                 |
                                 v
                            os.Exit(code)
```

Cross-cutting:
- `SIGINT` handler installed in `cmd/mclio/main.go`. On signal, the in-flight transport call is cancelled (best-effort `notifications/cancelled`) and the process exits 130 per §6.2.
- `SIGPIPE` handler per §6.3.
- A single `slog` logger writes to stderr. Verbosity controlled by `--verbose`/`--quiet`. Never writes to stdout.

## Conformance-level boundaries inside the codebase

Module conformance maps to Go build tags AND/OR feature flags. `mclio` ships Core + all 8 modules in v0; the boundaries are kept clean so a downstream consumer can re-target the binary to Core-only if needed.

| Module | Where it lives | Toggle |
|---|---|---|
| MCLIP-Core | All packages above except `internal/safety` for the optional confirmation prompt | always on |
| MCLIP-Resources | `internal/verbs/resources` | build tag `mclip_resources` (default on) |
| MCLIP-Prompts | `internal/verbs/prompts` | build tag `mclip_prompts` (default on) |
| MCLIP-HTTP | `internal/transport/http` | build tag `mclip_http` (default on) |
| MCLIP-Streaming | `internal/verbs/resources/watch.go` + progress handling | build tag `mclip_streaming` (default on) |
| MCLIP-Safety | `internal/safety/{prompt,dryrun,trust}.go` | build tag `mclip_safety` (default on) |
| MCLIP-Auth | `internal/credentials/{keychain,env,config}.go` | build tag `mclip_auth` (default on) |
| MCLIP-Metadata | `internal/verbs/_metadata` (stubbed until Extensions SEP lands) | build tag `mclip_metadata` (default off) |
| MCLIP-Discovery | `internal/config/inherited.go` | build tag `mclip_discovery` (default on) |

Reasoning for default-off MCLIP-Metadata: the Extensions SEP is unfiled. `[MCLIP-16-01]` is explicit that claiming MCLIP-Metadata before the SEP lands is premature. The build tag exists so the package compiles once the SEP is in.

## Build / release pipeline

- **Build matrix:** Go 1.22+ (the SDK's minimum at the time of writing; verify before locking). Build for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`. Reproducible builds via `-trimpath -ldflags='-buildid='`.
- **Static binary by default.** `CGO_ENABLED=0` for Linux; macOS keychain integration requires CGO so macOS build is the exception.
- **Release artifacts:** binary + SHA256 + minisign signature. Distributed via GitHub Releases.
- **Versioning:** the binary's own version is independent of the profile version; both surface via `<binary> --version` per `[MCLIP-15-09]`. CI verifies the JSON-form `mcpBaseline` matches the date in the profile header (drift would be a release-blocker bug).

## Testing strategy

Three layers:

1. **Unit tests** in each `internal/*` package. Standard `go test`. Coverage target: 80% for non-trivial branches; 100% for the §5.2 envelope construction and the §6.1 exit-code mapping (since correctness of these is the entire point of the reference).
2. **CLI scenario tests** using `testscript`. Each scenario is a script with stdin / stdout / stderr / exit-code assertions. Lives in `internal/testscripts/`. Covers all `--help`, `--version`, error-shape cases without needing real MCP servers.
3. **Conformance harness** — the external deliverable per `fixtures-spec.md`. Driven by the harness, not by `go test`. The CI workflow chains: unit → testscript → conformance harness against built fixture servers.

## Coder hand-off — what to do on day 1

1. Verify the Tier 1 Go SDK package path with `go list -m github.com/modelcontextprotocol/go-sdk@latest` (research agent's `real-mcp-servers.md` will confirm).
2. Scaffold the package layout above. Empty stubs for each package; `cmd/mclio/main.go` prints "not implemented" to stderr and exits 70.
3. Build the §6.1 exit-code table in `internal/exits/codes.go` and the §5.2 envelope in `internal/envelope/envelope.go` first — every other package depends on these.
4. Build `argv.Parse` second — it's the highest-leverage piece; once it works, every subsequent test can drive the CLI by argv.
5. Pick one verb (`tools list`) and implement the full pipeline end-to-end against the `fx-echo` fixture server. That single path proves the architecture.
6. Then expand verb-by-verb, gating each on its own conformance fixtures.

## Open implementation questions for the day-1 coder

1. **Keychain on Linux.** `libsecret` is the conventional path but it's a runtime dep that not every Linux environment has. Decide whether to: (a) require libsecret and document it, (b) gracefully degrade to env-var-only when libsecret is missing, (c) ship a pure-Go alternative. Recommendation: (b), matching how `go-keyring` already behaves.
2. **HTTP timeouts.** §12.5 says timeout for non-streaming requests is implementation-defined but MUST be finite. Recommend 60s default with `--timeout` override. Streaming requests (resources/watch) MUST NOT time out on idle output per §12.5; verify the SDK's HTTP client doesn't impose an idle timeout under the hood.
3. **SDK reading of `_meta`.** §16.2 says non-`mclip.*` `_meta` keys pass through to consumers in `tools describe` output but MUST NOT influence command shape. Confirm the SDK exposes the raw `_meta` map; if it strips unknown keys, we need a workaround.
4. **Config file format detection.** `[MCLIP-13-04]` requires accepting `mcpServers` as a synonym for `servers`. Decide whether to detect-and-coerce on load or detect-and-prefer at lookup time. The former is simpler; the latter is closer to "warn if both present".
