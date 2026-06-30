# Scenario

**Feature**: `agent-run` CLI — headless run, web server on localhost, session storage

```
go build agent-run + fake-codex → execCmd / runAgentRun
AGENT_RUN_HOME = filepath.Join(tempDir, ".agent-run")
web tests: startWebServer (background) → httpGet → stop in defer
```

## Preconditions

- Repository contains `cmd/agent-run` and `cmd/fake-codex` (build may fail until implemented).
- Each test uses an isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- `fake-codex` is on `PATH` for runner integration tests.

## Steps

1. Root `Setup` builds binaries, sets `AGENT_RUN_HOME`, prepends `bin/` to `PATH`.
2. Grouping `Setup` sets partial `req.Args` (subcommand prefix).
3. Leaf `Setup` finalizes `req.Args` and mode-specific fields.
4. `Run` executes `req.AgentRun` with `req.Args` (or performs HTTP probe for web mode).
5. Leaf `Assert` checks exit code, output, HTTP status, or filesystem isolation.

```go
import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func execCmd(t *testing.T, command string, args []string, dir string, env []string, stdin string, timeout time.Duration) (*Response, error) {
	t.Helper()
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}
	if err == nil {
		return resp, nil
	}
	if ctx.Err() != nil {
		return resp, ctx.Err()
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		resp.ExitCode = exitErr.ExitCode()
		return resp, nil
	}
	return resp, err
}

func runAgentRun(t *testing.T, req *Request, args ...string) (*Response, error) {
	t.Helper()
	if len(args) == 0 {
		args = req.Args
	}
	if req.Sidecar != nil {
		go req.Sidecar()
	}
	return execCmd(t, req.AgentRun, args, req.TempDir, req.Env, "", req.ExecTimeout)
}

func Setup(t *testing.T, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.AgentRun = filepath.Join(req.TempDir, "bin", "agent-run")
	req.FakeCodex = filepath.Join(req.TempDir, "bin", "fake-codex")

	if err := os.MkdirAll(filepath.Dir(req.AgentRun), 0755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}

	build := exec.Command("go", "build", "-o", req.AgentRun, "./cmd/agent-run")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build agent-run: %w\n%s", err, string(out))
	}

	build2 := exec.Command("go", "build", "-o", req.FakeCodex, "./cmd/fake-codex")
	build2.Dir = req.RepoRoot
	if out, err := build2.CombinedOutput(); err != nil {
		return fmt.Errorf("build fake-codex: %w\n%s", err, string(out))
	}

	req.Env = append(req.Env,
		"AGENT_RUN_HOME="+req.Home,
		"PATH="+filepath.Dir(req.AgentRun)+":"+os.Getenv("PATH"),
	)
	return nil
}

func assertSuccess(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s", resp.ExitCode, resp.Stderr)
	}
}

func assertExitCode(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp.ExitCode != want {
		t.Fatalf("expected exit code %d, got %d, stderr:\n%s", want, resp.ExitCode, resp.Stderr)
	}
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got string, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Fatalf("unexpected %q in:\n%s", want, got)
	}
}

func assertOutput(t *testing.T, resp *Response, stream string, contains ...string) {
	t.Helper()
	var got string
	switch stream {
	case "stdout":
		got = resp.Stdout
	case "stderr":
		got = resp.Stderr
	default:
		t.Fatalf("unknown stream %q (want stdout or stderr)", stream)
	}
	for _, want := range contains {
		assertContains(t, got, want)
	}
}

func parseJSONLines(t *testing.T, text string) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Fatalf("invalid JSON line: %v\n%s", err, line)
		}
		out = append(out, obj)
	}
	return out
}

func isJSONObjectLine(line string) bool {
	line = strings.TrimSpace(line)
	return strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}")
}

func parseWebListenURL(stderr string) (string, bool) {
	re := regexp.MustCompile(`https?://127\.0\.0\.1:(\d+)`)
	m := re.FindStringSubmatch(stderr)
	if len(m) < 2 {
		return "", false
	}
	return "http://127.0.0.1:" + m[1], true
}

func waitForListenPort(stderr *bytes.Buffer, timeout time.Duration) (int, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if u, ok := parseWebListenURL(stderr.String()); ok {
			re := regexp.MustCompile(`:(\d+)`)
			m := re.FindStringSubmatch(u)
			if len(m) >= 2 {
				var port int
				fmt.Sscanf(m[1], "%d", &port)
				return port, nil
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return 0, fmt.Errorf("timeout waiting for listen port in stderr")
}

func portOpen(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func waitForTCP(t *testing.T, host string, port int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if portOpen(host, port) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s:%d", host, port)
}

func httpDo(t *testing.T, method, url, bearer, contentType, body string) (int, string) {
	t.Helper()
	client := &http.Client{Timeout: 15 * time.Second}
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	httpReq, err := http.NewRequest(method, url, bodyReader)
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
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(respBody)
}

func httpGet(t *testing.T, url string, bearer string) (int, string) {
	return httpDo(t, http.MethodGet, url, bearer, "", "")
}

func httpPostJSON(t *testing.T, url, bearer, jsonBody string) (int, string) {
	return httpDo(t, http.MethodPost, url, bearer, "application/json", jsonBody)
}

func waitForHealth(t *testing.T, baseURL, bearer string, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	url := strings.TrimRight(baseURL, "/") + "/api/agent-run/health"
	for time.Now().Before(deadline) {
		status, _ := httpGet(t, url, bearer)
		if status == 200 {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func waitForHealthStatus(t *testing.T, baseURL, bearer string, wantStatus int, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	url := strings.TrimRight(baseURL, "/") + "/api/agent-run/health"
	for time.Now().Before(deadline) {
		status, _ := httpGet(t, url, bearer)
		if status == wantStatus {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func webTokenMode(req *Request) string {
	switch req.WebTokenMode {
	case "omit", "auto", "explicit":
		return req.WebTokenMode
	default:
		return "explicit"
	}
}

func parseWebTokenFromStderr(stderr string) string {
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		const prefix = "agent-run web token: "
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

func appendWebTokenArgs(args []string, req *Request) []string {
	switch webTokenMode(req) {
	case "omit":
		return args
	case "auto":
		return append(args, "--token", "auto")
	default:
		if req.WebToken == "" {
			req.WebToken = "test"
		}
		return append(args, "--token", req.WebToken)
	}
}

func startWebServer(t *testing.T, req *Request) {
	t.Helper()
	args := appendWebTokenArgs([]string{"web", "--no-open"}, req)
	switch {
	case req.WebPort == 0:
		args = append(args, "--port", "0")
	case req.WebPort > 0:
		args = append(args, "--port", fmt.Sprintf("%d", req.WebPort))
	}
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	req.webProcessStderr = &bytes.Buffer{}
	req.webProcessStdout = &bytes.Buffer{}
	cmd.Stderr = req.webProcessStderr
	cmd.Stdout = req.webProcessStdout

	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("start web server: %v", err)
	}
	req.WebCmd = cmd

	var baseURL string
	switch {
	case req.WebPort == 0:
		if port, err := waitForListenPort(req.webProcessStderr, 10*time.Second); err == nil {
			baseURL = fmt.Sprintf("http://127.0.0.1:%d", port)
		} else if u, ok := parseWebListenURL(req.webProcessStderr.String()); ok {
			baseURL = u
		}
	case req.WebPort > 0:
		baseURL = fmt.Sprintf("http://127.0.0.1:%d", req.WebPort)
	default:
		baseURL = "http://127.0.0.1:8192"
		waitForTCP(t, "127.0.0.1", 8192, 10*time.Second)
	}
	if baseURL == "" {
		_ = cmd.Process.Kill()
		cancel()
		t.Fatalf("web server did not publish listen URL within timeout, stderr:\n%s", req.webProcessStderr.String())
	}
	req.WebBaseURL = strings.TrimRight(baseURL, "/")
	req.WebServerStderr = req.webProcessStderr.String()

	t.Cleanup(func() {
		cancel()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	})

	mode := webTokenMode(req)
	switch mode {
	case "omit":
		if !waitForHealth(t, req.WebBaseURL, "", 10*time.Second) {
			t.Fatalf("web server health check failed (open API), stderr:\n%s", req.webProcessStderr.String())
		}
	case "auto":
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			req.WebServerStderr = req.webProcessStderr.String()
			if tok := parseWebTokenFromStderr(req.WebServerStderr); tok != "" {
				req.WebToken = tok
				if waitForHealth(t, req.WebBaseURL, tok, 2*time.Second) {
					return
				}
			}
			time.Sleep(50 * time.Millisecond)
		}
		t.Fatalf("web server health check failed (auto token), stderr:\n%s", req.webProcessStderr.String())
	default:
		if !waitForHealth(t, req.WebBaseURL, req.WebToken, 10*time.Second) {
			t.Fatalf("web server health check failed, stderr:\n%s", req.webProcessStderr.String())
		}
	}
}

func webProcessStderrText(req *Request) string {
	if req.webProcessStderr != nil {
		return req.webProcessStderr.String()
	}
	return req.WebServerStderr
}

func webProcessStdoutText(req *Request) string {
	if req.webProcessStdout != nil {
		return req.webProcessStdout.String()
	}
	return ""
}

func stopWebServer(t *testing.T, req *Request) {
	t.Helper()
	if req.WebCmd == nil || req.WebCmd.Process == nil {
		return
	}
	_ = req.WebCmd.Process.Kill()
	_ = req.WebCmd.Wait()
}

func runWebHTTP(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	path := req.HTTPPath
	if path == "" {
		path = "/api/agent-run/health"
	}
	url := req.WebBaseURL + path
	method := strings.ToUpper(strings.TrimSpace(req.HTTPMethod))
	if method == "" {
		method = http.MethodGet
	}
	var status int
	var body string
	switch method {
	case http.MethodPost:
		status, body = httpPostJSON(t, url, req.HTTPAuth, req.HTTPBody)
	default:
		status, body = httpGet(t, url, req.HTTPAuth)
	}
	return &Response{HTTPStatus: status, HTTPBody: body}, nil
}

func findEventsJSONL(t *testing.T, home string) (string, []string) {
	t.Helper()
	sessionsRoot := filepath.Join(home, "sessions")
	var found string
	var lines []string
	_ = filepath.Walk(sessionsRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if info.Name() == "events.jsonl" {
			found = path
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatalf("read %s: %v", path, readErr)
			}
			for _, line := range strings.Split(string(data), "\n") {
				if strings.TrimSpace(line) != "" {
					lines = append(lines, line)
				}
			}
		}
		return nil
	})
	if found == "" {
		t.Fatal("events.jsonl not found under AGENT_RUN_HOME/sessions")
	}
	return found, lines
}

func filesOutsidePrefix(t *testing.T, root, prefix string) []string {
	t.Helper()
	prefix = filepath.Clean(prefix)
	var outside []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		clean := filepath.Clean(path)
		if clean != prefix && !strings.HasPrefix(clean, prefix+string(filepath.Separator)) {
			outside = append(outside, clean)
		}
		return nil
	})
	return outside
}

func collectSSESessionEvents(t *testing.T, req *Request, runner, sessionID string, afterOffset int64, maxWait time.Duration) []map[string]any {
	t.Helper()
	url := fmt.Sprintf("%s/api/agent-run/sessions/%s/%s/events/stream?after=%d",
		req.WebBaseURL, runner, sessionID, afterOffset)
	ctx, cancel := context.WithTimeout(context.Background(), maxWait)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if req.WebToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+req.WebToken)
	}
	httpReq.Header.Set("Accept", "text/event-stream")
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("SSE GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("SSE status=%d body=%q", resp.StatusCode, string(body))
	}

	var events []map[string]any
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" {
			continue
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(payload), &obj); err != nil {
			t.Fatalf("invalid SSE JSON: %v\n%s", err, payload)
		}
		events = append(events, obj)
	}
	if err := sc.Err(); err != nil && ctx.Err() == nil {
		t.Fatalf("read SSE: %v", err)
	}
	return events
}
```