# Scenario

**Bug**: web tty sessions need durable chat-to-PTY terminal identity

```
web chat id web_* + runner_session_id provider-resume-id + terminal_session_id session-1
  -> terminal resolver -> codex-tty-registry/session-1.json
```

## Preconditions

- Repository contains `cmd/agent-run`.
- Each test uses isolated `AGENT_RUN_HOME`.
- Browser leaf uses `playwright-debug` when selected with `--label ui-automation`.

## Steps

1. Build `agent-run` into a temporary `bin/` directory.
2. Start `agent-run web --token test --port 0 --no-open`.
3. Descendant setup writes session metadata with distinct chat, provider, and terminal ids.
4. Descendant setup writes live or stale tty registry entries.
5. `Run` probes HTTP, websocket, follow-up, or browser behavior.

## Context

- Default runner is `codex-tty` for fixture-only leaves; web-created session helpers use `grok-tty` + grok mock harness.
- Default chat id is `web_persistent_terminal`.
- Default provider id is `019f2233-004b-72a2-9a91-480507fb5398`.
- Default PTY registry id is `session-1`.

```go
import (
	"runtime"
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
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	repoRoot, err := findAgentProRoot(d.DOCTEST_ROOT)
	if err != nil {
		return err
	}
	req.RepoRoot = repoRoot
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.AgentRun = filepath.Join(req.TempDir, "bin", "agent-run")
	req.WebToken = "test"
	req.Runner = "codex-tty"
	req.ChatSessionID = "web_persistent_terminal"
	req.RunnerSessionID = "019f2233-004b-72a2-9a91-480507fb5398"
	req.TerminalSessionID = "session-1"
	req.Status = "finished"
	req.Prompt = "first tty prompt"
	req.FollowUpPrompt = "second tty prompt"
	if err := os.MkdirAll(filepath.Dir(req.AgentRun), 0755); err != nil {
		return err
	}
	build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", req.AgentRun, "./agent-run")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build agent-run: %w\n%s", err, string(out))
	}
	req.Env = append(req.Env,
		"AGENT_RUN_HOME="+req.Home,
		"PATH="+filepath.Dir(req.AgentRun)+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	if err := buildLLMMockRunGrok(t, req); err != nil {
		return err
	}
	startAgentRunWeb(t, req)
	return nil
}

const persistentGrokMockUUID = "b2222222-2222-4222-8222-222222222222"

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
	root := req.RepoRoot
	if resolved, err := findAgentProRoot(root); err == nil {
		root = resolved
	}
	build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", req.LLMMockRunGrok, "./agent/llm/llm-mock/llm-mock-run-grok")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build llm-mock-run-grok (dir=%s): %w\n%s", root, err, string(out))
	}
	sibling := filepath.Join(filepath.Dir(req.LLMMockRunGrok), "llm-mock")
	buildSibling := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", sibling, "./agent/llm/llm-mock")
	buildSibling.Dir = req.RepoRoot
	if out, err := buildSibling.CombinedOutput(); err != nil {
		return fmt.Errorf("build llm-mock sibling: %w\n%s", err, string(out))
	}
	req.GrokTTYRunnerBinary = req.LLMMockRunGrok
	return nil
}

func setGrokMockHook(t *testing.T, req *Request, hook string) {
	t.Helper()
	req.GrokMockHook = hook
	stripEnvPrefix(req, "LLM_MOCK_RUN_GROK_COMMAND=")
	stripEnvPrefix(req, "AGENT_RUN_GROK_TTY_GROK_SESSION_ID=")
	req.Env = append(req.Env,
		"LLM_MOCK_RUN_GROK_COMMAND="+hook,
		"AGENT_RUN_GROK_TTY_GROK_SESSION_ID="+persistentGrokMockUUID,
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

func llmMockGrokTTYHook(prompt, sessionUUID, assistantText string, sleepSec int) string {
	if sleepSec < 0 {
		sleepSec = 0
	}
	// Write updates before stdin so keep-tty sync can bind without waiting on read.
	return fmt.Sprintf(`sh -c '
printf "GROK_TTY_BANNER\nGrok › "
submitted=%q
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
'`, prompt, sessionUUID, sessionUUID, assistantText, sleepSec)
}

func llmMockGrokTTYTwoTurnHook(firstPrompt, sessionUUID string) string {
	// Seed turn-1 updates immediately; timed reads wait for follow-up inject.
	return fmt.Sprintf(`sh -c '
printf "GROK_TTY_BANNER\nGrok › "
first=%q
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
{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"$first"}}
{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"Response: Paris"}}
EOF
printf "\nResponse: Paris\nGrok › "
read -t 8 -r second || true
if [ -n "$second" ]; then
  printf "\nFOLLOWUP_RESPONSE: received %%s\n" "$second" >> "$dir/updates.jsonl"
fi
sleep 3
exit 0
'`, firstPrompt, sessionUUID, sessionUUID)
}

func restartWebWithGrokMock(t *testing.T, req *Request) {
	t.Helper()
	if req.WebCmd != nil && req.WebCmd.Process != nil {
		_ = req.WebCmd.Process.Kill()
		_ = req.WebCmd.Wait()
		req.WebCmd = nil
	}
	startAgentRunWeb(t, req)
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

func writeMappedSessionFixture(t *testing.T, req *Request) {
	t.Helper()
	dir := filepath.Join(req.Home, "sessions", req.ChatSessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir session: %v", err)
	}
	meta := map[string]any{
		"runner":              req.Runner,
		"session_id":          req.ChatSessionID,
		"runner_session_id":   req.RunnerSessionID,
		"terminal_session_id": req.TerminalSessionID,
		"status":              req.Status,
		"workspace":           req.TempDir,
		"created_at":          time.Now().UTC().Format(time.RFC3339),
		"updated_at":          time.Now().UTC().Format(time.RFC3339),
	}
	writeJSONFile(t, filepath.Join(dir, "meta.json"), meta)
	events := `{"type":"message","role":"user","text":"` + req.Prompt + `","timestamp":1}` + "\n" +
		`{"type":"message","role":"assistant","text":"assistant keeps transcript","timestamp":2}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, "events.jsonl"), []byte(events), 0644); err != nil {
		t.Fatalf("write events: %v", err)
	}
}

func writeTTYRegistryFixture(t *testing.T, req *Request, sessionID, listenAddr string) {
	t.Helper()
	dir := filepath.Join(req.Home, req.Runner+"-registry")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir registry: %v", err)
	}
	entry := map[string]any{
		"session_id":  sessionID,
		"listen_addr": listenAddr,
		"pid":         os.Getpid(),
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

func startMappedPtywrap(t *testing.T, req *Request) string {
	t.Helper()
	upgrader := websocket.Upgrader{}
	count := 0
	req.PTYConnectionCount = &count
	mux := http.NewServeMux()
	mux.HandleFunc("/api/terminal", func(w http.ResponseWriter, r *http.Request) {
		count++
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
			if mt == websocket.BinaryMessage {
				_ = conn.WriteMessage(websocket.BinaryMessage, []byte("echo:"+string(msg)))
			}
		}
	})
	mux.HandleFunc("/api/terminal/sessions", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"` + req.TerminalSessionID + `","status":"running"}]`))
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	req.RegistryListenAddr = strings.TrimPrefix(server.URL, "http://")
	return req.RegistryListenAddr
}

func startControlFramePtywrap(t *testing.T, req *Request) string {
	t.Helper()
	upgrader := websocket.Upgrader{}
	count := 0
	inputSeen := ""
	req.PTYConnectionCount = &count
	req.PTYInputSeen = &inputSeen
	mux := http.NewServeMux()
	mux.HandleFunc("/api/terminal", func(w http.ResponseWriter, r *http.Request) {
		count++
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		requestedID := r.URL.Query().Get("session_id")
		if requestedID == "" {
			requestedID = "created-unmapped"
		}
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"session_id","session_id":"`+requestedID+`"}`))
		if requestedID != req.TerminalSessionID {
			return
		}
		_ = conn.WriteMessage(websocket.BinaryMessage, []byte(req.RegistryTranscript))
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if mt == websocket.BinaryMessage {
				inputSeen += string(msg)
				_ = conn.WriteMessage(websocket.BinaryMessage, []byte("echo:"+string(msg)))
			}
		}
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
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
	return "/api/agent-run/sessions/" + sessionID + "/terminal"
}

func terminalWSPath(runner, sessionID string) string {
	return "/api/agent-run/sessions/" + sessionID + "/terminal/ws"
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

func assertHTTPStatus(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp.HTTPStatus != want {
		t.Fatalf("HTTP status=%d want=%d body=%s", resp.HTTPStatus, want, resp.HTTPBody)
	}
}

func requireTerminalMappingAvailable(t *testing.T, req *Request, body string) {
	t.Helper()
	obj := decodeJSONBody(t, body)
	if !boolField(obj, "available") {
		t.Fatalf("terminal should be available through terminal_session_id: %s", body)
	}
	if stringField(obj, "runner") != req.Runner {
		t.Fatalf("wrong runner: %s", body)
	}
	if stringField(obj, "session_id") != req.ChatSessionID {
		t.Fatalf("status should keep chat session id %q: %s", req.ChatSessionID, body)
	}
	if stringField(obj, "terminal_session_id") != req.TerminalSessionID {
		t.Fatalf("status should report terminal_session_id %q: %s", req.TerminalSessionID, body)
	}
}

func requireTerminalMappingUnavailable(t *testing.T, req *Request, body string) {
	t.Helper()
	obj := decodeJSONBody(t, body)
	if boolField(obj, "available") {
		t.Fatalf("stale mapped terminal should be unavailable: %s", body)
	}
	if stringField(obj, "terminal_session_id") != req.TerminalSessionID {
		t.Fatalf("unavailable response should preserve terminal mapping %q: %s", req.TerminalSessionID, body)
	}
}

func runWebSocketTwice(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	first, errText := readTerminalWSOnce(t, req)
	if errText != "" {
		return &Response{WSOutput: first, WSError: errText}, nil
	}
	_, _ = doHTTP(t, "GET", req.WebBaseURL+"/api/agent-run/sessions/"+req.ChatSessionID, req.WebToken, "", "")
	second, errText := readTerminalWSOnce(t, req)
	count := 0
	if req.PTYConnectionCount != nil {
		count = *req.PTYConnectionCount
	}
	return &Response{WSOutput: first + second, WSError: errText, PTYConnectionSeen: count}, nil
}

func readTerminalWSOnce(t *testing.T, req *Request) (string, string) {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(req.WebBaseURL, "http") + req.WSPath
	header := http.Header{"Authorization": []string{"Bearer " + req.WebToken}}
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return "", fmt.Sprintf("status=%d err=%v", statusFromWSResponse(resp), err)
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return "", err.Error()
	}
	return string(msg), ""
}

func statusFromWSResponse(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}

func runFollowUpReuseProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	statusPath := terminalStatusPath(req.Runner, req.ChatSessionID)
	firstStatus, firstBody := doHTTP(t, "GET", req.WebBaseURL+statusPath, req.WebToken, "", "")
	req.RegistryIDsBefore = registryIDs(t, req)
	payload := `{"text":` + jsQuote(req.FollowUpPrompt) + `}`
	followStatus, followBody := doHTTP(t, "POST", req.WebBaseURL+"/api/agent-run/sessions/"+req.ChatSessionID+"/messages", req.WebToken, "application/json", payload)
	time.Sleep(300 * time.Millisecond)
	sessionStatus, sessionBody := doHTTP(t, "GET", req.WebBaseURL+"/api/agent-run/sessions/"+req.ChatSessionID, req.WebToken, "", "")
	secondStatus, secondBody := doHTTP(t, "GET", req.WebBaseURL+statusPath, req.WebToken, "", "")
	return &Response{
		FirstHTTPStatus:  firstStatus,
		FirstHTTPBody:    firstBody,
		FollowUpStatus:   followStatus,
		FollowUpBody:     followBody,
		FollowUpSessionStatus: sessionStatus,
		FollowUpSessionBody:   sessionBody,
		SecondHTTPStatus: secondStatus,
		SecondHTTPBody:   secondBody,
		RegistryIDsAfter: registryIDs(t, req),
	}, nil
}

func registryIDs(t *testing.T, req *Request) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(req.Home, req.Runner+"-registry"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read registry dir: %v", err)
	}
	var ids []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		ids = append(ids, strings.TrimSuffix(entry.Name(), ".json"))
	}
	sort.Strings(ids)
	return ids
}

func waitForAnyRegistryID(t *testing.T, req *Request, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ids := registryIDs(t, req)
		if len(ids) > 0 {
			req.TerminalSessionID = ids[0]
			return ids[0]
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for tty registry entry under %s; web stderr:\n%s", filepath.Join(req.Home, req.Runner+"-registry"), req.webStderr.String())
	return ""
}

func requireSameStringSlice(t *testing.T, got, want []string) {
	t.Helper()
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("registry ids = %v, want %v", got, want)
	}
}

func jsQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func sessionBrowserScript(req *Request, body string) string {
	path := req.WebBaseURL + "/sessions/" + req.ChatSessionID
	return fmt.Sprintf(`
await page.setViewportSize({ width: 430, height: 860 });
await page.goto(%s, { waitUntil: 'domcontentloaded' });
await page.evaluate((token) => localStorage.setItem('agent-run-token', token), %s);
await page.goto(%s, { waitUntil: 'domcontentloaded' });
%s
`, jsQuote(req.WebBaseURL), jsQuote(req.WebToken), jsQuote(path), body)
}

func createRunningWebGrokTTYSessionThroughAPI(t *testing.T, req *Request) {
	t.Helper()
	prompt := "keep terminal alive before assistant response"
	setGrokMockHook(t, req, llmMockGrokTTYHook(prompt, persistentGrokMockUUID, "delayed terminal run completed", 5))
	restartWebWithGrokMock(t, req)
	body := `{"runner":"grok-tty","prompt":"keep terminal alive before assistant response"}`
	status, respBody := doHTTP(t, "POST", req.WebBaseURL+"/api/agent-run/sessions", req.WebToken, "application/json", body)
	if status != http.StatusAccepted {
		t.Fatalf("create running web grok-tty session status=%d body=%s", status, respBody)
	}
	created := decodeJSONBody(t, respBody)
	session, _ := created["session"].(map[string]any)
	sessionID, _ := session["session_id"].(string)
	if sessionID == "" {
		t.Fatalf("create response missing session_id: %s", respBody)
	}
	req.Runner = "grok-tty"
	req.ChatSessionID = sessionID
}

func createWebGrokTTYSessionThroughAPI(t *testing.T, req *Request) {
	t.Helper()
	prompt := "one word of France capital"
	setGrokMockHook(t, req, llmMockGrokTTYHook(prompt, persistentGrokMockUUID, "Response: Paris", 2))
	restartWebWithGrokMock(t, req)
	body := `{"runner":"grok-tty","prompt":"one word of France capital"}`
	status, respBody := doHTTP(t, "POST", req.WebBaseURL+"/api/agent-run/sessions", req.WebToken, "application/json", body)
	if status != http.StatusAccepted {
		t.Fatalf("create web grok-tty session status=%d body=%s", status, respBody)
	}
	created := decodeJSONBody(t, respBody)
	session, _ := created["session"].(map[string]any)
	sessionID, _ := session["session_id"].(string)
	if sessionID == "" {
		t.Fatalf("create response missing session_id: %s", respBody)
	}
	req.Runner = "grok-tty"
	req.ChatSessionID = sessionID
	waitForCreatedTTYSessionFinished(t, req)
}

func waitForCreatedTTYTerminalID(t *testing.T, req *Request, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var last string
	for time.Now().Before(deadline) {
		status, body := doHTTP(t, "GET", req.WebBaseURL+"/api/agent-run/sessions/"+req.ChatSessionID, req.WebToken, "", "")
		last = body
		if status == http.StatusOK {
			var parsed struct {
				Session struct {
					TerminalSessionID string `json:"terminal_session_id"`
				} `json:"session"`
			}
			if err := json.Unmarshal([]byte(body), &parsed); err == nil && parsed.Session.TerminalSessionID != "" {
				req.TerminalSessionID = parsed.Session.TerminalSessionID
				return
			}
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for web-created running tty session terminal_session_id; last detail=%s", last)
}

func waitForCreatedTTYSessionFinished(t *testing.T, req *Request) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	var last string
	for time.Now().Before(deadline) {
		status, body := doHTTP(t, "GET", req.WebBaseURL+"/api/agent-run/sessions/"+req.ChatSessionID, req.WebToken, "", "")
		last = body
		if status == http.StatusOK {
			var parsed struct {
				Session struct {
					Status            string `json:"status"`
					TerminalSessionID string `json:"terminal_session_id"`
				} `json:"session"`
			}
			if err := json.Unmarshal([]byte(body), &parsed); err == nil {
				if parsed.Session.TerminalSessionID != "" {
					req.TerminalSessionID = parsed.Session.TerminalSessionID
				}
				if parsed.Session.Status == "finished" && parsed.Session.TerminalSessionID != "" {
					return
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for web-created tty session to finish with terminal_session_id; last detail=%s", last)
}

func parsePort(addr string) int {
	_, portText, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portText)
	return port
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
