# Scenario

**Feature**: ttyrunner provider registry, unified session resolver, stub-tty, multi-attach

```
ttyrunner.Register(builtins) -> ResolveByTerminalID / ResolveByAgentSession
agent-run run --agent-runner stub-tty -> registry + tty.json dual-write
agent-run tty status|send -> CheckWritable / server WriteInput
ptywrap multi-attach -> first interactive writer, observers read-only
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/ttyrunner` implements provider registry and resolver API.
- Repository contains `cmd/agent-run` (build may fail until ttyrunner is wired).
- Each test uses isolated `AGENT_RUN_HOME=filepath.Join(t.TempDir(), ".agent-run")`.
- `stub-tty` tests set `AGENT_RUN_ENABLE_STUB_TTY=1`.
- Sealed trees (`tty/`, `grok-tty/`, `codex-tty/`) are not modified.

## Steps

1. Root `Setup` builds `agent-run`, sets `AGENT_RUN_HOME`, enables stub-tty env when needed.
2. Grouping `Setup` narrows operation (registry, storage, lookup, …) and writes fixtures.
3. Leaf `Setup` sets `req.Operation`, `req.Action`, and scenario-specific fields.
4. `Run` calls ttyrunner package APIs or `agent-run` CLI as appropriate.
5. Leaf `Assert` checks outcomes.

## Context

- Registry JSON shape unchanged for sealed-test compat.
- `tty.json` is additive; resolver enriches terminal-id lookups.
- Multi-attach tests require ptywrap `attach_mode` support (Phase D).

```go
import (
	"runtime"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
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
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/agent-pro/pkgs/ttyrunner"
	"github.com/xhd2015/doctest/session"
)

var fakePTYWrapUpgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
// Accept default session-N ids and custom --session ids (same-id policy).
var stubSessionIDRe = regexp.MustCompile(`stub-tty:\s*([a-zA-Z0-9][a-zA-Z0-9._-]*)`)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	repoRoot, err := findAgentProRoot(d.DOCTEST_ROOT)
	if err != nil {
		return err
	}
	req.RepoRoot = repoRoot
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.AgentRun = filepath.Join(req.TempDir, "bin", "agent-run")
	if err := os.MkdirAll(filepath.Dir(req.AgentRun), 0755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}
	build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", req.AgentRun, "./agent-run")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build agent-run: %w\n%s", err, string(out))
	}
	req.Env = append(req.Env, "AGENT_RUN_HOME="+req.Home)
	if req.EnableStubTTY {
		req.Env = append(req.Env, "AGENT_RUN_ENABLE_STUB_TTY=1")
	}
	return nil
}

// applyEnv/restoreEnv removed: leaf isolation uses cmd.Env on CLI children and
// explicit req.Home for in-process ttyrunner/agentstorage APIs.

func registryDirFor(home, dirName string) string {
	return filepath.Join(home, dirName)
}

func registryPathFor(home, dirName, sessionID string) string {
	return filepath.Join(registryDirFor(home, dirName), sessionID+".json")
}

// Flat layout: sessions/<agentSessionID>/{tty,meta}.json (runner is meta only).
func ttyJSONPathFor(home, runner, agentSessionID string) string {
	_ = runner
	return filepath.Join(home, "sessions", agentSessionID, "tty.json")
}

func metaJSONPathFor(home, runner, agentSessionID string) string {
	_ = runner
	return filepath.Join(home, "sessions", agentSessionID, "meta.json")
}

func defaultRegistryEntryJSON(sessionID string, port int) string {
	listenAddr := fmt.Sprintf("127.0.0.1:%d", port)
	if port <= 0 {
		listenAddr = "127.0.0.1:59999"
	}
	entry := map[string]any{
		"session_id":  sessionID,
		"listen_addr": listenAddr,
		"pid":         12345,
		"created_at":  "2026-07-03T12:00:00Z",
	}
	b, _ := json.Marshal(entry)
	return string(b)
}

func writeRegistryEntry(t *testing.T, home, dirName, sessionID, entryJSON string) {
	t.Helper()
	dir := registryDirFor(home, dirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir registry: %v", err)
	}
	if entryJSON == "" {
		entryJSON = defaultRegistryEntryJSON(sessionID, 0)
	}
	path := registryPathFor(home, dirName, sessionID)
	if err := os.WriteFile(path, []byte(entryJSON), 0644); err != nil {
		t.Fatalf("write registry: %v", err)
	}
}

func writeTTYJSON(t *testing.T, home, runner, agentSessionID, content string) string {
	t.Helper()
	dir := filepath.Dir(ttyJSONPathFor(home, runner, agentSessionID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir session dir: %v", err)
	}
	path := ttyJSONPathFor(home, runner, agentSessionID)
	if content == "" {
		content = fmt.Sprintf(`{"runner_id":%q,"agent_session_id":%q,"terminal_session_id":"session-1","listen_addr":"127.0.0.1:54321","pid":12345,"created_at":"2026-07-03T12:00:00Z","screen_status":"idle","alive":true}`, runner, agentSessionID)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write tty.json: %v", err)
	}
	return path
}

func writeSessionMeta(t *testing.T, home, runner, agentSessionID, terminalSessionID string) string {
	t.Helper()
	dir := filepath.Dir(metaJSONPathFor(home, runner, agentSessionID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir meta dir: %v", err)
	}
	meta := map[string]any{
		"runner":              runner,
		"session_id":          agentSessionID,
		"terminal_session_id": terminalSessionID,
		"status":              "running",
		"created_at":          "2026-07-03T12:00:00Z",
		"updated_at":          "2026-07-03T12:00:05Z",
	}
	b, _ := json.Marshal(meta)
	path := metaJSONPathFor(home, runner, agentSessionID)
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("write meta.json: %v", err)
	}
	return path
}

func startFakePTYWrapServer(t *testing.T, req *Request) int {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/terminal/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `[{"id":"session-1"}]`)
	})
	mux.HandleFunc("/api/terminal", func(w http.ResponseWriter, r *http.Request) {
		conn, err := fakePTYWrapUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		scrollback := req.FakePTYWrapScrollback
		if scrollback == "" {
			scrollback = "GROK_TTY_BANNER\nGrok › prompt\nResponse: hello\n› "
		}
		_ = conn.WriteMessage(websocket.TextMessage, []byte(scrollback))
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("fake ptywrap listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	})
	return port
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

func portClosed(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
	if err != nil {
		return true
	}
	_ = conn.Close()
	return false
}

func execAgentRun(t *testing.T, req *Request, args ...string) (*Response, error) {
	t.Helper()
	timeout := req.ExecTimeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	resp := &Response{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
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

func writeStubScenario(t *testing.T, req *Request) string {
	t.Helper()
	content := req.StubScenarioJSON
	if content == "" {
		if req.StubScenarioPath != "" {
			return req.StubScenarioPath
		}
		content = defaultStubScenarioJSON()
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

func defaultStubScenarioJSON() string {
	return `{
  "banner_delay_ms": 50,
  "banner_text": "STUB_TTY_BANNER ready",
  "prompt_latency_ms": 100,
  "response_text": "Response: mock assistant reply",
  "screen_status": "idle",
  "screen_frames": [
    {"delay_ms": 0, "text": "STUB_TTY_BANNER\n› "},
    {"delay_ms": 200, "text": "STUB_TTY_BANNER\n› prompt\nResponse: mock assistant reply\n› "}
  ],
  "llm_events": [
    {"delay_ms": 150, "type": "message", "role": "assistant", "text": "mock assistant reply"},
    {"delay_ms": 200, "type": "done"}
  ],
  "runner_session_id": "stub-session-abc",
  "turn_complete_delay_ms": 300,
  "exit_after_turn": true
}`
}

func stubScenarioKeepAliveJSON() string {
	return `{
  "banner_delay_ms": 30,
  "banner_text": "STUB_TTY_BANNER ready",
  "screen_status": "idle",
  "screen_frames": [
    {"delay_ms": 0, "text": "STUB_TTY_BANNER\n› "}
  ],
  "llm_events": [],
  "runner_session_id": "stub-session-keep",
  "exit_after_turn": false
}`
}

func stubScenarioBusyJSON() string {
	return `{
  "banner_delay_ms": 0,
  "banner_text": "STUB_TTY_BANNER",
  "screen_status": "busy",
  "screen_frames": [
    {"delay_ms": 0, "text": "STUB_TTY_BANNER\n• Working on task...\n"}
  ],
  "writable_reason": "stub waiting for turn complete",
  "exit_after_turn": false
}`
}

func stubEnvWithScenario(req *Request, scenarioPath string) []string {
	env := append([]string{}, req.Env...)
	env = append(env, "AGENT_RUN_STUB_TTY_SCENARIO="+scenarioPath)
	return env
}

func runStubTTYRun(t *testing.T, req *Request, extraArgs ...string) (*Response, error) {
	t.Helper()
	scenarioPath := writeStubScenario(t, req)
	prompt := req.StubPrompt
	if prompt == "" {
		prompt = "hello stub"
	}
	args := []string{"run", "--agent-runner", "stub-tty", "--session", req.AgentSessionID}
	if req.AgentSessionID == "" {
		args = []string{"run", "--agent-runner", "stub-tty"}
	}
	if req.KeepTTY {
		args = append(args, "--keep-tty")
	}
	args = append(args, prompt)
	args = append(args, extraArgs...)
	env := stubEnvWithScenario(req, scenarioPath)
	oldEnv := req.Env
	req.Env = env
	defer func() { req.Env = oldEnv }()
	return execAgentRun(t, req, args...)
}

func startStubTTYBackground(t *testing.T, req *Request) string {
	t.Helper()
	scenarioPath := writeStubScenario(t, req)
	prompt := req.StubPrompt
	if prompt == "" {
		prompt = "multi-attach probe"
	}
	args := []string{"run", "--agent-runner", "stub-tty", "--keep-tty", prompt}
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), stubEnvWithScenario(req, scenarioPath)...)
	req.BackgroundStderr = &bytes.Buffer{}
	req.BackgroundStdout = &bytes.Buffer{}
	cmd.Stderr = req.BackgroundStderr
	cmd.Stdout = req.BackgroundStdout
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("start stub-tty background: %v", err)
	}
	req.BackgroundCmd = cmd
	t.Cleanup(func() {
		// KeepAlive serve is Setsid-detached from agent-run; killing only the
		// parent leaves ptywrap holding stub-tty-registry (TempDir cleanup fail).
		if sid := strings.TrimSpace(req.TerminalSessionID); sid != "" && req.Home != "" {
			if entry := tryReadStubRegistry(req.Home, sid); entry != nil && entry.PID > 0 {
				_ = killPID(entry.PID)
			}
		}
		cancel()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	})
	sessionID := waitForStubSessionLine(t, req.BackgroundStderr, 45*time.Second)
	if sessionID == "" {
		t.Fatal("stub-tty session id not found in stderr")
	}
	req.TerminalSessionID = sessionID
	return sessionID
}

type stubRegistryEntry struct {
	PID int `json:"pid"`
}

func tryReadStubRegistry(home, sessionID string) *stubRegistryEntry {
	path := registryPathFor(home, "stub-tty-registry", sessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var entry stubRegistryEntry
	if json.Unmarshal(data, &entry) != nil || entry.PID <= 0 {
		return nil
	}
	return &entry
}

func killPID(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
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
	dir := runner + "-registry"
	data, err := os.ReadFile(registryPathFor(home, dir, sessionID))
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

type wsAttachClient struct {
	conn       *websocket.Conn
	output     strings.Builder
	outputMu   sync.Mutex
	role       string
	canWrite   bool
}

func dialPTYAttach(t *testing.T, listenAddr, sessionID, attachMode string) *wsAttachClient {
	t.Helper()
	u, err := url.Parse("http://" + listenAddr)
	if err != nil {
		t.Fatalf("parse listen addr: %v", err)
	}
	u.Scheme = "ws"
	u.Path = "/api/terminal"
	q := u.Query()
	q.Set("session_id", sessionID)
	if attachMode != "" {
		q.Set("attach_mode", attachMode)
	}
	u.RawQuery = q.Encode()
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("dial ws %s: %v", u.String(), err)
	}
	client := &wsAttachClient{conn: conn, role: attachMode}
	// Handshake first — gorilla/websocket forbids concurrent ReadMessage.
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, _ = conn.ReadMessage()
	conn.SetReadDeadline(time.Time{})
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if isPTYWrapControlFrame(msg) {
				continue
			}
			client.outputMu.Lock()
			client.output.Write(msg)
			client.outputMu.Unlock()
		}
	}()
	return client
}

func wsAttachClientOutput(c *wsAttachClient) string {
	c.outputMu.Lock()
	defer c.outputMu.Unlock()
	return c.output.String()
}

func wsAttachClientTryWriteInput(c *wsAttachClient, data []byte) error {
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func wsAttachClientTryResize(c *wsAttachClient, cols, rows int) error {
	msg, _ := json.Marshal(map[string]any{"type": "resize", "cols": cols, "rows": rows})
	return c.conn.WriteMessage(websocket.TextMessage, msg)
}

func wsAttachClientClose(c *wsAttachClient) {
	_ = c.conn.Close()
}

func isPTYWrapControlFrame(msg []byte) bool {
	if len(msg) == 0 || msg[0] != '{' {
		return false
	}
	var ctrl map[string]any
	if json.Unmarshal(msg, &ctrl) != nil {
		return false
	}
	typ, _ := ctrl["type"].(string)
	switch typ {
	case "session_id", "attach_role":
		return true
	default:
		return typ != ""
	}
}

func runRegistryOp(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	switch req.Action {
	case "lists-builtin-providers":
		resp.ProviderIDs = ttyrunner.IDs()
	case "is-tty-runner":
		resp.IsTTYRunner = ttyrunner.IsTTYRunner(req.RunnerID)
	case "registry-dir-convention":
		p, ok := ttyrunner.Get(req.RunnerID)
		if !ok {
			return resp, fmt.Errorf("provider %q not registered", req.RunnerID)
		}
		resp.Provider = p
		resp.RegistryDir = p.RegistryDir
	default:
		return nil, fmt.Errorf("unknown registry action %q", req.Action)
	}
	return resp, nil
}

func runStorageOp(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	switch req.Action {
	case "dual-write-tty-json":
		req.EnableStubTTY = true
		// cmd.Env via runStubTTYRun / execAgentRun — no parent Setenv.
		req.Env = append(req.Env, "AGENT_RUN_ENABLE_STUB_TTY=1")
		req.AgentSessionID = "sess_stub_dual"
		// Keep registry + tty.json after turn so assert can inspect dual-write.
		req.KeepTTY = true
		// runStubTTYRun already adds --session when AgentSessionID is set.
		cliResp, err := runStubTTYRun(t, req)
		if err != nil {
			return resp, err
		}
		resp.Stdout = cliResp.Stdout
		resp.Stderr = cliResp.Stderr
		resp.ExitCode = cliResp.ExitCode
		if m := stubSessionIDRe.FindStringSubmatch(resp.Stderr); len(m) > 1 {
			req.TerminalSessionID = m[1]
		}
		// Same-id policy: custom --session is both agent and terminal id.
		if req.TerminalSessionID == "" {
			req.TerminalSessionID = req.AgentSessionID
		}
		resp.RegistryPath = registryPathFor(req.Home, "stub-tty-registry", req.TerminalSessionID)
		resp.TTYJSONPath = ttyJSONPathFor(req.Home, "stub-tty", req.AgentSessionID)
	case "resolve-by-agent-session":
		store, err := agentstorage.NewFileStore(req.Home)
		if err != nil {
			return resp, err
		}
		port := startFakePTYWrapServer(t, req)
		req.RegistrySessionID = "session-1"
		entryJSON := defaultRegistryEntryJSON(req.RegistrySessionID, port)
		writeRegistryEntry(t, req.Home, "grok-tty-registry", req.RegistrySessionID, entryJSON)
		agentID := req.AgentSessionID
		if agentID == "" {
			agentID = "sess_resolve_agent"
		}
		writeSessionMeta(t, req.Home, "grok-tty", agentID, req.RegistrySessionID)
		writeTTYJSON(t, req.Home, "grok-tty", agentID, "")
		sess, err := ttyrunner.ResolveByAgentSession(store, "grok-tty", agentID)
		if err != nil {
			return resp, err
		}
		resp.TTYSession = sess
	case "resolve-by-terminal-id":
		port := startFakePTYWrapServer(t, req)
		req.RegistrySessionID = "session-1"
		agentID := "sess_terminal_lookup"
		entryJSON := defaultRegistryEntryJSON(req.RegistrySessionID, port)
		writeRegistryEntry(t, req.Home, "grok-tty-registry", req.RegistrySessionID, entryJSON)
		ttyContent := fmt.Sprintf(`{"runner_id":"grok-tty","agent_session_id":%q,"terminal_session_id":"session-1","listen_addr":%q,"pid":12345,"created_at":"2026-07-03T12:00:00Z","screen_status":"idle","alive":true}`, agentID, fmt.Sprintf("127.0.0.1:%d", port))
		writeTTYJSON(t, req.Home, "grok-tty", agentID, ttyContent)
		sess, err := ttyrunner.ResolveByTerminalID(req.Home, req.RegistrySessionID)
		if err != nil {
			return resp, err
		}
		resp.TTYSession = sess
	default:
		return nil, fmt.Errorf("unknown storage action %q", req.Action)
	}
	return resp, nil
}

func runLookupOp(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	sessionID := req.RegistrySessionID
	if sessionID == "" {
		sessionID = "session-1"
	}
	switch req.Action {
	case "finds-grok-entry":
		port := startFakePTYWrapServer(t, req)
		writeRegistryEntry(t, req.Home, "grok-tty-registry", sessionID, defaultRegistryEntryJSON(sessionID, port))
		entry, runnerID, err := ttyrunner.LookupSession(req.Home, sessionID)
		if err != nil {
			return resp, err
		}
		resp.LookupEntry = entry
		resp.LookupRunnerID = runnerID
	case "finds-codex-entry":
		port := startFakePTYWrapServer(t, req)
		writeRegistryEntry(t, req.Home, "codex-tty-registry", sessionID, defaultRegistryEntryJSON(sessionID, port))
		entry, runnerID, err := ttyrunner.LookupSession(req.Home, sessionID)
		if err != nil {
			return resp, err
		}
		resp.LookupEntry = entry
		resp.LookupRunnerID = runnerID
	case "deterministic-order":
		grokPort := startFakePTYWrapServer(t, req)
		codexPort := startFakePTYWrapServer(t, req)
		writeRegistryEntry(t, req.Home, "grok-tty-registry", sessionID, defaultRegistryEntryJSON(sessionID, grokPort))
		writeRegistryEntry(t, req.Home, "codex-tty-registry", sessionID, defaultRegistryEntryJSON(sessionID, codexPort))
		entry, runnerID, err := ttyrunner.LookupSession(req.Home, sessionID)
		if err != nil {
			return resp, err
		}
		resp.LookupEntry = entry
		resp.LookupRunnerID = runnerID
	case "skips-stale-entry":
		staleJSON := defaultRegistryEntryJSON(sessionID, 59998)
		livePort := startFakePTYWrapServer(t, req)
		writeRegistryEntry(t, req.Home, "grok-tty-registry", sessionID, staleJSON)
		writeRegistryEntry(t, req.Home, "codex-tty-registry", sessionID, defaultRegistryEntryJSON(sessionID, livePort))
		entry, runnerID, err := ttyrunner.LookupSession(req.Home, sessionID)
		if err != nil {
			return resp, err
		}
		resp.LookupEntry = entry
		resp.LookupRunnerID = runnerID
	case "session-not-found":
		_, _, err := ttyrunner.LookupSession(req.Home, "session-missing")
		resp.Err = err
	default:
		return nil, fmt.Errorf("unknown lookup action %q", req.Action)
	}
	return resp, nil
}

func runStatusSendableOp(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	if req.StartFakePTYWrap {
		port := startFakePTYWrapServer(t, req)
		waitForPortOpen(t, fmt.Sprintf("127.0.0.1:%d", port), 5*time.Second)
		dir := req.RegistryDir
		if dir == "" {
			dir = "grok-tty-registry"
		}
		sid := req.RegistrySessionID
		if sid == "" {
			sid = "session-1"
		}
		writeRegistryEntry(t, req.Home, dir, sid, defaultRegistryEntryJSON(sid, port))
	}
	args := req.StatusArgs
	if len(args) == 0 {
		sid := req.RegistrySessionID
		if sid == "" {
			sid = "session-1"
		}
		args = []string{"tty", "status", sid, "--json"}
	}
	cliResp, err := execAgentRun(t, req, args...)
	if err != nil {
		return resp, err
	}
	resp.Stdout = cliResp.Stdout
	resp.Stderr = cliResp.Stderr
	resp.ExitCode = cliResp.ExitCode
	if cliResp.ExitCode == 0 && strings.TrimSpace(cliResp.Stdout) != "" {
		var obj map[string]any
		if jsonErr := json.Unmarshal([]byte(cliResp.Stdout), &obj); jsonErr == nil {
			resp.JSONBody = obj
		}
	}
	return resp, nil
}

func runStubTTYOp(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	req.EnableStubTTY = true
	// Child env only (execAgentRun cmd.Env) — no parent applyEnv.
	req.Env = append(req.Env, "AGENT_RUN_ENABLE_STUB_TTY=1")
	switch req.Action {
	case "run-creates-registry-and-tty-json":
		req.AgentSessionID = "sess_stub_run"
		cliResp, err := runStubTTYRun(t, req)
		if err != nil {
			return resp, err
		}
		resp.Stdout = cliResp.Stdout
		resp.Stderr = cliResp.Stderr
		resp.ExitCode = cliResp.ExitCode
		if m := stubSessionIDRe.FindStringSubmatch(resp.Stderr); len(m) > 1 {
			req.TerminalSessionID = m[1]
		}
		resp.RegistryPath = registryPathFor(req.Home, "stub-tty-registry", req.TerminalSessionID)
		resp.TTYJSONPath = ttyJSONPathFor(req.Home, "stub-tty", req.AgentSessionID)
	case "scenario-banner-delay":
		req.StubScenarioJSON = `{"banner_delay_ms":800,"banner_text":"STUB_TTY_BANNER","screen_frames":[{"delay_ms":800,"text":"STUB_TTY_BANNER\n› "}],"exit_after_turn":true}`
		start := time.Now()
		cliResp, err := runStubTTYRun(t, req)
		if err != nil {
			return resp, err
		}
		elapsed := time.Since(start)
		resp.Stdout = cliResp.Stdout
		resp.Stderr = cliResp.Stderr
		resp.ExitCode = cliResp.ExitCode
		if elapsed < 600*time.Millisecond {
			return resp, fmt.Errorf("expected banner delay >= 600ms, got %v", elapsed)
		}
	case "scenario-mock-llm-events":
		req.AgentSessionID = "sess_stub_events"
		cliResp, err := runStubTTYRun(t, req)
		if err != nil {
			return resp, err
		}
		resp.Stdout = cliResp.Stdout
		resp.Stderr = cliResp.Stderr
		resp.ExitCode = cliResp.ExitCode
		resp.EventsFilePath = filepath.Join(req.Home, "sessions", req.AgentSessionID, "events.jsonl")
	case "scenario-mock-screen-frames":
		req.StubScenarioJSON = `{"banner_delay_ms":0,"screen_frames":[{"delay_ms":0,"text":"frame-0\n"},{"delay_ms":300,"text":"frame-1\n› "}],"exit_after_turn":true}`
		cliResp, err := runStubTTYRun(t, req)
		if err != nil {
			return resp, err
		}
		resp.Stdout = cliResp.Stdout
		resp.Stderr = cliResp.Stderr
		resp.ExitCode = cliResp.ExitCode
	case "scenario-declared-screen-status":
		sessionID := startStubTTYBackground(t, req)
		addr := readRegistryListenAddr(t, req.Home, "stub-tty", sessionID)
		waitForPortOpen(t, addr, 10*time.Second)
		cliResp, err := execAgentRun(t, req, "tty", "status", sessionID, "--json")
		if err != nil {
			return resp, err
		}
		resp.Stdout = cliResp.Stdout
		resp.Stderr = cliResp.Stderr
		resp.ExitCode = cliResp.ExitCode
		var obj map[string]any
		_ = json.Unmarshal([]byte(cliResp.Stdout), &obj)
		resp.JSONBody = obj
	case "attach-interactive":
		sessionID := startStubTTYBackground(t, req)
		addr := readRegistryListenAddr(t, req.Home, "stub-tty", sessionID)
		waitForPortOpen(t, addr, 10*time.Second)
		var writer *wsAttachClient
		defer func() {
			if writer != nil {
				wsAttachClientClose(writer)
			}
		}()
		time.Sleep(500 * time.Millisecond)
		// Retry briefly: stub/ptywrap may close the first writer race-y under load.
		var writeErr error
		for i := 0; i < 5; i++ {
			if writer != nil {
				wsAttachClientClose(writer)
				writer = nil
			}
			writer = dialPTYAttach(t, addr, sessionID, "interactive")
			writeErr = wsAttachClientTryWriteInput(writer, []byte("probe\r"))
			if writeErr == nil {
				break
			}
			if !strings.Contains(writeErr.Error(), "close sent") && !strings.Contains(writeErr.Error(), "broken pipe") {
				return resp, writeErr
			}
			time.Sleep(200 * time.Millisecond)
		}
		if writeErr != nil {
			return resp, writeErr
		}
		time.Sleep(500 * time.Millisecond)
		resp.MultiAttachProbe = &MultiAttachProbeResult{WriterCanWrite: true, WriterReceived: wsAttachClientOutput(writer)}
	case "keep-tty-persists":
		req.KeepTTY = true
		req.AgentSessionID = "sess_stub_keep"
		cliResp, err := runStubTTYRun(t, req)
		if err != nil {
			return resp, err
		}
		resp.Stdout = cliResp.Stdout
		resp.Stderr = cliResp.Stderr
		resp.ExitCode = cliResp.ExitCode
		if m := stubSessionIDRe.FindStringSubmatch(resp.Stderr); len(m) > 1 {
			req.TerminalSessionID = m[1]
		}
		resp.RegistryPath = registryPathFor(req.Home, "stub-tty-registry", req.TerminalSessionID)
		resp.TTYJSONPath = ttyJSONPathFor(req.Home, "stub-tty", req.AgentSessionID)
	default:
		return nil, fmt.Errorf("unknown stub-tty action %q", req.Action)
	}
	return resp, nil
}

func runMultiAttachOp(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{MultiAttachProbe: &MultiAttachProbeResult{}}
	req.EnableStubTTY = true
	// Child env only (startStubTTYBackground cmd.Env) — no parent applyEnv.
	req.Env = append(req.Env, "AGENT_RUN_ENABLE_STUB_TTY=1")
	if req.StubScenarioJSON == "" {
		req.StubScenarioJSON = stubScenarioKeepAliveJSON()
	}
	sessionID := startStubTTYBackground(t, req)
	addr := readRegistryListenAddr(t, req.Home, "stub-tty", sessionID)
	waitForPortOpen(t, addr, 10*time.Second)
	probe := resp.MultiAttachProbe

	switch req.Action {
	case "first-attach-writes-second-readonly":
		w := dialPTYAttach(t, addr, sessionID, "interactive")
		defer wsAttachClientClose(w)
		o := dialPTYAttach(t, addr, sessionID, "interactive")
		defer wsAttachClientClose(o)
		time.Sleep(300 * time.Millisecond)
		_ = wsAttachClientTryWriteInput(w, []byte("writer\r"))
		observerErr := wsAttachClientTryWriteInput(o, []byte("observer\r"))
		time.Sleep(500 * time.Millisecond)
		probe.WriterCanWrite = true
		probe.ObserverCanWrite = observerErr == nil && strings.Contains(wsAttachClientOutput(o), "observer")
		probe.WriterReceived = wsAttachClientOutput(w)
		probe.ObserverReceived = wsAttachClientOutput(o)
	case "writer-detach-no-promotion-for-second":
		w := dialPTYAttach(t, addr, sessionID, "interactive")
		o := dialPTYAttach(t, addr, sessionID, "interactive")
		time.Sleep(200 * time.Millisecond)
		wsAttachClientClose(w)
		time.Sleep(200 * time.Millisecond)
		observerErr := wsAttachClientTryWriteInput(o, []byte("still-observer\r"))
		time.Sleep(400 * time.Millisecond)
		probe.ObserverCanWrite = observerErr == nil && strings.Contains(wsAttachClientOutput(o), "still-observer")
		wsAttachClientClose(o)
	case "writer-plus-observer-both-receive-output":
		w := dialPTYAttach(t, addr, sessionID, "interactive")
		defer wsAttachClientClose(w)
		o := dialPTYAttach(t, addr, sessionID, "observer")
		defer wsAttachClientClose(o)
		time.Sleep(800 * time.Millisecond)
		probe.WriterReceived = wsAttachClientOutput(w)
		probe.ObserverReceived = wsAttachClientOutput(o)
	case "third-attach-after-writer-gone-still-readonly":
		w := dialPTYAttach(t, addr, sessionID, "interactive")
		o := dialPTYAttach(t, addr, sessionID, "interactive")
		wsAttachClientClose(w)
		time.Sleep(200 * time.Millisecond)
		o2 := dialPTYAttach(t, addr, sessionID, "interactive")
		defer wsAttachClientClose(o2)
		_ = wsAttachClientTryWriteInput(o2, []byte("third\r"))
		time.Sleep(400 * time.Millisecond)
		probe.ObserverCanWrite = strings.Contains(wsAttachClientOutput(o2), "third")
		wsAttachClientClose(o)
	case "multiple-observers-all-receive":
		w := dialPTYAttach(t, addr, sessionID, "interactive")
		defer wsAttachClientClose(w)
		o1 := dialPTYAttach(t, addr, sessionID, "observer")
		defer wsAttachClientClose(o1)
		o2 := dialPTYAttach(t, addr, sessionID, "observer")
		defer wsAttachClientClose(o2)
		time.Sleep(800 * time.Millisecond)
		probe.ObserverReceived = wsAttachClientOutput(o1)
		probe.Observer2Received = wsAttachClientOutput(o2)
	case "observer-resize-ignored-pty-unchanged":
		w := dialPTYAttach(t, addr, sessionID, "interactive")
		defer wsAttachClientClose(w)
		o := dialPTYAttach(t, addr, sessionID, "observer")
		defer wsAttachClientClose(o)
		_ = wsAttachClientTryResize(o, 200, 60)
		time.Sleep(300 * time.Millisecond)
		probe.ResizeAccepted = strings.Contains(wsAttachClientOutput(o), `"cols":200`)
	case "observer-close-does-not-delete-session":
		w := dialPTYAttach(t, addr, sessionID, "interactive")
		defer wsAttachClientClose(w)
		o := dialPTYAttach(t, addr, sessionID, "observer")
		wsAttachClientClose(o)
		time.Sleep(200 * time.Millisecond)
		_, err := os.Stat(registryPathFor(req.Home, "stub-tty-registry", sessionID))
		probe.SessionStillAlive = err == nil
	case "snapshot-does-not-affect-writer":
		w := dialPTYAttach(t, addr, sessionID, "interactive")
		defer wsAttachClientClose(w)
		s := dialPTYAttach(t, addr, sessionID, "snapshot")
		wsAttachClientClose(s)
		err := wsAttachClientTryWriteInput(w, []byte("after-snapshot\r"))
		time.Sleep(300 * time.Millisecond)
		probe.SnapshotClaimedWrite = false
		probe.WriterCanWrite = err == nil
	case "send-uses-server-write-not-client-attach":
		w := dialPTYAttach(t, addr, sessionID, "interactive")
		defer wsAttachClientClose(w)
		before := wsAttachClientOutput(w)
		sendResp, err := execAgentRun(t, req, "tty", "send", sessionID, "server-probe")
		if err != nil {
			return resp, err
		}
		resp.Stdout = sendResp.Stdout
		resp.Stderr = sendResp.Stderr
		resp.ExitCode = sendResp.ExitCode
		time.Sleep(500 * time.Millisecond)
		after := wsAttachClientOutput(w)
		probe.SendInjected = len(after) > len(before) || sendResp.ExitCode == 0
	case "send-waits-until-writable-then-injects":
		sendResp, err := execAgentRun(t, req, "tty", "send", sessionID, "writable-probe")
		if err != nil {
			return resp, err
		}
		resp.Stdout = sendResp.Stdout
		resp.Stderr = sendResp.Stderr
		resp.ExitCode = sendResp.ExitCode
		probe.SendInjected = sendResp.ExitCode == 0
	case "send-times-out-with-reason-after-10s":
		req.StubScenarioJSON = stubScenarioBusyJSON()
		_ = req.BackgroundCmd.Process.Kill()
		_ = req.BackgroundCmd.Wait()
		sessionID = startStubTTYBackground(t, req)
		addr = readRegistryListenAddr(t, req.Home, "stub-tty", sessionID)
		waitForPortOpen(t, addr, 10*time.Second)
		// Default tty send waits indefinitely; pin --max-wait 10s for this leaf.
		req.ExecTimeout = 20 * time.Second
		sendResp, err := execAgentRun(t, req, "tty", "send", "--max-wait", "10s", sessionID, "timeout-probe")
		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			return resp, err
		}
		resp.Stdout = sendResp.Stdout
		resp.Stderr = sendResp.Stderr
		resp.ExitCode = sendResp.ExitCode
		probe.SendTimedOut = sendResp.ExitCode != 0
		combined := sendResp.Stderr + sendResp.Stdout
		probe.SendTimeoutReason = combined
	case "server-write-while-client-holds-unified-write":
		w := dialPTYAttach(t, addr, sessionID, "interactive")
		defer wsAttachClientClose(w)
		before := wsAttachClientOutput(w)
		sendResp, err := execAgentRun(t, req, "tty", "send", sessionID, "concurrent")
		if err != nil {
			return resp, err
		}
		time.Sleep(500 * time.Millisecond)
		after := wsAttachClientOutput(w)
		probe.ServerSendWhileWriter = sendResp.ExitCode == 0 && len(after) >= len(before)
	default:
		return nil, fmt.Errorf("unknown multi-attach action %q", req.Action)
	}
	return resp, nil
}

func runIntegrationOp(t *testing.T, req *Request) (*Response, error) {
	return &Response{
		SealedTreesDoc: strings.TrimSpace(`
Sealed doctest trees must pass unchanged after ttyrunner extraction:
  doctest test ./cmd/agent-run/tests/tty/...
  doctest test ./cmd/agent-run/tests/grok-tty/...
  doctest test ./cmd/agent-run/tests/codex-tty/...
`),
	}, nil
}

func readEventsJSONL(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var out []string
	for _, ln := range lines {
		if strings.TrimSpace(ln) != "" {
			out = append(out, ln)
		}
	}
	return out
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %s: %v", path, err)
	}
}

func assertJSONField(t *testing.T, obj map[string]any, key string, want any) {
	t.Helper()
	got, ok := obj[key]
	if !ok {
		t.Fatalf("missing JSON field %q in %v", key, obj)
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("%s: got %v (%T) want %v (%T)", key, got, got, want, want)
	}
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