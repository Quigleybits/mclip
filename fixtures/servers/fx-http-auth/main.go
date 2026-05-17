// fx-http-auth: Streamable-HTTP MCP fixture server protected by a static
// Bearer token.
//
// All requests to the MCP endpoint require `Authorization: Bearer <token>`.
// The auth middleware (auth.RequireBearerToken) returns 401 for missing or
// wrong tokens. This includes the very first POST that carries `initialize`
// — i.e. the harness must present credentials from the start. (The
// fixtures-spec phrasing "after initialize" is read here as "from the moment
// the client starts using the server, which begins with initialize"; gating
// the initialize handshake itself would require body-parsing middleware that
// adds nothing the conformance suite actually exercises.)
//
// Flags:
//
//	--addr   default 127.0.0.1:0      — TCP address; :0 picks an ephemeral port
//	--token  default test-token-abc123 — the one accepted Bearer token
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	addr  = flag.String("addr", "127.0.0.1:0", "TCP address to listen on (use :0 for an ephemeral port)")
	token = flag.String("token", "test-token-abc123", "the one Bearer token the server will accept")
	path  = flag.String("path", "/mcp", "URL path the MCP handler is mounted on")
)

// whoamiTool returns the last 4 chars of the Bearer token the server saw.
// The conformance harness asserts on this so it can tell which credential
// source resolved (env var, config file, OS keychain).
func whoamiTool(_ context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	authz := ""
	if req.Extra != nil {
		authz = req.Extra.Header.Get("Authorization")
	}
	authz = strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
	suffix := authz
	if n := len(authz); n > 4 {
		suffix = authz[n-4:]
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: suffix}},
	}, nil, nil
}

type readOnlyArgs struct {
	Text string `json:"text" jsonschema:"the text to echo"`
}

func readOnlyTool(_ context.Context, _ *mcp.CallToolRequest, args readOnlyArgs) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: args.Text}},
	}, nil, nil
}

func newServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "fx-http-auth",
		Version: "0.1.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "whoami",
		Description: "Returns the last 4 chars of the Bearer token the server saw on this request.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, whoamiTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_only",
		Description: "Echoes text back. Used to exercise authed conformant invocations.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, readOnlyTool)

	return server
}

func main() {
	flag.Parse()

	server := newServer()

	verifier := func(_ context.Context, presented string, _ *http.Request) (*auth.TokenInfo, error) {
		if presented != *token {
			return nil, fmt.Errorf("%w: bad token", auth.ErrInvalidToken)
		}
		return &auth.TokenInfo{
			Scopes:     []string{"read", "write"},
			Expiration: time.Now().Add(24 * time.Hour),
		}, nil
	}

	requireBearer := auth.RequireBearerToken(verifier, &auth.RequireBearerTokenOptions{
		Scopes: []string{"read"},
	})

	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)

	mux := http.NewServeMux()
	mux.Handle(*path, requireBearer(mcpHandler))

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fx-http-auth: listen %s: %v\n", *addr, err)
		os.Exit(1)
	}
	// Print the bound address on stdout in a single deterministic line so the
	// harness can resolve the ephemeral port without scraping log output.
	fmt.Printf("fx-http-auth listening %s%s\n", listener.Addr().String(), *path)

	httpServer := &http.Server{Handler: mux}
	if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.SetOutput(os.Stderr)
		log.Printf("fx-http-auth: server exited: %v", err)
		os.Exit(1)
	}
}
