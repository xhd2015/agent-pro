# Scenario

**Feature**: shared harness for `agent-run web` mobile layout doctests

```
build agent-run → temp AGENT_RUN_HOME → optional background web server → playwright-debug
```

## Preconditions

- Repository contains `cmd/agent-run` and `go.mod` at repo root.
- Each test uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- `playwright-debug` must be on `PATH` (leaf Run/Assert skip via `requirePlaywright`).

## Steps

1. Root `Setup` resolves repo root, builds `agent-run` binary, sets `req.Home` and env.
2. Leaf `Setup` picks free port, starts web server in background (or relies on Run), seeds data, sets `req.PlaywrightScript`.
3. `Run` waits for health if needed, executes `playwright-debug run <script>`.
4. Leaf `Assert` checks playwright exit code and layout-specific DOM expectations.

## Context

- Default `WebTokenMode`: `explicit` with `Token` `test-token` unless a leaf sets `omit` (open API).
- Viewport: **390×844** (set inside each playwright script).
- Server binds **127.0.0.1** only; tests use `http://127.0.0.1:<port>/`.

```go
import (
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
	"strconv"
	"strings"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../.."))
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

func layoutWebTokenMode(req *Request) string {
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

func parseLayoutTokenFromStderr(stderr string) string {
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
	switch layoutWebTokenMode(req) {
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

	mode := layoutWebTokenMode(req)
	switch mode {
	case "omit":
		if err := waitForHealth(req.BaseURL, "", 30*time.Second); err != nil {
			stopWeb(t, req)
			return err
		}
	case "auto":
		deadline := time.Now().Add(30 * time.Second)
		for time.Now().Before(deadline) {
			if tok := parseLayoutTokenFromStderr(stderrBuf.String()); tok != "" {
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
		t.Skipf("skipping web layout test: playwright-debug not on PATH: %v", err)
	}
}

// mobileViewportScript wraps user script with viewport 390×844 and shared scroll check.
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

// assertComposerPinnedBottom verifies [data-testid="composer"] sits at the viewport bottom.
func assertComposerPinnedBottom() string {
	return `
const composer = page.locator('[data-testid="composer"]');
await composer.waitFor({ state: 'visible', timeout: 15000 });
const box = await composer.boundingBox();
const vp = page.viewportSize();
if (!box || !vp) throw new Error('missing composer box or viewport');
const bottomGap = vp.height - (box.y + box.height);
if (bottomGap > 4) {
  throw new Error('composer not pinned to bottom: gap=' + bottomGap + 'px');
}
`
}

// assertAuthPagePinnedBottom verifies auth shell primary control is near viewport bottom.
func assertAuthPagePinnedBottom() string {
	return `
const input = page.locator('[data-testid="auth-token-input"], [data-testid="auth-page"] input').first();
await input.waitFor({ state: 'visible', timeout: 15000 });
const box = await input.boundingBox();
const vp = page.viewportSize();
if (!box || !vp) throw new Error('missing auth input box or viewport');
const bottomGap = vp.height - (box.y + box.height);
if (bottomGap > 80) {
  throw new Error('auth input not near bottom: gap=' + bottomGap + 'px');
}
`
}

func seedRoleTimelineSession(t *testing.T, home, runner, sessionID, workspace string, userTS, assistantTS int64) error {
	t.Helper()
	sessDir := filepath.Join(home, "sessions", runner, sessionID)
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	meta := map[string]any{
		"runner":     runner,
		"session_id": sessionID,
		"status":     "idle",
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
	eventsNDJSON := fmt.Sprintf("{\"type\":\"message\",\"role\":\"user\",\"text\":\"You said hi\",\"timestamp\":%d}\n", userTS) +
		fmt.Sprintf("{\"type\":\"message\",\"role\":\"assistant\",\"text\":\"Agent replied\",\"timestamp\":%d}\n", assistantTS)
	return os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(eventsNDJSON), 0644)
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

func seedRunningSessionAwaitingAssistant(t *testing.T, home, runner, sessionID string) error {
	t.Helper()
	sessDir := filepath.Join(home, "sessions", runner, sessionID)
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return err
	}
	now := time.Now().UTC()
	updatedAt := now.Add(-5 * time.Second).Format(time.RFC3339)
	createdAt := now.Add(-2 * time.Minute).Format(time.RFC3339)
	meta := map[string]any{
		"runner":     runner,
		"session_id": sessionID,
		"status":     "running",
		"created_at": createdAt,
		"updated_at": updatedAt,
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(sessDir, "meta.json"), metaBytes, 0644); err != nil {
		return err
	}
	userTS := now.Add(-4 * time.Second).UnixMilli()
	eventsNDJSON := fmt.Sprintf("{\"type\":\"message\",\"role\":\"user\",\"text\":\"Waiting for reply\",\"timestamp\":%d}\n", userTS)
	return os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(eventsNDJSON), 0644)
}

func assertInlineAssistantLoadingBelowUser() string {
	return `
const user = page.locator('[data-testid="message-item-user"]').last();
await user.waitFor({ state: 'visible', timeout: 15000 });
const loading = page.locator('[data-testid="message-item-assistant-loading"]');
await loading.waitFor({ state: 'visible', timeout: 15000 });
const userBox = await user.boundingBox();
const loadBox = await loading.boundingBox();
if (!userBox || !loadBox) throw new Error('missing user or loading bubble box');
if (loadBox.y < userBox.y + userBox.height - 2) {
  throw new Error('loading bubble must appear below user message: userBottom=' + (userBox.y + userBox.height) + ' loadingTop=' + loadBox.y);
}
`
}

func assertAssistantBubbleTextGrows() string {
	return `
const assistant = page.locator('[data-testid="message-item-assistant"]').last();
let prevLen = 0;
let grew = false;
for (let i = 0; i < 80; i++) {
  const count = await assistant.count();
  if (count > 0) {
    const text = (await assistant.innerText()).trim();
    if (text.length > prevLen && prevLen > 0) {
      grew = true;
      break;
    }
    if (text.length > 0) prevLen = text.length;
  }
  await page.waitForTimeout(250);
}
if (!grew) {
  throw new Error('assistant bubble text did not grow during streaming poll window');
}
`
}

func seedRunningSession(t *testing.T, home, runner, sessionID string, runningFor time.Duration) error {
	t.Helper()
	sessDir := filepath.Join(home, "sessions", runner, sessionID)
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return err
	}
	now := time.Now().UTC()
	updatedAt := now.Add(-runningFor).Format(time.RFC3339)
	createdAt := now.Add(-runningFor - 2*time.Minute).Format(time.RFC3339)
	meta := map[string]any{
		"runner":     runner,
		"session_id": sessionID,
		"status":     "running",
		"created_at": createdAt,
		"updated_at": updatedAt,
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(sessDir, "meta.json"), metaBytes, 0644); err != nil {
		return err
	}
	eventsNDJSON := "{\"type\":\"message\",\"role\":\"assistant\",\"text\":\"Working on your request…\"}\n"
	return os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(eventsNDJSON), 0644)
}

func seedIdleSessionForRunningCardNegative(t *testing.T, home, runner, sessionID string) error {
	t.Helper()
	sessDir := filepath.Join(home, "sessions", runner, sessionID)
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	meta := map[string]any{
		"runner":     runner,
		"session_id": sessionID,
		"status":     "idle",
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
	return os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(""), 0644)
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

func assertAgentRunningCardVisible() string {
	return `
const card = page.locator('[data-testid="agent-running-card"]');
await card.waitFor({ state: 'visible', timeout: 15000 });
const dur = page.locator('[data-testid="agent-running-duration"]');
await dur.waitFor({ state: 'visible', timeout: 15000 });
const durText = (await dur.innerText()).trim();
if (!durText) throw new Error('agent-running-duration is empty');
if (!/\d/.test(durText)) throw new Error('duration missing digits: ' + durText);
if (!/(:|m|s|for|running)/i.test(durText)) {
  throw new Error('duration missing time pattern: ' + durText);
}
`
}

func assertAgentRunningCardAbsent() string {
	return `
const card = page.locator('[data-testid="agent-running-card"]');
const count = await card.count();
if (count !== 0) {
  const visible = await card.isVisible().catch(() => false);
  if (visible) throw new Error('agent-running-card must not be visible when not running');
}
`
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
```