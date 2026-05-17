// fx-destructive: stdio MCP fixture server exposing destructive and
// error-producing tools, plus a cancellable sleep tool.
//
// delete_thing deliberately carries no annotations — exercises the
// `[MCLIP-14-0]` rule "no annotations → potentially destructive".
// fails always errors — exercises FX-RAW-02 (CallToolResult.isError pass-through).
// sleep is cancellable — exercises FX-SIGINT-01 (the client-side cancellation
// path must reach the server as `notifications/cancelled`, which the SDK
// surfaces as context.Done()).
// safe_read is a positive control with readOnlyHint: true.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type deleteArgs struct {
	ID int `json:"id" jsonschema:"the id of the thing to delete"`
}

type deleteResult struct {
	OK bool `json:"ok" jsonschema:"true if the delete succeeded"`
}

func deleteThing(_ context.Context, _ *mcp.CallToolRequest, _ deleteArgs) (*mcp.CallToolResult, *deleteResult, error) {
	return nil, &deleteResult{OK: true}, nil
}

type failsArgs struct {
	Text string `json:"text" jsonschema:"ignored — the tool always errors"`
}

func fails(_ context.Context, _ *mcp.CallToolRequest, args failsArgs) (*mcp.CallToolResult, any, error) {
	return nil, nil, fmt.Errorf("fixture failure: %s", args.Text)
}

type sleepArgs struct {
	Seconds int `json:"seconds" jsonschema:"how long to sleep before returning"`
}

func sleepTool(ctx context.Context, _ *mcp.CallToolRequest, args sleepArgs) (*mcp.CallToolResult, any, error) {
	if args.Seconds < 0 {
		return nil, nil, errors.New("seconds must be >= 0")
	}
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	case <-time.After(time.Duration(args.Seconds) * time.Second):
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "ok"}},
		}, nil, nil
	}
}

func safeRead(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "safe"}},
	}, nil, nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "fx-destructive",
		Version: "0.1.0",
	}, nil)

	// delete_thing: NO annotations on purpose. The "no annotations →
	// potentially destructive" rule [MCLIP-14-0] depends on this absence.
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_thing",
		Description: "Deletes a thing. No annotations declared on purpose.",
	}, deleteThing)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fails",
		Description: "Always returns isError: true with a fixture-failure message.",
	}, fails)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "sleep",
		Description: "Sleeps server-side for the given number of seconds; cancellable via notifications/cancelled.",
	}, sleepTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "safe_read",
		Description: "Returns the literal string \"safe\". Positive control alongside the destructive tools.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, safeRead)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.SetOutput(os.Stderr)
		log.Printf("fx-destructive: server exited: %v", err)
		os.Exit(1)
	}
}
