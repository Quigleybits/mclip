# AGENTS.md

Guidance for AI coding agents working in this repository. (Humans: start with `README.md` and `CONTRIBUTING.md`.)

## What this is

MCLIP — the **M**CP **C**ommand-**L**ine **I**nterface **P**rofile: a CLI conformance profile over the Model Context Protocol. It is **not** a new protocol and requires no vendor-side work; it is a client-side profile that pins how any MCP server is rendered into a uniform CLI surface (command shape, flags, JSON envelope, exit codes). Full context: `README.md`.

## What is authoritative

`profile-v0.md` is the **normative specification**. The client config schema (`site/public/schemas/config/v0.json`), the profile manifest (`site/public/profile/v0.json`), and the conformance fixtures are **derived** — if any of them disagrees with `profile-v0.md`, the prose specification governs.

Public reading order: `profile-v0.md` → `security-model.md` → `conformance-fixtures.md` → `adoption-guide.md`.

## Rules that bite when editing

1. **Quote, don't paraphrase, normative text.** `profile-v0.md`, the SEP drafts, `security-model.md`, and `conformance-fixtures.md` carry rule IDs (`[MCLIP-§-NN]`) and conformance language. Silent rewording is how conformance drifts.
2. **Rule IDs are stable — never renumber or reuse.** A retired rule's ID stays retired (e.g. `MCLIP-9-06`).

Full house rules and the contribution flow: `CONTRIBUTING.md`.

## Commands

- Build the site: `cd site && npm install && npm run build` (Astro static → `site/dist/`).
- Verify the site: `node site/scripts/site-verify.js`.
- Fixture servers + verify harness: see `fixtures/` (Go, MCP Go SDK v1.6.0).

## Agent entry points

- Web front door: <https://mclip.dev/llms.txt>
- Machine-readable profile manifest: <https://mclip.dev/profile/v0.json>

## Status

Pre-SEP draft. The specification set and the nine synthetic MCP fixture servers + Go verify harness are built. The `mclio` reference CLI and the full conformance harness are in development, not yet released — do not assume a working reference binary exists.

This status sentence is mirrored in four places — `site/public/llms.txt`, `site/public/profile/v0.json` (`status`), `site/src/pages/index.astro`, and here. When the project's status changes, update all four together.
