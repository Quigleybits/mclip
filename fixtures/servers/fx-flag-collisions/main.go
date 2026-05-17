// fx-flag-collisions: stdio MCP fixture server that advertises tools whose
// input schemas are designed to trigger MCLIP flag-name collisions in the
// CLI under test.
//
// Property names cannot be expressed as Go struct fields without losing case
// distinction (snake_case vs snakeCase) or accepting dots ("foo.bar"). The
// server therefore uses the SDK's low-level Server.AddTool path with a raw
// JSON Schema (map[string]any) as InputSchema. Tool handlers return a fixed
// text content — the conformance suite exercises the CLIENT's flag generator,
// not the server's argument handling.
package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func textOK(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func makeWidget(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return textOK("ok"), nil
}

func dump(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return textOK("ok"), nil
}

func polyglot(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return textOK("ok"), nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "fx-flag-collisions",
		Version: "0.1.0",
	}, nil)

	makeWidgetSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"snake_case": map[string]any{"type": "string"},
			"snakeCase":  map[string]any{"type": "string"},
		},
		"required":             []string{"snake_case", "snakeCase"},
		"additionalProperties": false,
	}
	server.AddTool(&mcp.Tool{
		Name:        "make_widget",
		Description: "Two properties whose normalised flag forms collide on `--snake-case`. Tests FX-COLLIDE-01 (fallback to `--input`).",
		InputSchema: makeWidgetSchema,
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, makeWidget)

	dumpSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"output": map[string]any{"type": "string"},
		},
		"required":             []string{"output"},
		"additionalProperties": false,
	}
	server.AddTool(&mcp.Tool{
		Name:        "dump",
		Description: "Property `output` collides with reserved global flag. Tests FX-COLLIDE-02 (renamed to `--arg-output`).",
		InputSchema: dumpSchema,
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, dump)

	polyglotSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"union": map[string]any{
				"oneOf": []any{
					map[string]any{"type": "string"},
					map[string]any{"type": "integer"},
				},
			},
			"tuple": map[string]any{
				"type": "array",
				"prefixItems": []any{
					map[string]any{"type": "string"},
					map[string]any{"type": "integer"},
				},
				"items": false,
			},
			"foo.bar": map[string]any{"type": "string"},
		},
		"required":             []string{"union", "tuple", "foo.bar"},
		"additionalProperties": false,
	}
	server.AddTool(&mcp.Tool{
		Name:        "polyglot",
		Description: "Complex-shape properties (oneOf union, tuple via prefixItems, dotted name). Tests FX-COLLIDE-03 (complex-shape fallback to `--input`).",
		InputSchema: polyglotSchema,
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, polyglot)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.SetOutput(os.Stderr)
		log.Printf("fx-flag-collisions: server exited: %v", err)
		os.Exit(1)
	}
}
