# Scenario

**Feature**: shared harness for agent-run web workspace path tap-to-expand doctests

```
# build binary, isolated AGENT_RUN_HOME, free port, optional deep cwd
build agent-run -> temp AGENT_RUN_HOME
  -> agent-run web --port <free> --no-open  (cmd.Dir = WebWorkingDir when set)
  -> playwright-debug @ 390×844 -> WorkspacePath expand/collapse/copy
```

## Preconditions

- Repository contains `cmd/agent-run` and `go.mod` at repo root.
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- `playwright-debug` must be on `PATH` (skip via `requirePlaywright` otherwise).
- Server binds **127.0.0.1** only; tests use `http://127.0.0.1:<port>/`.
- Viewport: **390×844** (set inside each playwright script via `mobileViewportScript`).
- Session storage is **flat**: `AGENT_RUN_HOME/sessions/<session_id>/`.

## Steps

1. Root `Setup` resolves repo root, builds `agent-run`, sets `req.Home` and env.
2. Grouping / leaf `Setup` sets surface defaults, workspace path, seeds session if needed, starts web, builds Playwright script.
3. `Run` executes `playwright-debug -e <script>`.
4. Leaf `Assert` checks playwright exit code and scenario id.

## Context

- Default `WebTokenMode`: `explicit` with `Token` `test-token` unless a grouping sets `omit`.
- Home long-path leaves set `WebWorkingDir` via `makeDeepWorkspaceDir` so
  `GET /api/agent-run/status` → long `workspace`.
- Session long-path leaves seed `meta.workspace` to a deep path under `TempDir`.
- Shared helpers mirror `cmd/agent-run/tests/web-layout/` (deep workspace, runner
  viewport assert, startWebBackground).

```go
import (
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

func buildAgentRun(t *testing.T, repoRoot, out string) error {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}
	cmd := exec.Command("go", "build", "-o", out, "./cmd/agent-run")
	cmd.Dir = repoRoot
	if outBytes, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build agent-run: %w\n%s", err, string(outBytes))
	}
	return nil
}

func workspaceWebTokenMode(req *Request) string {
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

func parseWorkspaceTokenFromStderr(stderr string) string {
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
	switch workspaceWebTokenMode(req) {
	case "omit":
		// open API — no --token flag
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
	if req.WebWorkingDir != "" {
		cmd.Dir = req.WebWorkingDir
	}
	cmd.Env = append(os.Environ(), req.Env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start agent-run web: %w", err)
	}
	req.webCmd = cmd
	req.BaseURL = "http://127.0.0.1:" + strconv.Itoa(req.Port)
	t.Cleanup(func() { stopWeb(t, req) })

	mode := workspaceWebTokenMode(req)
	switch mode {
	case "omit":
		if err := waitForHealth(req.BaseURL, "", 30*time.Second); err != nil {
			stopWeb(t, req)
			return err
		}
	case "auto":
		deadline := time.Now().Add(30 * time.Second)
		for time.Now().Before(deadline) {
			if tok := parseWorkspaceTokenFromStderr(stderrBuf.String()); tok != "" {
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
		t.Skipf("skipping web workspace-path test: playwright-debug not on PATH: %v", err)
	}
}

// mobileViewportScript wraps user script with viewport 390×844 and shared no-h-scroll check.
func mobileViewportScript(body string) string {
	return fmt.Sprintf(`
await page.setViewportSize({ width: 390, height: 844 });
%s
const docScrollWidth = await page.evaluate(() => document.documentElement.scrollWidth);
const docClientWidth = await page.evaluate(() => document.documentElement.clientWidth);
if (docScrollWidth > docClientWidth + 1) {
  throw new Error('horizontal scroll: scrollWidth=' + docScrollWidth + ' clientWidth=' + docClientWidth);
}
`, body)
}

func makeDeepWorkspaceDir(t *testing.T, base string) string {
	t.Helper()
	segments := []string{
		"agent-run", "very", "long", "nested", "workspace", "path", "segment",
		"for", "mobile", "header", "layout", "overflow", "regression", "case",
	}
	p := base
	for _, seg := range segments {
		p = filepath.Join(p, seg)
	}
	if err := os.MkdirAll(p, 0755); err != nil {
		t.Fatalf("makeDeepWorkspaceDir: %v", err)
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatalf("makeDeepWorkspaceDir abs: %v", err)
	}
	return abs
}

// makeShortWorkspaceDir creates a ≤2-segment path under base (…/ws/ab style).
func makeShortWorkspaceDir(t *testing.T, base string) string {
	t.Helper()
	p := filepath.Join(base, "ws", "ab")
	if err := os.MkdirAll(p, 0755); err != nil {
		t.Fatalf("makeShortWorkspaceDir: %v", err)
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatalf("makeShortWorkspaceDir abs: %v", err)
	}
	return abs
}

// shortWorkspaceLabel mirrors frontend-agent-run shortWorkspaceLabel for expects.
func shortWorkspaceLabel(workspace string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(workspace), "/\\")
	if trimmed == "" {
		return ""
	}
	parts := strings.FieldsFunc(trimmed, func(r rune) bool {
		return r == '/' || r == '\\'
	})
	if len(parts) == 0 {
		return trimmed
	}
	if len(parts) <= 2 {
		return strings.Join(parts, "/")
	}
	return "…/" + strings.Join(parts[len(parts)-2:], "/")
}

func jsString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	return s
}

func assertRunnerPickerWithinViewport() string {
	return `
const runnerPicker = page.locator('[data-testid="runner-picker"]');
await runnerPicker.waitFor({ state: 'visible', timeout: 15000 });
const runnerSelect = page.locator('[data-testid="runner-select"]');
await runnerSelect.waitFor({ state: 'visible', timeout: 15000 });
const pickerBox = await runnerPicker.boundingBox();
const pickerVp = page.viewportSize();
if (!pickerBox || !pickerVp) throw new Error('missing runner-picker box or viewport');
if (pickerBox.x < -1) throw new Error('runner picker off-screen left: x=' + pickerBox.x);
if (pickerBox.x + pickerBox.width > pickerVp.width + 1) {
  throw new Error('runner picker off-screen right: right=' + (pickerBox.x + pickerBox.width) + ' vpWidth=' + pickerVp.width);
}
`
}

func seedTokenInPage(token string) string {
	escaped := jsString(token)
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

// seedFlatSessionWithWorkspace writes flat sessions/<id> with meta.workspace.
func seedFlatSessionWithWorkspace(t *testing.T, home, sessionID, runner, status, workspace string) error {
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
		"workspace":  workspace,
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
	events := `{"type":"message","role":"user","text":"workspace path seed","timestamp":1719691200000}` + "\n" +
		`{"type":"message","role":"assistant","text":"ok","timestamp":1719691201000}` + "\n"
	return os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(events), 0644)
}

// playwright helpers: locate visible path text (prefer workspace-label, fallback workspace).
func jsWorkspaceLabelText() string {
	return `
async function workspaceLabelText() {
  const label = page.locator('[data-testid="workspace-label"]');
  if (await label.count() > 0) {
    return (await label.first().innerText()).trim();
  }
  const root = page.locator('[data-testid="workspace"]');
  return (await root.innerText()).trim();
}
`
}

func jsWaitWorkspaceVisible() string {
	return `
const workspace = page.locator('[data-testid="workspace"]');
await workspace.waitFor({ state: 'visible', timeout: 15000 });
`
}

func assertPlaywrightOK(t *testing.T, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run error: %v\nstderr:\n%s", err, resp.PlaywrightStderr)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright-debug exit %d\nstdout:\n%s\nstderr:\n%s",
			resp.PlaywrightExit, resp.PlaywrightStdout, resp.PlaywrightStderr)
	}
	if strings.TrimSpace(resp.PlaywrightStderr) != "" {
		t.Logf("playwright stderr: %s", resp.PlaywrightStderr)
	}
}
```
