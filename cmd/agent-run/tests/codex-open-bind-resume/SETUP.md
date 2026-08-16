# Scenario

**Feature**: codex-tty open must bind `runner_session_id`; auto-send-or-resume
must resume the same Codex id (not open a second conversation).

```
build agent-run + llm-mock + llm-mock-run-codex
  -> isolated AGENT_RUN_HOME + CODEX_HOME + workspace
  -> AGENT_RUN_OPEN_ATTACH_INSTANT=1
  -> LLM_MOCK_CONFIG_FILE exchanges

agent-run run --open --session-id S ... "OPEN_MARKER"
  -> meta.runner_session_id bound; one rollout

end (send /exit; kill serve if needed)

agent-run run --auto-send-or-resume --open ... "FOLLOW_UP"
  -> same runner_session_id / same codex rollout uuid
```

## Preconditions

- Nested DOCTEST root (does not inherit `cmd/agent-run/tests` Setup/Run).
- Repo root: `d.DOCTEST_ROOT/../../../..`
  (`codex-open-bind-resume` → `tests` → `agent-run` → `cmd` → module root).
- Real `codex` on PATH (else `t.Skip`).
- Build:
  - `agent-run` from nested module `cmd/` (`go build ./agent-run`)
  - `llm-mock` + `llm-mock-run-codex` from repo root

## Steps

1. Root Setup: homes, mock config, binaries, skip without codex.
2. Leaf Setup: session id / prompts.
3. Run: open → record meta + rollouts → end → auto-send-or-resume → record again.
4. Assert: bound after open; same codex id after resume.

## Context

```go
import (
	"runtime"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	envOpenAttachInstant = "AGENT_RUN_OPEN_ATTACH_INSTANT"
	envLLMMockCodexHome  = "LLM_MOCK_CODEX_HOME"
	envLLMMockConfigFile = "LLM_MOCK_CONFIG_FILE"

	defaultSessionID = "codex-open-bind-resume-s1"
	defaultOpenPrompt = "OPEN_MARKER"
	defaultFollowup   = "FOLLOW_UP"

	defaultExecTimeout = 2 * time.Minute
)

// Types Request / CmdResult / Response live in DOCTEST.md (single declaration).

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "codex-open-bind-resume-doctest-"+d.DOCTEST_SESSION_ID)
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
		// agent-run lives under nested cmd module (cmd/go.mod).
		cmdDir := filepath.Join(repoRoot, "cmd")
		b1 := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", agentRun, "./agent-run")
		b1.Dir = cmdDir
		if out, err := b1.CombinedOutput(); err != nil {
			return fmt.Errorf("build agent-run: %w\n%s", err, out)
		}
		// llm-mock packages live in the root module.
		for _, b := range []struct {
			out string
			pkg string
		}{
			{llmMock, "./agent/llm/llm-mock"},
			{llmMockRunCodex, "./agent/llm/llm-mock/llm-mock-run-codex"},
		} {
			cmd := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", b.out, b.pkg)
			cmd.Dir = repoRoot
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

func writeDefaultMockConfig(t *testing.T, path string) {
	t.Helper()
	const body = `{
  "exchanges": [
    {
      "request": {"role": "user", "content": "OPEN_MARKER", "index": -1},
      "response": {"content": "OPEN_MOCK_REPLY", "finish_reason": "stop"}
    },
    {
      "request": {"role": "user", "content": "FOLLOW_UP", "index": -1},
      "response": {"content": "FOLLOW_UP_MOCK_REPLY", "finish_reason": "stop"}
    }
  ]
}
`
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write mock config: %v", err)
	}
}

func baseEnv(req *Request) []string {
	env := append([]string(nil), req.Env...)
	if len(env) == 0 {
		env = os.Environ()
	}
	// Prefer leaf bin for sibling llm-mock next to llm-mock-run-codex.
	binDir := filepath.Dir(req.LLMMockRunCodex)
	path := binDir + string(os.PathListSeparator) + os.Getenv("PATH")
	env = withoutEnvKey(env, "PATH")
	env = append(env, "PATH="+path)
	env = withoutEnvKey(env, "AGENT_RUN_HOME")
	env = append(env, "AGENT_RUN_HOME="+req.Home)
	env = withoutEnvKey(env, envLLMMockConfigFile)
	env = append(env, envLLMMockConfigFile+"="+req.MockConfigFile)
	env = withoutEnvKey(env, envLLMMockCodexHome)
	env = append(env, envLLMMockCodexHome+"="+req.CodexHome)
	env = withoutEnvKey(env, "CODEX_HOME")
	env = append(env, "CODEX_HOME="+req.CodexHome)
	env = withoutEnvKey(env, envOpenAttachInstant)
	env = append(env, envOpenAttachInstant+"=1")
	// Real Codex UI — do not inject fake TTY command.
	env = withoutEnvKey(env, "AGENT_RUN_CODEX_TTY_COMMAND")
	return env
}

func execAgentRun(t *testing.T, req *Request, timeout time.Duration, args ...string) CmdResult {
	t.Helper()
	if timeout <= 0 {
		timeout = req.ExecTimeout
	}
	if timeout <= 0 {
		timeout = defaultExecTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.Workspace
	cmd.Env = baseEnv(req)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	res := CmdResult{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
	if err == nil {
		return res
	}
	if ctx.Err() != nil {
		res.Err = ctx.Err()
		res.ExitCode = -1
		return res
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		res.ExitCode = ee.ExitCode()
		res.Err = nil // process ran; inspect ExitCode
		return res
	}
	return res
}

func metaPath(req *Request) string {
	return filepath.Join(req.Home, "sessions", req.SessionID, "meta.json")
}

func readMeta(req *Request) (map[string]any, string) {
	data, err := os.ReadFile(metaPath(req))
	if err != nil {
		return nil, ""
	}
	var m map[string]any
	if json.Unmarshal(data, &m) != nil {
		return nil, ""
	}
	rid, _ := m["runner_session_id"].(string)
	return m, strings.TrimSpace(rid)
}

var rolloutUUIDRe = regexp.MustCompile(`rollout-.*-([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\.jsonl$`)

func listCodexIDs(codexHome string) []string {
	var ids []string
	_ = filepath.Walk(filepath.Join(codexHome, "sessions"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		m := rolloutUUIDRe.FindStringSubmatch(filepath.Base(path))
		if len(m) == 2 {
			ids = append(ids, m[1])
		}
		return nil
	})
	// stable unique order
	seen := map[string]bool{}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func statusText(t *testing.T, req *Request) string {
	t.Helper()
	r := execAgentRun(t, req, 20*time.Second, "status", req.SessionID)
	return r.Stdout + r.Stderr
}

func parseServePID(status string) int {
	// "  pid:     17050"
	re := regexp.MustCompile(`(?m)^\s*pid:\s+(\d+)\s*$`)
	m := re.FindStringSubmatch(status)
	if len(m) < 2 {
		return 0
	}
	var n int
	fmt.Sscanf(m[1], "%d", &n)
	return n
}

func killServeIfAlive(t *testing.T, req *Request, status string) bool {
	t.Helper()
	pid := parseServePID(status)
	if pid <= 0 {
		return false
	}
	// Best-effort: children then serve (simulate terminal close / OpenCloseExits).
	_ = exec.Command("pkill", "-P", fmt.Sprintf("%d", pid)).Run()
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	_ = proc.Signal(syscall.SIGTERM)
	time.Sleep(500 * time.Millisecond)
	_ = proc.Signal(syscall.SIGKILL)
	return true
}

func cleanupRegistryServes(home string) {
	// Kill any leftover registry PIDs under codex-tty-registry.
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

func openArgs(req *Request) []string {
	return []string{
		"run",
		"--agent-runner=codex-tty",
		"--agent-runner-binary=" + req.LLMMockRunCodex,
		"--agent-runner-config-home=" + req.CodexHome,
		"--session-id=" + req.SessionID,
		"--dir=" + req.Workspace,
		"--open",
		"--",
		req.OpenPrompt,
	}
}

func resumeAutoArgs(req *Request) []string {
	return []string{
		"run",
		"--agent-runner=codex-tty",
		"--agent-runner-binary=" + req.LLMMockRunCodex,
		"--agent-runner-config-home=" + req.CodexHome,
		"--session-id=" + req.SessionID,
		"--dir=" + req.Workspace,
		"--allow-relocate-resume-session-dir",
		"--auto-send-or-resume",
		"--open",
		"--",
		req.FollowupPrompt,
	}
}

func runOpenThenAutoResume(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	start := time.Now()
	resp := &Response{SessionID: req.SessionID}
	t.Cleanup(func() { cleanupRegistryServes(req.Home) })

	// 1) Open
	resp.Open = execAgentRun(t, req, 90*time.Second, openArgs(req)...)
	// Brief settle for rollout file + status.
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		ids := listCodexIDs(req.CodexHome)
		_, rid := readMeta(req)
		if len(ids) > 0 {
			resp.CodexIDsAfterOpen = ids
			resp.RunnerSessionIDAfterOpen = rid
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if len(resp.CodexIDsAfterOpen) == 0 {
		resp.CodexIDsAfterOpen = listCodexIDs(req.CodexHome)
	}
	resp.MetaAfterOpen, resp.RunnerSessionIDAfterOpen = readMeta(req)
	resp.StatusAfterOpen = statusText(t, req)

	// 2) End: try /exit, then kill serve (terminal-close simulation).
	resp.SendExit = execAgentRun(t, req, 45*time.Second, "send", "--max-wait", "30s", req.SessionID, "/exit")
	time.Sleep(1 * time.Second)
	st := statusText(t, req)
	// Always try force end so auto-send-or-resume is not ModeSend into a live TTY.
	if killServeIfAlive(t, req, st) {
		resp.KilledServe = true
		time.Sleep(1 * time.Second)
	}
	resp.StatusAfterEnd = statusText(t, req)

	// 3) Auto-send-or-resume (product path for local-bot ForceNew follow-up)
	resp.Resume = execAgentRun(t, req, 90*time.Second, resumeAutoArgs(req)...)
	deadline = time.Now().Add(30 * time.Second)
	before := append([]string(nil), resp.CodexIDsAfterOpen...)
	for time.Now().Before(deadline) {
		ids := listCodexIDs(req.CodexHome)
		if len(ids) > len(before) {
			resp.CodexIDsAfterResume = ids
			break
		}
		// even if count same, re-sample after a short wait
		resp.CodexIDsAfterResume = ids
		time.Sleep(500 * time.Millisecond)
	}
	if len(resp.CodexIDsAfterResume) == 0 {
		resp.CodexIDsAfterResume = listCodexIDs(req.CodexHome)
	}
	resp.MetaAfterResume, resp.RunnerSessionIDAfterResume = readMeta(req)
	resp.StatusAfterResume = statusText(t, req)
	resp.Elapsed = time.Since(start)
	return resp, nil
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
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
	if err := os.WriteFile(filepath.Join(req.Workspace, "README.md"), []byte("codex-open-bind-resume\n"), 0644); err != nil {
		return err
	}

	req.AgentRun, req.LLMMock, req.LLMMockRunCodex = ensureSessionBinaries(t, d, req.RepoRoot)
	// Place sibling binaries in one dir so llm-mock-run-codex finds llm-mock next to itself.
	// Session cache already has both; also symlink into leaf bin for clarity.
	leafBin := filepath.Join(req.TempDir, "bin")
	if err := os.MkdirAll(leafBin, 0755); err != nil {
		return err
	}
	_ = os.Symlink(req.AgentRun, filepath.Join(leafBin, "agent-run"))
	// Hard-copy or re-link mock pair into same directory (required: sibling lookup).
	// Use the session-cache paths as-is: they already sit in the same cache dir.
	if filepath.Dir(req.LLMMock) != filepath.Dir(req.LLMMockRunCodex) {
		return fmt.Errorf("llm-mock and llm-mock-run-codex must share a directory")
	}

	req.MockConfigFile = filepath.Join(req.TempDir, "llm-mock-config.json")
	writeDefaultMockConfig(t, req.MockConfigFile)

	if req.SessionID == "" {
		req.SessionID = defaultSessionID
	}
	if req.OpenPrompt == "" {
		req.OpenPrompt = defaultOpenPrompt
	}
	if req.FollowupPrompt == "" {
		req.FollowupPrompt = defaultFollowup
	}
	if req.ExecTimeout <= 0 {
		req.ExecTimeout = defaultExecTimeout
	}
	return nil
}

```
