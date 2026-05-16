# MCLIP Product Brief

MCLIP, the MCP Command-Line Interface Profile, is a standard for deriving a predictable CLI surface from the Model Context Protocol. If a service exposes capabilities through MCP, MCLIP defines how those capabilities should appear in the command line.

The problem is fragmentation. MCP-to-CLI wrappers already exist, but they use different command shapes, argument formats, JSON output modes, config conventions, auth behavior, and exit codes. A script or agent workflow written for one wrapper often cannot run on another. That undermines one of MCP's biggest promises: a common interface to external tools.

MCLIP standardizes the translation layer. It maps MCP discovery data into portable CLI behavior: server aliases become addressable targets, MCP tools become callable commands, JSON Schema inputs become flags or JSON input, and MCP results become stable text, JSON, or NDJSON output. The same MCP server should produce the same command grammar, flags, output envelope, errors, and exit codes through any MCLIP-conformant client.

This is especially useful for the agentic future, but not only for agents. Humans, scripts, CI systems, and autonomous agents all need deterministic interfaces they can call safely from shells and automation environments. MCP is primarily a protocol for model-tool interaction; MCLIP makes the same product surfaces usable through a standard command-line contract. That means agents can automate reliably, scripts can stay portable, and humans can inspect or operate services directly without each vendor inventing and documenting a bespoke CLI.

MCLIP is not a new protocol and does not require vendor-specific command definitions for Core. It is a profile over MCP: command mechanics, schema-to-flag mapping, stdin handling, output formats, error envelopes, exit codes, server discovery, and baseline safety. The outcome is simple: if a service supports MCP, humans and agents should know how to use it from the CLI.
