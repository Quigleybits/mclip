# profile-v0.md — adversarial consistency review (pre-implementation)

Date: 2026-07-06 · Reviewer: Fable 5 (final frontier-access day) · Target: draft v0 rev 2.2
(964 lines, all 16 sections + appendices read in full) · Disposition: fixes applied same day as
**draft revision 2.3** (see change log); judgment calls recorded here with alternatives.

Purpose: this is the cheapest moment to catch cross-module contradictions — mclio's two-phase
parser is about to translate these rules into Go, and every ambiguity below is a place two
conformant implementations would have forked.

## Findings

### F1 · HIGH · `resources watch` is torn between two modules (normative contradiction)
- §1.3 assigns `watch` to **Resources**; §0.7's Resources row includes all of §7, so claiming
  Resources binds `[MCLIP-7-05]`'s MUST: stream NDJSON to stdout.
- But §0.7's Core row says the `ndjson` format is "only required when **MCLIP-Streaming** is
  claimed", and `[MCLIP-9-04]` frames watch's NDJSON output as a *Streaming-module* rule
  ("the only v0 profile-defined streaming command is `resources watch`").
- A Resources-only client simultaneously MUST and need-not implement NDJSON. Two
  implementations fork exactly here.
- **Fix applied (judgment call):** `resources watch` now requires claiming **both**
  MCLIP-Resources **and** MCLIP-Streaming — verb table Module column, §0.7 Resources row, and
  §7.4 all updated. Alternative considered: moving watch wholly into Streaming — rejected
  because the verb's surface (subscribe semantics, capability check) is Resources domain; the
  joint requirement keeps both concerns where they live.

### F2 · HIGH · Exit code 141 violates the canonical-code MUST (normative contradiction)
- `[MCLIP-6-01]`: clients "MUST exit with one of the following codes" — the table has no 141.
- `[MCLIP-6-04]` tells clients to exit 141 on SIGPIPE; Appendix B lists 141 as if canonical.
- Worse: `[MCLIP-14-15]` requires non-interactive exits to come from the §6.1 table — and
  SIGPIPE in a pipeline (`mclip … | head`) is the canonical *non-interactive* event, making
  this a direct MUST-conflict in exactly the CI case §14.4 exists to protect.
- **Fix applied:** `141 · Broken pipe` added to the §6.1 table.

### F3 · MED · Flag-collision rule misses the negation and remap surfaces (gap)
- `[MCLIP-2-13]` only catches same-kebab collisions. Unhandled: (a) a property named `no_foo`
  colliding with the auto-generated `--no-foo` negation of boolean property `foo`
  (`[MCLIP-2-11]`); (b) a property named `arg_output` colliding with the `--arg-output` remap
  that `[MCLIP-2-07]` produces for a reserved-colliding property `output`.
- **Fix applied:** 2-13 broadened — collision detection runs over the FULL generated surface
  (kebab forms + `--no-` negation forms of booleans + `--arg-` remapped forms).

### F4 · MED · Prompt arguments don't inherit §2's collision rules; synopsis misleads (gap)
- `[MCLIP-8-02]`'s synopsis literal `[--arg-name value...]` reads like the `--arg-` prefix
  convention; the body means `--<argument-name>`. Nothing binds prompt-argument flags to the
  §2 reserved-collision remap or ambiguity rules — yet a prompt argument named `output` or
  `input` collides with reserved flags, and Appendix A's intro covered *tool* schemas only.
- **Fix applied:** 8-02 rewritten (synopsis `[--<argument-name> <value>...]`; §2.5/§2.7
  remap + broadened 2-13 apply to prompt arguments identically); Appendix A intro now covers
  prompt argument names.

### F5 · MED · `--raw` under *defaulted* text output (ambiguity → fork)
- `[MCLIP-4-07]`: "`--raw` with `text` or `ndjson` output MUST exit 64." With no `-o` on a
  TTY the default is `text` — so the same command line works piped but exits 64 interactively.
  As written that's legal but hostile, and implementations would "helpfully" imply json.
- **Fix applied (judgment call):** explicit selection of text/ndjson with `--raw` exits 64;
  when `--output` is unset, `--raw` implies `--output json`. Alternative (keep strict 64 even
  when defaulted) rejected: it punishes the interactive user for a TTY they didn't choose and
  makes behaviour depend on stdout's TTY-ness for no safety gain.

### F6 · LOW · Intra-§13.2 precedence unspecified
Duplicate alias across Claude Desktop vs `.vscode/mcp.json` vs `.cursor/mcp.json` had no
ordering. **Fix applied:** the list order in `[MCLIP-13-02]` is now normative precedence.

### F7 · LOW · `[MCLIP-9-01]` scope wording
"every outgoing request that could plausibly produce a progress notification (i.e. …)" made an
exhaustive list normative across modules the client may not claim. **Fix applied:** scoped to
requests the client actually issues from that set.

### F8 · LOW · Reserved-alias cross-reference drift
`[MCLIP-13-05]` pointed at "reserved root command names (§1.4)" while the actual reserved-name
list lives in `[MCLIP-1-08]`. **Fix applied:** cross-reference now names 1-08.

### F9 · INFO · Worked example C.3 implies the confirmation prompt is universal
The interactive prompt is MCLIP-Safety behaviour (`[MCLIP-14-03]`); a Core-only client MAY run
with no prompt (`[MCLIP-14-02]`). **Fix applied:** C.3 captioned "(client claiming
MCLIP-Safety)".

### F10 · INFO · `mclipDraft` post-freeze: "null or absent"
Two spellings of absence undermine the §14.4 determinism goal for `--version -o json` diffing.
**Fix applied:** "MUST be `null`" (absence dropped).

## Verified clean (checked, no action)
Envelope/exit-code cross-coverage (§5 ↔ §6, incl. `--raw` error path and the 100/dual-key tool
envelope) · nullable mapping 2-08 ↔ 2-15 · auth priority 11-01 vs config precedence 13-01 ·
project-local consent symmetry 13-06 ↔ 13-02 · redaction consistency 14-09/14-10/14-13 ·
§14.4's import of 11-02 into Core (14-12 is itself the binding Core rule — deliberate
subsumption, not a leak) · rule-ID stability policy (9-06 hole documented) · §15 claim grammar.

## For the mclio build (day-1 notes)
- The two-phase parser must implement the **broadened 2-13 surface** (kebab + negation + remap)
  as one collision set computed per tool schema — cheapest as a set-union check at schema load.
- `--raw` implies-json lands in phase 1 (pre-discovery parse) since it's a global-flag
  interaction, not schema-aware.
- Conformance fixtures to add: negation-collision tool (F3), prompt with reserved-name argument
  (F4), SIGPIPE-in-pipe (F2), watch claimed-modules matrix (F1).
