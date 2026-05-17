// fx-echo: minimal stdio MCP fixture server.
//
// Exposes a single read-only `echo` tool that returns its input unchanged.
// Used as the baseline server for FX-GLOBAL-01/02/03, FX-RAW-01, FX-CI-02
// (determinism), and FX-CI-04 (exit-code variants).
package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type echoArgs struct {
	Text string `json:"text" jsonschema:"the text to echo back unchanged"`
}

func echo(_ context.Context, _ *mcp.CallToolRequest, args echoArgs) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: args.Text}},
	}, nil, nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "fx-echo",
		Version: "0.1.0",
	}, nil)

	readOnly := true
	mcp.AddTool(server, &mcp.Tool{
		Name:        "echo",
		Description: "Echo input text back unchanged.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: readOnly},
	}, echo)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.SetOutput(os.Stderr)
		log.Printf("fx-echo: server exited: %v", err)
		os.Exit(1)
	}
}
