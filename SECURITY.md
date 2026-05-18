# Security policy

MCLIP is a draft specification, not a runtime service. "Security" here means concerns about the spec text itself, not about a particular MCP server or wrapper implementation.

## In scope

Please report:

- An ambiguity in `profile-v0.md` or `security-model.md` where two conformant clients could both pass a fixture but expose users to credential leakage, command injection, or unsafe destructive-action behaviour.
- A MUST or SHOULD that, if followed literally, would create an exploitable footgun.
- A missing constraint that would let a malicious MCP server abuse `mclip.*` metadata, `--dry-run`, the trusted-server escape hatch, or the project-local config consent rule.
- An ambiguity in `conformance-fixtures.md` where the fixture-pass criteria don't actually exercise the security-relevant behaviour the rule was meant to cover.

The [`security-model.md`](security-model.md) document defines the trust boundaries the spec is trying to enforce. Anything that weakens those boundaries — quietly — is in scope here.

## Out of scope for this repo

- Implementation bugs in third-party wrappers (`mcp2cli`, `MCPorter`, `MCPShim`, `f/mcptools`, `FastMCP generate-cli`, `mcpc`, `developit/mcp-cmd`, IBM `mcp-cli`). Report those to the wrapper's own security contact.
- Vulnerabilities in the MCP protocol itself. Report at [modelcontextprotocol/modelcontextprotocol](https://github.com/modelcontextprotocol/modelcontextprotocol).
- Vulnerabilities in any real MCP server (GitHub MCP Server, Context7, `server-everything`, etc.). Report to those projects directly.

## How to report

For non-sensitive findings (spec wording, fixture coverage gap), open a public issue. Most spec-level security issues are public-by-design — being able to discuss the threat openly is part of what fixes them.

For findings that you believe would be unsafe to disclose publicly before a fix lands (rare for a draft spec, but possible), email `aidey@mclip.dev` (in setup) with `[SECURITY]` in the subject line. Until that mailbox is live, use `aidanjohnquigley@gmail.com` with the same subject prefix.

Expected response time:

- Public issues: best-effort, within 7 days of filing.
- Private email: best-effort acknowledgement within 7 days; if the finding is substantive, a tracking issue (or, for issues that genuinely cannot be public yet, a private working note) within 14 days.

This is a solo-maintained pre-SEP project. There is no formal SLA.

## Disclosure

The default is coordinated public disclosure once a spec fix has landed. There is no embargo program. The spec is in draft (v0); rule-text changes during draft do not break stable rule IDs.

## See also

- [`security-model.md`](security-model.md) — the full normative trust model the spec is enforcing.
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — general contribution flow (non-security).
