// fx-http-error-data: Streamable-HTTP MCP fixture server that protects its
// MCP endpoint with a static Bearer token (same shape as fx-http-auth) and
// exposes one tool, `verbose_fail`, which returns a JSON-RPC error whose
// `data` field echoes the inbound request headers.
//
// The conformance suite uses this to verify the client filters credentials
// (Authorization, resolved token literal) from `error.data` before emitting
// it to the user [MCLIP-14-13 / FX-AUTH-07] and from audit-event payloads
// [FX-AUDIT-01].
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	addr  = flag.String("addr", "127.0.0.1:0", "TCP address to listen on (use :0 for an ephemeral port)")
	token = flag.String("token", "test-token-abc123", "the one Bearer token the server will accept")
	path  = flag.String("path", "/mcp", "URL path the MCP handler is mounted on")
)

type verboseFailArgs struct {
	Text string `json:"text" jsonschema:"input text — ignored; the tool always errors"`
}

// verboseFailTool always fails. The JSON-RPC error includes a `data` field
// populated with the inbound request headers (including any Authorization
// header). The conformance suite asserts the client's `error.data` redaction
// path strips credentials before display / audit-log emission.
func verboseFailTool(_ context.Context, req *mcp.CallToolRequest, args verboseFailArgs) (*mcp.CallToolResult, any, error) {
	headers := map[string][]string{}
	if req.Extra != nil {
		for k, v := range req.Extra.Header {
			headers[k] = v
		}
	}
	payload := map[string]any{
		"received_headers": headers,
		"received_text":    args.Text,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal error data: %w", err)
	}
	return nil, nil, &jsonrpc.Error{
		Code:    -32000,
		Message: "verbose fail",
		Data:    data,
	}
}

func newServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "fx-http-error-data",
		Version: "0.1.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "verbose_fail",
		Description: "Always returns a JSON-RPC error whose data field echoes the inbound request headers.",
	}, verboseFailTool)

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
		fmt.Fprintf(os.Stderr, "fx-http-error-data: listen %s: %v\n", *addr, err)
		os.Exit(1)
	}
	fmt.Printf("fx-http-error-data listening %s%s\n", listener.Addr().String(), *path)

	httpServer := &http.Server{Handler: mux}
	if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.SetOutput(os.Stderr)
		log.Printf("fx-http-error-data: server exited: %v", err)
		os.Exit(1)
	}
}
