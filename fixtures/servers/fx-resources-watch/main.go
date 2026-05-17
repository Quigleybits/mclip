// fx-resources-watch: stdio MCP fixture server exposing one subscribable
// resource (`test://changes`) and emitting deterministic
// `notifications/resources/updated` events after a client subscribes.
//
// Flags:
//
//	--event-interval=DURATION  (default 200ms) — cadence between events.
//	--event-count=N            (default 5)
//	    N >= 0: emit N events, then close the connection (FX-SIGINT-03 → exit 0).
//	    N == -1: emit forever until the client disconnects (FX-SIGINT-02).
//
// The payload of each event is `{"uri":"test://changes"}` with the sequence
// number carried in the standard `_meta.seq` field (MCP `_meta` is the
// protocol-reserved channel for ancillary metadata; the
// `ResourceUpdatedNotificationParams` schema does not define a top-level
// `seq` field, so `_meta` is the spec-compliant place to put it).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const watchedURI = "test://changes"

var (
	eventInterval = flag.Duration("event-interval", 200*time.Millisecond, "cadence between notifications/resources/updated events")
	eventCount    = flag.Int("event-count", 5, "number of events to emit; -1 = forever")
)

func readChanges(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "text/plain",
			Text:     "fx-resources-watch test resource",
		}},
	}, nil
}

func runTicker(ctx context.Context, cancel context.CancelFunc, server *mcp.Server, interval time.Duration, count int) {
	t := time.NewTicker(interval)
	defer t.Stop()
	var seq int
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			seq++
			_ = server.ResourceUpdated(ctx, &mcp.ResourceUpdatedNotificationParams{
				URI:  watchedURI,
				Meta: mcp.Meta{"seq": seq},
			})
			if count >= 0 && seq >= count {
				// Give the last notification a brief flush window before tearing
				// down the transport so the client observes all N events.
				time.Sleep(50 * time.Millisecond)
				cancel()
				return
			}
		}
	}
}

func main() {
	flag.Parse()

	mainCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var server *mcp.Server
	var startOnce sync.Once

	opts := &mcp.ServerOptions{
		Capabilities: &mcp.ServerCapabilities{
			Resources: &mcp.ResourceCapabilities{Subscribe: true},
		},
		SubscribeHandler: func(_ context.Context, req *mcp.SubscribeRequest) error {
			if req.Params.URI != watchedURI {
				return fmt.Errorf("unknown subscription URI: %s", req.Params.URI)
			}
			startOnce.Do(func() {
				go runTicker(mainCtx, cancel, server, *eventInterval, *eventCount)
			})
			return nil
		},
		UnsubscribeHandler: func(_ context.Context, _ *mcp.UnsubscribeRequest) error {
			return nil
		},
	}

	server = mcp.NewServer(&mcp.Implementation{
		Name:    "fx-resources-watch",
		Version: "0.1.0",
	}, opts)

	server.AddResource(&mcp.Resource{
		Name:        "changes",
		Description: "Subscribable test resource. Emits deterministic notifications/resources/updated events after a client subscribes.",
		MIMEType:    "text/plain",
		URI:         watchedURI,
	}, readChanges)

	if err := server.Run(mainCtx, &mcp.StdioTransport{}); err != nil {
		log.SetOutput(os.Stderr)
		log.Printf("fx-resources-watch: server exited: %v", err)
		os.Exit(1)
	}
}
