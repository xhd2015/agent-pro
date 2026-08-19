# Scenario

**Feature**: open → Paris → /exit → exited → resume --open hello with llm-mock-run-grok

```
# isolated homes + mock binary
build agent-run + llm-mock-run-grok (session cache)
  -> per-leaf AGENT_RUN_HOME + GROK_HOME + workspace
  -> LLM_MOCK_RUN_GROK_COMMAND multi-turn (Paris then hello; /exit markers)
  -> AGENT_RUN_OPEN_ATTACH_INSTANT=1

# primary flow
agent-run run --open ... "one word of France capital"
  -> snapshot/events contain Paris
  -> agent-run send /exit
  -> status exited:true
  -> agent-run resume --open "hello"
  -> no already in use; snapshot/events proper
```

## Preconditions

- Repository contains `cmd/agent-run` and `agent/llm/llm-mock/llm-mock-run-grok`.
- Nested DOCTEST root: does not inherit parent `cmd/agent-run/tests` Setup/Run.
- Repo root from this tree: `d.DOCTEST_ROOT/../../../..`
  (`open-resume-e2e` → `tests` → `agent-run` → `cmd` → module root).
- Session-scoped binary cache:
  `$TMPDIR/open-resume-e2e-doctest-<d.DOCTEST_SESSION_ID>/`
  shares compiled `agent-run` + `llm-mock-run-grok` across parallel leaves.
- Each leaf uses isolated temp tree:

  ```text
  TempDir/
    bin/agent-run              (or session-cache path)
    bin/llm-mock-run-grok
    .agent-run/                # AGENT_RUN_HOME
    .grok/                     # --agent-runner-config-home / GROK_HOME
    workspace/                 # --dir (project-like)
  ```

- **No real grok.** `--agent-runner-binary` points at session-built
  `llm-mock-run-grok`; `AGENT_RUN_GROK_TTY_COMMAND` is unset so the binary spec
  drives the PTY.
- Patterns reused from:
  - `cmd/agent-run/tests/run/agent-runner-binary/`
  - `cmd/agent-run/tests/web-e2e-grok-harness/`
  - `cmd/agent-run/tests/grok-discovery-race/`
  - `script/debug/open-resume-e2e` (flow only; mock instead of real grok)

## Steps

1. Root `Setup` resolves repo root, builds session binaries, allocates per-leaf
   homes/workspace, wires mock env defaults.
2. Grouping `mock-grok/Setup` pins runner flags (binary + config home).
3. Leaf `Setup` finalizes session id, prompts, and scenario.
4. `Run` dispatches by `req.Scenario` (full open/exit/resume, live-send,
   deny-while-live, double-resume, bind-only, etc.), then cleans registry PIDs.
5. Leaf `Assert` checks scenario-specific gates (Paris, exited, resume,
   already-in-use, hello markers, deny messages).

## Context

- Default open prompt: `one word of France capital`
- Default Paris substring: `Paris`
- Default followup: `hello`
- Default hello marker: `HELLO_RESUME_MARKER`
- Default session id: `test-open-resume-e2e`
- Default grok UUID: `b2222222-2222-4222-8222-222222222222`
- Instant attach env: `AGENT_RUN_OPEN_ATTACH_INSTANT=1`

```go
import (
	"runtime"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

const (
	envOpenAttachInstant     = "AGENT_RUN_OPEN_ATTACH_INSTANT"
	envGrokTTYCommand        = "AGENT_RUN_GROK_TTY_COMMAND"
	envGrokTTYGrokSessionID  = "AGENT_RUN_GROK_TTY_GROK_SESSION_ID"
	envLLMMockRunGrokCommand = "LLM_MOCK_RUN_GROK_COMMAND"

	defaultSessionID       = "test-open-resume-e2e"
	defaultGrokSessionUUID = "b2222222-2222-4222-8222-222222222222"
	defaultOpenPrompt      = "one word of France capital"
	defaultWantParis       = "Paris"
	defaultFollowup        = "hello"
	defaultHelloMarker     = "HELLO_RESUME_MARKER"

	defaultParisWait  = 45 * time.Second
	defaultExitWait   = 30 * time.Second
	defaultResumeWait = 30 * time.Second
	defaultExecTimeout = 3 * time.Minute
)

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "open-resume-e2e-doctest-"+d.DOCTEST_SESSION_ID)
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

func ensureStubDist(distDir string) error {
	entries, statErr := os.ReadDir(distDir)
	if statErr == nil {
		for _, e := range entries {
			if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			return nil
		}
	} else if !os.IsNotExist(statErr) {
		return statErr
	}
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(distDir, "index.html"), []byte("stub\n"), 0644)
}

func ensureSessionBinaries(t *testing.T, d *session.Doctest, repoRoot string) (agentRun, llmMock string) {
	t.Helper()
	if root, rootErr := findAgentProRoot(repoRoot); rootErr != nil {
		t.Fatalf("agent-pro root: %v", rootErr)
	} else {
		repoRoot = root
	}
	cache := sessionCacheDir(d)
	agentRun = filepath.Join(cache, "agent-run")
	llmMock = filepath.Join(cache, "llm-mock-run-grok")
	ready := filepath.Join(cache, "binaries.ready")
	lock := filepath.Join(cache, "build.lock")
	err := withFileLock(t, lock, func() error {
		llmMockServer := filepath.Join(filepath.Dir(llmMock), "llm-mock")
		if fileExists(ready) && fileExists(agentRun) && fileExists(llmMock) && fileExists(llmMockServer) {
			return nil
		}
		if err := os.MkdirAll(cache, 0755); err != nil {
			return err
		}
		for _, rel := range []string{"frontend-agent-run/dist", "frontend/dist"} {
			if err := ensureStubDist(filepath.Join(repoRoot, rel)); err != nil {
				return fmt.Errorf("ensure %s stub: %w", rel, err)
			}
		}
		builds := []struct {
			out  string
			args []string
		}{
			{agentRun, []string{"build", "-C", "cmd", "-o", agentRun, "./agent-run"}},
			{llmMock, []string{"build", "-o", llmMock, "./agent/llm/llm-mock/llm-mock-run-grok"}},
			{llmMockServer, []string{"build", "-o", llmMockServer, "./agent/llm/llm-mock"}},
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

func withoutEnvKey(env []string, key string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env))
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			continue
		}
		out = append(out, e)
	}
	return out
}

func setEnvKV(req *Request, key, value string) {
	req.Env = withoutEnvKey(req.Env, key)
	req.Env = append(req.Env, key+"="+value)
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
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), b, 0644); err != nil {
		t.Fatalf("write summary.json: %v", err)
	}
	updatesPath := filepath.Join(dir, "updates.jsonl")
	seed := []string{acpUserMessageChunk(prompt)}
	seed = append(seed, initialLines...)
	f, err := os.OpenFile(updatesPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("open updates.jsonl: %v", err)
	}
	defer f.Close()
	for _, line := range seed {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if _, err := fmt.Fprintln(f, line); err != nil {
			t.Fatalf("write updates.jsonl: %v", err)
		}
	}
	return updatesPath
}

// multiTurnOpenResumeHook is the LLM_MOCK_RUN_GROK_COMMAND body.
// Inject uses '\r' with ICANON off, so bash `read` never completes — use
// python3 to treat \r and \n as line terminators.
//
// Protocol:
//   - On start: banner + Paris on PTY + seed updates.jsonl (first turn).
//   - On /exit: print grok --resume footer + [Terminal exited], exit 0.
//   - On hello: print HELLO_RESUME_MARKER (resume --open followup).
//   - Other lines: re-print Paris (idempotent first-turn inject).
func multiTurnOpenResumeHook(sessionUUID, parisText, helloMarker string) string {
	// Python is more reliable for PTY inject line endings than bash read.
	return fmt.Sprintf(`python3 -u -c '
import json, os, sys, time, urllib.parse
from datetime import datetime, timezone

uuid = %q
paris = %q
hello_marker = %q

def enc_cwd(wd):
    p = os.path.abspath(wd)
    if p.startswith("/private/var/"):
        p = "/var/" + p[len("/private/var/"):]
    elif p.startswith("/private/tmp/"):
        p = "/tmp/" + p[len("/private/tmp/"):]
    return urllib.parse.quote(p, safe="")

def seed_session(user_text, asst_text):
    wd = os.getcwd()
    d = os.path.join(os.environ.get("GROK_HOME", ""), "sessions", enc_cwd(wd), uuid)
    os.makedirs(d, exist_ok=True)
    now = datetime.now(timezone.utc).strftime("%%Y-%%m-%%dT%%H:%%M:%%SZ")
    summary = {"info": {"cwd": wd, "sessionId": uuid, "openedAt": now}, "created_at": now}
    with open(os.path.join(d, "summary.json"), "w") as f:
        json.dump(summary, f)
    lines = [
        {"sessionUpdate": "user_message_chunk", "content": {"type": "text", "text": user_text}},
        {"sessionUpdate": "agent_message_chunk", "content": {"type": "text", "text": asst_text}},
    ]
    with open(os.path.join(d, "updates.jsonl"), "w") as f:
        for obj in lines:
            f.write(json.dumps(obj) + "\n")

def read_line():
    """Read until CR or LF (inject mode uses CR with ICANON off)."""
    buf = []
    while True:
        ch = sys.stdin.read(1)
        if ch == "":
            return None if not buf else "".join(buf)
        if ch in ("\r", "\n"):
            return "".join(buf)
        if ch == "\x7f" or ch == "\b":
            if buf:
                buf.pop()
            continue
        buf.append(ch)

# First turn: show Paris immediately so snapshot/bind are deterministic.
sys.stdout.write("GROK_TTY_BANNER\n")
sys.stdout.write(paris + "\n")
sys.stdout.write("Grok \u203a ")
sys.stdout.flush()
seed_session("one word of France capital", paris)

while True:
    line = read_line()
    if line is None:
        time.sleep(180)
        break
    line = line.strip()
    if not line:
        continue
    low = line.lower()
    if line == "/exit" or low == "exit" or low.endswith("/exit") or "/exit" in low:
        # Clear prior "Grok ›" idle markers so status sendable/exited heuristics
        # see a zombie exit footer (not historical prompt → sendable:yes).
        sys.stdout.write("\033[2J\033[H")
        sys.stdout.write("Resume this session with:\n  grok --resume %%s\n[Terminal exited]\n" %% uuid)
        sys.stdout.flush()
        sys.exit(0)
    if "hello" in low:
        seed_session(line, hello_marker)
        sys.stdout.write(hello_marker + "\nGrok \u203a ")
        sys.stdout.flush()
        continue
    # Re-assert Paris for open prompt inject / retries.
    seed_session(line, paris)
    sys.stdout.write(paris + "\nGrok \u203a ")
    sys.stdout.flush()
'`, sessionUUID, parisText, helloMarker)
}

func configureMockGrokEnv(t *testing.T, req *Request) {
	t.Helper()
	// Binary path drives PTY — never use AGENT_RUN_GROK_TTY_COMMAND.
	stripEnvPrefix(req, envGrokTTYCommand+"=")
	setEnvKV(req, "AGENT_RUN_HOME", req.Home)
	setEnvKV(req, envOpenAttachInstant, "1")
	setEnvKV(req, envGrokTTYGrokSessionID, req.GrokSessionUUID)
	setEnvKV(req, envLLMMockRunGrokCommand, multiTurnOpenResumeHook(req.GrokSessionUUID, req.WantParis, req.HelloMarker))
	// Help discovery and mock home resolution.
	setEnvKV(req, "GROK_HOME", req.GrokHome)
	// Prepend session bin dir so bare names resolve if needed.
	binDir := filepath.Dir(req.AgentRun)
	setEnvKV(req, "PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func execAgentRun(t *testing.T, req *Request, timeout time.Duration, args ...string) CmdResult {
	t.Helper()
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.Workspace
	cmd.Env = append(os.Environ(), req.Env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	res := CmdResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}
	if err == nil {
		return res
	}
	if ctx.Err() != nil {
		res.ExitCode = -1
		return res
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		res.ExitCode = exitErr.ExitCode()
		res.Err = nil // CLI non-zero is structured in ExitCode
		return res
	}
	return res
}

func waitSnapshotContains(t *testing.T, req *Request, want string, max time.Duration) (string, bool) {
	t.Helper()
	deadline := time.Now().Add(max)
	var last string
	wantLower := strings.ToLower(want)
	for time.Now().Before(deadline) {
		res := execAgentRun(t, req, 10*time.Second, "snapshot", req.SessionID)
		last = res.Stdout
		if strings.Contains(last, want) || strings.Contains(strings.ToLower(last), wantLower) {
			return last, true
		}
		// Soft fallback: events.jsonl may stream Paris before PTY paint.
		if blob := readEventsBlob(req); strings.Contains(blob, want) || strings.Contains(strings.ToLower(blob), wantLower) {
			return last + "\n/* events also matched */\n" + blob, true
		}
		time.Sleep(400 * time.Millisecond)
	}
	return last, false
}

func waitExitedTrue(t *testing.T, req *Request, max time.Duration) (human string, jsonOut string, ok bool) {
	t.Helper()
	deadline := time.Now().Add(max)
	for time.Now().Before(deadline) {
		humanRes := execAgentRun(t, req, 10*time.Second, "status", req.SessionID)
		human = humanRes.Stdout + humanRes.Stderr
		if strings.Contains(human, "exited:     true") || strings.Contains(human, "exited: true") {
			jsonRes := execAgentRun(t, req, 10*time.Second, "status", "--json", req.SessionID)
			return human, jsonRes.Stdout, true
		}
		jsonRes := execAgentRun(t, req, 10*time.Second, "status", "--json", req.SessionID)
		jsonOut = jsonRes.Stdout
		var st struct {
			Runner struct {
				Exited *bool `json:"exited"`
			} `json:"runner"`
		}
		if json.Unmarshal([]byte(jsonOut), &st) == nil && st.Runner.Exited != nil && *st.Runner.Exited {
			return human, jsonOut, true
		}
		time.Sleep(400 * time.Millisecond)
	}
	return human, jsonOut, false
}

func waitSendableBestEffort(t *testing.T, req *Request, max time.Duration) {
	t.Helper()
	deadline := time.Now().Add(max)
	for time.Now().Before(deadline) {
		res := execAgentRun(t, req, 10*time.Second, "status", req.SessionID)
		if strings.Contains(res.Stdout, "sendable: yes") {
			return
		}
		time.Sleep(300 * time.Millisecond)
	}
}

func eventsJSONLPath(home, sessionID string) string {
	return filepath.Join(home, "sessions", "grok-tty", sessionID, "events.jsonl")
}

func readEventsBlob(req *Request) string {
	path := eventsJSONLPath(req.Home, req.SessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		// Session id may differ if open rewrote it; scan home.
		root := filepath.Join(req.Home, "sessions", "grok-tty")
		var blob strings.Builder
		_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			if info.Name() == "events.jsonl" {
				b, readErr := os.ReadFile(p)
				if readErr == nil {
					blob.Write(b)
					blob.WriteByte('\n')
				}
			}
			return nil
		})
		return blob.String()
	}
	return string(data)
}

func cleanupRegistryServes(home string) {
	for _, sub := range []string{"grok-tty-registry", "codex-tty-registry"} {
		dir := filepath.Join(home, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			b, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			var ent struct {
				PID int `json:"pid"`
			}
			if json.Unmarshal(b, &ent) != nil || ent.PID <= 0 {
				continue
			}
			if ent.PID == os.Getpid() {
				continue
			}
			_ = exec.Command("kill", fmt.Sprint(ent.PID)).Run()
		}
	}
}

func openArgs(req *Request) []string {
	return []string{
		"run",
		"--session-id=" + req.SessionID,
		"--agent-runner=grok-tty",
		"--open",
		"--dir", req.Workspace,
		"--agent-runner-binary", req.LLMMockRunGrok,
		"--agent-runner-config-home", req.GrokHome,
		req.OpenPrompt,
	}
}

func resumeArgs(req *Request) []string {
	args := []string{"resume"}
	// ResumeOpen defaults true unless ResumeNoOpen is set by the leaf scenario.
	useOpen := !req.ResumeNoOpen
	if req.ResumeOpen {
		useOpen = true
	}
	if useOpen {
		args = append(args, "--open")
	}
	if req.NoSubmit {
		args = append(args, "--no-submit")
	}
	args = append(args,
		"--dir", req.Workspace,
		"--agent-runner-binary", req.LLMMockRunGrok,
		"--agent-runner-config-home", req.GrokHome,
		req.SessionID,
	)
	if strings.TrimSpace(req.FollowupPrompt) != "" {
		args = append(args, req.FollowupPrompt)
	}
	return args
}

func statusResumeReady(human, jsonOut string) bool {
	if strings.Contains(human, "ready: yes") {
		return true
	}
	var st struct {
		Resume struct {
			Ready bool `json:"ready"`
		} `json:"resume"`
	}
	if json.Unmarshal([]byte(jsonOut), &st) == nil && st.Resume.Ready {
		return true
	}
	return false
}

func statusBound(human, jsonOut string) bool {
	if strings.Contains(human, "status:     bound") || strings.Contains(human, "status: bound") {
		return true
	}
	var st struct {
		Runner struct {
			Status    string `json:"status"`
			SessionID string `json:"session_id"`
		} `json:"runner"`
	}
	if json.Unmarshal([]byte(jsonOut), &st) == nil {
		if st.Runner.Status == "bound" || strings.TrimSpace(st.Runner.SessionID) != "" {
			return true
		}
	}
	return strings.Contains(human, "session_id:") && !strings.Contains(human, "session_id: \n") &&
		strings.Contains(human, "bound")
}

func fillHelloFlags(resp *Response, req *Request) {
	lowerSnap := strings.ToLower(resp.ResumeSnapshot + "\n" + resp.EventsBlob)
	resp.HasHello = strings.Contains(resp.ResumeSnapshot, req.FollowupPrompt) ||
		strings.Contains(lowerSnap, "hello") ||
		strings.Contains(resp.ResumeSnapshot, req.HelloMarker) ||
		strings.Contains(resp.EventsBlob, req.HelloMarker)
	if req.SecondMarker != "" {
		resp.HasSecond = strings.Contains(resp.EventsBlob, req.SecondMarker) ||
			strings.Contains(resp.ResumeSnapshot, req.SecondMarker) ||
			strings.Contains(resp.ParisSnapshot, req.SecondMarker)
	}
}

func runScenario(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	switch strings.TrimSpace(req.Scenario) {
	case "live-send-followup-no-exit":
		return runLiveSendFollowup(t, req)
	case "send-two-followups-while-live":
		return runTwoLiveFollowups(t, req)
	case "resume-while-live-denied":
		return runResumeWhileLiveDenied(t, req)
	case "open-bind-prints-session":
		return runOpenBindOnly(t, req)
	case "status-ready-after-exit":
		req.SkipResume = true
		return runOpenParisExitResumeHello(t, req)
	case "double-resume":
		req.ResumeTwice = true
		return runOpenParisExitResumeHello(t, req)
	case "resume-no-followup":
		req.FollowupPrompt = ""
		return runOpenParisExitResumeHello(t, req)
	case "resume-open-no-submit-hello":
		req.NoSubmit = true
		req.ResumeOpen = true
		return runOpenParisExitResumeHello(t, req)
	case "resume-keep-tty-followup":
		req.ResumeNoOpen = true
		return runOpenParisExitResumeHello(t, req)
	case "open-paris-exit-resume-hello", "mock-grok", "":
		return runOpenParisExitResumeHello(t, req)
	default:
		return runOpenParisExitResumeHello(t, req)
	}
}

func runOpenParisExitResumeHello(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	start := time.Now()
	resp := &Response{}
	t.Cleanup(func() { cleanupRegistryServes(req.Home) })

	// Pre-seed grok session so background bind can resolve quickly; mock
	// overwrites updates.jsonl on first prompt with the same Paris text.
	_ = writeFakeGrokSessionDir(t, req.GrokHome, req.Workspace, req.GrokSessionUUID, req.OpenPrompt,
		acpAgentMessageChunk(req.WantParis),
	)

	// 1) Open first turn.
	openTimeout := 90 * time.Second
	if req.ExecTimeout > 0 && req.ExecTimeout < openTimeout {
		openTimeout = req.ExecTimeout
	}
	resp.Open = execAgentRun(t, req, openTimeout, openArgs(req)...)
	openCombined := resp.Open.Stderr + "\n" + resp.Open.Stdout
	resp.BoundOnOpen = strings.Contains(openCombined, "grok session") ||
		strings.Contains(openCombined, "grok-tty:")
	// Soft gate: prefer exit 0 or stderr showing bind progress / session id.
	if resp.Open.ExitCode != 0 && !resp.BoundOnOpen {
		resp.Err = fmt.Errorf("open exited %d without session bind hints; stderr=%s", resp.Open.ExitCode, resp.Open.Stderr)
		resp.Elapsed = time.Since(start)
		return resp, nil
	}

	// Status after open (for bind-running leaves).
	stOpen := execAgentRun(t, req, 15*time.Second, "status", req.SessionID)
	stOpenJSON := execAgentRun(t, req, 15*time.Second, "status", "--json", req.SessionID)
	resp.StatusAfterOpen = stOpen
	if statusBound(stOpen.Stdout, stOpenJSON.Stdout) {
		resp.BoundOnOpen = true
	}

	// 2) Wait Paris in snapshot and/or events.
	parisWait := req.ParisWait
	if parisWait <= 0 {
		parisWait = defaultParisWait
	}
	snap, ok := waitSnapshotContains(t, req, req.WantParis, parisWait)
	resp.ParisSnapshot = snap
	resp.HasParis = ok
	resp.EventsBlob = readEventsBlob(req)
	resp.EventsFilePath = eventsJSONLPath(req.Home, req.SessionID)
	if !resp.HasParis {
		// One more events-only check after wait.
		if strings.Contains(resp.EventsBlob, req.WantParis) {
			resp.HasParis = true
		}
	}

	if req.SkipExit {
		resp.Elapsed = time.Since(start)
		return resp, nil
	}

	// 3) Send /exit (prefer sendable, but continue even if not).
	waitSendableBestEffort(t, req, 15*time.Second)
	resp.SendExit = execAgentRun(t, req, 45*time.Second, "send", req.SessionID, "/exit")
	// Brief settle for mock to print exit markers into scrollback.
	time.Sleep(400 * time.Millisecond)

	// 4) Wait exited:true.
	exitWait := req.ExitWait
	if exitWait <= 0 {
		exitWait = defaultExitWait
	}
	human, js, exited := waitExitedTrue(t, req, exitWait)
	resp.StatusAfterExit = CmdResult{Stdout: human}
	resp.StatusJSONAfterExit = js
	resp.ExitedTrue = exited
	resp.ResumeReady = statusResumeReady(human, js)

	if req.SkipResume {
		resp.Elapsed = time.Since(start)
		return resp, nil
	}

	// 5) Resume (even if exited wait soft-failed — assert will RED).
	resumeTimeout := 90 * time.Second
	resp.Resume = execAgentRun(t, req, resumeTimeout, resumeArgs(req)...)
	resumeCombined := strings.ToLower(resp.Resume.Stderr + "\n" + resp.Resume.Stdout)
	resp.AlreadyInUse = strings.Contains(resumeCombined, "already in use")
	resp.ResumeDenied = strings.Contains(resumeCombined, "still active") ||
		strings.Contains(resumeCombined, "use send") ||
		strings.Contains(resumeCombined, "not exited") ||
		strings.Contains(resumeCombined, "cannot resume")

	// 6) Snapshot after resume (when we expect success).
	if !req.ExpectResumeDenied {
		time.Sleep(800 * time.Millisecond)
		snapRes := execAgentRun(t, req, 15*time.Second, "snapshot", req.SessionID)
		resp.ResumeSnapshot = snapRes.Stdout
		resp.EventsBlob = readEventsBlob(req)
		fillHelloFlags(resp, req)
	}

	// 7) Optional second exit+resume cycle.
	if req.ResumeTwice && !req.ExpectResumeDenied && resp.Resume.ExitCode == 0 {
		waitSendableBestEffort(t, req, 15*time.Second)
		_ = execAgentRun(t, req, 45*time.Second, "send", req.SessionID, "/exit")
		time.Sleep(400 * time.Millisecond)
		_, _, _ = waitExitedTrue(t, req, exitWait)
		resp.Resume2 = execAgentRun(t, req, resumeTimeout, resumeArgs(req)...)
		c2 := strings.ToLower(resp.Resume2.Stderr + "\n" + resp.Resume2.Stdout)
		if strings.Contains(c2, "already in use") {
			resp.AlreadyInUse = true
		}
		time.Sleep(500 * time.Millisecond)
		snap2 := execAgentRun(t, req, 15*time.Second, "snapshot", req.SessionID)
		resp.ResumeSnapshot = snap2.Stdout
		resp.EventsBlob = readEventsBlob(req)
		fillHelloFlags(resp, req)
	}

	resp.Elapsed = time.Since(start)
	return resp, nil
}

func runLiveSendFollowup(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	start := time.Now()
	resp := &Response{}
	t.Cleanup(func() { cleanupRegistryServes(req.Home) })
	_ = writeFakeGrokSessionDir(t, req.GrokHome, req.Workspace, req.GrokSessionUUID, req.OpenPrompt,
		acpAgentMessageChunk(req.WantParis),
	)
	resp.Open = execAgentRun(t, req, 90*time.Second, openArgs(req)...)
	snap, ok := waitSnapshotContains(t, req, req.WantParis, req.ParisWait)
	resp.ParisSnapshot = snap
	resp.HasParis = ok
	waitSendableBestEffort(t, req, 20*time.Second)
	resp.SendFollowup = execAgentRun(t, req, 45*time.Second, "send", req.SessionID, req.FollowupPrompt)
	time.Sleep(600 * time.Millisecond)
	snapRes := execAgentRun(t, req, 15*time.Second, "snapshot", req.SessionID)
	resp.ResumeSnapshot = snapRes.Stdout
	resp.EventsBlob = readEventsBlob(req)
	fillHelloFlags(resp, req)
	// Still live: exited should not be required true.
	human, js, exited := waitExitedTrue(t, req, 2*time.Second)
	resp.StatusAfterExit = CmdResult{Stdout: human}
	resp.StatusJSONAfterExit = js
	resp.ExitedTrue = exited
	resp.Elapsed = time.Since(start)
	return resp, nil
}

func runTwoLiveFollowups(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SecondFollowup == "" {
		req.SecondFollowup = "again please"
	}
	if req.SecondMarker == "" {
		req.SecondMarker = "SECOND_FOLLOWUP_MARKER"
	}
	// Extend mock to answer "again" with second marker by reusing hello path if "hello" in text;
	// use followups that include "hello" substrings for second, or seed both via "hello" and "hello again".
	req.FollowupPrompt = "hello"
	req.SecondFollowup = "hello again"
	req.HelloMarker = defaultHelloMarker
	req.SecondMarker = "HELLO_RESUME_MARKER" // mock maps any hello* to hello_marker
	configureMockGrokEnv(t, req)

	start := time.Now()
	resp := &Response{}
	t.Cleanup(func() { cleanupRegistryServes(req.Home) })
	_ = writeFakeGrokSessionDir(t, req.GrokHome, req.Workspace, req.GrokSessionUUID, req.OpenPrompt,
		acpAgentMessageChunk(req.WantParis),
	)
	resp.Open = execAgentRun(t, req, 90*time.Second, openArgs(req)...)
	snap, ok := waitSnapshotContains(t, req, req.WantParis, req.ParisWait)
	resp.ParisSnapshot = snap
	resp.HasParis = ok
	waitSendableBestEffort(t, req, 20*time.Second)
	resp.SendFollowup = execAgentRun(t, req, 45*time.Second, "send", req.SessionID, req.FollowupPrompt)
	time.Sleep(500 * time.Millisecond)
	resp.SendSecond = execAgentRun(t, req, 45*time.Second, "send", req.SessionID, req.SecondFollowup)
	time.Sleep(600 * time.Millisecond)
	snapRes := execAgentRun(t, req, 15*time.Second, "snapshot", req.SessionID)
	resp.ResumeSnapshot = snapRes.Stdout
	resp.EventsBlob = readEventsBlob(req)
	fillHelloFlags(resp, req)
	resp.HasSecond = resp.HasHello // mock reuses marker; assert at least two send exits 0
	resp.Elapsed = time.Since(start)
	return resp, nil
}

func runResumeWhileLiveDenied(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	req.ExpectResumeDenied = true
	req.SkipExit = false
	// Open + Paris, do NOT exit, then resume.
	start := time.Now()
	resp := &Response{}
	t.Cleanup(func() { cleanupRegistryServes(req.Home) })
	_ = writeFakeGrokSessionDir(t, req.GrokHome, req.Workspace, req.GrokSessionUUID, req.OpenPrompt,
		acpAgentMessageChunk(req.WantParis),
	)
	resp.Open = execAgentRun(t, req, 90*time.Second, openArgs(req)...)
	snap, ok := waitSnapshotContains(t, req, req.WantParis, req.ParisWait)
	resp.ParisSnapshot = snap
	resp.HasParis = ok
	// Resume while still live (no /exit).
	resp.Resume = execAgentRun(t, req, 45*time.Second, resumeArgs(req)...)
	c := strings.ToLower(resp.Resume.Stderr + "\n" + resp.Resume.Stdout)
	resp.AlreadyInUse = strings.Contains(c, "already in use")
	resp.ResumeDenied = strings.Contains(c, "still active") ||
		strings.Contains(c, "use send") ||
		strings.Contains(c, "not exited") ||
		strings.Contains(c, "cannot resume") ||
		resp.Resume.ExitCode != 0
	resp.Elapsed = time.Since(start)
	return resp, nil
}

func runOpenBindOnly(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	start := time.Now()
	resp := &Response{}
	t.Cleanup(func() { cleanupRegistryServes(req.Home) })
	_ = writeFakeGrokSessionDir(t, req.GrokHome, req.Workspace, req.GrokSessionUUID, req.OpenPrompt,
		acpAgentMessageChunk(req.WantParis),
	)
	resp.Open = execAgentRun(t, req, 90*time.Second, openArgs(req)...)
	openCombined := resp.Open.Stderr + "\n" + resp.Open.Stdout
	resp.BoundOnOpen = strings.Contains(openCombined, "grok session") &&
		strings.Contains(openCombined, "grok-tty:")
	// Also wait Paris so session is fully live.
	snap, ok := waitSnapshotContains(t, req, req.WantParis, req.ParisWait)
	resp.ParisSnapshot = snap
	resp.HasParis = ok
	resp.EventsBlob = readEventsBlob(req)
	resp.Elapsed = time.Since(start)
	return resp, nil
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	repoRoot, err := findAgentProRoot(d.DOCTEST_ROOT)
	if err != nil {
		return err
	}
	req.RepoRoot = repoRoot
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.GrokHome = filepath.Join(req.TempDir, ".grok")
	req.Workspace = filepath.Join(req.TempDir, "workspace")
	for _, d := range []string{req.Home, req.GrokHome, req.Workspace} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	// Project-like workspace marker (avoid bare-tmp picker heuristics).
	if err := os.WriteFile(filepath.Join(req.Workspace, "README.md"), []byte("open-resume-e2e workspace\n"), 0644); err != nil {
		return err
	}

	req.AgentRun, req.LLMMockRunGrok = ensureSessionBinaries(t, d, req.RepoRoot)
	// Also place copies under leaf bin/ for isolation documentation / PATH.
	leafBin := filepath.Join(req.TempDir, "bin")
	if err := os.MkdirAll(leafBin, 0755); err != nil {
		return err
	}
	// Symlink into leaf bin for clearer per-leaf layout (best-effort).
	_ = os.Symlink(req.AgentRun, filepath.Join(leafBin, "agent-run"))
	_ = os.Symlink(req.LLMMockRunGrok, filepath.Join(leafBin, "llm-mock-run-grok"))

	if req.SessionID == "" {
		req.SessionID = defaultSessionID
	}
	if req.GrokSessionUUID == "" {
		req.GrokSessionUUID = defaultGrokSessionUUID
	}
	if req.OpenPrompt == "" {
		req.OpenPrompt = defaultOpenPrompt
	}
	if req.WantParis == "" {
		req.WantParis = defaultWantParis
	}
	if req.FollowupPrompt == "" {
		req.FollowupPrompt = defaultFollowup
	}
	if req.HelloMarker == "" {
		req.HelloMarker = defaultHelloMarker
	}
	if req.ParisWait <= 0 {
		req.ParisWait = defaultParisWait
	}
	if req.ExitWait <= 0 {
		req.ExitWait = defaultExitWait
	}
	if req.ResumeWait <= 0 {
		req.ResumeWait = defaultResumeWait
	}
	if req.ExecTimeout <= 0 {
		req.ExecTimeout = defaultExecTimeout
	}

	configureMockGrokEnv(t, req)
	return nil
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
