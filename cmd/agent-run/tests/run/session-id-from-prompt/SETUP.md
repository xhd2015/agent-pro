# Scenario

**Feature**: `agent-run run --session-id-from-prompt` generates a shared storage/TTY session id

```
# auto-id path
agent-run run --session-id-from-prompt "prompt"
  -> slug-YYYYMMDD-HHMMSS[-N]
  -> sessions/<runner>/<id>/
  -> [TTY] grok-tty-registry/<id>.json + stderr grok-tty: <id>

# explicit session same-id policy (TTY)
agent-run run --session my-task "prompt"
  -> sessions/<runner>/my-task/
  -> [TTY] registry + stderr use my-task
```

## Preconditions

- Repository contains `cmd/agent-run` and `cmd/fake-codex`.
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Non-TTY leaves use `fake-codex` on `PATH`.
- TTY leaves set `AGENT_RUN_GROK_TTY_COMMAND` to a fake interactive TUI that
  prints `GROK_TTY_BANNER` then reads one line and exits.
- Session-scoped build cache may share compiled binaries across parallel leaves.

## Steps

1. Root `Setup` resolves repo root, creates temp home, builds `agent-run` and
   `fake-codex` (session cache), sets `AGENT_RUN_HOME` + `PATH`.
2. Grouping `Setup` narrows session-id policy (`auto-id` / `explicit-session` /
   `flag-conflict` / `help`) and runner class.
3. Leaf `Setup` finalizes flags, prompt, collision seeds, and TTY hook.
4. `Run` executes `agent-run` with `req.Args`.
5. Leaf `Assert` checks exit code, storage paths, id shape, stderr, registry, meta.

## Context

- Nested DOCTEST root: does not inherit parent `cmd/agent-run/tests` Setup/Run.
- Repo root from this tree: `d.DOCTEST_ROOT/../../../../..`.
- Session cache dir: `$TMPDIR/agent-run-session-id-from-prompt-doctest-<d.DOCTEST_SESSION_ID>/`.

```go
import (
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
	"unicode/utf8"
	"github.com/xhd2015/doctest/session"
)

const grokTTYBannerMarker = "GROK_TTY_BANNER"

// autoSessionIDShape matches generated ids: base-YYYYMMDD-HHMMSS or …-N.
// Base starts with [a-z0-9] and may contain [a-z0-9._-]; timestamp is 8+6 digits.
var autoSessionIDShape = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*-\d{8}-\d{6}(-\d+)?$`)

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "agent-run-session-id-from-prompt-doctest-"+d.DOCTEST_SESSION_ID)
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

func buildOnce(t *testing.T, d *session.Doctest) (agentRun, fakeCodex string, err error) {
	t.Helper()
	cache := sessionCacheDir(d)
	agentRun = filepath.Join(cache, "agent-run")
	fakeCodex = filepath.Join(cache, "fake-codex")
	lock := filepath.Join(cache, "build.lock")
	ready := filepath.Join(cache, "binaries.ready")
	repoRoot := filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../../.."))
	err = withFileLock(t, lock, func() error {
		if fileExists(ready) && fileExists(agentRun) && fileExists(fakeCodex) {
			return nil
		}
		if err := os.MkdirAll(cache, 0755); err != nil {
			return err
		}
		build := exec.Command("go", "build", "-o", agentRun, "./cmd/agent-run")
		build.Dir = repoRoot
		if out, err := build.CombinedOutput(); err != nil {
			return fmt.Errorf("build agent-run: %w\n%s", err, string(out))
		}
		build2 := exec.Command("go", "build", "-o", fakeCodex, "./cmd/fake-codex")
		build2.Dir = repoRoot
		if out, err := build2.CombinedOutput(); err != nil {
			return fmt.Errorf("build fake-codex: %w\n%s", err, string(out))
		}
		return os.WriteFile(ready, []byte("ok"), 0644)
	})
	return agentRun, fakeCodex, err
}

// fakeTUIRespondHi prints the grok banner and a turn response without blocking on
// stdin. Headless new sessions put the prompt on argv and do not re-inject, so a
// `read`-based fake hangs forever under --keep-tty wait-for-turn.
func fakeTUIRespondHi() string {
	return `sh -c 'printf "GROK_TTY_BANNER\nGrok › \nResponse: hi\n› "; sleep 0.2'`
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

func setGrokTTYCommand(req *Request, cmd string) {
	req.GrokTTYCommand = cmd
	req.Env = withoutEnvKey(req.Env, "AGENT_RUN_GROK_TTY_COMMAND")
	req.Env = append(req.Env, "AGENT_RUN_GROK_TTY_COMMAND="+cmd)
}

func execCmd(t *testing.T, command string, args []string, dir string, env []string, timeout time.Duration) (*Response, error) {
	t.Helper()
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}
	if err == nil {
		return resp, nil
	}
	if ctx.Err() != nil {
		return resp, ctx.Err()
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		resp.ExitCode = exitErr.ExitCode()
		return resp, nil
	}
	return resp, err
}

func runAgentRun(t *testing.T, req *Request, args ...string) (*Response, error) {
	t.Helper()
	if len(args) == 0 {
		args = req.Args
	}
	return execCmd(t, req.AgentRun, args, req.TempDir, req.Env, req.ExecTimeout)
}

func assertSuccess(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s\nstdout:\n%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

// parseGrokTTYSessionID extracts the registry session id from stderr.
// Skips multi-word lines like "grok-tty: grok session …" / "grok updates …".
func parseGrokTTYSessionID(stderr string) (string, bool) {
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		const prefix = "grok-tty:"
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, prefix))
		if rest == "" {
			continue
		}
		// Multi-word diagnostic lines are not the registry id.
		if strings.Contains(rest, " ") {
			continue
		}
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`, rest); matched {
			return rest, true
		}
	}
	return "", false
}

// Flat layout: sessions/<session_id>/ (runner is meta only; keep runner args for call sites).

func sessionsRootDir(home string) string {
	return filepath.Join(home, "sessions")
}

// sessionsRunnerDir is retained for call-site compatibility; flat layout ignores runner path segment.
func sessionsRunnerDir(home, runner string) string {
	_ = runner
	return sessionsRootDir(home)
}

func sessionDir(home, runner, id string) string {
	_ = runner
	return filepath.Join(sessionsRootDir(home), id)
}

func sessionMetaPath(home, runner, id string) string {
	return filepath.Join(sessionDir(home, runner, id), "meta.json")
}

func grokTTYRegistryDir(home string) string {
	return filepath.Join(home, "grok-tty-registry")
}

func grokTTYRegistryPath(home, sessionID string) string {
	return filepath.Join(grokTTYRegistryDir(home), sessionID+".json")
}

// listSessionIDs returns session directory names under sessions/ (flat layout).
// When runner is non-empty, only ids whose meta.runner matches are returned.
func listSessionIDs(t *testing.T, home, runner string) []string {
	t.Helper()
	dir := sessionsRootDir(home)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("readdir %s: %v", dir, err)
	}
	var ids []string
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		id := e.Name()
		if runner != "" {
			path := sessionMetaPath(home, runner, id)
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			var m sessionMeta
			if err := json.Unmarshal(data, &m); err != nil {
				continue
			}
			if strings.TrimSpace(m.Runner) != runner {
				continue
			}
		}
		ids = append(ids, id)
	}
	return ids
}

// singleSessionID expects exactly one matching session directory.
func singleSessionID(t *testing.T, home, runner string) string {
	t.Helper()
	ids := listSessionIDs(t, home, runner)
	if len(ids) != 1 {
		t.Fatalf("expected exactly 1 session under sessions/ (runner=%s), got %d: %v", runner, len(ids), ids)
	}
	return ids[0]
}

type sessionMeta struct {
	Runner            string `json:"runner"`
	SessionID         string `json:"session_id"`
	TerminalSessionID string `json:"terminal_session_id"`
	Status            string `json:"status"`
}

func readSessionMeta(t *testing.T, home, runner, id string) sessionMeta {
	t.Helper()
	path := sessionMetaPath(home, runner, id)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read meta %s: %v", path, err)
	}
	var m sessionMeta
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse meta %s: %v\n%s", path, err, string(data))
	}
	return m
}

// seedStorageSession creates sessions/<id>/meta.json so the id is taken (flat layout).
func seedStorageSession(t *testing.T, home, runner, id string) {
	t.Helper()
	dir := sessionDir(home, runner, id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir seed session %s: %v", dir, err)
	}
	meta := map[string]any{
		"runner":     runner,
		"session_id": id,
		"status":     "finished",
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}
	b, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal seed meta: %v", err)
	}
	path := filepath.Join(dir, "meta.json")
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("write seed meta %s: %v", path, err)
	}
}

// seedStorageCollisionsForBase pre-creates sessions for base-YYYYMMDD-HHMMSS
// across a wall-clock window so the next auto-id must append -N.
func seedStorageCollisionsForBase(t *testing.T, home, runner, base string) {
	t.Helper()
	now := time.Now()
	// Wide window: parallel slow machines + second boundary.
	for d := -5; d <= 30; d++ {
		ts := now.Add(time.Duration(d) * time.Second).Format("20060102-150405")
		seedStorageSession(t, home, runner, base+"-"+ts)
	}
}

// splitAutoSessionID separates base, timestamp, and optional numeric suffix.
// Returns ok=false if shape is wrong.
func splitAutoSessionID(id string) (base, ts, suffix string, ok bool) {
	if !autoSessionIDShape.MatchString(id) {
		return "", "", "", false
	}
	// Optional trailing -N (not part of timestamp).
	reSuf := regexp.MustCompile(`^(.*)-(\d{8}-\d{6})-(\d+)$`)
	if m := reSuf.FindStringSubmatch(id); m != nil {
		return m[1], m[2], m[3], true
	}
	re := regexp.MustCompile(`^(.*)-(\d{8}-\d{6})$`)
	m := re.FindStringSubmatch(id)
	if m == nil {
		return "", "", "", false
	}
	return m[1], m[2], "", true
}

func runeCount(s string) int {
	return utf8.RuneCountInString(s)
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	if err := os.MkdirAll(req.Home, 0755); err != nil {
		return fmt.Errorf("mkdir home: %w", err)
	}
	agentRun, fakeCodex, err := buildOnce(t, d)
	if err != nil {
		return err
	}
	// Copy/link into per-leaf bin for PATH isolation (reuse session-built bits).
	binDir := filepath.Join(req.TempDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}
	req.AgentRun = filepath.Join(binDir, "agent-run")
	req.FakeCodex = filepath.Join(binDir, "fake-codex")
	if out, err := exec.Command("cp", agentRun, req.AgentRun).CombinedOutput(); err != nil {
		return fmt.Errorf("cp agent-run: %w\n%s", err, string(out))
	}
	if out, err := exec.Command("cp", fakeCodex, req.FakeCodex).CombinedOutput(); err != nil {
		return fmt.Errorf("cp fake-codex: %w\n%s", err, string(out))
	}
	if err := os.Chmod(req.AgentRun, 0755); err != nil {
		return err
	}
	if err := os.Chmod(req.FakeCodex, 0755); err != nil {
		return err
	}
	req.Env = append(req.Env,
		"AGENT_RUN_HOME="+req.Home,
		"PATH="+binDir+":"+os.Getenv("PATH"),
	)
	req.Args = []string{"run"}
	return nil
}
```
