# Scenario

**Bug**: grok-tty `--keep-tty` discovery race — PTY chrome ends wait early and cancels `DiscoverSession`

```
llm-mock-run-grok + real-like chrome + --keep-tty
  -> waitForPersistentTurnRemote must not complete on scrollback alone
  -> DiscoverSession polls until updates.jsonl streams or real timeout
  -> events.jsonl is the assertion surface
```

## Preconditions

- Repository contains `cmd/agent-run` and `agent/llm/llm-mock/llm-mock-run-grok`.
- Session-scoped cache under `$TMPDIR/grok-discovery-race-doctest-<d.DOCTEST_SESSION_ID>/`
  shares compiled binaries across parallel leaves.
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Harness reuses patterns from `cmd/agent-run/tests/grok-tty/SETUP.md`
  (`delayedGrokSessionSchedule`, `fakeTUIHoldSeconds` timing model) and
  `cmd/agent-run/tests/enhance-chat/SETUP.md` (`LLM_MOCK_RUN_GROK_COMMAND` chrome hook).
- Reference repro: `script/debug/grok-tty-discovery-cancel/main.go`.

## Steps

1. Root `Setup` builds session binaries and default env.
2. Grouping `keep-tty/Setup` sets `--keep-tty` CLI flags and llm-mock binary wiring.
3. Leaf `Setup` configures delayed session schedule or empty-home chrome failure.
4. `Run` starts `agent-run run`, polls `events.jsonl`, records think→error timing.
5. Leaf `Assert` checks marker streaming or extended discovery window.

## Context

- Resolve think text: `Resolve session id...`
- Resolve error prefix: `Cannot resolve session id:`
- Delayed marker: `DELAYED_SESSION_MARKER`
- Discovery minimum poll window (chrome false-complete guard): 3s
- Early cancel guard window (delayed success): 10s

```go
import (
	"runtime"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

const (
	resolveThinkText          = "Resolve session id..."
	resolveErrorPrefix        = "Cannot resolve session id:"
	delayedSessionMarker      = "DELAYED_SESSION_MARKER"
	delayedSessionGrokUUID    = "dddddddd-dddd-dddd-dddd-dddddddddddd"
	chromeWaitPrompt          = "one word of France capital"
	discoveryMinWindow        = 3 * time.Second
	earlyCancelGuardWindow    = 10 * time.Second
	defaultChromeHoldSeconds  = 30
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.AgentRun, req.LLMMockRunGrok = ensureSessionBinaries(t, d, req.RepoRoot)
	req.Env = append(req.Env,
		"AGENT_RUN_HOME="+req.Home,
		"PATH="+filepath.Dir(req.AgentRun)+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	return nil
}

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "grok-discovery-race-doctest-"+d.DOCTEST_SESSION_ID)
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
			cmd := exec.Command(runtime.GOROOT()+"/bin/go", b.args...)
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

// realLikeChromeHook mirrors script/debug/grok-tty-discovery-cancel/main.go — PTY chrome
// that can falsely satisfy persistentTurnComplete when keep-tty scrollback polling runs.
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

// delayedGrokSessionSchedule mirrors cmd/agent-run/tests/grok-tty/SETUP.md — session dir
// appears after delay while the PTY chrome hook stays alive.
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

func configureLLMMockChromeEnv(t *testing.T, req *Request, prompt string, holdSec int) {
	t.Helper()
	if holdSec <= 0 {
		holdSec = defaultChromeHoldSeconds
	}
	if prompt == "" {
		prompt = chromeWaitPrompt
	}
	stripEnvPrefix(req, "LLM_MOCK_RUN_GROK_COMMAND=")
	req.Env = append(req.Env, "LLM_MOCK_RUN_GROK_COMMAND="+realLikeChromeHook(holdSec, prompt))
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
		}
	}()
}

func analyzeEvents(events []map[string]any) Response {
	var resp Response
	for _, ev := range events {
		typ, _ := ev["type"].(string)
		text, _ := ev["text"].(string)
		role, _ := ev["role"].(string)
		ts, _ := ev["timestamp"].(float64)
		tsMilli := int64(ts)

		switch typ {
		case "think":
			if strings.Contains(text, resolveThinkText) && resp.ThinkTimestamp == 0 {
				resp.ThinkTimestamp = tsMilli
			}
		case "error":
			if strings.HasPrefix(text, resolveErrorPrefix) {
				if resp.ErrorTimestamp == 0 {
					resp.ErrorTimestamp = tsMilli
				}
				if strings.Contains(text, "context canceled") {
					resp.HasContextCancel = true
					if resp.ThinkTimestamp > 0 && tsMilli > 0 {
						gap := time.Duration(tsMilli-resp.ThinkTimestamp) * time.Millisecond
						if gap < earlyCancelGuardWindow {
							resp.EarlyContextCancel = true
						}
					}
				}
			}
		case "message":
			if role == "assistant" && strings.Contains(text, delayedSessionMarker) {
				resp.HasDelayedMarker = true
			}
		}
	}
	if resp.ThinkTimestamp > 0 && resp.ErrorTimestamp > 0 {
		resp.ThinkToErrorGap = time.Duration(resp.ErrorTimestamp-resp.ThinkTimestamp) * time.Millisecond
	}
	return resp
}

func runKeepTTYEventsProbe(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SessionID == "" {
		return nil, fmt.Errorf("SessionID is required")
	}
	if req.Prompt == "" {
		req.Prompt = chromeWaitPrompt
	}
	timeout := req.ExecTimeout
	if timeout <= 0 {
		timeout = 45 * time.Second
	}

	args := []string{
		"run",
		"--agent-runner", "grok-tty",
		"--keep-tty",
		"--session", req.SessionID,
		"--agent-runner-binary", req.LLMMockRunGrok,
	}
	if req.GrokHome != "" {
		args = append(args, "--agent-runner-config-home", req.GrokHome)
	}
	args = append(args, req.Prompt)

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
			analysis = analyzeEvents(parseEventLines(t, lines))
			switch req.Scenario {
			case "delayed-session-streams":
				if analysis.HasDelayedMarker {
					goto finish
				}
				if analysis.EarlyContextCancel {
					goto finish
				}
			case "chrome-wait-exceeds-discovery-window":
				if analysis.ErrorTimestamp > 0 {
					goto finish
				}
			}
		}
		select {
		case waitErr := <-done:
			if lines, err := readEventsJSONL(eventsPath); err == nil {
				lastLines = lines
				analysis = analyzeEvents(parseEventLines(t, lines))
			}
			resp := &Response{
				Stdout:           stdoutBuf.String(),
				Stderr:           stderrBuf.String(),
				Elapsed:          time.Since(start),
				EventsFilePath:   eventsPath,
				EventsFileLines:  lastLines,
				EventsParsed:     parseEventLines(t, lastLines),
				ThinkTimestamp:   analysis.ThinkTimestamp,
				ErrorTimestamp:   analysis.ErrorTimestamp,
				ThinkToErrorGap:  analysis.ThinkToErrorGap,
				HasDelayedMarker: analysis.HasDelayedMarker,
				HasContextCancel: analysis.HasContextCancel,
				EarlyContextCancel: analysis.EarlyContextCancel,
				Err:              waitErr,
			}
			if waitErr != nil {
				var exitErr *exec.ExitError
				if errors.As(waitErr, &exitErr) {
					resp.ExitCode = exitErr.ExitCode()
				}
			}
			return resp, nil
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

finish:
	_ = cmd.Process.Kill()
	waitErr := <-done
	if lines, err := readEventsJSONL(eventsPath); err == nil {
		lastLines = lines
		analysis = analyzeEvents(parseEventLines(t, lines))
	}

	resp := &Response{
		Stdout:             stdoutBuf.String(),
		Stderr:             stderrBuf.String(),
		Elapsed:            time.Since(start),
		EventsFilePath:     eventsPath,
		EventsFileLines:    lastLines,
		EventsParsed:       parseEventLines(t, lastLines),
		ThinkTimestamp:     analysis.ThinkTimestamp,
		ErrorTimestamp:     analysis.ErrorTimestamp,
		ThinkToErrorGap:    analysis.ThinkToErrorGap,
		HasDelayedMarker:   analysis.HasDelayedMarker,
		HasContextCancel:   analysis.HasContextCancel,
		EarlyContextCancel: analysis.EarlyContextCancel,
		Err:                waitErr,
	}
	if waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			resp.ExitCode = exitErr.ExitCode()
		}
	}
	return resp, nil
}

func eventsHaveThinkText(events []map[string]any, text string) bool {
	for _, ev := range events {
		if ev["type"] == "think" && ev["text"] == text {
			return true
		}
	}
	return false
}

func eventsHaveErrorPrefix(events []map[string]any, prefix string) bool {
	for _, ev := range events {
		if ev["type"] != "error" {
			continue
		}
		text, _ := ev["text"].(string)
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return false
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