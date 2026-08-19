# Scenario

**Feature**: web CLI subset refactor — shared attach, events, send, lifecycle harness

```
session-scoped build agent-run + fake-codex + tty-watch
isolated AGENT_RUN_HOME per leaf
fake ptywrap / stub-tty / web server / CLI / SSE / playwright probes
```

## Preconditions

- Repository contains `cmd/agent-run`, `cmd/fake-codex`; tty-watch from module `github.com/xhd2015/tty-watch`.
- Session-scoped cache under `$TMPDIR/web-cli-subset-doctest-<d.DOCTEST_SESSION_ID>/`
  shares compiled binaries across parallel leaves.
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- UI leaves require `playwright-debug` on PATH.

## Steps

1. Root `Setup` builds session binaries and sets default env.
2. Grouping `Setup` narrows `req.Area` and shared fixtures.
3. Leaf `Setup` configures runner, session, mode, and probes.
4. `Run` performs HTTP, websocket, SSE, CLI, tty-watch, or Playwright action.
5. Leaf `Assert` checks CLI-parity contracts.

## Context

- Default web auth token is `test`.
- Default tty runner for attach/send isolation is `stub-tty`.
- Default web codex-tty lifecycle runner is `codex-tty`.

```go
import (
	"runtime"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/doctest/session"
)

var stubSessionIDRe = regexp.MustCompile(`stub-tty:\s*(session-\d+)`)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	repoRoot, err := findAgentProRoot(d.DOCTEST_ROOT)
	if err != nil {
		return err
	}
	req.RepoRoot = repoRoot
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.AgentRun, req.FakeCodex, req.TTYWatch = ensureSessionBinaries(t, d, req.RepoRoot)
	req.WebToken = "test"
	req.WebPort = 0
	req.Runner = "stub-tty"
	req.Env = append(req.Env,
		"AGENT_RUN_HOME="+req.Home,
		"PATH="+filepath.Dir(req.AgentRun)+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	return nil
}

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "web-cli-subset-doctest-"+d.DOCTEST_SESSION_ID)
}

func withFileLock(t *testing.T, lockPath string, fn func() error) error {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return fn()
}

func ensureSessionBinaries(t *testing.T, d *session.Doctest, repoRoot string) (agentRun, fakeCodex, ttyWatch string) {
	t.Helper()
	cache := sessionCacheDir(d)
	agentRun = filepath.Join(cache, "agent-run")
	fakeCodex = filepath.Join(cache, "fake-codex")
	ttyWatch = filepath.Join(cache, "tty-watch")
	ready := filepath.Join(cache, "binaries.ready")
	lock := filepath.Join(cache, "build.lock")
	err := withFileLock(t, lock, func() error {
		if fileExists(ready) && fileExists(agentRun) && fileExists(fakeCodex) && fileExists(ttyWatch) {
			return nil
		}
		if err := os.MkdirAll(cache, 0755); err != nil {
			return err
		}
		builds := []struct {
			out  string
			args []string
		}{
			{agentRun, []string{"build", "-C", "cmd", "-o", agentRun, "./agent-run"}},
			{fakeCodex, []string{"build", "-C", "cmd", "-o", fakeCodex, "./fake-codex"}},
			{ttyWatch, []string{"build", "-o", ttyWatch, "github.com/xhd2015/tty-watch/cmd/tty-watch"}},
		}
		for _, b := range builds {
			cmd := exec.Command(runtime.GOROOT()+"/bin/go", b.args...)
			cmd.Dir = repoRoot
			if out, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("go %v: %w\n%s", b.args, err, string(out))
			}
		}
		return os.WriteFile(ready, []byte("ok"), 0644)
	})
	if err != nil {
		t.Fatalf("session binaries: %v", err)
	}
	return agentRun, fakeCodex, ttyWatch
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

const cliSubsetGrokMockUUID = "d4444444-4444-4444-8444-444444444444"

func buildLLMMockRunGrok(t *testing.T, req *Request) error {
	t.Helper()
	if req.LLMMockRunGrok == "" {
		req.LLMMockRunGrok = filepath.Join(req.TempDir, "bin", "llm-mock-run-grok")
	}
	if req.GrokHome == "" {
		req.GrokHome = filepath.Join(req.TempDir, "web-grok-home")
	}
	if err := os.MkdirAll(filepath.Dir(req.LLMMockRunGrok), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		return err
	}
	build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", req.LLMMockRunGrok, "./agent/llm/llm-mock/llm-mock-run-grok")
	if resolved, err := findAgentProRoot(req.RepoRoot); err == nil {
		req.RepoRoot = resolved
	}
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build llm-mock-run-grok (dir=%s): %w\n%s", req.RepoRoot, err, string(out))
	}
	req.GrokTTYRunnerBinary = req.LLMMockRunGrok
	return nil
}

func llmMockGrokCLIHook(prompt, sessionUUID, marker string, sleepSec int) string {
	if sleepSec <= 0 {
		sleepSec = 2
	}
	return fmt.Sprintf(`sh -c '
printf "GROK_TTY_BANNER\nGrok › "
read -r line || true
submitted="${line:-%s}"
wd=$(pwd)
enc=$(python3 -c '"'"'import os,sys,urllib.parse
p=os.path.abspath(sys.argv[1])
if p.startswith("/private/var/"): p="/var/"+p[len("/private/var/"):]
elif p.startswith("/private/tmp/"): p="/tmp/"+p[len("/private/tmp/"):]
print(urllib.parse.quote(p, safe=""))'"'"' "$wd")
dir="$GROK_HOME/sessions/$enc/%s"
mkdir -p "$dir"
now=$(date -u +%%Y-%%m-%%dT%%H:%%M:%%SZ)
cat > "$dir/summary.json" <<EOF
{"info":{"cwd":"$wd","sessionId":"%s","openedAt":"$now"},"created_at":"$now"}
EOF
cat > "$dir/updates.jsonl" <<EOF
{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"$submitted"}}
{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"%s"}}
EOF
sleep %d
exit 0
'`, prompt, sessionUUID, sessionUUID, marker, sleepSec)
}

func startGrokTTYWebMockEnv(t *testing.T, req *Request, prompt, marker string, sleepSec int) {
	t.Helper()
	if err := buildLLMMockRunGrok(t, req); err != nil {
		t.Fatalf("build grok mock: %v", err)
	}
	hook := llmMockGrokCLIHook(prompt, cliSubsetGrokMockUUID, marker, sleepSec)
	stripEnvPrefix(req, "LLM_MOCK_RUN_GROK_COMMAND=")
	stripEnvPrefix(req, "AGENT_RUN_GROK_TTY_GROK_SESSION_ID=")
	req.Env = append(req.Env,
		"LLM_MOCK_RUN_GROK_COMMAND="+hook,
		"AGENT_RUN_GROK_TTY_GROK_SESSION_ID="+cliSubsetGrokMockUUID,
	)
}

func stripEnvPrefix(req *Request, prefix string) {
	var kept []string
	for _, e := range req.Env {
		if strings.HasPrefix(e, prefix) {
			continue
		}
		kept = append(kept, e)
	}
	req.Env = kept
}

func startAgentRunWeb(t *testing.T, req *Request) {
	t.Helper()
	args := []string{"web", "--no-open", "--port", "0", "--token", req.WebToken}
	if req.GrokTTYRunnerBinary != "" {
		args = append(args,
			"--agent-runner", "grok-tty",
			"--grok-home", req.GrokHome,
			"--grok-tty-runner-binary", req.GrokTTYRunnerBinary,
		)
	}
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
	req.WebBaseURL = waitForWebBaseURL(t, req.webStderr, 25*time.Second)
	waitForHTTPStatus(t, req.WebBaseURL+"/api/agent-run/health", req.WebToken, http.StatusOK, 25*time.Second)
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

func waitForHTTPStatus(t *testing.T, rawURL, bearer string, want int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, _ := doHTTP(t, http.MethodGet, rawURL, bearer, "", "")
		if status == want {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s status %d", rawURL, want)
}

func doHTTP(t *testing.T, method, rawURL, bearer, contentType, body string) (int, string) {
	t.Helper()
	client := &http.Client{Timeout: 15 * time.Second}
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	httpReq, err := http.NewRequest(method, rawURL, reader)
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
		t.Fatalf("http %s %s: %v", method, rawURL, err)
	}
	return resp.StatusCode, drainAndClose(resp)
}

func drainAndClose(resp *http.Response) string {
	if resp == nil || resp.Body == nil {
		return ""
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return string(data)
}

func runHTTP(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	method := strings.ToUpper(strings.TrimSpace(req.HTTPMethod))
	if method == "" {
		method = http.MethodGet
	}
	path := req.HTTPPath
	if path == "" {
		path = "/api/agent-run/health"
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

func runCLI(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	timeout := req.ExecTimeout
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	if req.Sidecar != nil {
		go req.Sidecar()
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	args := req.CLIArgs
	if len(args) == 0 {
		args = req.CLIArgs
	}
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	if req.CLIStdin != "" {
		cmd.Stdin = strings.NewReader(req.CLIStdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	resp := &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: commandErrorExit(err),
		Err:      err,
	}
	if err != nil && ctx.Err() != nil {
		return resp, ctx.Err()
	}
	return resp, nil
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

func runSSE(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	maxWait := req.SSEMaxWait
	if maxWait <= 0 {
		maxWait = 45 * time.Second
	}
	events := collectSSESessionEvents(t, req, req.Runner, req.SessionID, req.SSEAfterOffset, maxWait)
	return &Response{SSEEvents: events}, nil
}

func collectSSESessionEvents(t *testing.T, req *Request, runner, sessionID string, afterOffset int64, maxWait time.Duration) []map[string]any {
	t.Helper()
	rawURL := fmt.Sprintf("%s/api/agent-run/sessions/%s/events/stream?after=%d",
		req.WebBaseURL, sessionID, afterOffset)
	ctx, cancel := context.WithTimeout(context.Background(), maxWait)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
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

func runTerminalWebSocket(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(req.WebBaseURL, "http") + req.WSPath
	header := http.Header{}
	auth := strings.TrimSpace(req.WSAuth)
	if auth == "" {
		auth = strings.TrimSpace(req.WebToken)
	}
	if auth != "" {
		header.Set("Authorization", "Bearer "+auth)
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
	deadline := time.Now().Add(6 * time.Second)
	var out strings.Builder
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		_, msg, readErr := conn.ReadMessage()
		if readErr != nil {
			continue
		}
		out.Write(msg)
		got := out.String()
		if strings.Contains(got, req.RegistryTranscript) ||
			strings.Contains(got, "echo:"+strings.TrimSpace(req.WSInput)) ||
			strings.Contains(got, "resize-ok") {
			return &Response{WSOutput: got, WSResize: req.RegistryResize, WSUpstreamURL: req.RegistryServerURL}, nil
		}
	}
	return &Response{WSOutput: out.String(), WSResize: req.RegistryResize, WSUpstreamURL: req.RegistryServerURL}, nil
}

func statusFromWSResponse(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}

func runTTYWatchAttach(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	home := filepath.Join(req.TempDir, ".tty-watch")
	if err := os.MkdirAll(home, 0755); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.TTYWatch, "run", "--detach", "sh", "-c", "printf 'TTY_WATCH_MARKER\\n'; cat")
	cmd.Env = append(os.Environ(), "TTY_WATCH_HOME="+home)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return &Response{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: commandErrorExit(err), Err: err}, nil
	}
	sessionLine := strings.TrimSpace(stdout.String())
	sessionID := strings.TrimSpace(strings.TrimPrefix(sessionLine, "session-id:"))
	if sessionID == "" {
		return nil, fmt.Errorf("tty-watch detach missing session-id stdout: %q", stdout.String())
	}
	attachCtx, attachCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer attachCancel()
	attachCmd := exec.CommandContext(attachCtx, req.TTYWatch, "attach", sessionID)
	attachCmd.Env = append(os.Environ(), "TTY_WATCH_HOME="+home)
	attachCmd.Stdin = strings.NewReader("ATTACH_STDIN_MARKER\n")
	var attachOut, attachErr bytes.Buffer
	attachCmd.Stdout = &attachOut
	attachCmd.Stderr = &attachErr
	attachErrRun := attachCmd.Run()
	return &Response{
		Stdout:       attachOut.String(),
		Stderr:       attachErr.String(),
		ExitCode:     commandErrorExit(attachErrRun),
		InjectedText: attachOut.String(),
	}, nil
}

func startFakePtywrap(t *testing.T, req *Request) string {
	t.Helper()
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	var mu sync.Mutex
	var inputs []string
	var resizeSeen string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/terminal", func(w http.ResponseWriter, r *http.Request) {
		mode := r.URL.Query().Get("attach_mode")
		if req.WSAttachMode != "" && mode != req.WSAttachMode {
			http.Error(w, "unexpected attach_mode="+mode, http.StatusBadRequest)
			return
		}
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
			mu.Lock()
			if mt == websocket.TextMessage && strings.Contains(string(msg), "resize") {
				resizeSeen = string(msg)
				req.RegistryResize = resizeSeen
				_ = conn.WriteMessage(websocket.BinaryMessage, []byte("resize-ok"))
				mu.Unlock()
				continue
			}
			if mt == websocket.BinaryMessage {
				inputs = append(inputs, string(msg))
				req.RegistryInputs = append([]string{}, inputs...)
				_ = conn.WriteMessage(websocket.BinaryMessage, []byte("echo:"+string(msg)))
			}
			mu.Unlock()
		}
	})
	mux.HandleFunc("/api/terminal/sessions", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"` + req.SessionID + `","status":"running"}]`))
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	req.RegistryServerURL = server.URL
	req.RegistryListenAddr = strings.TrimPrefix(server.URL, "http://")
	return req.RegistryListenAddr
}

func writeSessionFixture(t *testing.T, req *Request, runner, sessionID, status, terminalSessionID string) {
	t.Helper()
	dir := filepath.Join(req.Home, "sessions", sessionID)
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
	if terminalSessionID != "" {
		meta["terminal_session_id"] = terminalSessionID
	}
	writeJSONFile(t, filepath.Join(dir, "meta.json"), meta)
	events := `{"type":"message","role":"user","text":"` + req.Prompt + `","timestamp":1}` + "\n" +
		`{"type":"message","role":"assistant","text":"assistant keeps transcript","timestamp":2}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, "events.jsonl"), []byte(events), 0644); err != nil {
		t.Fatalf("write events: %v", err)
	}
}

func writeTTYRegistryFixture(t *testing.T, req *Request, runner, terminalSessionID, listenAddr string) {
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
		"session_id":  terminalSessionID,
		"listen_addr": listenAddr,
		"pid":         pid,
		"created_at":  time.Now().Format(time.RFC3339Nano),
	}
	writeJSONFile(t, filepath.Join(dir, terminalSessionID+".json"), entry)
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

func terminalStatusPath(runner, sessionID string) string {
	return "/api/agent-run/sessions/" + sessionID + "/terminal"
}

func terminalWSPath(runner, sessionID string) string {
	return "/api/agent-run/sessions/" + sessionID + "/terminal/ws"
}

func postCreateSession(t *testing.T, req *Request, runner, prompt string) (string, int, string) {
	t.Helper()
	rawURL := req.WebBaseURL + "/api/agent-run/sessions"
	payload, err := json.Marshal(map[string]string{"runner": runner, "prompt": prompt})
	if err != nil {
		t.Fatalf("marshal create session: %v", err)
	}
	status, body := doHTTP(t, http.MethodPost, rawURL, req.WebToken, "application/json", string(payload))
	if status != http.StatusAccepted && status != http.StatusOK {
		t.Fatalf("POST sessions: status=%d body=%q", status, body)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("parse create response: %v body=%q", err, body)
	}
	sess, _ := parsed["session"].(map[string]any)
	if sess == nil {
		t.Fatalf("missing session in response: %q", body)
	}
	id, _ := sess["session_id"].(string)
	if strings.TrimSpace(id) == "" {
		t.Fatalf("empty session_id in response: %q", body)
	}
	return id, status, body
}

func postFollowUpMessage(t *testing.T, req *Request, runner, sessionID, message string) (int, string) {
	t.Helper()
	rawURL := fmt.Sprintf("%s/api/agent-run/sessions/%s/messages", req.WebBaseURL, sessionID)
	payload, err := json.Marshal(map[string]string{"message": message})
	if err != nil {
		t.Fatalf("marshal follow-up: %v", err)
	}
	return doHTTP(t, http.MethodPost, rawURL, req.WebToken, "application/json", string(payload))
}

func getSessionDetail(t *testing.T, req *Request, runner, sessionID string) (int, string) {
	t.Helper()
	rawURL := fmt.Sprintf("%s/api/agent-run/sessions/%s", req.WebBaseURL, sessionID)
	return doHTTP(t, http.MethodGet, rawURL, req.WebToken, "", "")
}

func sessionStatusFromDetail(detailJSON string) string {
	var parsed map[string]any
	if err := json.Unmarshal([]byte(detailJSON), &parsed); err != nil {
		return ""
	}
	sess, _ := parsed["session"].(map[string]any)
	if sess == nil {
		return ""
	}
	status, _ := sess["status"].(string)
	return status
}

func terminalSessionIDFromDetail(detailJSON string) string {
	var parsed map[string]any
	if err := json.Unmarshal([]byte(detailJSON), &parsed); err != nil {
		return ""
	}
	sess, _ := parsed["session"].(map[string]any)
	if sess == nil {
		return ""
	}
	id, _ := sess["terminal_session_id"].(string)
	return id
}

func waitForSessionStatus(t *testing.T, req *Request, runner, sessionID, wantStatus string, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, body := getSessionDetail(t, req, runner, sessionID)
		if status == http.StatusNotFound {
			t.Fatalf("session detail 404 while waiting for %q: %s", wantStatus, body)
		}
		if sessionStatusFromDetail(body) == wantStatus {
			return body
		}
		time.Sleep(50 * time.Millisecond)
	}
	_, body := getSessionDetail(t, req, runner, sessionID)
	t.Fatalf("timeout waiting for session status %q, got %q: %s", wantStatus, sessionStatusFromDetail(body), body)
	return body
}

func waitForTerminalSessionID(t *testing.T, req *Request, runner, sessionID string, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, body := getSessionDetail(t, req, runner, sessionID)
		if status == http.StatusNotFound {
			t.Fatalf("session detail 404 while waiting for terminal_session_id: %s", body)
		}
		if id := terminalSessionIDFromDetail(body); id != "" {
			return id
		}
		time.Sleep(50 * time.Millisecond)
	}
	_, body := getSessionDetail(t, req, runner, sessionID)
	t.Fatalf("timeout waiting for terminal_session_id in: %s", body)
	return ""
}

func readEventsJSONL(t *testing.T, home, runner, sessionID string) (string, []string) {
	t.Helper()
	path := filepath.Join(home, "sessions", sessionID, "events.jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return path, lines
}

func parseEventLines(t *testing.T, lines []string) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range lines {
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Fatalf("invalid event JSON: %v\n%s", err, line)
		}
		out = append(out, obj)
	}
	return out
}

func eventsHavePhaseField(events []map[string]any) bool {
	for _, ev := range events {
		if _, ok := ev["phase"]; ok {
			return true
		}
	}
	return false
}

func sseHasEventType(events []map[string]any, wantType string) bool {
	for _, ev := range events {
		if ev["type"] == wantType {
			return true
		}
	}
	return false
}

func sseHasUserPrompt(events []map[string]any, prompt string) bool {
	for _, ev := range events {
		if ev["type"] == "message" && ev["role"] == "user" && ev["text"] == prompt {
			return true
		}
	}
	return false
}

func queueFilePath(home, runner, terminalSessionID string) string {
	return filepath.Join(home, "send-queue", runner, terminalSessionID+".jsonl")
}

func queueContainsText(t *testing.T, path, want string) bool {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		t.Fatalf("read queue: %v", err)
	}
	return strings.Contains(string(data), want)
}

func listRegistryIDs(t *testing.T, home, runner string) []string {
	t.Helper()
	dir := filepath.Join(home, runner+"-registry")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read registry dir: %v", err)
	}
	var ids []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".json") {
			ids = append(ids, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	return ids
}

func stubScenarioKeepAliveJSON() string {
	return `{
  "banner_delay_ms": 30,
  "banner_text": "STUB_TTY_BANNER ready",
  "screen_status": "idle",
  "screen_frames": [{"delay_ms": 0, "text": "STUB_TTY_BANNER\n› "}],
  "llm_events": [],
  "runner_session_id": "stub-session-keep",
  "exit_after_turn": false
}`
}

func writeStubScenario(t *testing.T, req *Request) string {
	t.Helper()
	content := req.StubScenarioJSON
	if content == "" {
		content = stubScenarioKeepAliveJSON()
	}
	path := req.StubScenarioPath
	if path == "" {
		path = filepath.Join(req.TempDir, "stub-scenario.json")
		req.StubScenarioPath = path
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write stub scenario: %v", err)
	}
	return path
}

func startStubTTYBackground(t *testing.T, req *Request) string {
	t.Helper()
	scenarioPath := writeStubScenario(t, req)
	args := []string{"run", "--agent-runner", "stub-tty", "--keep-tty", "web-cli-subset-probe"}
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	cmd.Env = append(cmd.Env,
		"AGENT_RUN_ENABLE_STUB_TTY=1",
		"AGENT_RUN_STUB_TTY_SCENARIO="+scenarioPath,
	)
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("start stub-tty: %v", err)
	}
	req.BackgroundCmd = cmd
	t.Cleanup(func() {
		cancel()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	})
	sessionID := waitForStubSessionLine(t, stderr, 45*time.Second)
	addr := readRegistryListenAddr(t, req.Home, req.Runner, sessionID)
	waitForPortOpen(t, addr, 10*time.Second)
	return sessionID
}

func waitForStubSessionLine(t *testing.T, r io.Reader, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	buf := make([]byte, 4096)
	acc := ""
	for time.Now().Before(deadline) {
		n, err := r.Read(buf)
		if n > 0 {
			acc += string(buf[:n])
			if m := stubSessionIDRe.FindStringSubmatch(acc); len(m) > 1 {
				return m[1]
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			return ""
		}
	}
	if m := stubSessionIDRe.FindStringSubmatch(acc); len(m) > 1 {
		return m[1]
	}
	return ""
}

func readRegistryListenAddr(t *testing.T, home, runner, sessionID string) string {
	t.Helper()
	path := filepath.Join(home, runner+"-registry", sessionID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	var entry map[string]any
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("parse registry: %v", err)
	}
	addr, _ := entry["listen_addr"].(string)
	return addr
}

func waitForPortOpen(t *testing.T, addr string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s", addr)
}

func seedAgentSessionForStubTTY(t *testing.T, req *Request, agentSessionID, terminalSessionID string) {
	t.Helper()
	dir := filepath.Join(req.Home, "sessions", agentSessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir session: %v", err)
	}
	meta := map[string]any{
		"runner":              req.Runner,
		"session_id":          agentSessionID,
		"status":              "running",
		"terminal_session_id": terminalSessionID,
		"workspace":           req.TempDir,
		"created_at":          time.Now().UnixMilli(),
	}
	writeJSONFile(t, filepath.Join(dir, "meta.json"), meta)
	events := `{"type":"message","role":"user","text":"seed prompt","timestamp":1}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, "events.jsonl"), []byte(events), 0644); err != nil {
		t.Fatalf("write events: %v", err)
	}
}

func fakeCodexTTYCommand() string {
	return `sh -c 'printf "CODEX_TTY_BANNER\nCodex › "; read line; echo "Response: $line"; sleep 2'`
}

func startCodexTTYWebEnv(t *testing.T, req *Request) {
	t.Helper()
	req.CodexTTYCommand = fakeCodexTTYCommand()
	req.Env = append(req.Env, "AGENT_RUN_CODEX_TTY_COMMAND="+req.CodexTTYCommand)
}

func createWebCodexTTYSession(t *testing.T, req *Request, prompt string) {
	t.Helper()
	startAgentRunWeb(t, req)
	startCodexTTYWebEnv(t, req)
	if prompt == "" {
		prompt = "web codex tty lifecycle"
	}
	req.Prompt = prompt
	req.Runner = "codex-tty"
	chatID, _, _ := postCreateSession(t, req, req.Runner, prompt)
	req.ChatSessionID = chatID
	req.SessionID = chatID
}

func createWebGrokTTYSession(t *testing.T, req *Request, prompt string) {
	t.Helper()
	if prompt == "" {
		prompt = "web grok tty lifecycle"
	}
	marker := "WEB_CLI_GROK_MARKER:" + prompt
	startGrokTTYWebMockEnv(t, req, prompt, marker, 2)
	startAgentRunWeb(t, req)
	req.Prompt = prompt
	req.Runner = "grok-tty"
	chatID, _, _ := postCreateSession(t, req, req.Runner, prompt)
	req.ChatSessionID = chatID
	req.SessionID = chatID
}

func appendEventWhileSSE(t *testing.T, req *Request, runner, sessionID, text string) {
	t.Helper()
	store, err := agentstorage.NewFileStore(req.Home)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := store.AppendEvent(sessionID, types.AgentEvent{
		Type: types.ActionMessage,
		Role: "assistant",
		Text: text,
	}); err != nil {
		t.Fatalf("append event: %v", err)
	}
}

func openAgentStore(t *testing.T, req *Request) agentstorage.Store {
	t.Helper()
	store, err := agentstorage.NewFileStore(req.Home)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return store
}

func seedRunningSessionForPrint(t *testing.T, req *Request, runner, sessionID string) {
	t.Helper()
	store := openAgentStore(t, req)
	meta := agentstorage.SessionMeta{
		Runner:    runner,
		SessionID: sessionID,
		Status:    "running",
		Workspace: req.TempDir,
	}
	if err := store.CreateSession(sessionID, meta); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if err := store.AppendEvent(sessionID, types.AgentEvent{
		Type: types.ActionMessage,
		Role: "assistant",
		Text: "First running event",
	}); err != nil {
		t.Fatalf("append event: %v", err)
	}
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

func containsAny(text string, values ...string) bool {
	for _, v := range values {
		if strings.Contains(text, v) {
			return true
		}
	}
	return false
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
	path := req.WebBaseURL + "/sessions/" + req.SessionID
	return fmt.Sprintf(`
await page.setViewportSize({ width: 430, height: 860 });
await page.goto(%s, { waitUntil: 'domcontentloaded' });
await page.evaluate((token) => localStorage.setItem('agent-run-token', token), %s);
await page.goto(%s, { waitUntil: 'domcontentloaded' });
%s
`, jsQuote(req.WebBaseURL), jsQuote(req.WebToken), jsQuote(path), body)
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

func upstreamAttachModeFromRegistryURL(serverURL string) string {
	u, err := url.Parse(serverURL)
	if err != nil {
		return ""
	}
	return u.Query().Get("attach_mode")
}

func findAgentProRoot(start string) (string, error) {
	if start == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		start = wd
	}
	for dir := start; ; dir = filepath.Dir(dir) {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.TrimSpace(line) == "module github.com/xhd2015/agent-pro" {
					return dir, nil
				}
			}
		}
		if filepath.Dir(dir) == dir {
			return "", fmt.Errorf("could not find agent-pro module root above %s", start)
		}
	}
}

```