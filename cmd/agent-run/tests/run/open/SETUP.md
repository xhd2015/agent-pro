# Scenario

**Feature**: `agent-run run --open` opens a keep-alive TTY session, auto-attaches
silently, then prints the terminal session id after detach

```
# validation
agent-run run --open --agent-runner fake-codex "x" -> error (non-TTY)
agent-run run --open --json --agent-runner grok-tty "x" -> error
agent-run run --agent-runner grok-tty -> prompt is required
agent-run run --agent-runner grok-tty --open -> empty prompt allowed

# open lifecycle (TTY)
agent-run run --agent-runner grok-tty --open ["prompt"]
  -> silent start + auto-attach
  -> on attach exit: stderr once "grok-tty: <id>"
  -> registry/PTY kept alive
```

## Preconditions

- Repository contains `cmd/agent-run` and `cmd/fake-codex`.
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Non-TTY reject leaves use `fake-codex` on `PATH`.
- TTY lifecycle leaves set `AGENT_RUN_GROK_TTY_COMMAND` (or codex hook) to a fake
  interactive TUI that prints `GROK_TTY_BANNER` then reads/holds.
- Lifecycle leaves set `AGENT_RUN_OPEN_ATTACH_INSTANT=1` so auto-attach returns
  without a real controlling TTY (test contract; implementer must honor).
- Session-scoped build cache may share compiled binaries across parallel leaves.
- Discovery hang / “Resolve session id…” **fixing** is out of scope; open mode
  must simply not print that progress to the screen.

## Steps

1. Root `Setup` resolves repo root, creates temp home, builds `agent-run` and
   `fake-codex` (session cache), sets `AGENT_RUN_HOME` + `PATH`.
2. Grouping `Setup` narrows outcome class (`help` / `reject` / `prompt-policy` /
   `tty-lifecycle`) and runner.
3. Leaf `Setup` finalizes flags, prompt, TTY hooks, and open-attach instant env.
4. `Run` executes `agent-run` with `req.Args` (optional registry post-read).
5. Leaf `Assert` checks exit code, error text, silence, session id line, registry.

## Context

- Nested DOCTEST root: does not inherit parent `cmd/agent-run/tests` Setup/Run.
- Repo root from this tree: `DOCTEST_ROOT/../../../../..`.
- Session cache dir: `$TMPDIR/agent-run-run-open-doctest-<DOCTEST_SESSION_ID>/`.
- Test hook env: `AGENT_RUN_OPEN_ATTACH_INSTANT=1` — auto-attach returns
  immediately so lifecycle leaves complete without interactive stdin.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"
)

const (
	grokTTYBannerMarker     = "GROK_TTY_BANNER"
	envOpenAttachInstant    = "AGENT_RUN_OPEN_ATTACH_INSTANT"
	envGrokTTYCommand       = "AGENT_RUN_GROK_TTY_COMMAND"
	envCodexTTYCommand      = "AGENT_RUN_CODEX_TTY_COMMAND"
)

func sessionCacheDir() string {
	return filepath.Join(os.TempDir(), "agent-run-run-open-doctest-"+DOCTEST_SESSION_ID)
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

func buildOnce(t *testing.T) (agentRun, fakeCodex string, err error) {
	t.Helper()
	cache := sessionCacheDir()
	agentRun = filepath.Join(cache, "agent-run")
	fakeCodex = filepath.Join(cache, "fake-codex")
	lock := filepath.Join(cache, "build.lock")
	ready := filepath.Join(cache, "binaries.ready")
	repoRoot := filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../../.."))
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

func fakeTUIRespondHi() string {
	return `sh -c 'printf "GROK_TTY_BANNER\nGrok › "; read line; echo "Response: $line"'`
}

func fakeTUIHoldSeconds(sec int) string {
	return fmt.Sprintf(`sh -c 'printf "GROK_TTY_BANNER\nGrok › "; sleep %d'`, sec)
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

func setGrokTTYCommand(req *Request, cmd string) {
	req.GrokTTYCommand = cmd
	setEnvKV(req, envGrokTTYCommand, cmd)
}

func setCodexTTYCommand(req *Request, cmd string) {
	req.CodexTTYCommand = cmd
	setEnvKV(req, envCodexTTYCommand, cmd)
}

func applyOpenInstantAttach(req *Request) {
	if !req.OpenInstantAttach {
		return
	}
	setEnvKV(req, envOpenAttachInstant, "1")
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
	applyOpenInstantAttach(req)
	resp, err := execCmd(t, req.AgentRun, args, req.TempDir, req.Env, req.ExecTimeout)
	if err != nil {
		return resp, err
	}
	if resp != nil {
		if id, ok := parsePrefixedSessionID(resp.Stderr, "grok-tty"); ok {
			resp.SessionID = id
		} else if id, ok := parsePrefixedSessionID(resp.Stderr, "codex-tty"); ok {
			resp.SessionID = id
		}
		if req.Mode == "open-registry-after" && resp.SessionID != "" && resp.ExitCode == 0 {
			runner := req.Runner
			if runner == "" {
				runner = "grok-tty"
			}
			entry, rerr := readRegistryEntryOptional(req.Home, runner, resp.SessionID)
			if rerr == nil {
				resp.RegistryEntry = entry
			}
		}
	}
	return resp, nil
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

// parsePrefixedSessionID extracts "<runner>: <id>" where id is a single token.
// Skips multi-word diagnostic lines (e.g. "grok-tty: grok session …").
func parsePrefixedSessionID(stderr, runner string) (string, bool) {
	prefix := runner + ":"
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, prefix))
		if rest == "" || strings.Contains(rest, " ") {
			continue
		}
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`, rest); matched {
			return rest, true
		}
	}
	return "", false
}

func countPrefixedSessionIDLines(stderr, runner string) int {
	prefix := runner + ":"
	n := 0
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, prefix))
		if rest == "" || strings.Contains(rest, " ") {
			continue
		}
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`, rest); matched {
			n++
		}
	}
	return n
}

func registryDir(home, runner string) string {
	return filepath.Join(home, runner+"-registry")
}

func registryPath(home, runner, sessionID string) string {
	return filepath.Join(registryDir(home, runner), sessionID+".json")
}

func readRegistryEntryOptional(home, runner, sessionID string) (*RegistryEntry, error) {
	path := registryPath(home, runner, sessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entry RegistryEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	if entry.SessionID == "" {
		entry.SessionID = sessionID
	}
	return &entry, nil
}

func portOpen(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// forbiddenOpenNoise reports human-facing discovery/event stream noise that
// must not appear on stdout/stderr during --open.
func forbiddenOpenNoise(combined string) []string {
	checks := []string{
		"Resolve session id",
		"💭",
		"💬",
		"[done]",
		`"type":"think"`,
		`"type":"message"`,
		`"type":"done"`,
	}
	var found []string
	for _, c := range checks {
		if strings.Contains(combined, c) {
			found = append(found, c)
		}
	}
	return found
}

func Setup(t *testing.T, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	if err := os.MkdirAll(req.Home, 0755); err != nil {
		return fmt.Errorf("mkdir home: %w", err)
	}
	agentRun, fakeCodex, err := buildOnce(t)
	if err != nil {
		return err
	}
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
