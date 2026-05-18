# SEP — MCLIP CLI Metadata Extension

SEP: TBD — to be assigned on PR open
Title: MCLIP CLI Metadata Extension
Author: Aidan Quigley (`@Quigleybits`, aidanjohnquigley@gmail.com)
Status: Pre-Draft — gated on the SEP-2133 Extensions Track prerequisites being satisfied (see "Filing prerequisites" below); will move to Draft upon those + sponsor signal
Type: Extensions Track (per SEP-2133)
Created: 2026-05-16
Requires: MCP 2025-11-25
Targets: MCLIP profile v0 §16 (MCLIP-Metadata module)
PR: TBD — added on PR open
Sponsor: TBD
Extension Maintainer: TBD — see "Filing prerequisites"
Reference Implementation Host: TBD — see "Filing prerequisites"

## Filing prerequisites

Per SEP-2133, the Extensions Track requires:

1. **A named extension maintainer** committed to ongoing stewardship of the `mclip.*` keys defined here.
2. **An official SDK reference implementation** of the extension's keys (not just a downstream CLI implementation).

Neither is satisfied at the time of drafting. This SEP cannot move beyond Pre-Draft until both are resolved. The cleanest path is to host the reference implementation in the official Go SDK (`github.com/modelcontextprotocol/go-sdk`) as opt-in helper methods that parse `mclip.*` keys; the Go SDK maintainers would be the natural extension-maintainer candidates. Alternative: file the metadata keys under a different SEP track that does not have the official-SDK requirement (e.g. an Informational SEP describing recommended `_meta` conventions), with the trade-off that downstream conformance machinery becomes weaker.

This draft proceeds on the assumption that the prerequisite path will be resolved before formal filing; the technical content below is design-complete and ready for review even though the procedural gate is not yet met.

## Abstract

This SEP defines the `mclip.*` metadata key namespace under MCP's `_meta` field. The keys are optional hints MCP servers may attach to tools, resources, and prompts to improve their CLI surface when consumed by a [MCLIP-Metadata](profile-v0.md#16-mclip-metadata-module--cli-hint-extension)-conformant client. Keys are additive: vanilla MCP servers that publish none of them remain fully MCLIP-Core-conformant, and clients that ignore them remain MCLIP-Core-conformant. The MCLIP profile reserves the namespace; this Extensions SEP carries the key catalogue, types, and semantics.

## Motivation

MCLIP-Core derives its CLI surface entirely from MCP server discovery: `tools/list`, `resources/list`, `prompts/list`, and their associated JSON schemas. This works correctly for any MCP server without requiring vendor work. It is, however, **mechanically deterministic** — it cannot inject domain-specific judgement that a vendor reasonably wants surfaced on the CLI: a friendlier alias for a long tool name, a worked example for an unobvious flag combination, a heightened confirmation message for a tool whose blast radius is bigger than the schema can express.

The optional CLI-hint extension lets vendors opt into providing these affordances without changing the wire protocol and without breaking clients that don't know about them. This is a quality lift, never a prerequisite.

The MCLIP profile (v0 §16) reserves the `mclip.*` namespace and constrains its security posture (server-supplied metadata MAY tighten client safety behaviour but MUST NOT relax it). This SEP defines the specific keys.

## Specification

### Namespace

All keys defined here live under the `mclip.` prefix in any MCP `_meta` field. A MCLIP-Metadata-conformant client honours these keys per the rules below. A non-MCLIP-Metadata client ignores them per `[MCLIP-16-04]` (forward compatibility — unknown `mclip.*` keys are ignored silently).

### Key catalogue

The following keys are defined in this SEP. Each key declares its target (which MCP object can carry it: tool, resource, prompt, or any), its JSON type, whether it influences safety behaviour, and a one-line semantic statement.

#### `mclip.aliases`

- **Target:** tools, resources, prompts.
- **Type:** array of strings.
- **Safety-relevant:** no.
- **Semantic:** alternative names a MCLIP client MAY accept as command targets for this object. Each alias MUST be a lower-kebab-case identifier matching the regex `^[a-z][a-z0-9-]*$`. Aliases MUST NOT contain whitespace. The original (canonical) name MUST continue to work; aliases are additive.
- **Conflict resolution:** if two distinct objects (within the same server, same category) declare overlapping aliases, the client MUST refuse all of them — exit 64 with a clear error — rather than silently choosing.

#### `mclip.examples`

- **Target:** tools, prompts.
- **Type:** array of objects. Each object: `{ "description": string, "input": object }`.
- **Safety-relevant:** no.
- **Semantic:** worked invocation examples a client MAY render in `<binary> <server> tools describe <tool>` output. The `input` is a JSON object whose keys match the tool's `inputSchema` properties. Clients SHOULD validate examples against the schema and SHOULD omit examples that fail validation (logging a warning to stderr).

#### `mclip.destructive`

- **Target:** tools.
- **Type:** boolean.
- **Safety-relevant:** **YES — TIGHTENING ONLY.**
- **Semantic:** **`true`** declares the tool destructive in addition to whatever `annotations` already say. Clients MUST treat a `mclip.destructive: true` tool as potentially destructive per profile §14.0 even if `annotations.readOnlyHint == true`. **`false` has no effect** — it MUST NOT relax safety. (Profile §14.0 and §16.3 are explicit: server-supplied metadata cannot loosen client safety defaults; only `safety.trustAnnotations` per-server config can.)
- **Why one-way:** the asymmetric design prevents a hostile or buggy server from disabling confirmations by setting `mclip.destructive: false` on an actually-dangerous tool.

#### `mclip.sensitiveProperties`

- **Target:** tools.
- **Type:** array of strings. Each string is a JSON Pointer (per RFC 6901) into the tool's `inputSchema` properties tree, identifying which input properties carry sensitive values. Example: `["/password", "/auth/token", "/db/connection_string"]`.
- **Location:** lives under the tool's `_meta` per the §16 namespace rule (`_meta.mclip.sensitiveProperties`), NOT as a JSON-Schema vocabulary keyword inside `inputSchema`. This preserves the namespace promise from profile-v0.md §16.2 (MCLIP keys live under `_meta`, full stop).
- **Safety-relevant:** YES — TIGHTENING ONLY.
- **Semantic:** every property whose JSON Pointer is listed here MUST have its value redacted in any client-emitted artefact: `--dry-run` output (per profile `[MCLIP-14-10]`), audit log events, error envelopes that echo input. Redaction MUST use the literal string `"<redacted>"`. An empty array has no effect; absence is equivalent to an empty array.
- **Pointers that do not resolve to a property in `inputSchema`** MUST be ignored (a server that lists `/nonexistent` is not a conformance failure; the client just has nothing to redact). Clients MUST NOT error on unresolvable pointers.
- **One-way:** there is no `mclip.nonSensitiveProperties` opposite — sensitivity can only be added, never removed. The asymmetric design matches `mclip.destructive` and the overall tightening-only posture.

Example:

```json
{
  "name": "rotate_api_key",
  "inputSchema": {
    "type": "object",
    "properties": {
      "current_key": { "type": "string" },
      "label": { "type": "string" }
    },
    "required": ["current_key"]
  },
  "_meta": {
    "mclip.destructive": true,
    "mclip.sensitiveProperties": ["/current_key"]
  }
}
```

#### `mclip.confirm_message`

- **Target:** tools.
- **Type:** string.
- **Safety-relevant:** YES — ADDITIVE-ONLY (cannot weaken the client's safety signal).
- **Semantic:** server-supplied **supplementary** warning text that MCLIP-Safety clients MAY render *in addition to* (NEVER in place of) the client's own confirmation prompt. The client's prompt — naming the server, the tool, and the `[y/N]` choice — is always emitted unchanged so the user's safety signal does not depend on server text. The server-supplied string is appended below the prompt as escaped warning text.
- **Rendering rules** (all MUST):
  - The client's standard prompt MUST be emitted first, with its full `<server> tools call <tool>` identification and the `[y/N]` choice.
  - The server-supplied string MUST be escaped: control characters (ASCII 0x00–0x1F except `\n`), ANSI escape sequences, and terminal-cursor-movement bytes MUST be stripped before rendering. The result MUST NOT be able to overwrite the prompt or its `[y/N]` choice.
  - The server-supplied string is appended on the line(s) below the client prompt, clearly delimited (e.g. prefixed with `> ` or a `[server note]` marker).
  - Clients MUST truncate the server string to 280 characters with an ellipsis; longer messages are not a conformance failure but truncation is.
- **No placeholder substitution.** Earlier drafts allowed `{tool}` and `{server}` placeholders; those are removed. Server text is opaque to the client; the client already knows the tool and server names and emits them in its own prompt.
- **Why additive-only:** a replace-mode `confirm_message` would let a hostile or buggy server replace `Proceed with linear tools call delete_issue? [y/N]:` with `Continue? [y/N]:`, stripping the user's safety signal. That violates the tightening-only posture every other MCLIP metadata key obeys. Additive-only is the safe equivalent.
- **Untrusted-server reminder.** Server text is rendered to the user but treated as untrusted; clients SHOULD make the visual distinction between client prompt and server warning unambiguous (e.g. dim colour, label) so a sophisticated user can tell at a glance what the server said vs what the client said.

### What this SEP does NOT define

Several keys were considered and deferred:

- **`mclip.preferred_verb`** — would let servers suggest CLI verb names different from their MCP tool names. Deferred because the cross-server verb-collision rules become hard to specify and the value-vs-complexity ratio is unclear. Revisit in a v1 of this SEP if vendor demand emerges.
- **`mclip.hidden`** — would let servers hide tools from `<binary> <server> tools list` output. Rejected because it conflicts with MCLIP's "MCP server is the source of truth for discovery" principle (profile §0.5). Hiding tools would let servers shape the CLI surface beyond what their `tools/list` advertises, which is a form of vendor lock-in this profile is explicitly designed to avoid.
- **`mclip.color`, `mclip.icon`** — pure-presentation hints. Out of scope for an interoperability standard; left to each implementation's UI.

### Worked examples

A tool annotated for a CLI-friendly alias and a worked example:

```json
{
  "name": "create_repository_collaborator_with_permissions",
  "description": "Add a collaborator to a repo with explicit permissions",
  "inputSchema": { "...": "..." },
  "_meta": {
    "mclip.aliases": ["add-collab"],
    "mclip.examples": [
      {
        "description": "Add user as triage on this repo",
        "input": { "repo": "modelcontextprotocol/modelcontextprotocol", "user": "alice", "permission": "triage" }
      }
    ]
  }
}
```

A destructive tool with a custom (additive) confirmation message and sensitive input properties — note all MCLIP metadata lives under `_meta`, never as JSON-Schema vocabulary keywords:

```json
{
  "name": "rotate_api_key",
  "description": "Generate a new API key and invalidate the previous one",
  "inputSchema": {
    "type": "object",
    "properties": {
      "current_key": { "type": "string" },
      "label": { "type": "string" }
    },
    "required": ["current_key"]
  },
  "_meta": {
    "mclip.destructive": true,
    "mclip.sensitiveProperties": ["/current_key"],
    "mclip.confirm_message": "Previous key becomes invalid immediately; downstream services using it will start failing."
  }
}
```

## Rationale

### Why an Extensions Track SEP, not part of the main MCLIP profile?

Decoupling lets the main profile (the Standards Track SEP) advance through review without being held up by key-by-key bikeshed on this catalogue. Both SEPs reference the namespace reservation in profile §16 but the key set evolves on its own cadence; future revisions of this SEP can add keys without re-opening profile review. Per SEP-2133, Extensions Track is the canonical home for this kind of additive vocabulary.

### Why tightening-only safety semantics?

MCP spec is explicit: clients MUST consider tool annotations untrusted unless they come from trusted servers. Applying the same posture to MCLIP metadata is the only safe default. A hostile server that could set `mclip.destructive: false` to bypass confirmation would be a footgun. The trusted-server escape hatch (`safety.trustAnnotations` per profile §14.3) is the user-controlled mechanism for relaxing defaults; server-supplied metadata is not.

### Why `mclip.sensitiveProperties` under `_meta`, not a JSON-Schema vocabulary keyword?

An earlier draft put sensitivity inside `inputSchema` as a custom JSON-Schema vocabulary keyword (`{"type": "string", "mclip.sensitive": true}`). That violated the namespace promise in profile-v0.md §16 (MCLIP keys live under `_meta`). The fix: keep sensitivity at the tool level under `_meta`, using JSON Pointer (RFC 6901) to identify which properties are sensitive. This preserves a single source of truth for the namespace rule, lets clients that don't claim MCLIP-Metadata ignore the key cleanly per `[MCLIP-16-04]`, and avoids any need to extend JSON Schema's vocabulary mechanism. Tools with sensitive properties in deeply nested structures still work via pointer paths like `/auth/token`.

### Why no domain semantics?

This SEP — like the MCLIP profile itself (§0.6) — defines mechanics, not semantics. There is no `mclip.category` or `mclip.purpose` because cross-vendor semantic taxonomies are out of scope for an interoperability standard. Vendors who want to expose taxonomy data can use their own `_meta` namespace (e.g. `vendor.foo.category`) which MCLIP clients pass through to consumers per profile `[MCLIP-16-02]` but do not interpret.

## Backwards Compatibility

This SEP is purely additive.

- MCP servers that publish no `mclip.*` keys are unaffected.
- MCLIP-Core clients (not claiming MCLIP-Metadata) ignore these keys per `[MCLIP-16-04]`.
- Future Extensions SEP revisions can add new keys; clients that don't recognise the new keys ignore them silently per the same rule.

## Reference Implementation

The MCLIP reference CLI (per `mclio-architecture.md`) will implement MCLIP-Metadata as a build-tag-controlled package (`internal/verbs/_metadata`). The package is stubbed pending this SEP reaching Accepted status; the build tag is off by default until then. Once this SEP is Accepted, the implementation will:

1. Honour `mclip.aliases` in command resolution.
2. Render `mclip.examples` in `tools describe` output.
3. Treat `mclip.destructive: true` as a positive destructive signal per profile §14.0.
4. Redact every property value whose pointer is listed in `_meta.mclip.sensitiveProperties` in `--dry-run` output per profile `[MCLIP-14-10]`, using the literal `"<redacted>"` sentinel.
5. Append `mclip.confirm_message` as escaped, control-character-stripped, 280-char-truncated supplementary warning text *below* the client's own MCLIP-Safety prompt (never replacing it).

Conformance fixtures for this SEP will be added to `conformance-fixtures.md` (currently v0.2) once the reference implementation is in.

## Security Considerations

- The tightening-only semantics for `mclip.destructive` and `mclip.sensitiveProperties` are the load-bearing security property. `mclip.destructive: false` MUST be ignored, NOT treated as "this is safe". `mclip.sensitiveProperties` has no negative form at all (no opposite key, no way to mark properties as "definitely-not-sensitive"). Implementers MUST resist the temptation to extend either with relaxing semantics.
- `mclip.confirm_message` is additive-only: the client's own prompt (with `<server> tools call <tool>` identification and the `[y/N]` choice) is always rendered unchanged, and the server-supplied text is appended below as escaped supplementary warning text. Replace-mode behaviour would let a server strip the user's safety signal; additive-only is the only safe shape. Implementations MUST strip control characters, ANSI escape sequences, and cursor-movement bytes from the server string before rendering, to prevent terminal-injection attacks where a server crafts a message that overwrites or hides the prompt above it.
- `mclip.examples` content is server-supplied JSON. Implementations rendering examples MUST treat them as untrusted input — never `eval`, never shell-substitute, never URL-decode. The `input` object is a literal example, not an executable.
- `mclip.aliases` MUST be regex-validated before use; an alias containing whitespace or shell metacharacters could enable command-injection-style confusion in user-facing help text.

## Open Questions

- **Should `mclip.aliases` be allowed on the server alias itself** (i.e. server-supplied alternative names for the server, not just for individual tools)? Currently not in scope; the user controls server aliases via config. Worth revisiting if users find themselves coining the same alias for every server.
- **Should there be a `mclip.deprecated` boolean** to let servers mark tools as deprecated without removing them? Tempting, but the MCP spec already has a separate mechanism for deprecation. Defer until MCP spec stance is clearer.

## References

- MCLIP profile v0 §16 (MCLIP-Metadata module): `profile-v0.md`
- MCLIP security model: `security-model.md`
- SEP-2133 (Extensions framework): https://modelcontextprotocol.io/seps/2133-extensions
- MCP 2025-11-25 spec: https://modelcontextprotocol.io/specification/2025-11-25
