# Scenario

**Bug**: session page polls GET /terminal perpetually instead of stopping after discovery

```
seed session + optional registry -> agent-run web -> playwright network monitor -> assert terminal GET bounds
```

## Preconditions

- Repository contains `cmd/agent-run` and `frontend-agent-run/`.
- Session-scoped cache under `$TMPDIR/terminal-poll-doctest-<d.DOCTEST_SESSION_ID>/` shares
  compiled `agent-run` across parallel leaves.
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Playwright leaves require `playwright-debug` on PATH.

## Steps

1. Root `Setup` builds `agent-run`, sets default env and fixture ids.
2. Leaf `Setup` starts fake ptywrap, writes session/registry fixtures, starts web.
3. `Run` optionally delays registry write, executes playwright script with network monitor.
4. Leaf `Assert` checks playwright exit code and scenario tag.

## Context

- Default runner: `grok-tty`.
- Default web token: `test`.
- Default chat id: `web_terminal_poll_fixture`.
- Default terminal registry id: `session-1`.
- Terminal GET monitor regex: `/api/agent-run/sessions/{runner}/{id}/terminal`.

```go
import (
	"runtime"
	"bytes"
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
	"syscall"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.AgentRun = ensureAgentRunBinary(t, d, req.RepoRoot)
	req.Port = 0
	req.Token = "test"
	req.Runner = "grok-tty"
	req.ChatSessionID = "web_terminal_poll_fixture"
	req.RunnerSessionID = "019f2233-004b-72a2-9a91-480507fb5398"
	req.TerminalSessionID = "session-1"
	req.Status = "finished"
	req.Prompt = "terminal poll fixture prompt"
	req.RegistryTranscript = "terminal-poll-ready\n"
	req.WatchWindowMs = 8000
	req.Env = append(req.Env,
		"AGENT_RUN_HOME="+req.Home,
		"PATH="+filepath.Dir(req.AgentRun)+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	return nil
}

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "terminal-poll-doctest-"+d.DOCTEST_SESSION_ID)
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func ensureAgentRunBinary(t *testing.T, d *session.Doctest, repoRoot string) string {
	t.Helper()
	cache := sessionCacheDir(d)
	bin := filepath.Join(cache, "agent-run")
	ready := filepath.Join(cache, "binaries.ready")
	lock := filepath.Join(cache, "build.lock")
	err := withFileLock(t, lock, func() error {
		if fileExists(ready) && fileExists(bin) {
			return nil
		}
		if err := os.MkdirAll(cache, 0755); err != nil {
			return err
		}
		cmd := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", bin, "./agent-run")
		cmd.Dir = repoRoot
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("build agent-run: %w\n%s", err, string(out))
		}
		return os.WriteFile(ready, []byte("ok"), 0644)
	})
	if err != nil {
		t.Fatalf("ensure agent-run: %v", err)
	}
	return bin
}

func requirePlaywright(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("playwright-debug"); err != nil {
		t.Skipf("skipping terminal-poll test: playwright-debug not on PATH: %v", err)
	}
}

func findFreePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("find free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}

func waitForHealth(baseURL, bearer string, timeout time.Duration) error {
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(timeout)
	url := strings.TrimRight(baseURL, "/") + "/api/agent-run/health"
	for time.Now().Before(deadline) {
		httpReq, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		if bearer != "" {
			httpReq.Header.Set("Authorization", "Bearer "+bearer)
		}
		resp, err := client.Do(httpReq)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("health check timed out after %s: %s", timeout, url)
}

func stopWeb(t *testing.T, req *Request) {
	t.Helper()
	if req.webCmd == nil || req.webCmd.Process == nil {
		return
	}
	_ = req.webCmd.Process.Signal(os.Interrupt)
	done := make(chan error, 1)
	go func() { done <- req.webCmd.Wait() }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		_ = req.webCmd.Process.Kill()
		<-done
	}
	req.webCmd = nil
}

func startWebBackground(t *testing.T, req *Request) error {
	t.Helper()
	if req.Port == 0 {
		req.Port = findFreePort(t)
	}
	args := []string{"web", "--port", strconv.Itoa(req.Port), "--no-open", "--token", req.Token}
	cmd := exec.Command(req.AgentRun, args...)
	cmd.Env = append(os.Environ(), req.Env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start agent-run web: %w", err)
	}
	req.webCmd = cmd
	req.BaseURL = "http://127.0.0.1:" + strconv.Itoa(req.Port)
	t.Cleanup(func() { stopWeb(t, req) })
	if err := waitForHealth(req.BaseURL, req.Token, 30*time.Second); err != nil {
		stopWeb(t, req)
		return err
	}
	return nil
}

func startMappedPtywrap(t *testing.T, req *Request) string {
	t.Helper()
	upgrader := websocket.Upgrader{}
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

func writeMappedSessionFixture(t *testing.T, req *Request) {
	t.Helper()
	writeSessionFixture(t, req, true)
}

func writeSessionFixtureWithoutTerminalID(t *testing.T, req *Request) {
	t.Helper()
	writeSessionFixture(t, req, false)
}

func writeSessionFixture(t *testing.T, req *Request, includeTerminalID bool) {
	t.Helper()
	dir := filepath.Join(req.Home, "sessions", req.Runner, req.ChatSessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir session: %v", err)
	}
	meta := map[string]any{
		"runner":            req.Runner,
		"session_id":        req.ChatSessionID,
		"runner_session_id": req.RunnerSessionID,
		"status":            req.Status,
		"workspace":         req.TempDir,
		"created_at":        time.Now().UTC().Format(time.RFC3339),
		"updated_at":        time.Now().UTC().Format(time.RFC3339),
	}
	if includeTerminalID && req.TerminalSessionID != "" {
		meta["terminal_session_id"] = req.TerminalSessionID
	}
	writeJSONFile(t, filepath.Join(dir, "meta.json"), meta)
	var events string
	if req.Status == "running" {
		events = fmt.Sprintf(`{"type":"message","role":"user","text":"%s","timestamp":1}`+"\n", req.Prompt)
	} else {
		events = fmt.Sprintf(`{"type":"message","role":"user","text":"%s","timestamp":1}`+"\n", req.Prompt) +
			`{"type":"message","role":"assistant","text":"assistant keeps transcript","timestamp":2}` + "\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "events.jsonl"), []byte(events), 0644); err != nil {
		t.Fatalf("write events: %v", err)
	}
}

func writeRegistryEntry(home, runner, sessionID, listenAddr string) error {
	dir := filepath.Join(home, runner+"-registry")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	entry := map[string]any{
		"session_id":  sessionID,
		"listen_addr": listenAddr,
		"pid":         os.Getpid(),
		"created_at":  time.Now().Format(time.RFC3339Nano),
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, sessionID+".json"), data, 0644)
}

func writeTTYRegistryFixture(t *testing.T, req *Request, sessionID, listenAddr string) {
	t.Helper()
	if err := writeRegistryEntry(req.Home, req.Runner, sessionID, listenAddr); err != nil {
		t.Fatalf("write registry: %v", err)
	}
}

func startDelayedRegistryWriter(t *testing.T, req *Request) {
	t.Helper()
	home := req.Home
	runner := req.Runner
	sessionID := req.TerminalSessionID
	listenAddr := req.RegistryListenAddr
	delay := req.RegistryDelay
	go func() {
		time.Sleep(delay)
		if err := writeRegistryEntry(home, runner, sessionID, listenAddr); err != nil {
			t.Logf("delayed registry write failed: %v", err)
		}
	}()
}

func jsQuote(s string) string {
	return strconv.Quote(s)
}

func sessionBrowserScript(req *Request, body string) string {
	path := req.BaseURL + "/sessions/" + req.ChatSessionID
	return fmt.Sprintf(`
await page.setViewportSize({ width: 390, height: 844 });
await page.goto(%s, { waitUntil: 'domcontentloaded' });
await page.evaluate((token) => localStorage.setItem('agent-run-token', token), %s);
await page.goto(%s, { waitUntil: 'domcontentloaded' });
%s
`, jsQuote(req.BaseURL), jsQuote(req.Token), jsQuote(path), body)
}

// initTerminalPollMonitor registers page.on('request'|'response') before navigation.
func initTerminalPollMonitor() string {
	return `
let __terminalPollCount = 0;
let __terminalPollTimestamps = [];
let __terminalAvailableAt = 0;
let __terminalPollAfterAvailable = 0;

const __isTerminalStatusGet = (url, method) => {
  if (method !== 'GET') return false;
  return /\/api\/agent-run\/sessions\/[^/]+\/[^/]+\/terminal(?:\?|$)/.test(url);
};

page.on('request', (req) => {
  const url = req.url();
  if (!__isTerminalStatusGet(url, req.method())) return;
  const ts = Date.now();
  __terminalPollCount++;
  __terminalPollTimestamps.push(ts);
  if (__terminalAvailableAt > 0 && ts > __terminalAvailableAt + 250) {
    __terminalPollAfterAvailable++;
  }
});

page.on('response', async (resp) => {
  const url = resp.url();
  if (!__isTerminalStatusGet(url, 'GET')) return;
  if (resp.status() !== 200) return;
  try {
    const body = await resp.json();
    if (body && body.available === true && __terminalAvailableAt === 0) {
      __terminalAvailableAt = Date.now();
    }
  } catch (_) {}
});
`
}

// assertKnownTerminalNoRepeatPoll waits windowMs then asserts terminal GET count <= maxCount.
func assertKnownTerminalNoRepeatPoll(maxCount, windowMs int) string {
	return fmt.Sprintf(`
await page.waitForTimeout(%d);
if (__terminalPollCount > %d) {
  throw new Error('expected terminal GET count <= %d during %dms passive watch, got ' + __terminalPollCount + ' timestamps=' + JSON.stringify(__terminalPollTimestamps));
}
`, windowMs, maxCount, maxCount, windowMs)
}

// assertDiscoveryPollStopsAfterAvailable waits windowMs then asserts bounded discovery polls that stop after available.
func assertDiscoveryPollStopsAfterAvailable(maxTotal, windowMs int) string {
	return fmt.Sprintf(`
await page.waitForTimeout(%d);
if (__terminalPollCount === 0) {
  throw new Error('expected terminal GET count > 0 during discovery window, got 0');
}
if (__terminalPollCount > %d) {
  throw new Error('expected terminal GET count <= %d during %dms discovery window (not perpetual 500ms poll), got ' + __terminalPollCount + ' timestamps=' + JSON.stringify(__terminalPollTimestamps));
}
if (__terminalAvailableAt === 0) {
  throw new Error('terminal never reported available:true during discovery window');
}
if (__terminalPollAfterAvailable > 0) {
  throw new Error('expected 0 terminal GETs after available:true, got ' + __terminalPollAfterAvailable + ' timestamps=' + JSON.stringify(__terminalPollTimestamps) + ' availableAt=' + __terminalAvailableAt);
}
if (__terminalPollTimestamps.length > 1) {
  for (let i = 1; i < __terminalPollTimestamps.length; i++) {
    const gap = __terminalPollTimestamps[i] - __terminalPollTimestamps[i - 1];
    if (gap >= 400 && gap <= 600) {
      throw new Error('terminal GETs at ~500ms interval (perpetual fast poll): gapMs=' + gap + ' timestamps=' + JSON.stringify(__terminalPollTimestamps));
    }
  }
}
`, windowMs, maxTotal, maxTotal, windowMs)
}

func seedFinishedKnownTerminalSession(t *testing.T, req *Request) {
	t.Helper()
	req.Status = "finished"
	listenAddr := startMappedPtywrap(t, req)
	writeMappedSessionFixture(t, req)
	writeTTYRegistryFixture(t, req, req.TerminalSessionID, listenAddr)
}
```