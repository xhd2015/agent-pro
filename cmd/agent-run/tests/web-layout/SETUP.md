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
- Live-run leaves set `req.GrokTTYRunnerBinary` via `ensureLayoutGrokMockEnv` before `startWebBackground`.
- Viewport: **390×844** (set inside each playwright script).
- Server binds **127.0.0.1** only; tests use `http://127.0.0.1:<port>/`.

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
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

const layoutGrokMockUUID = "a1111111-1111-4111-8111-111111111111"
const layoutGrokStreamMarker = "WEB_LAYOUT_STREAM_MARKER"
const layoutGrokAssistantPrefix = "WEB_MOCK_ASSISTANT:"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	repoRoot, err := findAgentProRoot(d.DOCTEST_ROOT)
	if err != nil {
		return err
	}
	req.RepoRoot = repoRoot
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
	cmd := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", out, "./agent-run")
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
	if req.GrokTTYRunnerBinary != "" {
		args = append(args,
			"--agent-runner", "grok-tty",
			"--grok-home", req.GrokHome,
			"--grok-tty-runner-binary", req.GrokTTYRunnerBinary,
		)
	}
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

func seedProgressCompactionSession(t *testing.T, home, runner, sessionID, workspace string) error {
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
	toolID := "call-compact-demo"
	longOutput := strings.Repeat("line-output\n", 40)
	eventsNDJSON := strings.Join([]string{
		`{"type":"message","role":"user","text":"run tools","timestamp":1719691200000}`,
		`{"type":"think","timestamp":1719691201000,"text":"First think pass"}`,
		`{"type":"think","timestamp":1719691202000,"text":"Second think pass should replace first"}`,
		fmt.Sprintf(`{"type":"tool_call","timestamp":1719691203000,"text":"Shell","tool":"tool","tool_call_id":%q}`, toolID),
		fmt.Sprintf(`{"type":"tool_call","timestamp":1719691204000,"text":"Shell","tool":"tool","tool_call_id":%q}`, toolID),
		`{"type":"think","timestamp":1719691204500,"text":"Think between duplicate tool updates"}`,
		fmt.Sprintf(`{"type":"tool_call","timestamp":1719691205000,"text":"Shell","tool":"tool","tool_call_id":%q,"output":%q}`, toolID, longOutput),
		`{"type":"message","role":"assistant","text":"Done","timestamp":1719691234567}`,
	}, "\n") + "\n"
	return os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(eventsNDJSON), 0644)
}

func seedProgressMultiToolOrderingSession(t *testing.T, home, runner, sessionID, workspace string) error {
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
	toolA := "call-order-alpha"
	toolB := "call-order-beta"
	eventsNDJSON := strings.Join([]string{
		`{"type":"message","role":"user","text":"run two tools","timestamp":1719691200000}`,
		fmt.Sprintf(`{"type":"tool_call","timestamp":1719691201000,"text":"Shell","tool":"tool","tool_call_id":%q}`, toolA),
		fmt.Sprintf(`{"type":"tool_call","timestamp":1719691202000,"text":"Shell","tool":"tool","tool_call_id":%q}`, toolB),
		fmt.Sprintf(`{"type":"tool_call","timestamp":1719691203000,"text":"Shell","tool":"tool","tool_call_id":%q,"output":"alpha done"}`, toolA),
		`{"type":"message","role":"assistant","text":"Both tools finished","timestamp":1719691234567}`,
	}, "\n") + "\n"
	return os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(eventsNDJSON), 0644)
}

func seedGrokTTYMessageCardSession(t *testing.T, home, sessionID, workspace string) error {
	t.Helper()
	runner := "grok-tty"
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
	toolID := "call-grok-demo"
	eventsNDJSON := strings.Join([]string{
		`{"type":"message","role":"user","timestamp":1783493436429,"text":"run ls and pwd","extensions":{"grok_session":{}}}`,
		`{"type":"think","timestamp":1783493436887,"text":"Run ls and pwd in workspace.","extensions":{"grok_session":{}}}`,
		fmt.Sprintf(`{"type":"tool_call","timestamp":1783493436889,"text":"Shell","tool":"tool","tool_call_id":%q,"extensions":{"grok_session":{"status":"pending"}}}`, toolID),
		fmt.Sprintf(`{"type":"tool_call","timestamp":1783493437572,"text":"Shell","tool":"tool","tool_call_id":%q,"output":"pwd and ls output","extensions":{"grok_session":{"status":"completed"}}}`, toolID),
		`{"type":"think","timestamp":1783493440380,"text":"Summarize results for user.","extensions":{"grok_session":{}}}`,
		`{"type":"message","role":"assistant","timestamp":1783493442157,"text":"**pwd:** /tmp/workspace\n\n**ls:** agent cmd frontend","extensions":{"grok_session":{}}}`,
		`{"type":"done","timestamp":1783493442160,"extensions":{"grok_session":{}}}`,
		`{"type":"message","role":"user","timestamp":1783493456134,"text":"what did I say","extensions":{"grok_session":{"turn_index":1}}}`,
		`{"type":"think","timestamp":1783493456847,"text":"Recall prior user messages.","extensions":{"grok_session":{"turn_index":1}}}`,
		`{"type":"message","role":"assistant","timestamp":1783493458483,"text":"You said: run ls and pwd, then what did I say.","extensions":{"grok_session":{"turn_index":1}}}`,
		`{"type":"done","timestamp":1783493458573,"extensions":{"grok_session":{"turn_index":1}}}`,
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(eventsNDJSON), 0644); err != nil {
		return err
	}
	syncCheckpoint := map[string]any{
		"grok_session_id": "",
		"updates_path":    "",
		"updates_offset":  999999,
		"turn_index":      2,
	}
	syncBytes, err := json.Marshal(syncCheckpoint)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(sessDir, "grok-sync.json"), syncBytes, 0644)
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

func assertUserMessageCount(expected int) string {
	return fmt.Sprintf(`{
const userBubbles = page.locator('[data-testid="message-item-user"]');
const userCount = await userBubbles.count();
if (userCount !== %d) {
  throw new Error('expected %d user message bubbles, got ' + userCount);
}
}
`, expected, expected)
}

// waitForUserMessageCount polls until user bubble count equals expected (does not require idle/assistant).
func waitForUserMessageCount(expected int) string {
	return fmt.Sprintf(`
{
  const userBubbles = page.locator('[data-testid="message-item-user"]');
  let ok = false;
  for (let i = 0; i < 240; i++) {
    const userCount = await userBubbles.count();
    if (userCount === %d) {
      ok = true;
      break;
    }
    if (userCount > %d) {
      throw new Error('user message count exceeded expected: count=' + userCount + ' expected=%d');
    }
    await page.waitForTimeout(250);
  }
  if (!ok) {
    const finalCount = await userBubbles.count();
    throw new Error('timed out waiting for %d user message bubbles, got ' + finalCount);
  }
}
`, expected, expected, expected, expected)
}

func assertDistinctUserPromptsOnce(prompts []string) string {
	var b strings.Builder
	b.WriteString(`{
const userBubbles = page.locator('[data-testid="message-item-user"]');
const userTexts = await userBubbles.allInnerTexts();
const normalized = userTexts.map((t) => t.trim());
`)
	for _, p := range prompts {
		escaped := strings.ReplaceAll(p, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `'`, `\'`)
		b.WriteString(fmt.Sprintf(`
{
  const needle = '%s';
  const hits = normalized.filter((t) => t.includes(needle)).length;
  if (hits !== 1) {
    throw new Error('expected prompt to appear once: ' + needle + ' hits=' + hits);
  }
}
`, escaped))
	}
	b.WriteString("}\n")
	return b.String()
}

// assertNoDuplicateUserMessagesDuringRun polls while the session is running and fails
// immediately if user bubble count exceeds maxExpected (catches mid-run duplication).
func assertNoDuplicateUserMessagesDuringRun(maxExpected int) string {
	return fmt.Sprintf(`{
const userBubbles = page.locator('[data-testid="message-item-user"]');
const runningCard = page.locator('[data-testid="agent-running-card"]');
const inlineLoading = page.locator('[data-testid="message-item-assistant-loading"]');
const statusPill = page.locator('.status-pill');
const layoutIsSessionRunning = async () => {
  if (await runningCard.isVisible().catch(() => false)) return true;
  if (await inlineLoading.isVisible().catch(() => false)) return true;
  if (await statusPill.isVisible().catch(() => false)) {
    const text = (await statusPill.innerText()).trim().toLowerCase();
    if (text === 'running') return true;
  }
  return false;
};
let sawRunning = false;
for (let i = 0; i < 40; i++) {
  if (await layoutIsSessionRunning()) {
    sawRunning = true;
    break;
  }
  await page.waitForTimeout(250);
}
if (!sawRunning) {
  throw new Error('session run never started after follow-up');
}
for (let i = 0; i < 120; i++) {
  const userCount = await userBubbles.count();
  if (userCount > %d) {
    throw new Error('duplicate user messages during run: count=' + userCount + ' max=%d');
  }
  if (!(await layoutIsSessionRunning())) break;
  await page.waitForTimeout(250);
}
}
`, maxExpected, maxExpected)
}

// sendComposerMessage is safe to concatenate multiple times (block-scoped bindings).
// Waits for send enabled only after fill — empty draft keeps the button disabled.
func sendComposerMessage(text string) string {
	escaped := strings.ReplaceAll(text, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	return fmt.Sprintf(`
{
  const composerInput = page.locator('[data-testid="composer-input"]');
  await composerInput.waitFor({ state: 'visible', timeout: 15000 });
  await composerInput.fill('%s');
  const sendBtn = page.locator('[data-testid="send-button"]');
  await sendBtn.waitFor({ state: 'visible', timeout: 15000 });
  let enabled = false;
  for (let i = 0; i < 40; i++) {
    const disabled = await sendBtn.isDisabled().catch(() => true);
    if (!disabled) {
      enabled = true;
      break;
    }
    await page.waitForTimeout(100);
  }
  if (!enabled) {
    throw new Error('send-button stayed disabled after fill');
  }
  await sendBtn.click();
}
`, escaped)
}

// waitForSessionRunComplete is block-scoped so it can appear more than once in one script.
// Complete when either:
//   (1) not running and ≥1 assistant bubble (streaming success), or
//   (2) not running (no running card / inline loading / status=running) after having
//       observed a run start — covers follow-ups where mock bind never lands assistant text.
func waitForSessionRunComplete() string {
	return `
{
  const runningCard = page.locator('[data-testid="agent-running-card"]');
  const inlineLoading = page.locator('[data-testid="message-item-assistant-loading"]');
  const statusPill = page.locator('.status-pill');
  const layoutIsSessionRunning = async () => {
    if (await runningCard.isVisible().catch(() => false)) return true;
    if (await inlineLoading.isVisible().catch(() => false)) return true;
    if (await statusPill.isVisible().catch(() => false)) {
      const text = (await statusPill.innerText()).trim().toLowerCase();
      if (text === 'running') return true;
    }
    return false;
  };
  let sawRunning = false;
  let runDone = false;
  for (let i = 0; i < 120; i++) {
    const running = await layoutIsSessionRunning();
    if (running) {
      sawRunning = true;
      await page.waitForTimeout(500);
      continue;
    }
    const assistantCount = await page.locator('[data-testid="message-item-assistant"]').count();
    // Success path: assistant present and idle.
    if (assistantCount >= 1) {
      runDone = true;
      break;
    }
    // Idle after a run started (or never showed chrome but enough time elapsed with no running).
    // Avoid false-complete on the first idle tick before the run begins.
    if (sawRunning || i >= 6) {
      runDone = true;
      break;
    }
    await page.waitForTimeout(500);
  }
  if (!runDone) {
    throw new Error('session run did not complete within poll window');
  }
}
`
}

// initStreamingTransportMonitor registers page.on('request') counters. Call before navigation.
func initStreamingTransportMonitor() string {
	return `
let __layoutDetailGetCount = 0;
let __layoutSSECount = 0;
page.on('request', (req) => {
  const url = req.url();
  if (!url.includes('/api/agent-run/sessions/')) return;
  if (url.includes('/events/stream')) {
    __layoutSSECount++;
    return;
  }
  if (req.method() !== 'GET') return;
  if (/\/api\/agent-run\/sessions\/[^/]+\/[^/?]+(?:\?|$)/.test(url)) {
    __layoutDetailGetCount++;
  }
});
`
}

// assertStreamingTransportProfile waits 8s then asserts SSE used and session-detail GETs are bounded.
func assertStreamingTransportProfile() string {
	return `
await page.waitForTimeout(8000);
if (__layoutSSECount < 1) {
  throw new Error('expected >=1 events/stream request during streaming window, got ' + __layoutSSECount);
}
if (__layoutDetailGetCount > 3) {
  throw new Error('expected session-detail GET count <= 3 during streaming window, got ' + __layoutDetailGetCount);
}
`
}

// initSSEPersistenceMonitor registers page.on('request') + requestfailed before navigation.
// Tracks stream request starts and client-side aborts (net::ERR_ABORTED) on /events/stream URLs.
func initSSEPersistenceMonitor() string {
	return `
let __layoutDetailGetCount = 0;
let __layoutSSECount = 0;
let __layoutSSEAbortedCount = 0;

page.on('request', (req) => {
  const url = req.url();
  if (!url.includes('/api/agent-run/sessions/')) return;
  if (url.includes('/events/stream')) {
    __layoutSSECount++;
    return;
  }
  if (req.method() !== 'GET') return;
  if (/\/api\/agent-run\/sessions\/[^/]+\/[^/?]+(?:\?|$)/.test(url)) {
    __layoutDetailGetCount++;
  }
});

page.on('requestfailed', (req) => {
  const url = req.url();
  if (!url.includes('/events/stream')) return;
  const failure = req.failure();
  const errText = failure ? failure.errorText : '';
  if (errText.includes('net::ERR_ABORTED') || /abort|cancel/i.test(errText)) {
    __layoutSSEAbortedCount++;
  }
});
`
}

// assertSSEStaysConnectedDuringRun waits 8s then asserts one persistent stream with no aborts.
func assertSSEStaysConnectedDuringRun() string {
	return `
await page.waitForTimeout(8000);
if (__layoutSSECount !== 1) {
  throw new Error('expected exactly 1 events/stream request during run, got ' + __layoutSSECount);
}
if (__layoutSSEAbortedCount !== 0) {
  throw new Error('expected 0 aborted/cancelled events/stream requests, got ' + __layoutSSEAbortedCount);
}
if (__layoutDetailGetCount > 3) {
  throw new Error('expected session-detail GET count <= 3 during run window, got ' + __layoutDetailGetCount);
}
`
}

// initSessionDetailPollMonitor registers page.on('request') + requestfailed before navigation.
// Tallies exact session-detail GETs for /api/agent-run/sessions/:runner/:id (excludes stream, messages, list).
func initSessionDetailPollMonitor() string {
	return `
let __layoutDetailGetCount = 0;
let __layoutDetailGetTimestamps = [];
let __layoutSSECount = 0;
let __layoutSSEAbortedCount = 0;

const __layoutIsSessionDetailGet = (url, method) => {
  if (!url.includes('/api/agent-run/sessions/')) return false;
  if (url.includes('/events/stream')) return false;
  if (url.includes('/messages')) return false;
  if (method !== 'GET') return false;
  return /\/api\/agent-run\/sessions\/[^/]+\/[^/?]+(?:\?|$)/.test(url);
};

page.on('request', (req) => {
  const url = req.url();
  if (!url.includes('/api/agent-run/sessions/')) return;
  if (url.includes('/events/stream')) {
    __layoutSSECount++;
    return;
  }
  if (__layoutIsSessionDetailGet(url, req.method())) {
    __layoutDetailGetCount++;
    __layoutDetailGetTimestamps.push(Date.now());
  }
});

page.on('requestfailed', (req) => {
  const url = req.url();
  if (!url.includes('/events/stream')) return;
  const failure = req.failure();
  const errText = failure ? failure.errorText : '';
  if (errText.includes('net::ERR_ABORTED') || /abort|cancel/i.test(errText)) {
    __layoutSSEAbortedCount++;
  }
});
`
}

// assertNoSessionDetailPollWhileRunning waits windowMs then asserts no meta-poll detail GETs while SSE is active.
func assertNoSessionDetailPollWhileRunning(expectedInitial int, windowMs int) string {
	return fmt.Sprintf(`
await page.waitForTimeout(%d);
if (__layoutDetailGetCount !== %d) {
  throw new Error('expected exactly %d session-detail GET during run window, got ' + __layoutDetailGetCount + ' timestamps=' + JSON.stringify(__layoutDetailGetTimestamps));
}
if (__layoutSSECount !== 1) {
  throw new Error('expected exactly 1 events/stream request during run, got ' + __layoutSSECount);
}
if (__layoutSSEAbortedCount !== 0) {
  throw new Error('expected 0 aborted/cancelled events/stream requests, got ' + __layoutSSEAbortedCount);
}
if (__layoutDetailGetTimestamps.length > 1) {
  for (let i = 1; i < __layoutDetailGetTimestamps.length; i++) {
    const gap = __layoutDetailGetTimestamps[i] - __layoutDetailGetTimestamps[i - 1];
    if (gap >= 4000 && gap <= 6000) {
      throw new Error('session-detail GETs at ~5s interval (meta poll): gapMs=' + gap + ' timestamps=' + JSON.stringify(__layoutDetailGetTimestamps));
    }
  }
}
`, windowMs, expectedInitial, expectedInitial)
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

// seedIdleSessionWithUserMessage writes flat sessions/<sessionID>/{meta.json,events.jsonl}
// (store layout is flat; runner lives in meta only).
func seedIdleSessionWithUserMessage(t *testing.T, home, runner, sessionID, userText string) error {
	t.Helper()
	return seedIdleSessionWithUserAndAssistant(t, home, runner, sessionID, userText, "")
}

// seedIdleSessionWithUserAndAssistant writes flat sessions/<sessionID> with one user message
// and optional assistant reply (status=idle). Used for follow-up leaves that must not depend
// on a live first-turn grok bind.
func seedIdleSessionWithUserAndAssistant(t *testing.T, home, runner, sessionID, userText, assistantText string) error {
	t.Helper()
	sessDir := filepath.Join(home, "sessions", sessionID)
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
	escapeJSON := func(s string) string {
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `"`, `\"`)
		return s
	}
	userTS := time.Now().UTC().Add(-2 * time.Minute).UnixMilli()
	assistantTS := userTS + 30_000
	var b strings.Builder
	b.WriteString(fmt.Sprintf("{\"type\":\"message\",\"role\":\"user\",\"text\":\"%s\",\"timestamp\":%d}\n",
		escapeJSON(userText), userTS))
	if strings.TrimSpace(assistantText) != "" {
		b.WriteString(fmt.Sprintf("{\"type\":\"message\",\"role\":\"assistant\",\"text\":\"%s\",\"timestamp\":%d}\n",
			escapeJSON(assistantText), assistantTS))
	}
	return os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(b.String()), 0644)
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

// seedLayoutScrollSession seeds an idle session with alternating user/assistant messages
// so message-list overflows on mobile viewport (≥15 messages by default).
func seedLayoutScrollSession(t *testing.T, home, runner, sessionID string, messageCount int) error {
	t.Helper()
	if messageCount < 15 {
		messageCount = 15
	}
	sessDir := filepath.Join(home, "sessions", runner, sessionID)
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	meta := map[string]any{
		"runner":     runner,
		"session_id": sessionID,
		"status":     "idle",
		"workspace":  "/tmp/layout-scroll-workspace",
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
	var b strings.Builder
	baseTS := time.Now().UTC().Add(-time.Duration(messageCount) * time.Minute).UnixMilli()
	for i := 0; i < messageCount; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		text := fmt.Sprintf("Layout scroll seed message %d — enough text to push the transcript well past one mobile viewport height.", i+1)
		ts := baseTS + int64(i)*60000
		escaped := strings.ReplaceAll(text, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		b.WriteString(fmt.Sprintf("{\"type\":\"message\",\"role\":\"%s\",\"text\":\"%s\",\"timestamp\":%d}\n", role, escaped, ts))
	}
	return os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(b.String()), 0644)
}

func buildFakeCodexIntoPath(t *testing.T, req *Request) error {
	t.Helper()
	fakeCodex := filepath.Join(req.TempDir, "bin", "fake-codex")
	if err := os.MkdirAll(filepath.Dir(fakeCodex), 0755); err != nil {
		return err
	}
	build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", fakeCodex, "./fake-codex")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build fake-codex: %w\n%s", err, string(out))
	}
	req.Env = append(req.Env, "PATH="+filepath.Dir(fakeCodex)+string(os.PathListSeparator)+os.Getenv("PATH"))
	return nil
}

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

func llmMockGrokLayoutHook(prompt, sessionUUID, marker string, sleepSec int) string {
	if sleepSec <= 0 {
		sleepSec = 2
	}
	streamPart1, streamPart2, streamPart3 := layoutGrokStreamParts(marker)
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
updates="$dir/updates.jsonl"
printf %%s\\n "{\"sessionUpdate\":\"user_message_chunk\",\"content\":{\"type\":\"text\",\"text\":\"$submitted\"}}" > "$updates"
stream_chunk() { printf %%s\\n "$1" >> "$updates"; }
stream_chunk "{\"sessionUpdate\":\"agent_message_chunk\",\"content\":{\"type\":\"text\",\"text\":\"%s\"}}"
sleep 0.55
stream_chunk "{\"sessionUpdate\":\"agent_message_chunk\",\"content\":{\"type\":\"text\",\"text\":\"%s\"}}"
sleep 0.55
stream_chunk "{\"sessionUpdate\":\"agent_message_chunk\",\"content\":{\"type\":\"text\",\"text\":\"%s\"}}"
sleep 0.55
stream_chunk "{\"sessionUpdate\":\"turn_completed\"}"
sleep %d
exit 0
'`, prompt, sessionUUID, sessionUUID, streamPart1, streamPart2, streamPart3, sleepSec)
}

func layoutGrokStreamParts(marker string) (string, string, string) {
	marker = strings.TrimSpace(marker)
	if marker == "" {
		marker = layoutGrokStreamMarker
	}
	n := len(marker)
	if n < 3 {
		return marker, marker, marker
	}
	third := n / 3
	twoThird := (2 * n) / 3
	if third < 1 {
		third = 1
	}
	if twoThird <= third {
		twoThird = third + 1
	}
	if twoThird >= n {
		twoThird = n - 1
	}
	return marker[:third], marker[:twoThird], marker
}

func ensureLayoutGrokMockEnv(t *testing.T, req *Request, prompt, marker string, sleepSec int) error {
	t.Helper()
	if err := buildLLMMockRunGrok(t, req); err != nil {
		return err
	}
	hook := llmMockGrokLayoutHook(prompt, layoutGrokMockUUID, marker, sleepSec)
	req.Env = append(req.Env,
		"LLM_MOCK_RUN_GROK_COMMAND="+hook,
		"AGENT_RUN_GROK_TTY_GROK_SESSION_ID="+layoutGrokMockUUID,
	)
	return nil
}

func layoutGrokAssistantMarker(prompt string) string {
	return layoutGrokAssistantPrefix + prompt
}

func openLiveGrokTTYSession(baseURL, prompt string) string {
	escaped := strings.ReplaceAll(prompt, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	base := strings.TrimRight(baseURL, "/")
	// Flat SPA route: /sessions/:sessionId (no runner segment).
	// Block-scoped so bindings never collide with later waitForChatActive / sendComposerMessage.
	return fmt.Sprintf(`
{
  const res = await fetch('%s/api/agent-run/sessions', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ runner: 'grok-tty', prompt: '%s' }),
  });
  if (!res.ok && res.status !== 202) throw new Error('create session failed: ' + res.status);
  const data = await res.json();
  const sid = data.session.session_id;
  const runner = data.session.runner;
  if (runner !== 'grok-tty') throw new Error('expected grok-tty runner, got ' + runner);
  await page.goto('%s/sessions/' + encodeURIComponent(sid), { waitUntil: 'domcontentloaded' });
  const chat = page.locator('[data-testid="chat-active"]');
  await chat.waitFor({ state: 'visible', timeout: 15000 });
  const messages = page.locator('[data-testid="message-list"]');
  await messages.waitFor({ state: 'visible', timeout: 15000 });
}
`, base, escaped, base)
}

// openHomeWithComposer opens `/` and waits for composer + runner-select (empty or list chrome).
func openHomeWithComposer(baseURL string) string {
	base := strings.TrimRight(baseURL, "/")
	return fmt.Sprintf(`
{
  await page.goto('%s/', { waitUntil: 'domcontentloaded' });
  const composer = page.locator('[data-testid="composer"]');
  await composer.waitFor({ state: 'visible', timeout: 15000 });
  const runnerSelectHome = page.locator('[data-testid="runner-select"]');
  await runnerSelectHome.waitFor({ state: 'visible', timeout: 15000 });
}
`, base)
}

// selectRunnerOption sets the home/session runner <select> to the given value.
// Waits until the option exists (runners list hydrates from /api/agent-run/runners).
func selectRunnerOption(runner string) string {
	escaped := strings.ReplaceAll(runner, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	return fmt.Sprintf(`
{
  const runnerSelect = page.locator('[data-testid="runner-select"]');
  await runnerSelect.waitFor({ state: 'visible', timeout: 15000 });
  const option = page.locator('[data-testid="runner-select"] option[value="%s"]');
  let hasOption = false;
  for (let i = 0; i < 60; i++) {
    if (await option.count() > 0) {
      hasOption = true;
      break;
    }
    await page.waitForTimeout(250);
  }
  if (!hasOption) {
    throw new Error('runner option never appeared: %s');
  }
  await runnerSelect.selectOption('%s').catch(async () => {
    await page.evaluate((value) => {
      const el = document.querySelector('[data-testid="runner-select"]');
      if (!el) return;
      el.value = value;
      el.dispatchEvent(new Event('change', { bubbles: true }));
    }, '%s');
  });
  const selected = await runnerSelect.inputValue().catch(() => '');
  if (selected !== '%s') {
    throw new Error('runner-select value=' + selected + ', expected %s');
  }
}
`, escaped, escaped, escaped, escaped, escaped, escaped)
}

// waitForSendButtonEnabled polls until the composer send control is clickable.
func waitForSendButtonEnabled() string {
	return `
{
  const sendBtn = page.locator('[data-testid="send-button"]');
  let enabled = false;
  for (let i = 0; i < 80; i++) {
    const disabled = await sendBtn.isDisabled().catch(() => true);
    if (!disabled) {
      enabled = true;
      break;
    }
    await page.waitForTimeout(250);
  }
  if (!enabled) {
    throw new Error('send-button stayed disabled');
  }
}
`
}

// waitForSessionIdle polls until running card / inline loading / running status are gone.
func waitForSessionIdle() string {
	return `
{
  const runningCard = page.locator('[data-testid="agent-running-card"]');
  const inlineLoading = page.locator('[data-testid="message-item-assistant-loading"]');
  const statusPill = page.locator('.status-pill');
  let idle = false;
  for (let i = 0; i < 160; i++) {
    const cardVisible = await runningCard.isVisible().catch(() => false);
    const loadingVisible = await inlineLoading.isVisible().catch(() => false);
    let statusRunning = false;
    if (await statusPill.isVisible().catch(() => false)) {
      const text = (await statusPill.innerText()).trim().toLowerCase();
      statusRunning = text === 'running';
    }
    if (!cardVisible && !loadingVisible && !statusRunning) {
      idle = true;
      break;
    }
    await page.waitForTimeout(250);
  }
  if (!idle) {
    throw new Error('session never became idle within poll window');
  }
}
`
}

// assertLiveFollowUpMessageCardOrder polls the live timeline after a second user prompt.
// Locks: chronological users (prompt1 then prompt2), each appears once, and no non-empty
// assistant bubble appears before the first user (anti-regression for strip-all-users merge).
// Optional reload checks the same order after full page refresh.
func assertLiveFollowUpMessageCardOrder(prompt1, prompt2 string, includeReload bool) string {
	p1 := strings.ReplaceAll(prompt1, `\`, `\\`)
	p1 = strings.ReplaceAll(p1, `'`, `\'`)
	p2 := strings.ReplaceAll(prompt2, `\`, `\\`)
	p2 = strings.ReplaceAll(p2, `'`, `\'`)
	reloadBlock := ""
	if includeReload {
		reloadBlock = `
await page.reload({ waitUntil: 'domcontentloaded', timeout: 30000 });
await page.locator('[data-testid="chat-active"]').waitFor({ state: 'visible', timeout: 20000 });
{
  let reloaded = false;
  for (let i = 0; i < 40; i++) {
    const tl = await readTimeline();
    if (findUserIndex(tl, PROMPT1) >= 0 && findUserIndex(tl, PROMPT2) >= 0) {
      assertHealthyOrder(tl, 'after reload');
      reloaded = true;
      break;
    }
    await page.waitForTimeout(500);
  }
  if (!reloaded) {
    throw new Error('after reload: both user prompts never appeared');
  }
}
`
	}
	return fmt.Sprintf(`
{
  const PROMPT1 = '%s';
  const PROMPT2 = '%s';
  const readTimeline = async () => page.evaluate(() => {
    const nodes = Array.from(
      document.querySelectorAll(
        '[data-testid="message-item-user"], [data-testid="message-item-assistant"]',
      ),
    );
    return nodes.map((el) => {
      const testid = el.getAttribute('data-testid') || '';
      const role = testid.includes('user') ? 'user' : 'assistant';
      const text = (el.querySelector('.message-body')?.textContent || '').trim();
      return { role, text };
    });
  });
  const findUserIndex = (timeline, needle) =>
    timeline.findIndex(
      (e) => e.role === 'user' && (e.text.includes(needle) || needle.includes(e.text)),
    );
  const countUserHits = (timeline, needle) =>
    timeline.filter(
      (e) => e.role === 'user' && (e.text.includes(needle) || needle.includes(e.text)),
    ).length;
  const assertHealthyOrder = (timeline, phase) => {
    const i1 = findUserIndex(timeline, PROMPT1);
    const i2 = findUserIndex(timeline, PROMPT2);
    if (i1 < 0) throw new Error(phase + ': missing user prompt ' + JSON.stringify(PROMPT1));
    if (i2 < 0) throw new Error(phase + ': missing user prompt ' + JSON.stringify(PROMPT2));
    if (i1 > i2) {
      throw new Error(
        phase + ': user order inverted: ' + JSON.stringify(PROMPT1) + ' at ' + i1 +
        ' after ' + JSON.stringify(PROMPT2) + ' at ' + i2,
      );
    }
    const hits1 = countUserHits(timeline, PROMPT1);
    const hits2 = countUserHits(timeline, PROMPT2);
    if (hits1 !== 1) {
      throw new Error(phase + ': expected prompt once: ' + JSON.stringify(PROMPT1) + ' hits=' + hits1);
    }
    if (hits2 !== 1) {
      throw new Error(phase + ': expected prompt once: ' + JSON.stringify(PROMPT2) + ' hits=' + hits2);
    }
    const firstAssistant = timeline.findIndex((e) => e.role === 'assistant' && e.text.length > 0);
    if (firstAssistant >= 0 && firstAssistant < i1) {
      throw new Error(
        phase + ': assistant at ' + firstAssistant +
        ' appears before first user ' + JSON.stringify(PROMPT1) + ' at ' + i1,
      );
    }
  };

  let sawSecondUser = false;
  let healthyLive = false;
  const pollDeadline = Date.now() + 60000;
  while (Date.now() < pollDeadline) {
    const tl = await readTimeline();
    if (findUserIndex(tl, PROMPT2) >= 0) {
      sawSecondUser = true;
      assertHealthyOrder(tl, 'live');
      healthyLive = true;
      const runningCard = page.locator('[data-testid="agent-running-card"]');
      const cardVisible = await runningCard.isVisible().catch(() => false);
      const statusPill = page.locator('.status-pill');
      let statusRunning = false;
      if (await statusPill.isVisible().catch(() => false)) {
        const text = (await statusPill.innerText()).trim().toLowerCase();
        statusRunning = text === 'running';
      }
      // After both users visible with healthy order, prefer to settle once not running.
      if (!cardVisible && !statusRunning) {
        await page.waitForTimeout(800);
        const settled = await readTimeline();
        assertHealthyOrder(settled, 'live-settled');
        break;
      }
    }
    await page.waitForTimeout(250);
  }
  if (!sawSecondUser) {
    throw new Error('second user prompt never appeared: ' + JSON.stringify(PROMPT2));
  }
  if (!healthyLive) {
    throw new Error('live follow-up never reached healthy message-card order');
  }
  %s
}
`, p1, p2, reloadBlock)
}

func waitForChatActive() string {
	return `
{
  const chat = page.locator('[data-testid="chat-active"]');
  await chat.waitFor({ state: 'visible', timeout: 15000 });
  const messages = page.locator('[data-testid="message-list"]');
  await messages.waitFor({ state: 'visible', timeout: 15000 });
}
`
}

func openSeededSessionPage(baseURL, sessionPath string) string {
	base := strings.TrimRight(baseURL, "/")
	return fmt.Sprintf(`
await page.goto('%s%s', { waitUntil: 'domcontentloaded' });
%s`, base, sessionPath, waitForChatActive())
}

func recordFrozenScrollTop() string {
	return `
let __layoutFrozenScrollTop = await page.locator('[data-testid="message-list"]').evaluate((el) => el.scrollTop);
`
}

// assertNoDocumentVerticalScroll verifies html/body do not scroll (±2px tolerance).
func assertNoDocumentVerticalScroll() string {
	return `
{
  const scrollHeight = await page.evaluate(() => document.documentElement.scrollHeight);
  const clientHeight = await page.evaluate(() => document.documentElement.clientHeight);
  if (Math.abs(scrollHeight - clientHeight) > 2) {
    throw new Error('document scrolls vertically: scrollHeight=' + scrollHeight + ' clientHeight=' + clientHeight);
  }
  const bodyScrollHeight = await page.evaluate(() => document.body.scrollHeight);
  const bodyClientHeight = await page.evaluate(() => document.body.clientHeight);
  if (Math.abs(bodyScrollHeight - bodyClientHeight) > 2) {
    throw new Error('body scrolls vertically: scrollHeight=' + bodyScrollHeight + ' clientHeight=' + bodyClientHeight);
  }
}
`
}

// assertMessageListOverflows verifies only the transcript region has vertical overflow.
func assertMessageListOverflows() string {
	return `
{
  const list = page.locator('[data-testid="message-list"]');
  const metrics = await list.evaluate((el) => ({
    scrollHeight: el.scrollHeight,
    clientHeight: el.clientHeight,
  }));
  if (metrics.scrollHeight <= metrics.clientHeight) {
    throw new Error('message-list must overflow: scrollHeight=' + metrics.scrollHeight + ' clientHeight=' + metrics.clientHeight);
  }
}
`
}

// waitForMessageListOverflow polls until message-list scrollHeight exceeds clientHeight (+10px slack).
func waitForMessageListOverflow() string {
	return `
{
  const list = page.locator('[data-testid="message-list"]');
  let overflow = false;
  for (let i = 0; i < 120; i++) {
    const metrics = await list.evaluate((el) => ({
      scrollHeight: el.scrollHeight,
      clientHeight: el.clientHeight,
    }));
    if (metrics.scrollHeight > metrics.clientHeight + 10) {
      overflow = true;
      break;
    }
    await page.waitForTimeout(250);
  }
  if (!overflow) {
    throw new Error('message-list never overflowed within 30s timeout');
  }
}
`
}

// scrollMessageListUpFromBottom scrolls the transcript up by at least minPx from the bottom.
func scrollMessageListUpFromBottom(minPx int) string {
	return fmt.Sprintf(`
{
  const list = page.locator('[data-testid="message-list"]');
  await list.evaluate((el, minPx) => {
    const maxTop = Math.max(0, el.scrollHeight - el.clientHeight);
    el.scrollTop = Math.max(0, maxTop - minPx);
  }, %d);
  await page.waitForTimeout(100);
}
`, minPx)
}

// assertChromeFixedWhileMessageListScrolls records chrome positions, scrolls list, asserts ±2px stability.
func assertChromeFixedWhileMessageListScrolls() string {
	return `
{
  const topBar = page.locator('.top-bar');
  const sessionHeader = page.locator('.session-header');
  const composer = page.locator('[data-testid="composer"]');
  await topBar.waitFor({ state: 'visible', timeout: 15000 });
  await sessionHeader.waitFor({ state: 'visible', timeout: 15000 });
  await composer.waitFor({ state: 'visible', timeout: 15000 });
  const beforeTop = await topBar.boundingBox();
  const beforeHeader = await sessionHeader.boundingBox();
  const beforeComposer = await composer.boundingBox();
  if (!beforeTop || !beforeHeader || !beforeComposer) {
    throw new Error('missing chrome bounding boxes before scroll');
  }
  const list = page.locator('[data-testid="message-list"]');
  await list.evaluate((el) => {
    el.scrollTop = Math.max(0, el.scrollHeight - el.clientHeight - 300);
  });
  await page.waitForTimeout(150);
  const afterTop = await topBar.boundingBox();
  const afterHeader = await sessionHeader.boundingBox();
  const afterComposer = await composer.boundingBox();
  if (!afterTop || !afterHeader || !afterComposer) {
    throw new Error('missing chrome bounding boxes after scroll');
  }
  const tol = 2;
  if (Math.abs(afterTop.y - beforeTop.y) > tol) {
    throw new Error('top-bar moved after message-list scroll: beforeY=' + beforeTop.y + ' afterY=' + afterTop.y);
  }
  if (Math.abs(afterHeader.y - beforeHeader.y) > tol) {
    throw new Error('session-header moved after message-list scroll: beforeY=' + beforeHeader.y + ' afterY=' + afterHeader.y);
  }
  if (Math.abs(afterComposer.y - beforeComposer.y) > tol) {
    throw new Error('composer moved after message-list scroll: beforeY=' + beforeComposer.y + ' afterY=' + afterComposer.y);
  }
}
`
}

const layoutBottomThresholdPx = 80

func assertMessageListAtBottom() string {
	return fmt.Sprintf(`
{
  const list = page.locator('[data-testid="message-list"]');
  const distance = await list.evaluate((el) => el.scrollHeight - el.scrollTop - el.clientHeight);
  if (distance > %d) {
    throw new Error('message-list not at bottom: distanceFromBottom=' + distance + ' threshold=%d');
  }
}
`, layoutBottomThresholdPx, layoutBottomThresholdPx)
}

func assertMessageListDetached() string {
	return fmt.Sprintf(`
{
  const list = page.locator('[data-testid="message-list"]');
  const distance = await list.evaluate((el) => el.scrollHeight - el.scrollTop - el.clientHeight);
  if (distance <= %d) {
    throw new Error('message-list still at bottom after scroll-up: distanceFromBottom=' + distance);
  }
}
`, layoutBottomThresholdPx)
}

// assertFollowAtBottomDuringStreaming polls assistant growth and requires scrollTop stay at bottom.
func assertFollowAtBottomDuringStreaming() string {
	return fmt.Sprintf(`
{
  const list = page.locator('[data-testid="message-list"]');
  const assistant = page.locator('[data-testid="message-item-assistant"]').last();
  let prevLen = 0;
  let bottomChecks = 0;
  for (let i = 0; i < 80; i++) {
    const count = await assistant.count();
    if (count > 0) {
      const text = (await assistant.innerText()).trim();
      if (text.length > prevLen && prevLen > 0) {
        const distance = await list.evaluate((el) => el.scrollHeight - el.scrollTop - el.clientHeight);
        if (distance > %d) {
          throw new Error('auto-follow failed during stream: distanceFromBottom=' + distance + ' textLen=' + text.length);
        }
        bottomChecks++;
      }
      if (text.length > 0) prevLen = text.length;
    }
    await page.waitForTimeout(250);
  }
  if (bottomChecks < 1) {
    throw new Error('assistant text grew but no at-bottom checks ran during streaming window');
  }
}
`, layoutBottomThresholdPx)
}

// assertScrollTopFrozenDuringStreaming asserts ±2px scrollTop stability while assistant text grows.
// Uses script-scoped __layoutFrozenScrollTop when set (e.g. before composer send).
func assertScrollTopFrozenDuringStreaming() string {
	return `
{
  const list = page.locator('[data-testid="message-list"]');
  const assistant = page.locator('[data-testid="message-item-assistant"]').last();
  const runningCard = page.locator('[data-testid="agent-running-card"]');
  const inlineLoading = page.locator('[data-testid="message-item-assistant-loading"]');
  const statusPill = page.locator('.status-pill');
  const layoutIsSessionRunning = async () => {
    if (await runningCard.isVisible().catch(() => false)) return true;
    if (await inlineLoading.isVisible().catch(() => false)) return true;
    if (await statusPill.isVisible().catch(() => false)) {
      const text = (await statusPill.innerText()).trim().toLowerCase();
      if (text === 'running') return true;
    }
    return false;
  };

  const frozenScrollTop = typeof __layoutFrozenScrollTop !== 'undefined'
    ? __layoutFrozenScrollTop
    : await list.evaluate((el) => el.scrollTop);

  let sawRunning = false;
  for (let i = 0; i < 40; i++) {
    if (await layoutIsSessionRunning()) {
      sawRunning = true;
      break;
    }
    await page.waitForTimeout(250);
  }
  if (!sawRunning) {
    throw new Error('session run never started after follow-up');
  }

  const assertFrozen = async () => {
    const scrollTop = await list.evaluate((el) => el.scrollTop);
    if (Math.abs(scrollTop - frozenScrollTop) > 2) {
      throw new Error('scrollTop moved while detached: frozen=' + frozenScrollTop + ' now=' + scrollTop);
    }
  };

  let baselineCount = await assistant.count();
  let initialLastText = '';
  let prevLen = 0;
  if (baselineCount > 0) {
    initialLastText = (await assistant.innerText()).trim();
    prevLen = initialLastText.length;
  }
  let freezeChecks = 0;
  let sawTextGrowth = false;

  for (let i = 0; i < 120; i++) {
    const stillRunning = await layoutIsSessionRunning();
    let grewThisTick = false;
    const assistantCount = await assistant.count();
    if (assistantCount > baselineCount) {
      sawTextGrowth = true;
      baselineCount = assistantCount;
      prevLen = (await assistant.innerText()).trim().length;
      grewThisTick = true;
    } else if (assistantCount > 0) {
      const current = (await assistant.innerText()).trim();
      if (current !== initialLastText) sawTextGrowth = true;
      if (current.length > prevLen) grewThisTick = true;
      prevLen = current.length;
    }
    if (await inlineLoading.isVisible().catch(() => false)) {
      sawTextGrowth = true;
      grewThisTick = true;
    }

    if (stillRunning || grewThisTick) {
      await assertFrozen();
      freezeChecks++;
    }

    if (!stillRunning && sawTextGrowth && freezeChecks >= 1) break;
    await page.waitForTimeout(250);
  }

  if (freezeChecks < 1) {
    throw new Error('no frozen-scrollTop checks ran during streaming window');
  }
  if (!sawTextGrowth) {
    throw new Error('assistant text did not grow during streaming window');
  }
}
`
}

// assertJumpToLatestChipFlow waits for chip while detached + streaming, taps it, asserts follow resumes.
func assertJumpToLatestChipFlow() string {
	return fmt.Sprintf(`
{
  const chip = page.locator('[data-testid="jump-to-latest"]');
  let chipVisible = false;
  for (let i = 0; i < 80; i++) {
    if (await chip.isVisible().catch(() => false)) {
      chipVisible = true;
      break;
    }
    await page.waitForTimeout(250);
  }
  if (!chipVisible) {
    throw new Error('jump-to-latest chip never became visible while detached with streaming content');
  }
  await chip.click();
  await page.waitForTimeout(200);
  const list = page.locator('[data-testid="message-list"]');
  const distance = await list.evaluate((el) => el.scrollHeight - el.scrollTop - el.clientHeight);
  if (distance > %d) {
    throw new Error('jump-to-latest did not scroll to bottom: distanceFromBottom=' + distance);
  }
  const stillVisible = await chip.isVisible().catch(() => false);
  if (stillVisible) {
    throw new Error('jump-to-latest chip still visible after tap');
  }
}
`, layoutBottomThresholdPx)
}

// recordMessageListScrollTop stores scrollTop in a playwright script-scoped let binding.
func recordMessageListScrollTop(varName string) string {
	return fmt.Sprintf(`
let __layout%s = await page.locator('[data-testid="message-list"]').evaluate((el) => el.scrollTop);
`, varName)
}

// assertMessageListScrollTopEqualsVar asserts scrollTop unchanged vs recorded let binding (±2px).
func assertMessageListScrollTopEqualsVar(varName string) string {
	return fmt.Sprintf(`
{
  const scrollTop = await page.locator('[data-testid="message-list"]').evaluate((el) => el.scrollTop);
  if (typeof __layout%s === 'undefined') throw new Error('missing recorded scrollTop let __layout%s');
  if (Math.abs(scrollTop - __layout%s) > 2) {
    throw new Error('message-list scrollTop changed: recorded=' + __layout%s + ' now=' + scrollTop);
  }
}
`, varName, varName, varName, varName)
}

const layoutScrollContainerSessionList = "session-list"

func homeSessionID(index int) string {
	return fmt.Sprintf("home-sess-%03d", index)
}

// seedManyHomeSessions creates count session dirs (home-sess-001 …) with distinct session_id
// and staggered updated_at. Home list is newest-first: higher index = newer = closer to top.
func seedManyHomeSessions(t *testing.T, home, runner string, count int) error {
	t.Helper()
	if count < 20 {
		count = 20
	}
	base := time.Now().UTC().Add(-time.Duration(count+1) * time.Minute)
	for i := 1; i <= count; i++ {
		if err := writeHomeSessionDir(home, runner, i, base.Add(time.Duration(i)*time.Minute)); err != nil {
			return err
		}
	}
	return nil
}

// writeHomeSessionDir writes flat sessions/<session_id>/meta.json (store layout is flat).
func writeHomeSessionDir(home, runner string, index int, updatedAt time.Time) error {
	sessionID := homeSessionID(index)
	sessDir := filepath.Join(home, "sessions", sessionID)
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return err
	}
	createdAt := updatedAt.Add(-2 * time.Minute).Format(time.RFC3339)
	updatedAtStr := updatedAt.Format(time.RFC3339)
	meta := map[string]any{
		"runner":         runner,
		"session_id":     sessionID,
		"status":         "idle",
		"initial_prompt": fmt.Sprintf("Home scroll seed session %d", index),
		"workspace":      "/tmp/home-scroll-workspace",
		"created_at":     createdAt,
		"updated_at":     updatedAtStr,
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(sessDir, "meta.json"), metaBytes, 0644); err != nil {
		return err
	}
	userTS := updatedAt.UnixMilli()
	eventsNDJSON := fmt.Sprintf("{\"type\":\"message\",\"role\":\"user\",\"text\":\"Home scroll seed session %d\",\"timestamp\":%d}\n", index, userTS)
	return os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(eventsNDJSON), 0644)
}

// appendHomeSessionDir writes one additional home session dir after page load (poll refresh tests).
func appendHomeSessionDir(t *testing.T, home, runner string, index int) error {
	t.Helper()
	updatedAt := time.Now().UTC()
	return writeHomeSessionDir(home, runner, index, updatedAt)
}

// scheduleAppendHomeSessionDir starts a goroutine that appends a session dir after a delay.
// Errors surface via t.Cleanup after the playwright script finishes.
func scheduleAppendHomeSessionDir(t *testing.T, req *Request, runner string, index int, after time.Duration) {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		time.Sleep(after)
		done <- appendHomeSessionDir(t, req.Home, runner, index)
	}()
	t.Cleanup(func() {
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("appendHomeSessionDir(%d): %v", index, err)
			}
		case <-time.After(30 * time.Second):
			t.Errorf("appendHomeSessionDir(%d) timed out", index)
		}
	})
}

func openHomePage(baseURL string) string {
	base := strings.TrimRight(baseURL, "/")
	return fmt.Sprintf(`
await page.goto('%s/', { waitUntil: 'domcontentloaded' });
const homeActive = page.locator('[data-testid="home-active"]');
await homeActive.waitFor({ state: 'visible', timeout: 15000 });
const list = page.locator('[data-testid="session-list"]');
await list.waitFor({ state: 'visible', timeout: 15000 });
`, base)
}

func waitForHomePollRefresh() string {
	return `
await page.waitForTimeout(4000);
`
}

func scrollContainerLocator(testid string) string {
	return fmt.Sprintf(`[data-testid="%s"]`, testid)
}

func assertScrollContainerOverflows(testid string) string {
	loc := scrollContainerLocator(testid)
	return fmt.Sprintf(`
{
  const list = page.locator('%s');
  const metrics = await list.evaluate((el) => ({
    scrollHeight: el.scrollHeight,
    clientHeight: el.clientHeight,
  }));
  if (metrics.scrollHeight <= metrics.clientHeight) {
    throw new Error('%s must overflow: scrollHeight=' + metrics.scrollHeight + ' clientHeight=' + metrics.clientHeight);
  }
}
`, loc, testid)
}

func waitForScrollContainerOverflow(testid string) string {
	loc := scrollContainerLocator(testid)
	return fmt.Sprintf(`
{
  const list = page.locator('%s');
  let overflow = false;
  for (let i = 0; i < 120; i++) {
    const metrics = await list.evaluate((el) => ({
      scrollHeight: el.scrollHeight,
      clientHeight: el.clientHeight,
    }));
    if (metrics.scrollHeight > metrics.clientHeight + 10) {
      overflow = true;
      break;
    }
    await page.waitForTimeout(250);
  }
  if (!overflow) {
    throw new Error('%s never overflowed within 30s timeout');
  }
}
`, loc, testid)
}

func scrollContainerToBottom(testid string) string {
	loc := scrollContainerLocator(testid)
	return fmt.Sprintf(`
{
  const list = page.locator('%s');
  await list.evaluate((el) => {
    el.scrollTop = Math.max(0, el.scrollHeight - el.clientHeight);
  });
  await page.waitForTimeout(100);
}
`, loc)
}

func scrollContainerUpFromBottom(testid string, minPx int) string {
	loc := scrollContainerLocator(testid)
	return fmt.Sprintf(`
{
  const list = page.locator('%s');
  await list.evaluate((el, minPx) => {
    const maxTop = Math.max(0, el.scrollHeight - el.clientHeight);
    el.scrollTop = Math.max(0, maxTop - minPx);
  }, %d);
  await page.waitForTimeout(100);
}
`, loc, minPx)
}

func assertScrollContainerAtBottom(testid string) string {
	loc := scrollContainerLocator(testid)
	return fmt.Sprintf(`
{
  const list = page.locator('%s');
  const distance = await list.evaluate((el) => el.scrollHeight - el.scrollTop - el.clientHeight);
  if (distance > %d) {
    throw new Error('%s not at bottom: distanceFromBottom=' + distance + ' threshold=%d');
  }
}
`, loc, layoutBottomThresholdPx, testid, layoutBottomThresholdPx)
}

func assertScrollContainerDetached(testid string) string {
	loc := scrollContainerLocator(testid)
	return fmt.Sprintf(`
{
  const list = page.locator('%s');
  const distance = await list.evaluate((el) => el.scrollHeight - el.scrollTop - el.clientHeight);
  if (distance <= %d) {
    throw new Error('%s still at bottom after scroll-up: distanceFromBottom=' + distance);
  }
}
`, loc, layoutBottomThresholdPx, testid)
}

func recordScrollContainerScrollTop(varName, testid string) string {
	loc := scrollContainerLocator(testid)
	return fmt.Sprintf(`
let __layout%s = await page.locator('%s').evaluate((el) => el.scrollTop);
`, varName, loc)
}

func assertScrollContainerScrollTopEqualsVar(varName, testid string) string {
	loc := scrollContainerLocator(testid)
	return fmt.Sprintf(`
{
  const scrollTop = await page.locator('%s').evaluate((el) => el.scrollTop);
  if (typeof __layout%s === 'undefined') throw new Error('missing recorded scrollTop let __layout%s');
  if (Math.abs(scrollTop - __layout%s) > 2) {
    throw new Error('%s scrollTop changed: recorded=' + __layout%s + ' now=' + scrollTop);
  }
}
`, loc, varName, varName, varName, testid, varName)
}

func assertSessionListOverflows() string {
	return assertScrollContainerOverflows(layoutScrollContainerSessionList)
}

func waitForSessionListOverflow() string {
	return waitForScrollContainerOverflow(layoutScrollContainerSessionList)
}

// Home session-list is newest-first: "latest" is at the TOP. Session message-list
// still uses bottom-follow helpers above; home wrappers use top-anchor semantics.

func scrollSessionListToTop() string {
	loc := scrollContainerLocator(layoutScrollContainerSessionList)
	return fmt.Sprintf(`
{
  const list = page.locator('%s');
  await list.evaluate((el) => { el.scrollTop = 0; });
  await page.waitForTimeout(100);
}
`, loc)
}

// scrollSessionListDownFromTop moves away from newest (top) to detach follow mode.
// Uses real mouse wheel so React onWheel → markUserScrollIntent runs before scroll.
func scrollSessionListDownFromTop(minPx int) string {
	loc := scrollContainerLocator(layoutScrollContainerSessionList)
	return fmt.Sprintf(`
{
  const list = page.locator('%s');
  await list.scrollIntoViewIfNeeded();
  const box = await list.boundingBox();
  if (!box) throw new Error('session-list bounding box missing');
  await page.mouse.move(box.x + box.width / 2, box.y + Math.min(40, box.height / 2));
  // Wheel down = away from newest (top); may need multiple ticks on some engines.
  let remaining = %d;
  while (remaining > 0) {
    const step = Math.min(remaining, 120);
    await page.mouse.wheel(0, step);
    remaining -= step;
    await page.waitForTimeout(40);
  }
  // Ensure we landed far enough from top even if wheel delta was clamped.
  await list.evaluate((el, minPx) => {
    if (el.scrollTop < minPx) {
      el.scrollTop = Math.min(Math.max(0, el.scrollHeight - el.clientHeight), minPx);
    }
  }, %d);
  await page.waitForTimeout(150);
  const distance = await list.evaluate((el) => el.scrollTop);
  if (distance <= 80) {
    throw new Error('scrollSessionListDownFromTop failed: still near top scrollTop=' + distance);
  }
}
`, loc, minPx, minPx)
}

func assertSessionListAtTop() string {
	loc := scrollContainerLocator(layoutScrollContainerSessionList)
	return fmt.Sprintf(`
{
  const list = page.locator('%s');
  const distance = await list.evaluate((el) => el.scrollTop);
  if (distance > %d) {
    throw new Error('%s not at top: distanceFromTop=' + distance + ' threshold=%d');
  }
}
`, loc, layoutBottomThresholdPx, layoutScrollContainerSessionList, layoutBottomThresholdPx)
}

func assertSessionListDetached() string {
	loc := scrollContainerLocator(layoutScrollContainerSessionList)
	return fmt.Sprintf(`
{
  const list = page.locator('%s');
  const distance = await list.evaluate((el) => el.scrollTop);
  if (distance <= %d) {
    throw new Error('%s still at top after scroll-down: distanceFromTop=' + distance);
  }
}
`, loc, layoutBottomThresholdPx, layoutScrollContainerSessionList)
}

// Deprecated aliases kept so older leaves compile if any still reference bottom helpers.
func scrollSessionListToBottom() string {
	return scrollSessionListToTop()
}

func scrollSessionListUpFromBottom(minPx int) string {
	return scrollSessionListDownFromTop(minPx)
}

func assertSessionListAtBottom() string {
	return assertSessionListAtTop()
}

func recordSessionListScrollTop(varName string) string {
	return recordScrollContainerScrollTop(varName, layoutScrollContainerSessionList)
}

func assertSessionListScrollTopEqualsVar(varName string) string {
	return assertScrollContainerScrollTopEqualsVar(varName, layoutScrollContainerSessionList)
}

// assertChromeFixedWhileSessionListScrolls records home chrome positions, scrolls session-list, asserts ±2px stability.
func assertChromeFixedWhileSessionListScrolls() string {
	return `
{
  const topBar = page.locator('.top-bar.top-bar-home');
  const composer = page.locator('[data-testid="composer"]');
  await topBar.waitFor({ state: 'visible', timeout: 15000 });
  await composer.waitFor({ state: 'visible', timeout: 15000 });
  const beforeTop = await topBar.boundingBox();
  const beforeComposer = await composer.boundingBox();
  if (!beforeTop || !beforeComposer) {
    throw new Error('missing home chrome bounding boxes before scroll');
  }
  const list = page.locator('[data-testid="session-list"]');
  await list.evaluate((el) => {
    el.scrollTop = Math.max(0, el.scrollHeight - el.clientHeight - 300);
  });
  await page.waitForTimeout(150);
  const afterTop = await topBar.boundingBox();
  const afterComposer = await composer.boundingBox();
  if (!afterTop || !afterComposer) {
    throw new Error('missing home chrome bounding boxes after scroll');
  }
  const tol = 2;
  if (Math.abs(afterTop.y - beforeTop.y) > tol) {
    throw new Error('top-bar-home moved after session-list scroll: beforeY=' + beforeTop.y + ' afterY=' + afterTop.y);
  }
  if (Math.abs(afterComposer.y - beforeComposer.y) > tol) {
    throw new Error('composer moved after session-list scroll: beforeY=' + beforeComposer.y + ' afterY=' + afterComposer.y);
  }
}
` + assertComposerPinnedBottom()
}

// assertHomeJumpToLatestChipFlow waits for chip while detached + poll refresh, taps,
// asserts follow resumes at TOP (newest-first home list).
func assertHomeJumpToLatestChipFlow() string {
	return fmt.Sprintf(`
{
  const chip = page.locator('[data-testid="jump-to-latest"]');
  let chipVisible = false;
  for (let i = 0; i < 80; i++) {
    if (await chip.isVisible().catch(() => false)) {
      chipVisible = true;
      break;
    }
    await page.waitForTimeout(250);
  }
  if (!chipVisible) {
    throw new Error('jump-to-latest chip never became visible while detached with new sessions above');
  }
  await chip.click();
  await page.waitForTimeout(200);
  const list = page.locator('[data-testid="session-list"]');
  const distance = await list.evaluate((el) => el.scrollTop);
  if (distance > %d) {
    throw new Error('jump-to-latest did not scroll session-list to top: distanceFromTop=' + distance);
  }
  const stillVisible = await chip.isVisible().catch(() => false);
  if (stillVisible) {
    throw new Error('jump-to-latest chip still visible after tap');
  }
}
`, layoutBottomThresholdPx)
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