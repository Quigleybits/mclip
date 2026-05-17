// fx-prompts: stdio MCP fixture server exposing the Prompts surface.
//
// One prompt — `greet({name, lang?})` — returns a single user-role message.
// Argument values are strings at the MCP protocol layer (`map[string]string`),
// which is what `[MCLIP-8-03]` requires; no client-side type conversion of
// prompt arguments is permissible.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func greet(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	args := req.Params.Arguments
	name, ok := args["name"]
	if !ok || name == "" {
		return nil, fmt.Errorf("prompt argument \"name\" is required")
	}
	lang := args["lang"]
	if lang == "" {
		lang = "en"
	}

	var hello string
	switch lang {
	case "en":
		hello = "Hello"
	case "fr":
		hello = "Bonjour"
	case "es":
		hello = "Hola"
	default:
		hello = "Hello"
	}

	return &mcp.GetPromptResult{
		Description: "A simple language-aware greeting.",
		Messages: []*mcp.PromptMessage{{
			Role:    "user",
			Content: &mcp.TextContent{Text: fmt.Sprintf("%s, %s.", hello, name)},
		}},
	}, nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "fx-prompts",
		Version: "0.1.0",
	}, nil)

	server.AddPrompt(&mcp.Prompt{
		Name:        "greet",
		Description: "Greet someone, optionally in a specific language.",
		Arguments: []*mcp.PromptArgument{
			{Name: "name", Description: "The person to greet.", Required: true},
			{Name: "lang", Description: "Language code (en|fr|es). Defaults to en."},
		},
	}, greet)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.SetOutput(os.Stderr)
		log.Printf("fx-prompts: server exited: %v", err)
		os.Exit(1)
	}
}
