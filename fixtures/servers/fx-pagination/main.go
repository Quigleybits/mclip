// fx-pagination: stdio MCP fixture server with 50 read-only tools to
// exercise `tools/list` auto-pagination.
//
// Tools are named `t01`..`t50`. Each returns its own name as text content.
// The server's page size is fixed at 10 so a client doing the full list
// must follow `nextCursor` four times.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const pageSize = 10
const totalTools = 50

func textOK(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func makeHandler(name string) mcp.ToolHandler {
	return func(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return textOK(name), nil
	}
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "fx-pagination",
		Version: "0.1.0",
	}, &mcp.ServerOptions{
		PageSize: pageSize,
	})

	emptyObjectSchema := map[string]any{
		"type":                 "object",
		"properties":           map[string]any{},
		"additionalProperties": false,
	}

	for i := 1; i <= totalTools; i++ {
		name := fmt.Sprintf("t%02d", i)
		server.AddTool(&mcp.Tool{
			Name:        name,
			Description: "Pagination-fixture tool. Returns its own name as text.",
			InputSchema: emptyObjectSchema,
			Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
		}, makeHandler(name))
	}

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.SetOutput(os.Stderr)
		log.Printf("fx-pagination: server exited: %v", err)
		os.Exit(1)
	}
}
