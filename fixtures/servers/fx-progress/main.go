// fx-progress: stdio MCP fixture server exposing one tool, `slow_count`,
// that emits a series of `notifications/progress` messages before returning.
//
// Backs the `[MCLIP-9-02]` rule that progress goes to stderr while the JSON
// result goes cleanly to stdout. The fixture is fully deterministic — no
// sleep between emits — so the client receives exactly N progress
// notifications in order, then the result.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type slowCountArgs struct {
	Target int `json:"target" jsonschema:"how many progress steps to emit before returning"`
}

func slowCount(ctx context.Context, req *mcp.CallToolRequest, args slowCountArgs) (*mcp.CallToolResult, any, error) {
	if args.Target < 0 {
		return nil, nil, errors.New("target must be >= 0")
	}
	token := req.Params.GetProgressToken()
	for i := 1; i <= args.Target; i++ {
		if token != nil {
			_ = req.Session.NotifyProgress(ctx, &mcp.ProgressNotificationParams{
				ProgressToken: token,
				Progress:      float64(i),
				Total:         float64(args.Target),
			})
		}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("done %d", args.Target)}},
	}, nil, nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "fx-progress",
		Version: "0.1.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "slow_count",
		Description: "Emit `target` progress notifications, then return the text \"done <target>\".",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, slowCount)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.SetOutput(os.Stderr)
		log.Printf("fx-progress: server exited: %v", err)
		os.Exit(1)
	}
}
