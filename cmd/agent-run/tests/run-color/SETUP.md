# Scenario

**Feature**: `agent-run run --color` forces TTY child color env last

```
# help
agent-run run --help -> documents --color

# color ON (TTY env-logger)
agent-run run --agent-runner grok-tty --color \
  --agent-runner-binary <env-logger> "prompt"
  (parent cmd.Env may have NO_COLOR / TERM=dumb / …)
  -> child: unset NO_COLOR; FORCE_COLOR=1 CLICOLOR=1 CLICOLOR_FORCE=1
  -> TERM rewritten only when empty/dumb → xterm-256color
  -> policy wins over -e NO_COLOR=…
  -> not stored on meta.json

# color ON non-TTY
--agent-runner fake-codex --color -> exit ≠ 0

# color OFF
run without --color -> no force from this feature
```

## Preconditions

- Repository contains `cmd/agent-run` (thin main → `pkgs/agentruncli`).
- Nested DOCTEST root: does not inherit parent `cmd/agent-run/tests` Setup/Run.
- Repo root from this tree: `d.DOCTEST_ROOT/../../../..`
  (`run-color` → `tests` → `agent-run` → `cmd` → module root).
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Session-scoped build cache:
  `$TMPDIR/agent-run-run-color-doctest-<d.DOCTEST_SESSION_ID>/` shares the
  compiled `agent-run` binary across parallel leaves.
- `frontend-agent-run/dist` (and `frontend/dist` if present) may be absent
  (gitignored). Build Setup stubs a minimal `index.html` so `//go:embed dist`
  compiles; these leaves do not serve UI assets.
- No real Grok CLI: env-logging fake runner via `--agent-runner-binary`.
- `AGENT_RUN_GROK_TTY_COMMAND` is cleared so the hook does not replace argv/env.
- **Parallel-safe:** no `t.Setenv` / process-global env mutation. Parent
  `NO_COLOR` / `TERM` for scenarios are applied only on the **agent-run child**
  `cmd.Env` (`req.Env` / `ParentTERM` / `ParentNoColor`).
- Host color keys (`NO_COLOR`, `FORCE_COLOR`, `CLICOLOR`, `CLICOLOR_FORCE`) are
  stripped from the agent-run process base env so leaves own those keys.

## Steps

1. Root `Setup` resolves repo root, builds `agent-run` once per session, sets
   isolated `AGENT_RUN_HOME`, strips grok-tty hook + color keys from cmd env.
2. Grouping / leaf `Setup` writes fake runners, sets parent env factors, finalizes
   `req.Args` (including optional `--color`).
3. `Run` executes `agent-run` and captures stdout/stderr/exit.
4. Leaf `Assert` checks exit code, env probe, and/or error text / meta absence.

## Context

- Default TTY runner: `grok-tty`.
- Color policy is applied last on the TTY agent runner child only.
- Probe keys: `NO_COLOR`, `FORCE_COLOR`, `CLICOLOR`, `CLICOLOR_FORCE`, `TERM`, `PATH`.

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
	"strings"
	"syscall"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

const (
	envGrokTTYCommand = "AGENT_RUN_GROK_TTY_COMMAND"
	defaultRunner     = "grok-tty"
)

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "agent-run-run-color-doctest-"+d.DOCTEST_SESSION_ID)
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
// //go:embed dist compiles when frontend dist is gitignored/absent.
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

func buildOnce(t *testing.T, d *session.Doctest) (agentRun string, err error) {
	t.Helper()
	cache := sessionCacheDir(d)
	agentRun = filepath.Join(cache, "agent-run")
	lock := filepath.Join(cache, "build.lock")
	ready := filepath.Join(cache, "binaries.ready")
	repoRoot := filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../.."))
	err = withFileLock(t, lock, func() error {
		if fileExists(ready) && fileExists(agentRun) {
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
		// Package path: thin main under cmd/agent-run (may need main.go form on some modules).
		build := exec.Command("go", "build", "-o", agentRun, "./cmd/agent-run")
		build.Dir = repoRoot
		if out, err := build.CombinedOutput(); err != nil {
			// Fallback: explicit main.go (some worktrees list package oddly).
			build2 := exec.Command("go", "build", "-o", agentRun, "./cmd/agent-run/main.go")
			build2.Dir = repoRoot
			if out2, err2 := build2.CombinedOutput(); err2 != nil {
				return fmt.Errorf("build agent-run: %w\n%s\nfallback: %v\n%s", err, string(out), err2, string(out2))
			}
		}
		return os.WriteFile(ready, []byte("ok"), 0644)
	})
	return agentRun, err
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

// mergeCmdEnv builds agent-run process env: host base, strip noise keys, last-win extra.
// Color-related keys and the grok-tty hook are stripped from the host base so leaves
// fully own them via req.Env (parallel-safe; no t.Setenv).
func mergeCmdEnv(extra []string) []string {
	base := append([]string(nil), os.Environ()...)
	for _, key := range []string{
		"NO_COLOR", "FORCE_COLOR", "CLICOLOR", "CLICOLOR_FORCE",
		envGrokTTYCommand, "GROK_HOME", "AGENT_RUNNER_CONFIG_HOME",
	} {
		base = withoutEnvKey(base, key)
	}
	for _, e := range extra {
		if key, _, ok := strings.Cut(e, "="); ok {
			base = withoutEnvKey(base, key)
		}
	}
	return append(base, extra...)
}

// writeEnvLoggingRunner dumps PATH and color-related keys to probePath, then
// prints a grok-tty banner and exits after one read (completes a short run).
func writeEnvLoggingRunner(t *testing.T, dir, probePath string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir fake-bin: %v", err)
	}
	path := filepath.Join(dir, "env-logger.sh")
	script := fmt.Sprintf(`#!/bin/sh
{
  echo "PATH=$PATH"
  env | grep -E '^(NO_COLOR|FORCE_COLOR|CLICOLOR|CLICOLOR_FORCE|TERM|FOO|A|NEW|BAZ|GROK_HOME|CODEX_HOME|AGENT_RUNNER_CONFIG_HOME)=' | sort
} > %q
printf "GROK_TTY_BANNER\nGrok › "
read -r line || true
echo "ENV_LOGGER_OK"
exit 0
`, probePath)
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write env logger: %v", err)
	}
	return path
}

func absPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatalf("abs %q: %v", p, err)
	}
	return abs
}

func metaJSONPath(home, sessionID string) string {
	return filepath.Join(home, "sessions", sessionID, "meta.json")
}

func readMetaJSON(t *testing.T, home, sessionID string) map[string]any {
	t.Helper()
	path := metaJSONPath(home, sessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read meta.json %s: %v", path, err)
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("parse meta.json: %v", err)
	}
	return obj
}

func readEnvProbe(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read env probe %s: %v", path, err)
	}
	return string(data)
}

func probeKV(probe, key string) (value string, ok bool) {
	prefix := key + "="
	for _, line := range strings.Split(probe, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, prefix) {
			return strings.TrimPrefix(line, prefix), true
		}
	}
	return "", false
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
	cmd.Env = mergeCmdEnv(env)
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

func assertNonZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0\nstderr:\n%s\nstdout:\n%s", resp.Stderr, resp.Stdout)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in:\n%s", want, got)
	}
}

func assertContainsAny(t *testing.T, got string, wants ...string) {
	t.Helper()
	for _, w := range wants {
		if strings.Contains(got, w) {
			return
		}
	}
	t.Fatalf("none of %v found in:\n%s", wants, got)
}

func assertTrailingNewline(t *testing.T, s, label string) {
	t.Helper()
	if s == "" || !strings.HasSuffix(s, "\n") {
		tail := s
		if len(tail) > 32 {
			tail = tail[len(tail)-32:]
		}
		t.Fatalf("%s must end with trailing newline; last bytes %q", label, tail)
	}
}

func assertProbeHasKV(t *testing.T, probe, key, value string) {
	t.Helper()
	want := key + "=" + value
	if !strings.Contains(probe, want) {
		t.Fatalf("env probe missing %q; probe:\n%s", want, probe)
	}
}

func assertProbeMissingKey(t *testing.T, probe, key string) {
	t.Helper()
	if _, ok := probeKV(probe, key); ok {
		t.Fatalf("env probe must not have key %q; probe:\n%s", key, probe)
	}
}

// assertColorForceOn checks the locked color-ON child env contract.
func assertColorForceOn(t *testing.T, probe string) {
	t.Helper()
	assertProbeMissingKey(t, probe, "NO_COLOR")
	assertProbeHasKV(t, probe, "FORCE_COLOR", "1")
	assertProbeHasKV(t, probe, "CLICOLOR", "1")
	assertProbeHasKV(t, probe, "CLICOLOR_FORCE", "1")
}

// assertMetaNoColorField ensures color is not persisted on meta.json.
func assertMetaNoColorField(t *testing.T, meta map[string]any) {
	t.Helper()
	for _, key := range []string{"color", "force_color", "forceColor", "Color"} {
		if _, ok := meta[key]; ok {
			t.Fatalf("meta must not persist color field %q; meta=%v", key, meta)
		}
	}
}

// prepareEnvLoggingRun writes the env-logger fake runner and records probe path.
func prepareEnvLoggingRun(t *testing.T, req *Request) {
	t.Helper()
	binDir := filepath.Join(req.TempDir, "fake-bin")
	req.EnvProbePath = filepath.Join(req.TempDir, "env-probe.log")
	req.RunnerScriptPath = writeEnvLoggingRunner(t, binDir, req.EnvProbePath)
	req.AgentRunnerBinary = req.RunnerScriptPath
	if req.Prompt == "" {
		req.Prompt = "run color probe"
	}
}

// applyParentEnvFactors appends ParentTERM / ParentNoColor onto req.Env.
func applyParentEnvFactors(req *Request) {
	if req.ParentNoColor {
		req.Env = withoutEnvKey(req.Env, "NO_COLOR")
		req.Env = append(req.Env, "NO_COLOR=1")
	}
	if strings.TrimSpace(req.ParentTERM) != "" {
		req.Env = withoutEnvKey(req.Env, "TERM")
		req.Env = append(req.Env, "TERM="+req.ParentTERM)
	}
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	if err := os.MkdirAll(req.Home, 0755); err != nil {
		return fmt.Errorf("mkdir home: %w", err)
	}
	cached, err := buildOnce(t, d)
	if err != nil {
		return err
	}
	binDir := filepath.Join(req.TempDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}
	req.AgentRun = filepath.Join(binDir, "agent-run")
	if out, err := exec.Command("cp", cached, req.AgentRun).CombinedOutput(); err != nil {
		return fmt.Errorf("cp agent-run: %w\n%s", err, string(out))
	}
	if err := os.Chmod(req.AgentRun, 0755); err != nil {
		return err
	}
	req.Env = append(req.Env,
		"AGENT_RUN_HOME="+req.Home,
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	req.Env = withoutEnvKey(req.Env, envGrokTTYCommand)
	req.Env = withoutEnvKey(req.Env, "GROK_HOME")
	req.Env = withoutEnvKey(req.Env, "AGENT_RUNNER_CONFIG_HOME")
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
