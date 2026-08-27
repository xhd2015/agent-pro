# Scenario

**Feature**: grok-tty `--exit-on-idle` on live boxed chrome via `llm-mock-run-grok`

```
agent-run run --detach --exit-on-idle --idle-timeout=2s
  --agent-runner=grok-tty
  --agent-runner-binary llm-mock-run-grok
  + LLM_MOCK_RUN_GROK_COMMAND finished-turn boxed ❯ chrome
  -> idle-policy.json
  -> tty status --json (screen / input_box / live)
```

## Preconditions

- Repository contains `cmd/agent-run` and `agent/llm/llm-mock/llm-mock-run-grok`.
- Session-scoped cache under
  `$TMPDIR/run-exit-on-idle-grok-tty-doctest-<d.DOCTEST_SESSION_ID>/` shares
  compiled binaries across parallel leaves.
- Each leaf uses isolated `AGENT_RUN_HOME` / `GROK_HOME` under `t.TempDir()`.
- Parallel-safe: no `os.Setenv` / `Chdir`; child `cmd.Env` / `cmd.Dir` only.
- No stand-in `grok`/`codex` binary on PATH — only `llm-mock-run-grok`.

## Steps

1. Root `Setup` builds session binaries and isolated homes.
2. `detach/` sets `--detach` + live chrome hook + 2s idle timeout.
3. Leaf sets `Op` / `ObserveAfter`.
4. `Run` detaches, sleeps, probes `tty status --json`.

## Context

- Crime scene: host `idle-probe-10s-detach-v2` — finished `pong` / `Worked for`
  / empty boxed composer; `screen=banner` `input_box=occupied`; did not exit.
- Desired: `screen=idle` `input_box=empty`; then TTY gone after timeout+grace.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

const (
	defaultIdleTimeout = 2 * time.Second
	defaultGrace       = 5 * time.Second
	classifySettle     = 2 * time.Second
	// Probe Cap (~1s×3) + schedule gaps + grace; leave slack for SnapshotText.
	exitObserve = defaultIdleTimeout + defaultGrace + 8*time.Second
	idlePrompt         = "reply with exactly: pong"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	if req == nil {
		return fmt.Errorf("nil Request")
	}
	req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	if err := os.MkdirAll(req.Home, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		return err
	}
	req.AgentRun, req.LLMMockRunGrok = ensureSessionBinaries(t, d, req.RepoRoot)
	req.IdleTimeout = defaultIdleTimeout
	req.Prompt = idlePrompt
	req.Env = append(req.Env,
		"AGENT_RUN_HOME="+req.Home,
		"GROK_HOME="+req.GrokHome,
		"PATH="+filepath.Dir(req.AgentRun)+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	return nil
}

const seededGrokUUID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

func seedGrokSessionForBind(t *testing.T, grokHome, workspace, prompt string) {
	t.Helper()
	abs, err := filepath.Abs(workspace)
	if err != nil {
		abs = workspace
	}
	dir := filepath.Join(grokHome, "sessions", url.PathEscape(abs), seededGrokUUID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir grok session: %v", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	summary, _ := json.Marshal(map[string]any{
		"info": map[string]any{
			"cwd":       abs,
			"sessionId": seededGrokUUID,
			"openedAt":  now,
		},
		"created_at": now,
	})
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), summary, 0644); err != nil {
		t.Fatalf("write summary.json: %v", err)
	}
	userLine, _ := json.Marshal(map[string]any{
		"sessionUpdate": "user_message_chunk",
		"content":       map[string]any{"type": "text", "text": prompt},
	})
	if err := os.WriteFile(filepath.Join(dir, "updates.jsonl"), append(userLine, '\n'), 0644); err != nil {
		t.Fatalf("write updates.jsonl: %v", err)
	}
}

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "run-exit-on-idle-grok-tty-doctest-"+d.DOCTEST_SESSION_ID)
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
	llmMockServer := filepath.Join(cache, "llm-mock")
	ready := filepath.Join(cache, "binaries.ready")
	lock := filepath.Join(cache, "build.lock")
	err := withFileLock(t, lock, func() error {
		if fileExists(ready) && fileExists(agentRun) && fileExists(llmMock) && fileExists(llmMockServer) {
			return nil
		}
		if err := os.MkdirAll(cache, 0755); err != nil {
			return err
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
			cmd.Env = append(os.Environ(), "GOWORK=off")
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

// liveBoxedIdleChromeHook paints the crime-scene finished Grok frame:
// Worked for + boxed empty ❯ with right border on the same line (no wrap).
// llm-mock-run-grok execs this via LLM_MOCK_RUN_GROK_COMMAND.
func liveBoxedIdleChromeHook(holdSec int) string {
	if holdSec <= 0 {
		holdSec = 30
	}
	// One physical composer line: │ ❯ + padding + │  (occupied today; want empty).
	// Read stdin so watchdog SoftExit `/exit` can stop the PTY (sleep-only
	// ignores inject, so __serve would stay up until holdSec).
	frame := "" +
		"reply with exactly: pong\\n" +
		"    Worked for 6.0s\\n" +
		"pong\\n" +
		" ╭----------------------------------------------------------------------╮\\n" +
		" │ ❯                                                        │\\n" +
		" ------------------------------------------------ Grok 4.6 (high) · always-approve -╯\\n" +
		" Shift+Tab:mode  │  Ctrl+x:shortcuts\\n"
	return fmt.Sprintf("sh -c 'printf \"%s\"; end=$((SECONDS+%d)); while [ \"$SECONDS\" -lt \"$end\" ]; do IFS= read -r -t 1 line || continue; case \"$line\" in *exit*) exit 0 ;; esac; done'", frame, holdSec)
}

func configureLiveChromeHook(req *Request, holdSec int) {
	kept := req.Env[:0]
	for _, e := range req.Env {
		if len(e) >= len("LLM_MOCK_RUN_GROK_COMMAND=") && e[:len("LLM_MOCK_RUN_GROK_COMMAND=")] == "LLM_MOCK_RUN_GROK_COMMAND=" {
			continue
		}
		kept = append(kept, e)
	}
	req.Env = append(kept, "LLM_MOCK_RUN_GROK_COMMAND="+liveBoxedIdleChromeHook(holdSec))
}

func runDetachThenStatus(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SessionID == "" {
		return nil, fmt.Errorf("SessionID is required")
	}
	if req.IdleTimeout <= 0 {
		req.IdleTimeout = defaultIdleTimeout
	}
	seedGrokSessionForBind(t, req.GrokHome, req.TempDir, req.Prompt)
	observe := req.ObserveAfter
	if observe <= 0 {
		observe = classifySettle
	}

	args := []string{
		"run",
		"--agent-runner", "grok-tty",
		"--session-id", req.SessionID,
		"--auto-send-or-resume",
		"--allow-relocate-resume-session-dir",
		"--dir", req.TempDir,
		"--agent-runner-binary", req.LLMMockRunGrok,
		"--agent-runner-config-home", req.GrokHome,
		"--color",
		"--detach",
		"--exit-on-idle",
		"--idle-timeout", req.IdleTimeout.String(),
		"--",
		req.Prompt,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	runErr := cmd.Run()
	resp := &Response{
		DetachStdout: stdoutBuf.String(),
		DetachStderr: stderrBuf.String(),
	}
	if cmd.ProcessState != nil {
		resp.DetachExit = cmd.ProcessState.ExitCode()
	} else if runErr != nil {
		resp.DetachExit = 1
	}
	if runErr != nil {
		return resp, fmt.Errorf("detach: %w (exit=%d)\nstdout:\n%s\nstderr:\n%s", runErr, resp.DetachExit, resp.DetachStdout, resp.DetachStderr)
	}

	policyPath := filepath.Join(req.Home, "sessions", req.SessionID, "idle-policy.json")
	if raw, err := os.ReadFile(policyPath); err == nil {
		resp.PolicyJSON = string(raw)
	}

	t.Cleanup(func() {
		killCtx, killCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer killCancel()
		k := exec.CommandContext(killCtx, req.AgentRun, "tty", "kill", req.SessionID)
		k.Dir = req.TempDir
		k.Env = append(os.Environ(), req.Env...)
		_ = k.Run()
	})

	time.Sleep(observe)
	fillStatus(t, req, resp)
	return resp, nil
}

func fillStatus(t *testing.T, req *Request, resp *Response) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.AgentRun, "tty", "status", "--json", req.SessionID)
	cmd.Dir = req.TempDir
	cmd.Env = append(os.Environ(), req.Env...)
	out, err := cmd.CombinedOutput()
	resp.StatusJSON = string(out)
	resp.StatusHuman = string(out)
	if cmd.ProcessState != nil {
		resp.StatusExit = cmd.ProcessState.ExitCode()
	}
	if err != nil {
		resp.Alive = false
		return
	}
	var st ttyStatusJSON
	if jerr := json.Unmarshal(out, &st); jerr != nil {
		resp.Alive = false
		return
	}
	resp.ScreenStatus = st.ScreenStatus
	resp.InputBox = st.InputBox
	resp.Sendable = st.Sendable
	resp.SendableState = st.SendableState
	resp.TCPReachable = st.TCPReachable
	resp.Alive = st.TCPReachable && st.PID > 0
}
```
