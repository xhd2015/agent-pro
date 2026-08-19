# Scenario

**Feature**: `--exit-on-idle` on live Codex TUI via `llm-mock-run-codex`

```
# isolated homes + session-cached binaries
agent-run run --detach --exit-on-idle --idle-timeout=10s
  --agent-runner=codex-tty
  --agent-runner-binary llm-mock-run-codex
  -- "reply with exactly: pong"
  -> idle-policy.json
  -> watchdog space-probe occupancy -> /exit or hold
```

## Preconditions

- Nested DOCTEST root (does not inherit `cmd/agent-run/tests` Setup/Run).
- Repo root: `d.DOCTEST_ROOT/../../../..`
  (`run-exit-on-idle-codex-tty` → `tests` → `agent-run` → `cmd` → module root).
- Real `codex` on PATH (else `t.Skip`).
- Session-scoped cache under
  `$TMPDIR/run-exit-on-idle-codex-tty-doctest-<d.DOCTEST_SESSION_ID>/` shares
  compiled binaries (`agent-run`, `llm-mock`, `llm-mock-run-codex`) across
  parallel leaves. Mock pair lives in the **same directory**.
- Each leaf uses isolated `AGENT_RUN_HOME` / `CODEX_HOME` /
  `LLM_MOCK_CODEX_HOME` / workspace under `t.TempDir()`. Never write
  `~/.agent-run`.
- Parallel-safe: no `os.Setenv` / `Unsetenv` / `t.Setenv` / `Chdir` /
  `t.Chdir`. Child `cmd.Env` / `cmd.Dir` only.
- Strip `AGENT_RUN_CODEX_TTY_COMMAND` and `LLM_MOCK_RUN_CODEX_COMMAND` so the
  real Codex TUI runs.

## Steps

1. Root `Setup` skips without `codex`, builds session binaries, writes mock
   config, isolates homes.
2. `detach/` sets `--detach` + 10s idle timeout.
3. Leaf sets `Op` / session id / draft.
4. `Run` detaches, optionally waits sendable and injects a no-submit draft,
   sleeps, probes `tty status --json` liveness only.

## Context

- Timeout 10s (< 30s) → watchdog samples at 0, 5s, 10s. Each occupancy probe
  may add up to 5s. Grace is 5s. Observe window is 10s + 5s + 20s probe slack
  (same budget as SPL `tests/seatalk-localbot-debugging-idle`).
- Observe occupancy through watchdog effect (gone vs still live). Do not
  treat `tty status input_box` as the probe result.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

const (
	defaultIdleTimeout = 10 * time.Second
	defaultGrace       = 5 * time.Second
	probeSlack         = 20 * time.Second
	shutdownSlack      = 15 * time.Second // serve TCP can linger briefly after /exit
	detachTimeout      = 90 * time.Second
	readyPoll          = 90 * time.Second
	idlePrompt         = "reply with exactly: pong"
	defaultDraft       = "DRAFT_OCCUPANCY_HOLD_zz9"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	if req == nil {
		return fmt.Errorf("nil Request")
	}
	if _, err := exec.LookPath("codex"); err != nil {
		t.Skip("codex not found in PATH")
	}
	req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.CodexHome = filepath.Join(req.TempDir, ".codex")
	req.Workspace = filepath.Join(req.TempDir, "workspace")
	for _, dir := range []string{req.Home, req.CodexHome, req.Workspace} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	_ = exec.Command("git", "init", "-q", req.Workspace).Run()
	if err := os.WriteFile(filepath.Join(req.Workspace, "README.md"), []byte("run-exit-on-idle-codex-tty\n"), 0644); err != nil {
		return err
	}

	req.AgentRun, req.LLMMock, req.LLMMockRunCodex = ensureSessionBinaries(t, d, req.RepoRoot)
	if filepath.Dir(req.LLMMock) != filepath.Dir(req.LLMMockRunCodex) {
		return fmt.Errorf("llm-mock and llm-mock-run-codex must share a directory")
	}

	req.MockConfigFile = filepath.Join(req.TempDir, "llm-mock-config.json")
	if err := os.WriteFile(req.MockConfigFile, []byte(`{
  "exchanges": [
    {
      "request": {"role": "user", "content": "reply with exactly: pong", "index": -1},
      "response": {"content": "pong", "finish_reason": "stop"}
    }
  ]
}
`), 0644); err != nil {
		return err
	}

	req.IdleTimeout = defaultIdleTimeout
	req.Prompt = idlePrompt
	req.Env = childEnv(req)
	return nil
}

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "run-exit-on-idle-codex-tty-doctest-"+d.DOCTEST_SESSION_ID)
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

func ensureSessionBinaries(t *testing.T, d *session.Doctest, repoRoot string) (agentRun, llmMock, llmMockRunCodex string) {
	t.Helper()
	cache := sessionCacheDir(d)
	agentRun = filepath.Join(cache, "agent-run")
	llmMock = filepath.Join(cache, "llm-mock")
	llmMockRunCodex = filepath.Join(cache, "llm-mock-run-codex")
	ready := filepath.Join(cache, "binaries.ready")
	lock := filepath.Join(cache, "build.lock")
	err := withFileLock(t, lock, func() error {
		if fileExists(ready) && fileExists(agentRun) && fileExists(llmMock) && fileExists(llmMockRunCodex) {
			return nil
		}
		if err := os.MkdirAll(cache, 0755); err != nil {
			return err
		}
		cmdDir := filepath.Join(repoRoot, "cmd")
		b1 := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", agentRun, "./agent-run")
		b1.Dir = cmdDir
		b1.Env = append(os.Environ(), "GOWORK=off")
		if out, err := b1.CombinedOutput(); err != nil {
			return fmt.Errorf("build agent-run: %w\n%s", err, out)
		}
		for _, b := range []struct {
			out string
			pkg string
		}{
			{llmMock, "./agent/llm/llm-mock"},
			{llmMockRunCodex, "./agent/llm/llm-mock/llm-mock-run-codex"},
		} {
			cmd := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", b.out, b.pkg)
			cmd.Dir = repoRoot
			cmd.Env = append(os.Environ(), "GOWORK=off")
			if out, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("build %s: %w\n%s", b.pkg, err, out)
			}
		}
		return os.WriteFile(ready, []byte("ok"), 0644)
	})
	if err != nil {
		t.Fatalf("session binaries: %v", err)
	}
	return agentRun, llmMock, llmMockRunCodex
}

func childEnv(req *Request) []string {
	env := append([]string(nil), os.Environ()...)
	binDir := filepath.Dir(req.LLMMockRunCodex)
	for _, key := range []string{
		"PATH",
		"AGENT_RUN_HOME",
		"CODEX_HOME",
		"LLM_MOCK_CODEX_HOME",
		"LLM_MOCK_CONFIG_FILE",
		"AGENT_RUN_CODEX_TTY_COMMAND",
		"LLM_MOCK_RUN_CODEX_COMMAND",
	} {
		env = withoutEnvKey(env, key)
	}
	return append(env,
		"AGENT_RUN_HOME="+req.Home,
		"CODEX_HOME="+req.CodexHome,
		"LLM_MOCK_CODEX_HOME="+req.CodexHome,
		"LLM_MOCK_CONFIG_FILE="+req.MockConfigFile,
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
}

func policyPath(req *Request) string {
	return filepath.Join(req.Home, "sessions", req.SessionID, "idle-policy.json")
}

func readPolicy(req *Request) string {
	raw, err := os.ReadFile(policyPath(req))
	if err != nil {
		return ""
	}
	return string(raw)
}

func execAgentRun(t *testing.T, req *Request, timeout time.Duration, args ...string) (stdout, stderr string, exitCode int, err error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.Workspace
	cmd.Env = append([]string(nil), req.Env...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	stdout, stderr = outBuf.String(), errBuf.String()
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	} else if err != nil {
		exitCode = 1
	}
	if ctx.Err() != nil {
		return stdout, stderr, exitCode, ctx.Err()
	}
	return stdout, stderr, exitCode, err
}

func cleanupRegistryServes(home string) {
	reg := filepath.Join(home, "codex-tty-registry")
	entries, err := os.ReadDir(reg)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(reg, e.Name()))
		if err != nil {
			continue
		}
		var st struct {
			PID int `json:"pid"`
		}
		if json.Unmarshal(data, &st) != nil || st.PID <= 0 {
			continue
		}
		_ = exec.Command("pkill", "-P", fmt.Sprintf("%d", st.PID)).Run()
		if p, err := os.FindProcess(st.PID); err == nil {
			_ = p.Signal(syscall.SIGTERM)
			time.Sleep(100 * time.Millisecond)
			_ = p.Signal(syscall.SIGKILL)
		}
	}
}

func fillStatus(t *testing.T, req *Request, resp *Response) {
	t.Helper()
	stdout, stderr, exit, err := execAgentRun(t, req, 5*time.Second, "tty", "status", "--json", req.SessionID)
	resp.StatusJSON = stdout + stderr
	resp.StatusExit = exit
	if err != nil {
		resp.Alive = false
		return
	}
	var st ttyStatusJSON
	if json.Unmarshal([]byte(stdout), &st) != nil {
		resp.Alive = false
		return
	}
	resp.TCPReachable = st.TCPReachable
	resp.PID = st.PID
	resp.Alive = st.TCPReachable && st.PID > 0
}

func waitSendable(t *testing.T, req *Request) bool {
	t.Helper()
	deadline := time.Now().Add(readyPoll)
	for time.Now().Before(deadline) {
		stdout, _, _, err := execAgentRun(t, req, 5*time.Second, "tty", "status", "--json", req.SessionID)
		if err == nil {
			var st ttyStatusJSON
			if json.Unmarshal([]byte(stdout), &st) == nil && st.Sendable && st.TCPReachable && st.PID > 0 {
				return true
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

func snapshotText(t *testing.T, req *Request) string {
	t.Helper()
	stdout, stderr, _, _ := execAgentRun(t, req, 5*time.Second, "tty", "snapshot", req.SessionID)
	return stdout + stderr
}

func runDetachThenObserve(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SessionID == "" {
		return nil, fmt.Errorf("SessionID is required")
	}
	if req.IdleTimeout <= 0 {
		req.IdleTimeout = defaultIdleTimeout
	}
	if req.Prompt == "" {
		req.Prompt = idlePrompt
	}
	observe := req.ObserveAfter
	if observe <= 0 {
		observe = defaultIdleTimeout + defaultGrace + probeSlack
	}
	if len(req.Env) == 0 {
		req.Env = childEnv(req)
	}

	args := []string{
		"run",
		"--detach",
		"--exit-on-idle",
		"--idle-timeout", req.IdleTimeout.String(),
		"--agent-runner=codex-tty",
		"--agent-runner-binary", req.LLMMockRunCodex,
		"--agent-runner-config-home", req.CodexHome,
		"--session-id", req.SessionID,
		"--dir", req.Workspace,
		"--prepend-path", filepath.Dir(req.LLMMockRunCodex),
		"--env", "LLM_MOCK_CODEX_HOME=" + req.CodexHome,
		"--env", "LLM_MOCK_CONFIG_FILE=" + req.MockConfigFile,
		"--",
		req.Prompt,
	}

	stdout, stderr, exit, runErr := execAgentRun(t, req, detachTimeout, args...)
	resp := &Response{
		DetachStdout: stdout,
		DetachStderr: stderr,
		DetachExit:   exit,
		PolicyJSON:   readPolicy(req),
	}
	// Detach may return before WaitReady; continue when idle-policy.json exists.
	if runErr != nil && resp.PolicyJSON == "" {
		return resp, fmt.Errorf("detach: %w (exit=%d)\nstdout:\n%s\nstderr:\n%s", runErr, exit, stdout, stderr)
	}
	if resp.PolicyJSON == "" {
		resp.PolicyJSON = readPolicy(req)
	}

	t.Cleanup(func() {
		kctx, kcancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer kcancel()
		k := exec.CommandContext(kctx, req.AgentRun, "tty", "kill", req.SessionID)
		k.Dir = req.Workspace
		k.Env = append([]string(nil), req.Env...)
		_ = k.Run()
		cleanupRegistryServes(req.Home)
	})

	if req.Op == "draft-hold" {
		draft := strings.TrimSpace(req.Draft)
		if draft == "" {
			draft = defaultDraft
		}
		if !waitSendable(t, req) {
			return resp, fmt.Errorf("session never became sendable before draft inject\nstdout:\n%s\nstderr:\n%s\nstatus:\n%s", resp.DetachStdout, resp.DetachStderr, snapshotText(t, req))
		}
		dOut, dErr, _, dRunErr := execAgentRun(t, req, 60*time.Second,
			"tty", "send", "--no-submit", "--max-wait", "45s", req.SessionID, draft)
		resp.DraftStdout = dOut
		resp.DraftStderr = dErr
		if dRunErr != nil {
			return resp, fmt.Errorf("tty send --no-submit: %w\nstdout:\n%s\nstderr:\n%s", dRunErr, dOut, dErr)
		}
		resp.DraftInjected = true
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			snap := snapshotText(t, req)
			resp.DraftSnapshot = snap
			if strings.Contains(snap, draft) {
				break
			}
			time.Sleep(400 * time.Millisecond)
		}
		// Hold window starts after the draft is in the composer.
		time.Sleep(observe)
		fillStatus(t, req, resp)
		return resp, nil
	}

	// placeholder-exit: SoftExit /exit can mark screen=exited before __serve
	// drops TCP. Poll until not live so we do not snapshot mid-grace.
	deadline := time.Now().Add(observe + shutdownSlack)
	for {
		fillStatus(t, req, resp)
		if !resp.Alive {
			return resp, nil
		}
		if !time.Now().Before(deadline) {
			break
		}
		time.Sleep(time.Second)
	}
	fillStatus(t, req, resp)
	return resp, nil
}
```
