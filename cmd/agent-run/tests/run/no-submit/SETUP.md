# Scenario

**Feature**: `agent-run run --no-submit` (with required `--open`) injects a prompt
into the TTY without auto-submit Enter

```
# validation
agent-run run --no-submit --agent-runner grok-tty "x" -> error (--no-submit requires --open)
agent-run run --open --no-submit --agent-runner fake-codex "x" -> error (non-TTY)

# open + no-submit inject (TTY)
agent-run run --agent-runner grok-tty --open --no-submit "draft"
  -> silent open start + inject without \r
  -> auto-attach (AGENT_RUN_OPEN_ATTACH_INSTANT=1 in tests)
  -> on attach exit: stderr once "grok-tty: <id>"
  -> CR-sensitive fake TUI must not print SUBMITTED:draft
```

## Preconditions

- Repository contains `cmd/agent-run` and `cmd/fake-codex`.
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Non-TTY reject leaves use `fake-codex` on `PATH`.
- With-open inject leaf sets `AGENT_RUN_GROK_TTY_COMMAND` to a CR-sensitive fake
  TUI that prints `SUBMITTED:<line>` only after Enter, and sets
  `AGENT_RUN_OPEN_ATTACH_INSTANT=1` so auto-attach returns without a real
  controlling TTY (test contract; implementer must honor).
- Session-scoped build cache may share compiled binaries across parallel leaves.

## Steps

1. Root `Setup` resolves repo root, creates temp home, builds `agent-run` and
   `fake-codex` (session cache), sets `AGENT_RUN_HOME` + `PATH`.
2. Grouping `Setup` narrows outcome class (`help` / `reject` / `with-open`) and
   runner / flags.
3. Leaf `Setup` finalizes flags, prompt, TTY hooks, and open-attach instant env.
4. `Run` executes `agent-run` with `req.Args` (optional registry / snapshot post-read).
5. Leaf `Assert` checks exit code, error text, help text, or no-submit inject proof.

## Context

- Nested DOCTEST root: does not inherit parent `cmd/agent-run/tests` Setup/Run.
- Repo root from this tree: `DOCTEST_ROOT/../../../../..` (same depth as `run/open`).
- Session cache dir: `$TMPDIR/agent-run-run-no-submit-doctest-<DOCTEST_SESSION_ID>/`.
- Test hook env: `AGENT_RUN_OPEN_ATTACH_INSTANT=1` — auto-attach returns
  immediately so open leaves complete without interactive stdin.

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

	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
)

const (
	grokTTYBannerMarker  = "GROK_TTY_BANNER"
	envOpenAttachInstant = "AGENT_RUN_OPEN_ATTACH_INSTANT"
	envGrokTTYCommand    = "AGENT_RUN_GROK_TTY_COMMAND"
	envCodexTTYCommand   = "AGENT_RUN_CODEX_TTY_COMMAND"
)

func sessionCacheDir() string {
	return filepath.Join(os.TempDir(), "agent-run-run-no-submit-doctest-"+DOCTEST_SESSION_ID)
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

// ensureStubDist makes sure distDir has at least one embeddable file so
// //go:embed dist compiles. frontend-agent-run/dist is gitignored and may be
// absent in a fresh checkout; the CLI paths under test do not serve UI assets.
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
		// Satisfy //go:embed dist in frontend-agent-run (and frontend if linked).
		for _, rel := range []string{"frontend-agent-run/dist", "frontend/dist"} {
			if err := ensureStubDist(filepath.Join(repoRoot, rel)); err != nil {
				return fmt.Errorf("ensure %s stub: %w", rel, err)
			}
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

// fakeTUIRequiresCR prints SUBMITTED:<line> only after Enter completes the line.
// Without trailing \r from inject, read blocks and SUBMITTED never appears —
// proving --no-submit (suffixCR=false). Sleep after submit keeps PTY briefly
// alive if a buggy inject still submits.
func fakeTUIRequiresCR() string {
	return `sh -c 'printf "GROK_TTY_BANNER\nGrok › "; read -r line; echo "SUBMITTED:$line"; sleep 30'`
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
		needRegistry := req.Mode == "open-registry-after" || req.Mode == "open-snapshot-after"
		if needRegistry && resp.SessionID != "" && resp.ExitCode == 0 {
			runner := req.Runner
			if runner == "" {
				runner = "grok-tty"
			}
			entry, rerr := readRegistryEntryOptional(req.Home, runner, resp.SessionID)
			if rerr == nil {
				resp.RegistryEntry = entry
			}
		}
		if req.Mode == "open-snapshot-after" && resp.RegistryEntry != nil && resp.RegistryEntry.ListenAddr != "" && resp.SessionID != "" {
			// Brief settle so a buggy submit-with-CR would print SUBMITTED before we snapshot.
			time.Sleep(400 * time.Millisecond)
			if text, serr := ttywatch.SnapshotText(resp.RegistryEntry.ListenAddr, resp.SessionID); serr == nil {
				resp.Snapshot = text
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
