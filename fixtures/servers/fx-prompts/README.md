# fx-prompts

Stdio MCP fixture server exposing the Prompts surface — a single language-aware
greeter — as scaffolding for the MCLIP-Prompts module conformance suite.

## Prompts

| Prompt | Arguments | Output |
|---|---|---|
| `greet` | `name` (string, **required**); `lang` (string, optional; one of `en\|fr\|es`, default `en`) | One `user`-role message: `"Hello, <name>."` / `"Bonjour, <name>."` / `"Hola, <name>."` |

## Backs

No FX-PROMPTS-* fixtures are defined yet in `conformance-fixtures.md`. This
server exists as scaffolding so the harness can exercise the prompts surface
once those fixtures are written.

The relevant rule it backs by construction is **[MCLIP-8-03]** — prompt
argument values are strings at the protocol layer (`map[string]string` in
the Go SDK's `GetPromptParams.Arguments`); the client must not silently
coerce values to other types.
