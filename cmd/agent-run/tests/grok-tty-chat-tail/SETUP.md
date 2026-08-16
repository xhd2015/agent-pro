# Scenario

**Bug**: grok-tty web chat incomplete — producer tail ends on first sync; consumer stops on finished status

```
updates.jsonl (tool-using turn) -> [producer grok tail] -> events.jsonl
  -> WatchEvents -> web SSE / CLI --print
keep-tty: tail must survive tailState.streamed; WatchEvents until ctx done
```

## Preconditions

- Repository contains `cmd/agent-run` and `agent/llm/llm-mock/llm-mock-run-grok`.
- Session-scoped cache under `$TMPDIR/grok-tty-chat-tail-doctest-<d.DOCTEST_SESSION_ID>/`
  shares compiled binaries across parallel leaves.
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Harness reuses ACP builders and delayed-scheduler patterns from
  `cmd/agent-run/tests/grok-tty/SETUP.md` and
  `cmd/agent-run/tests/grok-discovery-race/SETUP.md`.
- Web leaves reuse SSE helpers from `cmd/agent-run/tests/enhance-chat/SETUP.md`.

## Steps

1. Root `Setup` builds session binaries and default env.
2. Grouping `Setup` narrows layer (`producer`, `consumer`, `web`) and mode.
3. Leaf `Setup` configures partial `updates.jsonl` seed, completion schedule, or finished-session fixtures.
4. `Run` executes producer CLI probe, SSE collection, or CLI `--print` follow probe.
5. Leaf `Assert` checks assistant marker, completed tool, event ordering, or SSE delivery.

## Context

- Assistant marker: `CHAT_TAIL_ASSISTANT_MARKER`
- Think seed text: `Planning tool use...`
- Tool call id: `call_chat_tail_ls`
- Default prompt: `run ls and pwd chat tail probe`
- Grok session UUID: `cccccccc-cccc-cccc-cccc-cccccccccccc`

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

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/doctest/session"
)

const (
	chatTailPrompt          = "run ls and pwd chat tail probe"
	chatTailGrokUUID        = "cccccccc-cccc-cccc-cccc-cccccccccccc"
	chatTailAssistantMarker = "CHAT_TAIL_ASSISTANT_MARKER"
	chatTailThinkText       = "Planning tool use..."
	chatTailToolCallID      = "call_chat_tail_ls"
	defaultChromeHoldSec    = 30
	producerFinishTimeout   = 60 * time.Second
	sseFinishTimeout        = 75 * time.Second
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
	req.Env = append(req.Env,
		"AGENT_RUN_HOME="+req.Home,
		"PATH="+filepath.Dir(req.AgentRun)+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	return nil
}

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "grok-tty-chat-tail-doctest-"+d.DOCTEST_SESSION_ID)
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

func acpAgentThoughtChunk(text string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "agent_thought_chunk",
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

func acpToolCall(toolCallID, kind, title string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "tool_call",
		"toolCallId":    toolCallID,
		"kind":          kind,
		"title":         title,
		"status":        "pending",
	})
	return string(line)
}

func acpToolCallUpdate(toolCallID, status, output string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "tool_call_update",
		"toolCallId":    toolCallID,
		"status":        status,
		"content": []map[string]any{
			{
				"type": "content",
				"content": map[string]any{
					"type": "text",
					"text": output,
				},
			},
		},
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
		t.Fatalf("mkdir grok session dir %s: %v", dir, err)
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

func completionAppendLines() []string {
	return []string{
		acpToolCallUpdate(chatTailToolCallID, "completed", "agent\nagents"),
		acpAgentMessageChunk(chatTailAssistantMarker),
		acpTurnCompleted(),
	}
}

func configureProducerKeepTTYEnv(t *testing.T, req *Request) {
	t.Helper()
	if req.Prompt == "" {
		req.Prompt = chatTailPrompt
	}
	if req.GrokSessionUUID == "" {
		req.GrokSessionUUID = chatTailGrokUUID
	}
	if req.GrokHome == "" {
		req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	}
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		t.Fatalf("mkdir grok home: %v", err)
	}
	hold := req.ChromeHoldSeconds
	if hold <= 0 {
		hold = defaultChromeHoldSec
	}
	stripEnvPrefix(req, "GROK_HOME=")
	stripEnvPrefix(req, "LLM_MOCK_RUN_GROK_COMMAND=")
	stripEnvPrefix(req, "AGENT_RUN_GROK_TTY_GROK_SESSION_ID=")
	req.Env = append(req.Env,
		"GROK_HOME="+req.GrokHome,
		"LLM_MOCK_RUN_GROK_COMMAND="+realLikeChromeHook(hold, req.Prompt),
		"AGENT_RUN_GROK_TTY_GROK_SESSION_ID="+req.GrokSessionUUID,
	)
	req.GrokUpdatesPath = writeFakeGrokSessionDir(t, req.GrokHome, req.TempDir, req.GrokSessionUUID, req.Prompt,
		acpAgentThoughtChunk(chatTailThinkText),
		acpToolCall(chatTailToolCallID, "execute", "ls"),
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

func analyzeChatTailEvents(events []map[string]any) Response {
	var resp Response
	assistantIdx := -1
	doneIdx := -1
	pendingToolIdx := -1
	completedToolIdx := -1
	for i, ev := range events {
		typ, _ := ev["type"].(string)
		switch typ {
		case "message":
			role, _ := ev["role"].(string)
			text, _ := ev["text"].(string)
			if role == "assistant" && strings.Contains(text, chatTailAssistantMarker) {
				resp.HasAssistantMarker = true
				assistantIdx = i
			}
		case "done":
			doneIdx = i
		case "tool_call":
			status, _ := ev["status"].(string)
			if status == "completed" {
				resp.HasCompletedTool = true
				completedToolIdx = i
			} else if pendingToolIdx < 0 {
				pendingToolIdx = i
			}
		}
	}
	if pendingToolIdx >= 0 && (assistantIdx < 0 || pendingToolIdx < assistantIdx) {
		resp.HasPendingToolFirst = true
	}
	if assistantIdx >= 0 && doneIdx >= 0 && doneIdx > assistantIdx {
		resp.DoneAfterAssistant = true
	}
	if completedToolIdx < 0 {
		for _, ev := range events {
			if ev["type"] == "tool_call" {
				blob, _ := json.Marshal(ev)
				if strings.Contains(strings.ToLower(string(blob)), "completed") {
					resp.HasCompletedTool = true
					break
				}
			}
		}
	}
	return resp
}

func parseGrokTTYSessionID(stderr string) (string, bool) {
	re := regexp.MustCompile(`grok-tty:\s*(session-\d+)`)
	m := re.FindStringSubmatch(stderr)
	if len(m) < 2 {
		return "", false
	}
	return m[1], true
}

func startGrokUpdatesScheduler(t *testing.T, req *Request, sessionReady <-chan struct{}) {
	t.Helper()
	if len(req.GrokUpdatesSchedules) == 0 {
		return
	}
	go func() {
		<-sessionReady
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

func runProducerProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SessionID == "" {
		return nil, fmt.Errorf("SessionID is required")
	}
	timeout := req.ExecTimeout
	if timeout <= 0 {
		timeout = producerFinishTimeout
	}
	args := []string{
		"run",
		"--agent-runner", "grok-tty",
		"--keep-tty",
		"--session", req.SessionID,
		"--agent-runner-binary", req.LLMMockRunGrok,
		"--agent-runner-config-home", req.GrokHome,
		req.Prompt,
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf
	start := time.Now()
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	sessionReady := make(chan struct{})
	go func() {
		deadline := time.Now().Add(15 * time.Second)
		for time.Now().Before(deadline) {
			if _, ok := parseGrokTTYSessionID(stderrBuf.String()); ok {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		close(sessionReady)
	}()
	startGrokUpdatesScheduler(t, req, sessionReady)

	eventsPath := eventsJSONLPath(req.Home, req.SessionID)
	var lastLines []string
	var analysis Response
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	pollDeadline := time.Now().Add(timeout)
	for time.Now().Before(pollDeadline) {
		if lines, err := readEventsJSONL(eventsPath); err == nil && len(lines) > 0 {
			lastLines = lines
			parsed := parseEventLines(t, lines)
			analysis = analyzeChatTailEvents(parsed)
			if analysis.HasAssistantMarker && analysis.HasCompletedTool && analysis.DoneAfterAssistant {
				goto finish
			}
		}
		select {
		case waitErr := <-done:
			if lines, err := readEventsJSONL(eventsPath); err == nil {
				lastLines = lines
				analysis = analyzeChatTailEvents(parseEventLines(t, lines))
			}
			return buildProducerResponse(t, stdoutBuf, stderrBuf, start, eventsPath, lastLines, analysis, waitErr), nil
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

finish:
	_ = cmd.Process.Kill()
	waitErr := <-done
	if lines, err := readEventsJSONL(eventsPath); err == nil {
		lastLines = lines
		analysis = analyzeChatTailEvents(parseEventLines(t, lines))
	}
	return buildProducerResponse(t, stdoutBuf, stderrBuf, start, eventsPath, lastLines, analysis, waitErr), nil
}

func buildProducerResponse(t *testing.T, stdoutBuf, stderrBuf *bytes.Buffer, start time.Time, eventsPath string, lines []string, analysis Response, waitErr error) *Response {
	resp := &Response{
		Stdout:              stdoutBuf.String(),
		Stderr:              stderrBuf.String(),
		Elapsed:             time.Since(start),
		EventsFilePath:      eventsPath,
		EventsFileLines:     lines,
		HasAssistantMarker:  analysis.HasAssistantMarker,
		HasCompletedTool:    analysis.HasCompletedTool,
		HasPendingToolFirst: analysis.HasPendingToolFirst,
		DoneAfterAssistant:  analysis.DoneAfterAssistant,
		Err:                 waitErr,
	}
	if len(lines) > 0 {
		resp.EventsParsed = parseEventLines(t, lines)
	}
	if waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			resp.ExitCode = exitErr.ExitCode()
		}
	}
	return resp
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

func runWebDelayedAssistantSSE(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SessionID == "" {
		return nil, fmt.Errorf("sse probe requires SessionID")
	}
	delay := req.CompletionDelay
	if delay <= 0 {
		delay = 1200 * time.Millisecond
	}
	sessionReady := make(chan struct{})
	close(sessionReady)
	req.GrokUpdatesSchedules = []GrokUpdatesSchedule{{
		Delay: delay,
		Lines: completionAppendLines(),
	}}
	startGrokUpdatesScheduler(t, req, sessionReady)

	maxWait := req.SSEMaxWait
	if maxWait <= 0 {
		maxWait = sseFinishTimeout
	}
	events := collectSSESessionEvents(t, req, req.Runner, req.SessionID, req.SSEAfterOffset, maxWait)
	path, lines, _ := readSessionEventsOptional(eventsJSONLPath(req.Home, req.SessionID))
	analysis := analyzeChatTailEvents(parseEventLines(t, lines))
	return &Response{
		SSEEvents:          events,
		EventsFilePath:     path,
		EventsFileLines:    lines,
		EventsParsed:       parseEventLines(t, lines),
		HasAssistantMarker: analysis.HasAssistantMarker || sseHasAssistantMarker(events),
		HasCompletedTool:   analysis.HasCompletedTool,
		DoneAfterAssistant: analysis.DoneAfterAssistant,
	}, nil
}

func sseHasAssistantMarker(events []map[string]any) bool {
	for _, ev := range events {
		if ev["type"] == "message" && ev["role"] == "assistant" {
			text, _ := ev["text"].(string)
			if strings.Contains(text, chatTailAssistantMarker) {
				return true
			}
		}
	}
	return false
}

func readSessionEventsOptional(path string) (string, []string, error) {
	lines, err := readEventsJSONL(path)
	if err != nil {
		if os.IsNotExist(err) {
			return path, nil, nil
		}
		return path, nil, err
	}
	return path, lines, nil
}

func openAgentStore(t *testing.T, req *Request) agentstorage.Store {
	t.Helper()
	store, err := agentstorage.NewFileStore(req.Home)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return store
}

func seedRunningSessionForFollow(t *testing.T, req *Request, runner, sessionID string) {
	t.Helper()
	store := openAgentStore(t, req)
	meta := agentstorage.SessionMeta{
		Runner:    runner,
		SessionID: sessionID,
		Status:    "running",
		Workspace: req.TempDir,
	}
	if err := store.CreateSession(sessionID, meta); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if err := store.AppendEvent(sessionID, types.AgentEvent{
		Type: types.ActionMessage,
		Role: "assistant",
		Text: "Initial running event",
	}); err != nil {
		t.Fatalf("append event: %v", err)
	}
}

func markSessionFinished(t *testing.T, req *Request, runner, sessionID string) {
	t.Helper()
	store := openAgentStore(t, req)
	if err := store.UpdateSessionStatus(sessionID, "finished"); err != nil {
		t.Fatalf("UpdateSessionStatus finished: %v", err)
	}
}

func seedFinishedSession(t *testing.T, req *Request, runner, sessionID, firstText string) {
	t.Helper()
	store := openAgentStore(t, req)
	meta := agentstorage.SessionMeta{
		Runner:    runner,
		SessionID: sessionID,
		Status:    "finished",
		Workspace: req.TempDir,
	}
	if err := store.CreateSession(sessionID, meta); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if err := store.UpdateSessionStatus(sessionID, "finished"); err != nil {
		t.Fatalf("UpdateSessionStatus: %v", err)
	}
	if err := store.AppendEvent(sessionID, types.AgentEvent{
		Type: types.ActionMessage,
		Role: "assistant",
		Text: firstText,
	}); err != nil {
		t.Fatalf("append event: %v", err)
	}
}

func appendSessionEvent(t *testing.T, req *Request, runner, sessionID, text string) {
	t.Helper()
	store := openAgentStore(t, req)
	if err := store.AppendEvent(sessionID, types.AgentEvent{
		Type: types.ActionMessage,
		Role: "assistant",
		Text: text,
	}); err != nil {
		t.Fatalf("append event: %v", err)
	}
}

func runSSEFinishedAppendProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	maxWait := req.SSEMaxWait
	if maxWait <= 0 {
		maxWait = 12 * time.Second
	}
	done := make(chan []map[string]any, 1)
	go func() {
		done <- collectSSESessionEvents(t, req, req.Runner, req.SessionID, req.SSEAfterOffset, maxWait)
	}()
	if req.Sidecar != nil {
		req.Sidecar()
	}
	events := <-done
	return &Response{
		SSEEvents:          events,
		HasAssistantMarker: sseHasAssistantMarker(events),
	}, nil
}

func runCLIFollowProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	timeout := req.ExecTimeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.AgentRun, req.CLIArgs...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if req.Sidecar != nil {
		go req.Sidecar()
	}
	err := cmd.Run()
	resp := &Response{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			resp.ExitCode = exitErr.ExitCode()
			if ctx.Err() == context.DeadlineExceeded {
				return resp, nil
			}
		}
		if ctx.Err() == context.DeadlineExceeded {
			return resp, nil
		}
		return resp, err
	}
	return resp, nil
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