# Scenario

**Feature**: shared web tty-terminal harness

```
go build agent-run -> temp AGENT_RUN_HOME
agent-run web --token test --port 0 -> HTTP / WS / browser probes
fixture session metadata + tty registry -> terminal resolver
```

## Preconditions

- Repository contains `cmd/agent-run`.
- Each leaf uses isolated `AGENT_RUN_HOME`.
- Backend leaves use HTTP and websocket clients; UI leaves use `playwright-debug`.

## Steps

1. Build `agent-run` into a temp `bin/` directory.
2. Start `agent-run web --token test --port 0 --no-open`.
3. Leaf setup writes session and registry fixtures or prepares browser scripts.
4. `Run` performs HTTP, websocket, or Playwright action.
5. Leaf `Assert` checks contract-specific behavior.

## Context

- The default auth token is `test`.
- Default tty session id is `tty-session-1`.
- Fixture helpers intentionally write session metadata in the same on-disk layout
  named by the requirement.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func Setup(t *testing.T, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.AgentRun = filepath.Join(req.TempDir, "bin", "agent-run")
	req.WebToken = "test"
	req.WebPort = 0
	req.Runner = "codex-tty"
	req.SessionID = "tty-session-1"
	req.Prompt = "terminal feature check"
	if err := os.MkdirAll(filepath.Dir(req.AgentRun), 0755); err != nil {
		return err
	}
	build := exec.Command("go", "build", "-buildvcs=false", "-o", req.AgentRun, "./cmd/agent-run")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build agent-run: %w\n%s", err, string(out))
	}
	req.Env = append(req.Env,
		"AGENT_RUN_HOME="+req.Home,
		"PATH="+filepath.Dir(req.AgentRun)+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	startAgentRunWeb(t, req)
	return nil
}

func startAgentRunWeb(t *testing.T, req *Request) {
	t.Helper()
	args := []string{"web", "--no-open", "--port", "0", "--token", req.WebToken}
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	req.webStdout = &bytes.Buffer{}
	req.webStderr = &bytes.Buffer{}
	cmd.Stdout = req.webStdout
	cmd.Stderr = req.webStderr
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("start agent-run web: %v", err)
	}
	req.WebCmd = cmd
	t.Cleanup(func() {
		cancel()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	})
	req.WebBaseURL = waitForWebBaseURL(t, req.webStderr, 20*time.Second)
	waitForHTTPStatus(t, req.WebBaseURL+"/api/agent-run/health", req.WebToken, http.StatusOK, 20*time.Second)
}

func waitForWebBaseURL(t *testing.T, stderr *bytes.Buffer, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		text := stderr.String()
		idx := strings.Index(text, "http://127.0.0.1:")
		if idx >= 0 {
			rest := text[idx:]
			end := strings.IndexAny(rest, " \n\r\t")
			if end >= 0 {
				rest = rest[:end]
			}
			return strings.TrimRight(rest, "/")
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for web listen URL; stderr:\n%s", stderr.String())
	return ""
}

func waitForHTTPStatus(t *testing.T, url, bearer string, want int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, _ := doHTTP(t, http.MethodGet, url, bearer, "", "")
		if status == want {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s status %d", url, want)
}

func doHTTP(t *testing.T, method, url, bearer, contentType, body string) (int, string) {
	t.Helper()
	client := &http.Client{Timeout: 10 * time.Second}
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	httpReq, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if bearer != "" {
		httpReq.Header.Set("Authorization", "Bearer "+bearer)
	}
	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("http %s %s: %v", method, url, err)
	}
	return resp.StatusCode, drainAndClose(resp)
}

func runHTTP(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	method := strings.ToUpper(strings.TrimSpace(req.HTTPMethod))
	if method == "" {
		method = http.MethodGet
	}
	path := req.HTTPPath
	if path == "" {
		path = "/api/agent-run/runners"
	}
	auth := req.HTTPAuth
	if auth == "" {
		auth = req.WebToken
	}
	contentType := ""
	if req.HTTPBody != "" {
		contentType = "application/json"
	}
	status, body := doHTTP(t, method, req.WebBaseURL+path, auth, contentType, req.HTTPBody)
	return &Response{HTTPStatus: status, HTTPBody: body}, nil
}

func runPlaywright(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if _, err := exec.LookPath("playwright-debug"); err != nil {
		t.Skipf("playwright-debug not on PATH: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "playwright-debug", "-e", req.PlaywrightScript)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return &Response{
		PlaywrightStdout: stdout.String(),
		PlaywrightStderr: stderr.String(),
		PlaywrightExit:   commandErrorExit(err),
		Err:              err,
	}, nil
}

func runTerminalWebSocket(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(req.WebBaseURL, "http") + req.WSPath
	header := http.Header{}
	if req.WSAuth != "" {
		header.Set("Authorization", "Bearer "+req.WSAuth)
	}
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return &Response{HTTPStatus: statusFromWSResponse(resp), WSError: err.Error()}, nil
	}
	defer conn.Close()
	if req.WSResizeJSON != "" {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(req.WSResizeJSON)); err != nil {
			return nil, err
		}
	}
	if req.WSInput != "" {
		if err := conn.WriteMessage(websocket.BinaryMessage, []byte(req.WSInput)); err != nil {
			return nil, err
		}
	}
	deadline := time.Now().Add(5 * time.Second)
	var out strings.Builder
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		_, msg, readErr := conn.ReadMessage()
		if readErr != nil {
			continue
		}
		out.Write(msg)
		got := out.String()
		if strings.Contains(got, req.RegistryTranscript) || strings.Contains(got, "echo:"+strings.TrimSpace(req.WSInput)) {
			return &Response{WSOutput: got, WSResize: req.RegistryResize}, nil
		}
	}
	return &Response{WSOutput: out.String(), WSResize: req.RegistryResize}, nil
}

func statusFromWSResponse(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}

func writeSessionFixture(t *testing.T, req *Request, runner, sessionID, status string) {
	t.Helper()
	dir := filepath.Join(req.Home, "sessions", runner, sessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir session: %v", err)
	}
	meta := map[string]any{
		"runner":     runner,
		"session_id": sessionID,
		"status":     status,
		"workspace":  req.TempDir,
		"created_at": time.Now().UnixMilli(),
	}
	writeJSONFile(t, filepath.Join(dir, "meta.json"), meta)
	events := `{"type":"message","role":"user","text":"` + req.Prompt + `","timestamp":1}` + "\n" +
		`{"type":"message","role":"assistant","text":"assistant keeps transcript","timestamp":2}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, "events.jsonl"), []byte(events), 0644); err != nil {
		t.Fatalf("write events: %v", err)
	}
}

func writeTTYRegistryFixture(t *testing.T, req *Request, runner, sessionID, listenAddr string) {
	t.Helper()
	dir := filepath.Join(req.Home, runner+"-registry")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir registry: %v", err)
	}
	pid := req.RegistryPID
	if pid == 0 {
		pid = os.Getpid()
	}
	entry := map[string]any{
		"session_id":  sessionID,
		"listen_addr": listenAddr,
		"pid":         pid,
		"created_at":  time.Now().Format(time.RFC3339Nano),
	}
	writeJSONFile(t, filepath.Join(dir, sessionID+".json"), entry)
}

func writeJSONFile(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s: %v", path, err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func startFakePtywrap(t *testing.T, req *Request) string {
	t.Helper()
	upgrader := websocket.Upgrader{}
	var resizeSeen string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/terminal", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.WriteMessage(websocket.BinaryMessage, []byte(req.RegistryTranscript))
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if mt == websocket.TextMessage && strings.Contains(string(msg), "resize") {
				resizeSeen = string(msg)
				req.RegistryResize = resizeSeen
				_ = conn.WriteMessage(websocket.BinaryMessage, []byte("resize-ok"))
				continue
			}
			if mt == websocket.BinaryMessage {
				_ = conn.WriteMessage(websocket.BinaryMessage, []byte("echo:"+string(msg)))
			}
		}
	})
	mux.HandleFunc("/api/terminal/sessions", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"` + req.SessionID + `","status":"running"}]`))
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	req.RegistryServerURL = server.URL
	req.RegistryWSPath = "/api/terminal"
	req.RegistryListenAddr = strings.TrimPrefix(server.URL, "http://")
	return req.RegistryListenAddr
}

func unusedLocalAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

func terminalStatusPath(runner, sessionID string) string {
	return "/api/agent-run/sessions/" + runner + "/" + sessionID + "/terminal"
}

func terminalWSPath(runner, sessionID string) string {
	return "/api/agent-run/sessions/" + runner + "/" + sessionID + "/terminal/ws"
}

func decodeJSONBody(t *testing.T, body string) map[string]any {
	t.Helper()
	var parsed map[string]any
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("invalid JSON body: %v\n%s", err, body)
	}
	return parsed
}

func boolField(obj map[string]any, key string) bool {
	v, _ := obj[key].(bool)
	return v
}

func stringField(obj map[string]any, key string) string {
	v, _ := obj[key].(string)
	return v
}

func requireNoPathLeak(t *testing.T, body string) {
	t.Helper()
	if strings.Contains(body, reqHomeLeakSentinel()) || strings.Contains(body, "-registry/") || strings.Contains(body, ".json") {
		t.Fatalf("terminal response leaked local registry details: %s", body)
	}
}

func reqHomeLeakSentinel() string {
	return string(filepath.Separator) + ".agent-run"
}

func assertHTTPStatus(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp.HTTPStatus != want {
		t.Fatalf("HTTP status=%d want=%d body=%s", resp.HTTPStatus, want, resp.HTTPBody)
	}
}

func requirePlaywrightOK(t *testing.T, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d\nstdout:\n%s\nstderr:\n%s", resp.PlaywrightExit, resp.PlaywrightStdout, resp.PlaywrightStderr)
	}
}

func jsQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func browserScript(req *Request, body string) string {
	return fmt.Sprintf(`
await page.setViewportSize({ width: 430, height: 860 });
await page.goto(%s, { waitUntil: 'domcontentloaded' });
await page.evaluate((token) => localStorage.setItem('agent-run-token', token), %s);
await page.reload({ waitUntil: 'domcontentloaded' });
%s
`, jsQuote(req.WebBaseURL), jsQuote(req.WebToken), body)
}

func sessionBrowserScript(req *Request, body string) string {
	path := req.WebBaseURL + "/sessions/" + req.Runner + "/" + req.SessionID
	return fmt.Sprintf(`
await page.setViewportSize({ width: 430, height: 860 });
await page.goto(%s, { waitUntil: 'domcontentloaded' });
await page.evaluate((token) => localStorage.setItem('agent-run-token', token), %s);
await page.goto(%s, { waitUntil: 'domcontentloaded' });
%s
`, jsQuote(req.WebBaseURL), jsQuote(req.WebToken), jsQuote(path), body)
}

func containsAny(text string, values ...string) bool {
	for _, v := range values {
		if strings.Contains(text, v) {
			return true
		}
	}
	return false
}

func parsePort(addr string) int {
	_, portText, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portText)
	return port
}
```
