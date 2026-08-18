# Scenario

**Feature**: shared harness for agent-run web sessions list pagination + search doctests

```
# build binary, isolated AGENT_RUN_HOME, free port, seed flat metas
build agent-run -> temp AGENT_RUN_HOME
  -> seed sessions/<id>/meta.json (controlled updated_at / prompt / status)
  -> agent-run web --port <free> --no-open
  -> Mode=http multi-step GET  |  playwright-debug UI @ 390×844
```

## Preconditions

- Repository contains `cmd/agent-run` and `go.mod` at repo root.
- `frontend-agent-run/dist` may be gitignored. Root setup ensures a minimal SPA
  `index.html` with `#root` so `//go:embed dist` compiles. Real Vite builds
  (when present) are left untouched — **ui/** leaves need a real SPA.
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Session storage is **flat**: `AGENT_RUN_HOME/sessions/<session_id>/meta.json`.
- Playwright leaves require `playwright-debug` on `PATH` (skip otherwise).
- Server binds **127.0.0.1** only; tests use `http://127.0.0.1:<port>/`.
- Product page size under test: **30**. Poll window cap: **150**.
- **counts** semantics: independent of `q` (full-store status buckets);
  pagination `total` applies `q` + `status`.

## Steps

1. Root `Setup` resolves repo root, ensures dist shell, builds `agent-run`, sets home/env.
2. Grouping `Setup` sets `Mode` (`http` vs `ui`) and default token mode.
3. Leaf `Setup` seeds metas, starts web, sets `HTTPSteps` or `PlaywrightScript`.
4. `Run` executes multi-step HTTP or `playwright-debug -e <script>`.
5. Leaf `Assert` checks status codes, JSON fields, or playwright exit.

## Context

- Default `WebTokenMode`: `explicit` with `Token` `test-token` unless a leaf sets `omit`.
- Seed helpers write fixed RFC3339 timestamps so order asserts are deterministic.
- Shared fixture timestamps are relative to a fixed base so parallel leaves stay independent.

```go
import (
	"runtime"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

// pageSize is the product home list page size.
const pageSize = 30

// seedBase is a fixed UTC base for deterministic updated_at ordering.
var seedBase = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	if req.WebTokenMode == "" {
		req.WebTokenMode = "explicit"
	}
	if req.WebTokenMode == "explicit" && req.Token == "" {
		req.Token = "test-token"
	}
	if err := ensureFrontendDistShell(req.RepoRoot); err != nil {
		return err
	}
	req.AgentRun = filepath.Join(req.TempDir, "bin", "agent-run")
	if err := buildAgentRun(t, req.RepoRoot, req.AgentRun); err != nil {
		return err
	}
	req.Env = append(req.Env,
		"AGENT_RUN_HOME="+req.Home,
		"PATH="+filepath.Dir(req.AgentRun)+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	return nil
}

func ensureFrontendDistShell(repoRoot string) error {
	distDir := filepath.Join(repoRoot, "frontend-agent-run", "dist")
	indexPath := filepath.Join(distDir, "index.html")
	if st, err := os.Stat(indexPath); err == nil && !st.IsDir() {
		return nil
	}
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return err
	}
	const shell = `<!doctype html>
<html lang="en">
<head><meta charset="UTF-8"><title>agent-run</title></head>
<body>
<div id="root"></div>
</body>
</html>
`
	return os.WriteFile(indexPath, []byte(shell), 0644)
}

func buildAgentRun(t *testing.T, repoRoot, out string) error {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}
	cmd := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", out, "./agent-run")
	cmd.Dir = repoRoot
	if outBytes, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build agent-run: %w\n%s", err, string(outBytes))
	}
	return nil
}

func sessionsListWebTokenMode(req *Request) string {
	switch req.WebTokenMode {
	case "omit", "auto", "explicit":
		return req.WebTokenMode
	default:
		return "explicit"
	}
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

func parseTokenFromStderr(stderr string) string {
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		const prefix = "agent-run web token: "
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
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
	if req.AgentRun == "" {
		return fmt.Errorf("AgentRun binary not built")
	}

	args := []string{"web", "--port", strconv.Itoa(req.Port), "--no-open"}
	switch sessionsListWebTokenMode(req) {
	case "omit":
		// open API
	case "auto":
		args = append(args, "--token", "auto")
	default:
		if req.Token == "" {
			req.Token = "test-token"
		}
		args = append(args, "--token", req.Token)
	}
	var stderrBuf bytes.Buffer
	cmd := exec.Command(req.AgentRun, args...)
	cmd.Env = append(os.Environ(), req.Env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start agent-run web: %w", err)
	}
	req.webCmd = cmd
	req.BaseURL = "http://127.0.0.1:" + strconv.Itoa(req.Port)
	t.Cleanup(func() { stopWeb(t, req) })

	mode := sessionsListWebTokenMode(req)
	switch mode {
	case "omit":
		if err := waitForHealth(req.BaseURL, "", 30*time.Second); err != nil {
			stopWeb(t, req)
			return err
		}
	case "auto":
		deadline := time.Now().Add(30 * time.Second)
		for time.Now().Before(deadline) {
			if tok := parseTokenFromStderr(stderrBuf.String()); tok != "" {
				req.Token = tok
				if err := waitForHealth(req.BaseURL, tok, 2*time.Second); err == nil {
					return nil
				}
			}
			time.Sleep(50 * time.Millisecond)
		}
		stopWeb(t, req)
		return fmt.Errorf("health check timed out (auto token), stderr:\n%s", stderrBuf.String())
	default:
		if err := waitForHealth(req.BaseURL, req.Token, 30*time.Second); err != nil {
			stopWeb(t, req)
			return err
		}
	}
	return nil
}

func requirePlaywright(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("playwright-debug"); err != nil {
		t.Skipf("skipping web-sessions-list UI test: playwright-debug not on PATH: %v", err)
	}
}

// writeSeedSession writes flat sessions/<id>/meta.json (+ minimal events.jsonl).
func writeSeedSession(t *testing.T, home string, s SeedSession) error {
	t.Helper()
	id := strings.TrimSpace(s.SessionID)
	if id == "" {
		return fmt.Errorf("SessionID required")
	}
	runner := strings.TrimSpace(s.Runner)
	if runner == "" {
		runner = "fake-codex"
	}
	status := strings.TrimSpace(s.Status)
	if status == "" {
		status = "idle"
	}
	created := s.CreatedAt
	if created.IsZero() {
		created = s.UpdatedAt
	}
	if created.IsZero() {
		created = seedBase
	}
	updated := s.UpdatedAt
	if updated.IsZero() {
		updated = created
	}
	ws := s.Workspace
	if ws == "" {
		ws = "/tmp/sessions-list-ws"
	}
	sessDir := filepath.Join(home, "sessions", id)
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return err
	}
	meta := map[string]any{
		"runner":         runner,
		"session_id":     id,
		"status":         status,
		"initial_prompt": s.InitialPrompt,
		"workspace":      ws,
		"created_at":     created.UTC().Format(time.RFC3339),
		"updated_at":     updated.UTC().Format(time.RFC3339),
	}
	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(sessDir, "meta.json"), metaBytes, 0644); err != nil {
		return err
	}
	prompt := s.InitialPrompt
	if prompt == "" {
		prompt = "seed " + id
	}
	events := fmt.Sprintf(
		`{"type":"message","role":"user","text":%s,"timestamp":%d}`+"\n",
		mustJSONString(prompt),
		updated.UnixMilli(),
	)
	return os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(events), 0644)
}

func mustJSONString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func seedSessions(t *testing.T, req *Request, sessions []SeedSession) error {
	t.Helper()
	req.Seeded = sessions
	for _, s := range sessions {
		if err := writeSeedSession(t, req.Home, s); err != nil {
			return err
		}
	}
	return nil
}

// defaultFiveSessions is a shared 5-session fixture:
//   alpha (oldest idle) … epsilon (newest idle)
// with one running (beta), one finished (gamma), unique prompt/workspace for q.
// Order by updated_at ascending: alpha < beta < gamma < delta < epsilon.
func defaultFiveSessions() []SeedSession {
	return []SeedSession{
		{
			SessionID:     "sess-alpha",
			Runner:        "fake-codex",
			Status:        "idle",
			InitialPrompt: "hello world alpha",
			Workspace:     "/tmp/ws-alpha",
			CreatedAt:     seedBase.Add(0 * time.Minute),
			UpdatedAt:     seedBase.Add(0 * time.Minute),
		},
		{
			SessionID:     "sess-beta",
			Runner:        "fake-codex",
			Status:        "running",
			InitialPrompt: "running beta",
			Workspace:     "/tmp/ws-beta",
			CreatedAt:     seedBase.Add(1 * time.Minute),
			UpdatedAt:     seedBase.Add(1 * time.Minute),
		},
		{
			SessionID:     "sess-gamma",
			Runner:        "opencode",
			Status:        "finished",
			InitialPrompt: "done gamma",
			Workspace:     "/tmp/ws-gamma",
			CreatedAt:     seedBase.Add(2 * time.Minute),
			UpdatedAt:     seedBase.Add(2 * time.Minute),
		},
		{
			SessionID:     "sess-delta",
			Runner:        "fake-codex",
			Status:        "idle",
			InitialPrompt: "newest-ish delta UNIQUE-QUERY-TOKEN",
			Workspace:     "/tmp/ws-unique-path",
			CreatedAt:     seedBase.Add(3 * time.Minute),
			UpdatedAt:     seedBase.Add(3 * time.Minute),
		},
		{
			SessionID:     "sess-epsilon",
			Runner:        "fake-codex",
			Status:        "idle",
			InitialPrompt: "brand newest epsilon",
			Workspace:     "/tmp/ws-epsilon",
			CreatedAt:     seedBase.Add(4 * time.Minute),
			UpdatedAt:     seedBase.Add(4 * time.Minute),
		},
	}
}

// manySessions builds n sessions with ascending updated_at and prompt "page-seed-NNN".
// Newest has the highest index; session_id = page-sess-NNN.
func manySessions(n int) []SeedSession {
	out := make([]SeedSession, 0, n)
	for i := 1; i <= n; i++ {
		ts := seedBase.Add(time.Duration(i) * time.Minute)
		out = append(out, SeedSession{
			SessionID:     fmt.Sprintf("page-sess-%03d", i),
			Runner:        "fake-codex",
			Status:        "idle",
			InitialPrompt: fmt.Sprintf("page-seed-%03d", i),
			Workspace:     "/tmp/ws-page",
			CreatedAt:     ts.Add(-30 * time.Second),
			UpdatedAt:     ts,
		})
	}
	return out
}

func sessionsPath(query string) string {
	p := "/api/agent-run/sessions"
	if strings.TrimSpace(query) == "" {
		return p
	}
	return p + "?" + query
}

func findHTTPResult(resp *Response, name string) (HTTPResult, bool) {
	if resp == nil {
		return HTTPResult{}, false
	}
	for _, r := range resp.HTTPResults {
		if r.Name == name {
			return r, true
		}
	}
	return HTTPResult{}, false
}

func parseJSONMap(t *testing.T, body string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("parse JSON: %v body=%q", err, truncate(body, 400))
	}
	return m
}

func sessionsFromBody(t *testing.T, body string) []map[string]any {
	t.Helper()
	m := parseJSONMap(t, body)
	raw, ok := m["sessions"]
	if !ok || raw == nil {
		t.Fatalf("response missing sessions key: %q", truncate(body, 300))
	}
	arr, ok := raw.([]any)
	if !ok {
		t.Fatalf("sessions is not array: %T", raw)
	}
	out := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		sm, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("session entry not object: %T", item)
		}
		out = append(out, sm)
	}
	return out
}

func sessionIDs(sessions []map[string]any) []string {
	out := make([]string, 0, len(sessions))
	for _, s := range sessions {
		id, _ := s["session_id"].(string)
		out = append(out, id)
	}
	return out
}

func jsonFloat(m map[string]any, key string) (float64, bool) {
	if m == nil {
		return 0, false
	}
	v, ok := m[key]
	if !ok || v == nil {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}

func jsonBool(m map[string]any, key string) (bool, bool) {
	if m == nil {
		return false, false
	}
	v, ok := m[key]
	if !ok || v == nil {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

func jsonCounts(m map[string]any) (map[string]any, bool) {
	if m == nil {
		return nil, false
	}
	raw, ok := m["counts"]
	if !ok || raw == nil {
		return nil, false
	}
	cm, ok := raw.(map[string]any)
	return cm, ok
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func seedTokenInPage(token string) string {
	escaped := strings.ReplaceAll(token, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	return fmt.Sprintf(`
await page.addInitScript(() => {
  localStorage.setItem('agent-run-token', '%s');
});
`, escaped)
}

func mobileViewportScript(body string) string {
	return fmt.Sprintf(`
await page.setViewportSize({ width: 390, height: 844 });
%s
`, body)
}

func openHomeWithSessions(baseURL string) string {
	base := strings.TrimRight(baseURL, "/")
	return fmt.Sprintf(`
await page.goto(%s, { waitUntil: 'networkidle' });
const home = page.locator('[data-testid="home-active"]');
await home.waitFor({ state: 'visible', timeout: 20000 });
const list = page.locator('[data-testid="session-list"]');
await list.waitFor({ state: 'visible', timeout: 20000 });
`, mustJSONString(base+"/"))
}

func assertPlaywrightOK(t *testing.T, resp *Response, err error) {
	t.Helper()
	if err != nil {
		stderr := ""
		if resp != nil {
			stderr = resp.PlaywrightStderr
		}
		t.Fatalf("Run error: %v\nstderr:\n%s", err, stderr)
	}
	if resp == nil {
		t.Fatal("nil Response")
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d\nstdout:\n%s\nstderr:\n%s",
			resp.PlaywrightExit, resp.PlaywrightStdout, resp.PlaywrightStderr)
	}
}

// requireOK200 fails unless the named HTTP result is 200.
func requireOK200(t *testing.T, resp *Response, name string) HTTPResult {
	t.Helper()
	r, ok := findHTTPResult(resp, name)
	if !ok {
		t.Fatalf("missing HTTP result %q", name)
	}
	if r.Status != http.StatusOK {
		t.Fatalf("%s expected 200, got %d body=%q", name, r.Status, truncate(r.Body, 400))
	}
	return r
}

// queryEncode is a tiny helper for building query strings in leaves.
func queryEncode(pairs map[string]string) string {
	v := url.Values{}
	for k, val := range pairs {
		v.Set(k, val)
	}
	return v.Encode()
}
```
