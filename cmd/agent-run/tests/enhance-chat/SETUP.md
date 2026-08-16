# Scenario

**Feature**: enhance chat — pure session events, no PTY scrollback fallback

```
llm-mock-run-grok + grok-home + web POST grok-tty
  -> agenttty emits think/error to events.jsonl
  -> SSE tails file; React renders progress/error cards
PTY chrome stays in terminal modal only (never in events.jsonl)
```

## Preconditions

- Repository contains `cmd/agent-run` and `agent/llm/llm-mock/llm-mock-run-grok`.
- Session-scoped cache under `$TMPDIR/enhance-chat-doctest-<d.DOCTEST_SESSION_ID>/`
  shares compiled binaries across parallel leaves.
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- UI leaves require `playwright-debug` on PATH.

## Steps

1. Root `Setup` builds session binaries and default env.
2. Grouping `Setup` narrows `req.Area` and default `req.Mode`.
3. Leaf `Setup` configures binding outcome (success vs failure), starts web, POSTs session.
4. `Run` reads `events.jsonl`, collects SSE, or runs Playwright script.
5. Leaf `Assert` checks think/error emission and absence of PTY chrome in events.

## Context

- Default web auth token is `test`.
- Default runner is `grok-tty`.
- Resolve progress text: `Resolve session id...`
- Resolve error prefix: `Cannot resolve session id:`

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
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

const (
	enhanceChatGrokMockUUID      = "f6666666-6666-4666-8666-666666666666"
	enhanceChatSuccessMarker     = "ENHANCE_CHAT_SUCCESS_MARKER"
	resolveThinkText             = "Resolve session id..."
	resolveErrorPrefix           = "Cannot resolve session id:"
	bindingFailureFinishTimeout  = 100 * time.Second
	sessionFinishTimeout         = 45 * time.Second
)

var ptyChromeSubstrings = []string{
	"╭",
	"Grok Build",
	"Starting session",
	"Shift+Tab",
	"GROK_TTY_BANNER",
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.AgentRun, req.LLMMockRunGrok = ensureSessionBinaries(t, d, req.RepoRoot)
	req.GrokHome = filepath.Join(req.TempDir, "web-grok-home")
	req.GrokTTYRunnerBinary = req.LLMMockRunGrok
	req.WebToken = "test"
	req.Runner = "grok-tty"
	req.GrokSessionUUID = enhanceChatGrokMockUUID
	req.Env = append(req.Env,
		"AGENT_RUN_HOME="+req.Home,
		"PATH="+filepath.Dir(req.AgentRun)+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		return err
	}
	return nil
}

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "enhance-chat-doctest-"+d.DOCTEST_SESSION_ID)
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

func ensureSessionBinaries(t *testing.T, d *session.Doctest, repoRoot string) (agentRun, llmMock string) {
	t.Helper()
	cache := sessionCacheDir(d)
	agentRun = filepath.Join(cache, "agent-run")
	llmMock = filepath.Join(cache, "llm-mock-run-grok")
	ready := filepath.Join(cache, "binaries.ready")
	lock := filepath.Join(cache, "build.lock")
	err := withFileLock(t, lock, func() error {
		if fileExists(ready) && fileExists(agentRun) && fileExists(llmMock) {
			return nil
		}
		if err := os.MkdirAll(cache, 0755); err != nil {
			return err
		}
		builds := []struct {
			out  string
			args []string
		}{
			{agentRun, []string{"build", "-o", agentRun, "./cmd/agent-run"}},
			{llmMock, []string{"build", "-o", llmMock, "./agent/llm/llm-mock/llm-mock-run-grok"}},
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
	return agentRun, llmMock
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func llmMockGrokSuccessHook(prompt, sessionUUID, marker string) string {
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
{"sessionUpdate":"turn_completed"}
EOF
exit 0
'`, prompt, sessionUUID, sessionUUID, marker)
}

func llmMockGrokFailurePTYChromeHook() string {
	return `sh -c '
printf "GROK_TTY_BANNER\n"
printf "╭ Grok Build ╮\n"
printf "Starting session...\n"
printf "Shift+Tab\n"
printf "Grok › "
read -r line || true
printf "Response: hi\n"
exit 0
'`
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

func configureBindingSuccessEnv(t *testing.T, req *Request, prompt, marker string) {
	t.Helper()
	if prompt == "" {
		prompt = "enhance chat success bind"
	}
	if marker == "" {
		marker = enhanceChatSuccessMarker
	}
	req.BindingOutcome = "success"
	req.Prompt = prompt
	stripEnvPrefix(req, "LLM_MOCK_RUN_GROK_COMMAND=")
	stripEnvPrefix(req, "AGENT_RUN_GROK_TTY_GROK_SESSION_ID=")
	req.Env = append(req.Env,
		"LLM_MOCK_RUN_GROK_COMMAND="+llmMockGrokSuccessHook(prompt, enhanceChatGrokMockUUID, marker),
		"AGENT_RUN_GROK_TTY_GROK_SESSION_ID="+enhanceChatGrokMockUUID,
	)
}

func configureBindingFailureEnv(t *testing.T, req *Request, prompt string) {
	t.Helper()
	if prompt == "" {
		prompt = "enhance chat failure bind"
	}
	req.BindingOutcome = "failure"
	req.Prompt = prompt
	req.GrokHome = filepath.Join(req.TempDir, "empty-grok-home")
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		t.Fatalf("mkdir empty grok home: %v", err)
	}
	stripEnvPrefix(req, "LLM_MOCK_RUN_GROK_COMMAND=")
	stripEnvPrefix(req, "AGENT_RUN_GROK_TTY_GROK_SESSION_ID=")
	stripEnvPrefix(req, "GROK_HOME=")
	req.Env = append(req.Env,
		"GROK_HOME="+req.GrokHome,
		"LLM_MOCK_RUN_GROK_COMMAND="+llmMockGrokFailurePTYChromeHook(),
	)
	req.SSEMaxWait = bindingFailureFinishTimeout
}

func startAgentRunWeb(t *testing.T, req *Request) {
	t.Helper()
	args := []string{
		"web", "--no-open", "--port", "0", "--token", req.WebToken,
		"--agent-runner", "grok-tty",
		"--grok-home", req.GrokHome,
		"--grok-tty-runner-binary", req.GrokTTYRunnerBinary,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	req.webStderr = &bytes.Buffer{}
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

func postCreateSession(t *testing.T, req *Request, runner, prompt string) string {
	t.Helper()
	payload, err := json.Marshal(map[string]string{"runner": runner, "prompt": prompt})
	if err != nil {
		t.Fatalf("marshal create session: %v", err)
	}
	status, body := doHTTP(t, http.MethodPost, req.WebBaseURL+"/api/agent-run/sessions", req.WebToken, "application/json", string(payload))
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
	return id
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

func getSessionDetail(t *testing.T, req *Request, runner, sessionID string) (int, string) {
	t.Helper()
	rawURL := fmt.Sprintf("%s/api/agent-run/sessions/%s/%s", req.WebBaseURL, runner, sessionID)
	return doHTTP(t, http.MethodGet, rawURL, req.WebToken, "", "")
}

func waitForSessionStatus(t *testing.T, req *Request, runner, sessionID, wantStatus string, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, body := getSessionDetail(t, req, runner, sessionID)
		if sessionStatusFromDetail(body) == wantStatus {
			return body
		}
		time.Sleep(50 * time.Millisecond)
	}
	_, body := getSessionDetail(t, req, runner, sessionID)
	t.Fatalf("timeout waiting for session status %q, got %q: %s", wantStatus, sessionStatusFromDetail(body), body)
	return body
}

func sessionStatusTerminal(status string) bool {
	return status == "finished" || status == "error"
}

func waitForSessionComplete(t *testing.T, req *Request, runner, sessionID string, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, body := getSessionDetail(t, req, runner, sessionID)
		if sessionStatusTerminal(sessionStatusFromDetail(body)) {
			return body
		}
		time.Sleep(50 * time.Millisecond)
	}
	_, body := getSessionDetail(t, req, runner, sessionID)
	t.Fatalf("timeout waiting for session to complete, got %q: %s", sessionStatusFromDetail(body), body)
	return body
}

func readSessionEventsJSONL(t *testing.T, home, runner, sessionID string) (string, []string) {
	t.Helper()
	path := filepath.Join(home, "sessions", runner, sessionID, "events.jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read events.jsonl: %v", err)
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

func eventsHaveThinkText(events []map[string]any, text string) bool {
	for _, ev := range events {
		if ev["type"] == "think" && ev["text"] == text {
			return true
		}
	}
	return false
}

func eventsHaveErrorPrefix(events []map[string]any, prefix string) bool {
	for _, ev := range events {
		if ev["type"] != "error" {
			continue
		}
		text, _ := ev["text"].(string)
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return false
}

func eventsHaveAssistantMessage(events []map[string]any) bool {
	for _, ev := range events {
		if ev["type"] == "message" && ev["role"] == "assistant" {
			return true
		}
	}
	return false
}

func eventsLinesContainSubstring(lines []string, want string) bool {
	lower := strings.ToLower(want)
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), lower) {
			return true
		}
	}
	return false
}

func eventsLinesContainAnyChrome(lines []string) (string, bool) {
	for _, marker := range ptyChromeSubstrings {
		if eventsLinesContainSubstring(lines, marker) {
			return marker, true
		}
	}
	return "", false
}

func sseHasEventType(events []map[string]any, wantType string) bool {
	for _, ev := range events {
		if ev["type"] == wantType {
			return true
		}
	}
	return false
}

func collectSSESessionEvents(t *testing.T, req *Request, runner, sessionID string, afterOffset int64, maxWait time.Duration) []map[string]any {
	t.Helper()
	rawURL := fmt.Sprintf("%s/api/agent-run/sessions/%s/%s/events/stream?after=%d",
		req.WebBaseURL, runner, sessionID, afterOffset)
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

func runEventsProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SessionID == "" {
		return nil, fmt.Errorf("events probe requires SessionID")
	}
	finishTimeout := sessionFinishTimeout
	if req.BindingOutcome == "failure" {
		finishTimeout = bindingFailureFinishTimeout
		waitForSessionComplete(t, req, req.Runner, req.SessionID, finishTimeout)
	} else {
		waitForSessionStatus(t, req, req.Runner, req.SessionID, "finished", finishTimeout)
	}
	path, lines := readSessionEventsJSONL(t, req.Home, req.Runner, req.SessionID)
	return &Response{
		EventsFilePath:  path,
		EventsFileLines: lines,
		EventsParsed:    parseEventLines(t, lines),
	}, nil
}

func runSSEProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SessionID == "" {
		return nil, fmt.Errorf("sse probe requires SessionID")
	}
	maxWait := req.SSEMaxWait
	if maxWait <= 0 {
		maxWait = sessionFinishTimeout
		if req.BindingOutcome == "failure" {
			maxWait = bindingFailureFinishTimeout
		}
	}
	events := collectSSESessionEvents(t, req, req.Runner, req.SessionID, req.SSEAfterOffset, maxWait)
	return &Response{SSEEvents: events}, nil
}

func runPlaywrightProbe(t *testing.T, req *Request) (*Response, error) {
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

func commandErrorExit(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	if strings.Contains(err.Error(), "context deadline exceeded") {
		return 124
	}
	return 1
}

func jsQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
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

func requirePlaywrightOK(t *testing.T, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d\nstdout:\n%s\nstderr:\n%s", resp.PlaywrightExit, resp.PlaywrightStdout, resp.PlaywrightStderr)
	}
}

func startWebGrokSession(t *testing.T, req *Request) {
	t.Helper()
	startAgentRunWeb(t, req)
	req.SessionID = postCreateSession(t, req, req.Runner, req.Prompt)
}
```