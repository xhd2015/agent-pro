# Scenario

**Feature**: shared harness for agent-run web SPA static fallback + client-nav doctests

```
# build binary, isolated AGENT_RUN_HOME, free port
build agent-run -> temp AGENT_RUN_HOME -> agent-run web --port <free> --no-open
  -> HTTP GET static/API  |  playwright-debug client routes
```

## Preconditions

- Repository contains `cmd/agent-run` and `go.mod` at repo root.
- `frontend-agent-run/dist` may be gitignored. Root setup ensures a minimal SPA
  `index.html` with `#root` so `//go:embed dist` compiles and static HTML leaves
  can assert the shell. Real Vite builds (when present) are left untouched —
  **client-nav** leaves need a real SPA (or implementer frontend build) for
  React Router behavior.
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Playwright leaves require `playwright-debug` on `PATH` (skip otherwise).
- Server binds **127.0.0.1** only; tests use `http://127.0.0.1:<port>/`.

## Steps

1. Root `Setup` resolves repo root, ensures dist shell, builds `agent-run`, sets home/env.
2. Grouping `Setup` sets `Mode` (`http` vs `ui`) and default token mode.
3. Leaf `Setup` seeds sessions if needed, starts web, sets `HTTPPath(s)` or `PlaywrightScript`.
4. `Run` performs HTTP GETs or `playwright-debug -e <script>`.
5. Leaf `Assert` checks status/body/DOM contracts.

## Context

- Default `WebTokenMode`: `explicit` with `Token` `test-token` unless a leaf sets `omit`.
- Session storage is **flat**: `AGENT_RUN_HOME/sessions/<session_id>/`.
- Client routes: `/`, `/sessions/:sessionId`, `*` (NotFound). Auth is gate UI only.
- SPA nav marker: `window.__SPA_NAV_MARKER = 'alive'` via `addInitScript` — must survive soft nav.

```go
import (
	"runtime"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

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
	// Minimal SPA shell: #root required by static fallback leaves; </body> for bootstrap inject.
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

func spaWebTokenMode(req *Request) string {
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

func parseSPATokenFromStderr(stderr string) string {
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
	switch spaWebTokenMode(req) {
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

	mode := spaWebTokenMode(req)
	switch mode {
	case "omit":
		if err := waitForHealth(req.BaseURL, "", 30*time.Second); err != nil {
			stopWeb(t, req)
			return err
		}
	case "auto":
		deadline := time.Now().Add(30 * time.Second)
		for time.Now().Before(deadline) {
			if tok := parseSPATokenFromStderr(stderrBuf.String()); tok != "" {
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
		t.Skipf("skipping SPA client-nav test: playwright-debug not on PATH: %v", err)
	}
}

// seedFlatSession writes AGENT_RUN_HOME/sessions/<sessionID>/{meta.json,events.jsonl}.
func seedFlatSession(t *testing.T, home, sessionID, runner, status string) error {
	t.Helper()
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return fmt.Errorf("sessionID required")
	}
	if runner == "" {
		runner = "fake-codex"
	}
	if status == "" {
		status = "idle"
	}
	sessDir := filepath.Join(home, "sessions", sessionID)
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	meta := map[string]any{
		"runner":     runner,
		"session_id": sessionID,
		"status":     status,
		"workspace":  "/tmp/spa-structured-app",
		"created_at": now,
		"updated_at": now,
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(sessDir, "meta.json"), metaBytes, 0644); err != nil {
		return err
	}
	events := `{"type":"message","role":"user","text":"spa seed hello","timestamp":1719691200000}` + "\n" +
		`{"type":"message","role":"assistant","text":"spa seed reply","timestamp":1719691201000}` + "\n"
	return os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(events), 0644)
}

func htmlHasRootMount(body string) bool {
	// Match id="root" or id='root' on an element (typically div#root).
	lower := strings.ToLower(body)
	return strings.Contains(lower, `id="root"`) || strings.Contains(lower, `id='root'`)
}

func htmlHasSessionBootstrap(body string) bool {
	return strings.Contains(body, `id="agent-run-session-bootstrap"`) ||
		strings.Contains(body, `id='agent-run-session-bootstrap'`)
}

func bootstrapContainsSessionID(body, sessionID string) bool {
	if !htmlHasSessionBootstrap(body) {
		return false
	}
	// Payload is JSON inside the script tag; accept either nested meta field.
	return strings.Contains(body, sessionID) &&
		(strings.Contains(body, `"session_id"`) || strings.Contains(body, `"sessionId"`))
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

func clearTokenInPage() string {
	return `
await page.addInitScript(() => {
  localStorage.removeItem('agent-run-token');
});
`
}

func spaNavMarkerInit() string {
	return `
await page.addInitScript(() => {
  window.__SPA_NAV_MARKER = 'alive';
});
`
}

func assertSpaNavMarkerSurvives() string {
	return `
const marker = await page.evaluate(() => window.__SPA_NAV_MARKER);
if (marker !== 'alive') {
  throw new Error('expected window.__SPA_NAV_MARKER to survive soft navigation, got ' + JSON.stringify(marker));
}
`
}

func mobileViewportOptional(body string) string {
	// Client-nav is not layout-focused; still set a stable viewport for clicks.
	return fmt.Sprintf(`
await page.setViewportSize({ width: 390, height: 844 });
%s
`, body)
}
```
