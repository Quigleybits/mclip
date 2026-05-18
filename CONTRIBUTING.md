# Contributing to MCLIP

MCLIP is a draft conformance profile over the Model Context Protocol. The repo is the source of truth for the specification, the security model, the conformance fixture catalogue, and the synthetic fixture servers that back conformance testing. This guide is the operating contract for working on any of those.

## Where to start

If you are new to the project, read the specification set in this order:

1. [`profile-v0.md`](profile-v0.md) — the normative profile (Core + 8 modules).
2. [`security-model.md`](security-model.md) — trust boundaries and the destructive-action policy.
3. [`conformance-fixtures.md`](conformance-fixtures.md) — the rule-to-fixture catalogue.

Then [`adoption-guide.md`](adoption-guide.md) if you are a wrapper maintainer, or [`fixtures-spec.md`](fixtures-spec.md) / [`mclio-architecture.md`](mclio-architecture.md) if you want to engage on the reference implementation.

## What to file an issue for

Good fits for issues:

- A factual error in `real-mcp-servers.md` about an MCP server you maintain or know well.
- An ambiguity in `profile-v0.md` where two conformant implementations could both pass a fixture but produce different observable behaviour.
- A MUST in `profile-v0.md` that would force a breaking change to your wrapper's existing surface — push back with the specific rule ID.
- A gap in `conformance-fixtures.md` — a rule with no fixture, or a fixture whose pass criteria don't match the rule.
- An inconsistency between `profile-v0.md`, `security-model.md`, and `sep-standards-mclip-profile.md`.

Out of scope for this repo:

- Implementation bugs in third-party wrappers (`mcp2cli`, `MCPorter`, etc.) — file with the wrapper's own project.
- MCP protocol design questions — those belong in `modelcontextprotocol/modelcontextprotocol`.

## Editing the spec — house rules

These rules exist because the documents are normatively cited from each other and from the SEP draft. Drift is what makes conformance unenforceable.

1. **Quote normative text, don't paraphrase.** `profile-v0.md`, `security-model.md`, `sep-standards-mclip-profile.md`, and `sep-extensions-mclip-metadata.md` carry rule IDs (`[MCLIP-§-NN]`). Silent rewording across reads is how conformance drifts.
2. **Auto-bump version + date on every spec edit.** `profile-v0.md` uses `Draft v0 — revision X.Y (YYYY-MM-DD)`. Bump the revision and the date in the same commit as the edit. Same rule for `security-model.md` (`Draft X.Y`) and `conformance-fixtures.md` (`vX.Y`).
3. **One change per PR.** A spec edit, a fixture addition, and a doc reword are three PRs, not one.
4. **Rule IDs are stable.** Once a rule has a `[MCLIP-§-NN]` ID and the spec is past draft 2.x, the ID does not move. If the rule needs to be retired, mark it as superseded; do not renumber.

## Decision records (`decisions/`)

Significant project decisions — naming, scope changes, server-set ratification, governance routing — live as dated files under `decisions/`. The format is **Context · Decision · Consequences · Alternatives considered**.

- One file per decision. Filename: `decisions/YYYY-MM-DD-<short-slug>.md`.
- Append-only — once written, do not rewrite a decision file. To revise, write a new dated file that explicitly supersedes the old one. Both stay in history.

## PRs

- Keep PR titles short (under 70 characters). Use the description for detail.
- For spec edits, reference the rule ID in the PR title (e.g. `[MCLIP-14-13] clarify error-envelope redaction`).
- For fixture changes, run `fixtures/build.sh` (or `fixtures/build.ps1` on Windows) and confirm `fixtures/verify` passes against all 9 servers before opening the PR.

## Pre-SEP status

The repo is pre-SEP. The seven core specification documents are written and internally consistent; the 9 fixture servers + verify harness are built. The remaining build queue is the `mclio` binary, the conformance harness, and a companion consumer app. External coordination — Discord post, draft SEP PR, wrapper-maintainer outreach — is held until that build queue is complete.

If you read the spec and find something that should change, file an issue now. Earlier feedback is much cheaper than feedback after the SEP is filed.

## Contact

- GitHub issues — preferred for anything substantive.
- `aidey@mclip.dev` (in setup) — for things that genuinely don't fit a public issue.
- See [`SECURITY.md`](SECURITY.md) for security concerns about the spec itself.

## Licence

By submitting a contribution you agree to license it under the project's [Apache-2.0](LICENSE) terms.
