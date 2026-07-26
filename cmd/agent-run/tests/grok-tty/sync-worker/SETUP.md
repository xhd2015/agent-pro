# Scenario

**Bug**: web rapid follow-ups spawn overlapping grok tails → duplicate events.jsonl rows

```
agent-run web + llm-mock grok-tty keep-tty
  -> POST session (prompt A)
  -> scheduled updates.jsonl turn 1 completion
  -> wait done in events.jsonl
  -> POST follow-up B within overlap window
  -> scheduled turn 2 completion
  -> events.jsonl: one user line per prompt
```

## Preconditions

- Repository contains `cmd/agent-run` and `agent/llm/llm-mock/llm-mock-run-grok`.
- Session-scoped cache: `$TMPDIR/grok-tty-sync-worker-doctest-<d.DOCTEST_SESSION_ID>/`.
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Reuses ACP builders and web helpers (pattern from `grok-tty-chat-tail`).

## Steps

1. Root `Setup` builds session binaries and default env.
2. Grouping `Setup` narrows transport (`web`).
3. Leaf `Setup` configures prompts, completion schedules, follow-up gap.
4. `Run` drives web POST + follow-up + polls `events.jsonl`.
5. Leaf `Assert` counts user messages per prompt text.

## Context

- Default prompts mirror reproduced bug: `hello?` then `what did I say?`.
- `FollowUpGap` default 2s (within 90s tail overlap; repro was 17s).
- File-level assertion on `events.jsonl` — not Playwright/SSE.

```go
import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentsync"
	"github.com/xhd2015/doctest/session"
)

const (
	defaultPromptA       = "hello?"
	defaultPromptB       = "what did I say?"
	defaultReplyA        = "web-sync-reply-hello"
	defaultReplyB        = "web-sync-reply-recall"
	defaultGrokUUID      = "dddddddd-dddd-dddd-dddd-dddddddddddd"
	defaultChromeHold  = 45
	defaultProbeTimeout = 90 * time.Second
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.AgentRun, req.LLMMockRunGrok = ensureSessionBinaries(t, d, req.RepoRoot)
	req.Runner = "grok-tty"
	req.WebToken = "test"
	if req.PromptA == "" {
		req.PromptA = defaultPromptA
	}
	if req.PromptB == "" {
		req.PromptB = defaultPromptB
	}
	if req.ReplyA == "" {
		req.ReplyA = defaultReplyA
	}
	if req.ReplyB == "" {
		req.ReplyB = defaultReplyB
	}
	if req.GrokSessionUUID == "" {
		req.GrokSessionUUID = defaultGrokUUID
	}
	if req.FollowUpGap <= 0 {
		req.FollowUpGap = 2 * time.Second
	}
	if req.ProbeTimeout <= 0 {
		req.ProbeTimeout = defaultProbeTimeout
	}
	req.Env = append(req.Env,
		"AGENT_RUN_HOME="+req.Home,
		"PATH="+filepath.Dir(req.AgentRun)+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	return nil
}

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "grok-tty-sync-worker-doctest-"+d.DOCTEST_SESSION_ID)
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
			cmd := exec.Command("go", b.args...)
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

func realLikeChromeHook(holdSec int, prompt string) string {
	return fmt.Sprintf(`sh -c 'printf "⎇ master worktree ~/.wrk +1\n#1 %s\n⠦ Starting session… 0.9s\n╭----------------------------------------------------------------------------╮\n│ ❯                                                                          │\n╰---------------------------------------------- Grok Build · always-approve -╯\nShift+Tab:mode  │  Ctrl+;:queue  │  Ctrl+.:shortcuts\n"; sleep %d'`, prompt, holdSec)
}

func encodedGrokCwd(workspace string) string {
	abs, err := filepath.Abs(workspace)
	if err != nil {
		abs = workspace
	}
	return url.PathEscape(abs)
}

func grokSessionDir(grokHome, workspace, sessionUUID string) string {
	return filepath.Join(grokHome, "sessions", encodedGrokCwd(workspace), sessionUUID)
}

func grokSummaryJSON(workspace, sessionUUID string) string {
	abs, _ := filepath.Abs(workspace)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	payload := map[string]any{
		"info": map[string]any{
			"cwd":       abs,
			"sessionId": sessionUUID,
			"openedAt":  now,
		},
		"created_at": now,
	}
	b, _ := json.Marshal(payload)
	return string(b)
}

func acpUserMessageChunk(text string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "user_message_chunk",
		"content":       map[string]any{"type": "text", "text": text},
	})
	return string(line)
}

func acpAgentMessageChunk(text string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "agent_message_chunk",
		"content":       map[string]any{"type": "text", "text": text},
	})
	return string(line)
}

func acpTurnCompleted() string {
	return `{"sessionUpdate":"turn_completed"}`
}

func appendUpdatesJSONL(path string, lines ...string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if _, err := fmt.Fprintln(f, line); err != nil {
			return err
		}
	}
	return nil
}

func writeFakeGrokSessionDir(t *testing.T, grokHome, workspace, sessionUUID, prompt string, initialLines ...string) string {
	t.Helper()
	dir := grokSessionDir(grokHome, workspace, sessionUUID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir grok session dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), []byte(grokSummaryJSON(workspace, sessionUUID)), 0644); err != nil {
		t.Fatalf("write summary.json: %v", err)
	}
	updatesPath := filepath.Join(dir, "updates.jsonl")
	seed := []string{acpUserMessageChunk(prompt)}
	seed = append(seed, initialLines...)
	if err := appendUpdatesJSONL(updatesPath, seed...); err != nil {
		t.Fatalf("seed updates.jsonl: %v", err)
	}
	return updatesPath
}

func turnOneCompletionLines(reply string) []string {
	return []string{
		acpAgentMessageChunk(reply),
		acpTurnCompleted(),
	}
}

func turnTwoCompletionLines(userPrompt, reply string) []string {
	return []string{
		acpUserMessageChunk(userPrompt),
		acpAgentMessageChunk(reply),
		acpTurnCompleted(),
	}
}

func configureWebGrokEnv(t *testing.T, req *Request) {
	t.Helper()
	if req.GrokHome == "" {
		req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	}
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		t.Fatalf("mkdir grok home: %v", err)
	}
	hold := req.ChromeHoldSeconds
	if hold <= 0 {
		hold = defaultChromeHold
	}
	stripEnvPrefix(req, "GROK_HOME=")
	stripEnvPrefix(req, "LLM_MOCK_RUN_GROK_COMMAND=")
	stripEnvPrefix(req, "AGENT_RUN_GROK_TTY_GROK_SESSION_ID=")
	req.Env = append(req.Env,
		"GROK_HOME="+req.GrokHome,
		"LLM_MOCK_RUN_GROK_COMMAND="+realLikeChromeHook(hold, req.PromptA),
		"AGENT_RUN_GROK_TTY_GROK_SESSION_ID="+req.GrokSessionUUID,
	)
	req.GrokUpdatesPath = writeFakeGrokSessionDir(t, req.GrokHome, req.TempDir, req.GrokSessionUUID, req.PromptA)
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

func eventsJSONLPath(home, sessionID string) string {
	return filepath.Join(home, "sessions", "grok-tty", sessionID, "events.jsonl")
}

func readEventsJSONL(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
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

func countUserMessages(events []map[string]any, text string) int {
	count := 0
	for _, ev := range events {
		if ev["type"] == "message" && ev["role"] == "user" {
			msg, _ := ev["text"].(string)
			if msg == text {
				count++
			}
		}
	}
	return count
}

func countDoneEvents(events []map[string]any) int {
	count := 0
	for _, ev := range events {
		if ev["type"] == "done" {
			count++
		}
	}
	return count
}

func waitForDoneCount(t *testing.T, home, sessionID string, want int, timeout time.Duration) []string {
	t.Helper()
	path := eventsJSONLPath(home, sessionID)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		lines, err := readEventsJSONL(path)
		if err == nil {
			events := parseEventLines(t, lines)
			if countDoneEvents(events) >= want {
				return lines
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	lines, _ := readEventsJSONL(path)
	t.Fatalf("timeout waiting for %d done events in %s; got %d\n%s",
		want, path, countDoneEvents(parseEventLines(t, lines)), strings.Join(lines, "\n"))
	return nil
}

func startAgentRunWeb(t *testing.T, req *Request) {
	t.Helper()
	args := []string{
		"web", "--no-open", "--port", "0", "--token", req.WebToken,
		"--agent-runner", "grok-tty",
		"--grok-home", req.GrokHome,
		"--grok-tty-runner-binary", req.LLMMockRunGrok,
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

func postFollowUpMessage(t *testing.T, req *Request, sessionID, message string) (int, string) {
	t.Helper()
	payload, err := json.Marshal(map[string]string{"message": message})
	if err != nil {
		t.Fatalf("marshal follow-up: %v", err)
	}
	path := fmt.Sprintf("%s/api/agent-run/sessions/%s/%s/messages", req.WebBaseURL, req.Runner, sessionID)
	return doHTTP(t, http.MethodPost, path, req.WebToken, "application/json", string(payload))
}

func startGrokUpdatesScheduler(t *testing.T, req *Request) {
	t.Helper()
	if len(req.GrokUpdatesSchedules) == 0 {
		return
	}
	go func() {
		for _, sched := range req.GrokUpdatesSchedules {
			sched := sched
			time.Sleep(sched.Delay)
			if sched.OnFire != nil {
				sched.OnFire()
			}
			path := sched.UpdatesPath
			if path == "" {
				path = req.GrokUpdatesPath
			}
			if path != "" && len(sched.Lines) > 0 {
				if err := appendUpdatesJSONL(path, sched.Lines...); err != nil {
					t.Logf("append updates.jsonl: %v", err)
				}
			}
		}
	}()
}

func delayedGrokSessionSchedule(t *testing.T, delay time.Duration, grokHome, workspace, sessionUUID, prompt string, initialLines ...string) (GrokUpdatesSchedule, string) {
	t.Helper()
	updatesPath := filepath.Join(grokSessionDir(grokHome, workspace, sessionUUID), "updates.jsonl")
	return GrokUpdatesSchedule{
		Delay: delay,
		OnFire: func() {
			writeFakeGrokSessionDir(t, grokHome, workspace, sessionUUID, prompt, initialLines...)
		},
	}, updatesPath
}

func readSessionMetaRunnerSessionID(t *testing.T, home, sessionID string) string {
	t.Helper()
	path := filepath.Join(home, "sessions", "grok-tty", sessionID, "meta.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var meta map[string]any
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("parse meta.json: %v", err)
	}
	id, _ := meta["runner_session_id"].(string)
	return strings.TrimSpace(id)
}

func waitForUserAndAssistant(t *testing.T, home, sessionID, prompt, assistantSubstring string, timeout time.Duration) ([]string, []map[string]any) {
	t.Helper()
	path := eventsJSONLPath(home, sessionID)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		lines, err := readEventsJSONL(path)
		if err == nil {
			events := parseEventLines(t, lines)
			hasUser := countUserMessages(events, prompt) >= 1
			hasAssistant := false
			for _, ev := range events {
				if ev["type"] == "message" && ev["role"] == "assistant" {
					text, _ := ev["text"].(string)
					if strings.Contains(text, assistantSubstring) {
						hasAssistant = true
						break
					}
				}
			}
			if hasUser && hasAssistant {
				return lines, events
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	lines, _ := readEventsJSONL(path)
	t.Fatalf("timeout waiting for user+assistant in %s; prompt=%q\n%s", path, prompt, strings.Join(lines, "\n"))
	return nil, nil
}

func seedFinishedSessionEmptyEvents(t *testing.T, req *Request) {
	t.Helper()
	sessionDir := filepath.Join(req.Home, "sessions", req.Runner, req.SessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("mkdir session dir: %v", err)
	}
	absWorkspace, err := filepath.Abs(req.TempDir)
	if err != nil {
		absWorkspace = req.TempDir
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	meta := map[string]any{
		"runner":              req.Runner,
		"session_id":          req.SessionID,
		"status":              "finished",
		"initial_prompt":      req.PromptA,
		"runner_session_id":   req.GrokSessionUUID,
		"workspace":           absWorkspace,
		"created_at":          now,
		"updated_at":          now,
	}
	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal meta.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sessionDir, "meta.json"), data, 0644); err != nil {
		t.Fatalf("write meta.json: %v", err)
	}
	if err := os.Remove(eventsJSONLPath(req.Home, req.SessionID)); err != nil && !os.IsNotExist(err) {
		t.Fatalf("remove events.jsonl: %v", err)
	}
}

func getSessionDetail(t *testing.T, req *Request) (int, string) {
	t.Helper()
	path := fmt.Sprintf("%s/api/agent-run/sessions/%s/%s", req.WebBaseURL, req.Runner, req.SessionID)
	return doHTTP(t, http.MethodGet, path, req.WebToken, "", "")
}

func runWebOpenSessionStartsSync(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.GrokHome == "" {
		req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	}
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		return nil, err
	}
	if req.GrokSessionUUID == "" {
		req.GrokSessionUUID = "22222222-2222-2222-2222-222222222222"
	}
	if req.SessionID == "" {
		req.SessionID = "web_open_sync_seed"
	}
	stripEnvPrefix(req, "GROK_HOME=")
	stripEnvPrefix(req, "AGENT_RUN_GROK_TTY_GROK_SESSION_ID=")
	req.Env = append(req.Env, "GROK_HOME="+req.GrokHome)

	req.GrokUpdatesPath = writeFakeGrokSessionDir(t, req.GrokHome, req.TempDir, req.GrokSessionUUID, req.PromptA,
		turnOneCompletionLines(req.ReplyA)...,
	)
	seedFinishedSessionEmptyEvents(t, req)

	startAgentRunWeb(t, req)

	status, body := getSessionDetail(t, req)
	if status != http.StatusOK {
		return &Response{GetSessionStatus: status}, fmt.Errorf("GET session detail: status=%d body=%q", status, body)
	}

	lines, events := waitForUserAndAssistant(t, req.Home, req.SessionID, req.PromptA, req.ReplyA, req.ProbeTimeout)
	assistantFound := false
	for _, ev := range events {
		if ev["type"] == "message" && ev["role"] == "assistant" {
			text, _ := ev["text"].(string)
			if strings.Contains(text, req.ReplyA) {
				assistantFound = true
				break
			}
		}
	}
	return &Response{
		EventsFilePath:   eventsJSONLPath(req.Home, req.SessionID),
		EventsFileLines:  lines,
		EventsParsed:     events,
		UserCount:        countUserMessages(events, req.PromptA),
		AssistantFound:   assistantFound,
		RunnerSessionID:  readSessionMetaRunnerSessionID(t, req.Home, req.SessionID),
		WorkerActive:     agentsync.GrokSyncWorkerActive(req.Runner, req.SessionID),
		GetSessionStatus: status,
		DoneCount:        countDoneEvents(events),
	}, nil
}

func runWebCreateSessionEvents(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.GrokHome == "" {
		req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	}
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		return nil, err
	}
	stripEnvPrefix(req, "GROK_HOME=")
	stripEnvPrefix(req, "LLM_MOCK_RUN_GROK_COMMAND=")
	stripEnvPrefix(req, "AGENT_RUN_GROK_TTY_GROK_SESSION_ID=")
	req.Env = append(req.Env, "GROK_HOME="+req.GrokHome)

	hold := req.ChromeHoldSeconds
	if hold <= 0 {
		hold = 60
	}
	req.Env = append(req.Env, "LLM_MOCK_RUN_GROK_COMMAND="+realLikeChromeHook(hold, req.PromptA))

	delay := req.CompletionDelayTurn1
	if delay <= 0 {
		delay = 2 * time.Second
	}
	grokUUID := req.GrokSessionUUID
	if grokUUID == "" {
		grokUUID = "11111111-1111-1111-1111-111111111111"
	}
	sched, _ := delayedGrokSessionSchedule(t, delay, req.GrokHome, req.TempDir, grokUUID, req.PromptA,
		turnOneCompletionLines(req.ReplyA)...,
	)
	req.GrokUpdatesSchedules = []GrokUpdatesSchedule{sched}
	startGrokUpdatesScheduler(t, req)

	startAgentRunWeb(t, req)
	req.SessionID = postCreateSession(t, req, req.Runner, req.PromptA)

	lines, events := waitForUserAndAssistant(t, req.Home, req.SessionID, req.PromptA, req.ReplyA, req.ProbeTimeout)
	assistantFound := false
	for _, ev := range events {
		if ev["type"] == "message" && ev["role"] == "assistant" {
			text, _ := ev["text"].(string)
			if strings.Contains(text, req.ReplyA) {
				assistantFound = true
				break
			}
		}
	}
	return &Response{
		EventsFilePath:  eventsJSONLPath(req.Home, req.SessionID),
		EventsFileLines: lines,
		EventsParsed:    events,
		UserCount:       countUserMessages(events, req.PromptA),
		AssistantFound:  assistantFound,
		RunnerSessionID: readSessionMetaRunnerSessionID(t, req.Home, req.SessionID),
		DoneCount:       countDoneEvents(events),
	}, nil
}

func runWebRapidFollowups(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	configureWebGrokEnv(t, req)
	startAgentRunWeb(t, req)

	delay1 := req.CompletionDelayTurn1
	if delay1 <= 0 {
		delay1 = 1500 * time.Millisecond
	}
	delay2 := req.CompletionDelayTurn2
	if delay2 <= 0 {
		delay2 = 1500 * time.Millisecond
	}

	req.GrokUpdatesSchedules = []GrokUpdatesSchedule{
		{Delay: delay1, Lines: turnOneCompletionLines(req.ReplyA)},
	}
	startGrokUpdatesScheduler(t, req)

	req.SessionID = postCreateSession(t, req, req.Runner, req.PromptA)

	waitForDoneCount(t, req.Home, req.SessionID, 1, req.ProbeTimeout)

	time.Sleep(req.FollowUpGap)

	statusB, bodyB := postFollowUpMessage(t, req, req.SessionID, req.PromptB)
	if statusB != http.StatusAccepted && statusB != http.StatusOK {
		return &Response{FollowUpBStatus: statusB, FollowUpBBody: bodyB},
			fmt.Errorf("follow-up B status=%d body=%q", statusB, bodyB)
	}

	req.GrokUpdatesSchedules = []GrokUpdatesSchedule{
		{Delay: delay2, Lines: turnTwoCompletionLines(req.PromptB, req.ReplyB)},
	}
	startGrokUpdatesScheduler(t, req)

	waitForDoneCount(t, req.Home, req.SessionID, 2, req.ProbeTimeout)

	path := eventsJSONLPath(req.Home, req.SessionID)
	lines, err := readEventsJSONL(path)
	if err != nil {
		return nil, err
	}
	events := parseEventLines(t, lines)
	return &Response{
		EventsFilePath:  path,
		EventsFileLines: lines,
		EventsParsed:    events,
		UserCountA:      countUserMessages(events, req.PromptA),
		UserCountB:      countUserMessages(events, req.PromptB),
		DoneCount:       countDoneEvents(events),
		FollowUpBStatus: statusB,
		FollowUpBBody:   bodyB,
	}, nil
}

func parseGrokTTYSessionID(stderr string) (string, bool) {
	re := regexp.MustCompile(`grok-tty:\s*(session-\d+)`)
	m := re.FindStringSubmatch(stderr)
	if len(m) < 2 {
		return "", false
	}
	return m[1], true
}

func drainLines(r io.Reader) []string {
	var lines []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}
```
