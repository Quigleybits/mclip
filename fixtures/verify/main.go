// fixtures/verify: smoke-test the 9 conformance fixture servers.
//
// For each stdio server: spawn it, send `initialize` +
// `notifications/initialized` + a relevant list method (`tools/list`,
// `prompts/list`, or `resources/list`), assert the response shape and the
// expected feature names.
//
// For each HTTP server: spawn it, read the bound address from its stdout
// startup line, confirm a no-auth GET is rejected with 401, then POST a
// proper `initialize` with the Bearer token and assert the SSE-wrapped
// response carries the expected serverInfo.name.
//
// Exit 0 if all 9 pass; 1 otherwise. Per-server result lines go to stderr.
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type jsonRPC struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  any             `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data,omitempty"`
	} `json:"error,omitempty"`
}

func binName(server string) string {
	if runtime.GOOS == "windows" {
		return server + ".exe"
	}
	return server
}

type result struct {
	server string
	ok     bool
	detail string
}

func main() {
	root, err := repoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "verify: %v\n", err)
		os.Exit(1)
	}

	stdioCases := []struct {
		server     string
		listMethod string
		want       []string
	}{
		{"fx-echo", "tools/list", []string{"echo"}},
		{"fx-flag-collisions", "tools/list", []string{"make_widget", "dump", "polyglot"}},
		{"fx-destructive", "tools/list", []string{"delete_thing", "fails", "sleep", "safe_read"}},
		{"fx-resources-watch", "resources/list", []string{"changes"}},
		{"fx-pagination", "tools/list", nil}, // page-1 only; just assert non-empty
		{"fx-prompts", "prompts/list", []string{"greet"}},
		{"fx-progress", "tools/list", []string{"slow_count"}},
	}

	httpCases := []struct {
		server     string
		token      string
		wantServer string
	}{
		{"fx-http-auth", "test-token-abc123", "fx-http-auth"},
		{"fx-http-error-data", "test-token-abc123", "fx-http-error-data"},
	}

	var results []result
	for _, c := range stdioCases {
		r := verifyStdio(root, c.server, c.listMethod, c.want)
		results = append(results, r)
		printResult(r)
	}
	for _, c := range httpCases {
		r := verifyHTTP(root, c.server, c.token, c.wantServer)
		results = append(results, r)
		printResult(r)
	}

	passed := 0
	for _, r := range results {
		if r.ok {
			passed++
		}
	}
	fmt.Fprintf(os.Stderr, "\n%d/%d fixture servers verified.\n", passed, len(results))
	if passed != len(results) {
		os.Exit(1)
	}
}

func repoRoot() (string, error) {
	// verify is at <root>/fixtures/verify/, so root = ../../ from cwd at run time.
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// Walk up until we find fixtures/servers/.
	for cur := wd; cur != filepath.Dir(cur); cur = filepath.Dir(cur) {
		if _, err := os.Stat(filepath.Join(cur, "fixtures", "servers")); err == nil {
			return cur, nil
		}
	}
	return "", errors.New("could not locate repo root containing fixtures/servers/")
}

func printResult(r result) {
	mark := "PASS"
	if !r.ok {
		mark = "FAIL"
	}
	fmt.Fprintf(os.Stderr, "[%s] %-22s %s\n", mark, r.server, r.detail)
}

// verifyStdio spawns a stdio fixture server, performs the MCP handshake,
// and asserts the named list method returns the expected feature names.
func verifyStdio(root, server, listMethod string, want []string) result {
	bin := filepath.Join(root, "fixtures", "servers", server, binName(server))
	if _, err := os.Stat(bin); err != nil {
		return result{server, false, "binary missing: " + bin}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return result{server, false, "start: " + err.Error()}
	}
	defer func() {
		_ = stdin.Close()
		_ = cmd.Wait()
	}()

	go func() {
		fmt.Fprintln(stdin, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"mclip-verify","version":"0"}}}`)
		fmt.Fprintln(stdin, `{"jsonrpc":"2.0","method":"notifications/initialized"}`)
		fmt.Fprintf(stdin, `{"jsonrpc":"2.0","id":2,"method":%q}`+"\n", listMethod)
	}()

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	var initOK, listOK bool
	var listDetail string
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) && (!initOK || !listOK) {
		if !scanner.Scan() {
			break
		}
		var msg jsonRPC
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}
		if msg.ID == nil {
			continue
		}
		idFloat, _ := msg.ID.(float64)
		switch int(idFloat) {
		case 1:
			if msg.Error != nil {
				return result{server, false, "initialize error: " + msg.Error.Message}
			}
			initOK = true
		case 2:
			if msg.Error != nil {
				return result{server, false, listMethod + " error: " + msg.Error.Message}
			}
			ok, detail := checkListResult(listMethod, msg.Result, want)
			listOK = ok
			listDetail = detail
		}
	}
	if !initOK {
		return result{server, false, "no initialize response (stderr: " + truncate(stderrBuf.String(), 200) + ")"}
	}
	if !listOK {
		return result{server, false, listDetail}
	}
	return result{server, true, listDetail}
}

func checkListResult(method string, raw json.RawMessage, want []string) (bool, string) {
	var got []string
	switch method {
	case "tools/list":
		var r struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
			NextCursor string `json:"nextCursor"`
		}
		if err := json.Unmarshal(raw, &r); err != nil {
			return false, "unmarshal tools/list: " + err.Error()
		}
		for _, t := range r.Tools {
			got = append(got, t.Name)
		}
		if want == nil {
			if len(got) == 0 {
				return false, "tools/list returned 0 tools"
			}
			return true, fmt.Sprintf("page 1 of tools/list returned %d tools (nextCursor=%v)", len(got), r.NextCursor != "")
		}
	case "prompts/list":
		var r struct {
			Prompts []struct {
				Name string `json:"name"`
			} `json:"prompts"`
		}
		if err := json.Unmarshal(raw, &r); err != nil {
			return false, "unmarshal prompts/list: " + err.Error()
		}
		for _, p := range r.Prompts {
			got = append(got, p.Name)
		}
	case "resources/list":
		var r struct {
			Resources []struct {
				Name string `json:"name"`
			} `json:"resources"`
		}
		if err := json.Unmarshal(raw, &r); err != nil {
			return false, "unmarshal resources/list: " + err.Error()
		}
		for _, x := range r.Resources {
			got = append(got, x.Name)
		}
	default:
		return false, "unsupported list method " + method
	}
	missing := []string{}
	for _, w := range want {
		found := false
		for _, g := range got {
			if g == w {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, w)
		}
	}
	if len(missing) > 0 {
		return false, fmt.Sprintf("missing from %s: %v (got %v)", method, missing, got)
	}
	return true, fmt.Sprintf("%s returned %v", method, got)
}

// verifyHTTP spawns an HTTP fixture server, reads its "listening" line,
// asserts that an unauthed GET is 401, then POSTs an authed initialize and
// asserts the SSE-wrapped response carries the expected serverInfo.name.
func verifyHTTP(root, server, token, wantServer string) result {
	bin := filepath.Join(root, "fixtures", "servers", server, binName(server))
	if _, err := os.Stat(bin); err != nil {
		return result{server, false, "binary missing: " + bin}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "--addr", "127.0.0.1:0")
	stdout, _ := cmd.StdoutPipe()
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	if err := cmd.Start(); err != nil {
		return result{server, false, "start: " + err.Error()}
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	// Read first line of stdout for the bound URL.
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1<<16), 1<<16)
	urlCh := make(chan string, 1)
	go func() {
		if scanner.Scan() {
			line := scanner.Text()
			// Format: "<server> listening 127.0.0.1:PORT/PATH"
			parts := strings.Split(line, " listening ")
			if len(parts) == 2 {
				urlCh <- "http://" + parts[1]
				return
			}
		}
		urlCh <- ""
	}()
	var endpoint string
	select {
	case endpoint = <-urlCh:
	case <-time.After(3 * time.Second):
		return result{server, false, "timed out waiting for listening line"}
	}
	if endpoint == "" {
		return result{server, false, "could not parse listening line"}
	}

	// Unauthed GET → expect 401.
	getReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	getRes, err := http.DefaultClient.Do(getReq)
	if err != nil {
		return result{server, false, "unauthed GET: " + err.Error()}
	}
	_, _ = io.Copy(io.Discard, getRes.Body)
	getRes.Body.Close()
	if getRes.StatusCode != http.StatusUnauthorized {
		return result{server, false, fmt.Sprintf("unauthed GET expected 401, got %d", getRes.StatusCode)}
	}

	// Authed POST initialize → expect 200 with SSE payload.
	body := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"mclip-verify","version":"0"}}}`)
	postReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Accept", "application/json, text/event-stream")
	postReq.Header.Set("Authorization", "Bearer "+token)
	postRes, err := http.DefaultClient.Do(postReq)
	if err != nil {
		return result{server, false, "authed POST initialize: " + err.Error()}
	}
	defer postRes.Body.Close()
	if postRes.StatusCode != http.StatusOK {
		return result{server, false, fmt.Sprintf("authed POST initialize expected 200, got %d", postRes.StatusCode)}
	}
	raw, err := io.ReadAll(postRes.Body)
	if err != nil {
		return result{server, false, "read body: " + err.Error()}
	}

	// Response may be SSE (event-stream) or plain JSON; pull the first JSON object out of it.
	jsonBytes := extractFirstJSON(raw)
	if jsonBytes == nil {
		return result{server, false, "no JSON in response body: " + truncate(string(raw), 200)}
	}
	var msg jsonRPC
	if err := json.Unmarshal(jsonBytes, &msg); err != nil {
		return result{server, false, "unmarshal initialize: " + err.Error()}
	}
	if msg.Error != nil {
		return result{server, false, "initialize error: " + msg.Error.Message}
	}
	var ir struct {
		ServerInfo struct {
			Name string `json:"name"`
		} `json:"serverInfo"`
	}
	if err := json.Unmarshal(msg.Result, &ir); err != nil {
		return result{server, false, "unmarshal initialize result: " + err.Error()}
	}
	if ir.ServerInfo.Name != wantServer {
		return result{server, false, fmt.Sprintf("serverInfo.name = %q, want %q", ir.ServerInfo.Name, wantServer)}
	}
	return result{server, true, fmt.Sprintf("authed initialize OK at %s", endpoint)}
}

// extractFirstJSON returns the first JSON object found in body. Handles
// both raw JSON responses and SSE-framed `data: <json>` lines.
func extractFirstJSON(body []byte) []byte {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) > 0 && trimmed[0] == '{' {
		return trimmed
	}
	// Scan SSE lines.
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		line := scanner.Text()
		const prefix = "data:"
		if strings.HasPrefix(line, prefix) {
			payload := strings.TrimSpace(line[len(prefix):])
			if strings.HasPrefix(payload, "{") {
				return []byte(payload)
			}
		}
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
