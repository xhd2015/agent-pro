# Scenario

**Feature**: session-scoped `--prepend-path`, `-e`/`--env`, and persisted
`--agent-runner-config-home` for TTY runners (apply + meta + resume)

```
# run: flags → child Env + meta.json
agent-run run --agent-runner grok-tty \
  --prepend-path DIR -e KEY=VALUE --agent-runner-config-home PATH \
  --agent-runner-binary <env-logger> "prompt"
  -> PTY child PATH / KEY / GROK_HOME
  -> sessions/<id>/meta.json prepend_paths, env, agent_runner_config_home

# resume: reapply stored; optional append flags
seed bound+exited meta (+ prepend_paths/env/config_home)
  -> agent-run resume [--prepend-path|--env] <id> "followup"
  -> child Env from stored (+ append); meta updated
```

## Preconditions

- Repository contains `cmd/agent-run`.
- Nested DOCTEST root: does not inherit parent `cmd/agent-run/tests` Setup/Run.
- Repo root from this tree: `d.DOCTEST_ROOT/../../../..`
  (`session-env` → `tests` → `agent-run` → `cmd` → module root).
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Session-scoped build cache:
  `$TMPDIR/agent-run-session-env-doctest-<d.DOCTEST_SESSION_ID>/` shares the
  compiled `agent-run` binary across parallel leaves.
- `frontend-agent-run/dist` (and `frontend/dist` if present) may be absent
  (gitignored). Build Setup stubs a minimal `index.html` so `//go:embed dist`
  compiles; these leaves do not serve UI assets.
- No real Grok CLI: env-logging / hold fake runners via `--agent-runner-binary`.
- `AGENT_RUN_GROK_TTY_COMMAND` is cleared so the hook does not replace argv/env.
- Session layout is **flat**: `AGENT_RUN_HOME/sessions/<session_id>/meta.json`.

## Steps

1. Root `Setup` resolves repo root, builds `agent-run` once per session, sets
   isolated `AGENT_RUN_HOME`, strips grok-tty hook env.
2. Grouping / leaf `Setup` writes fake runners, seeds meta (resume), finalizes
   `req.Args`.
3. `Run` executes `agent-run` and captures stdout/stderr/exit.
4. Leaf `Assert` checks exit code, env probe, and/or meta.json fields.

## Context

- Default TTY runner: `grok-tty`.
- Child env contract: process `cmd.Env` (not parent `os.Setenv`); PATH prefix
  order is stored `prepend_paths` then any resume appends; env last-win per key.
- Meta keys: `prepend_paths`, `env`, `agent_runner_config_home` (abs paths when stored).

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
	"strings"
	"syscall"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

const (
	envGrokTTYCommand = "AGENT_RUN_GROK_TTY_COMMAND"
	defaultRunner     = "grok-tty"
	defaultModel      = "test-model"
)

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "agent-run-session-env-doctest-"+d.DOCTEST_SESSION_ID)
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
	// DistComplete needs non-empty index.html and at least one assets/* file.
	// placeholder.txt alone is not enough; always ensure a minimal SPA shell.
	if err := os.MkdirAll(filepath.Join(distDir, "assets"), 0755); err != nil {
		return err
	}
	const shell = `<!doctype html>
<html lang="en">
<head><meta charset="UTF-8"><title>agent-run</title></head>
<body>
<div id="root"></div>
</body>
</html>
`
	indexPath := filepath.Join(distDir, "index.html")
	needIndex := true
	if data, err := os.ReadFile(indexPath); err == nil {
		s := string(data)
		if strings.Contains(s, `id="root"`) || strings.Contains(s, "id='root'") {
			needIndex = false
		}
	}
	if needIndex {
		if err := os.WriteFile(indexPath, []byte(shell), 0644); err != nil {
			return err
		}
	}
	assetPath := filepath.Join(distDir, "assets", "doctest-stub.js")
	if st, err := os.Stat(assetPath); err != nil || st.Size() == 0 {
		if err := os.WriteFile(assetPath, []byte("/* doctest stub */\n"), 0644); err != nil {
			return err
		}
	}
	return nil
}

func buildOnce(t *testing.T, d *session.Doctest) (agentRun string, err error) {
	t.Helper()
	cache := sessionCacheDir(d)
	agentRun = filepath.Join(cache, "agent-run")
	lock := filepath.Join(cache, "build.lock")
	ready := filepath.Join(cache, "binaries.ready")
	repoRoot, err := findAgentProRoot(d.DOCTEST_ROOT)
	if err != nil {
		return "", err
	}
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
		build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", agentRun, "./agent-run")
		build.Dir = repoRoot
		if out, err := build.CombinedOutput(); err != nil {
			return fmt.Errorf("build agent-run: %w\n%s", err, string(out))
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

// writeEnvLoggingRunner dumps PATH and selected keys to probePath, then
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
  env | grep -E '^(FOO|A|NEW|BAZ|GROK_HOME|CODEX_HOME|AGENT_RUNNER_CONFIG_HOME)=' | sort
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

// findFirstMetaJSON walks sessions/ for the first meta.json (flat layout).
func findFirstMetaJSON(t *testing.T, home string) (sessionID string, meta map[string]any) {
	t.Helper()
	root := filepath.Join(home, "sessions")
	var foundPath string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		if info.Name() != "meta.json" {
			return nil
		}
		foundPath = path
		return errors.New("stop") // stop walk after first hit
	})
	if foundPath == "" {
		t.Fatalf("no meta.json under %s", root)
	}
	data, err := os.ReadFile(foundPath)
	if err != nil {
		t.Fatalf("read %s: %v", foundPath, err)
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("parse %s: %v", foundPath, err)
	}
	// sessions/<session_id>/meta.json
	sessionID = filepath.Base(filepath.Dir(foundPath))
	return sessionID, meta
}

func stringSliceField(meta map[string]any, key string) []string {
	raw, ok := meta[key]
	if !ok || raw == nil {
		return nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func metaString(meta map[string]any, key string) string {
	v, _ := meta[key].(string)
	return v
}

// seedBoundExitedMeta writes a finished bound session meta.json suitable for
// resume gate (runner_session_id set, no live terminal registry).
// Includes optional prepend_paths / env / agent_runner_config_home for reapply tests.
func seedBoundExitedMeta(t *testing.T, req *Request) {
	t.Helper()
	if req.SessionID == "" {
		t.Fatal("seedBoundExitedMeta requires SessionID")
	}
	if req.Runner == "" {
		req.Runner = defaultRunner
	}
	if req.MetaStatus == "" {
		req.MetaStatus = "finished"
	}
	if req.RunnerSessionID == "" {
		req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440500"
	}
	if req.TerminalSessionID == "" {
		req.TerminalSessionID = "term-session-env-1"
	}
	if req.Workspace == "" {
		req.Workspace = req.TempDir
	}
	if req.Model == "" {
		req.Model = defaultModel
	}
	if req.InitialPrompt == "" {
		req.InitialPrompt = "prior turn"
	}
	meta := map[string]any{
		"runner":              req.Runner,
		"session_id":          req.SessionID,
		"status":              req.MetaStatus,
		"runner_session_id":   req.RunnerSessionID,
		"terminal_session_id": req.TerminalSessionID,
		"workspace":           req.Workspace,
		"model":               req.Model,
		"initial_prompt":      req.InitialPrompt,
		"created_at":          "2026-07-03T12:00:00Z",
		"updated_at":          "2026-07-03T12:00:00Z",
	}
	if len(req.SeedPrependPaths) > 0 {
		meta["prepend_paths"] = req.SeedPrependPaths
	}
	if len(req.SeedEnv) > 0 {
		meta["env"] = req.SeedEnv
	}
	if strings.TrimSpace(req.SeedConfigHome) != "" {
		meta["agent_runner_config_home"] = req.SeedConfigHome
	}
	path := metaJSONPath(req.Home, req.SessionID)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir session dir: %v", err)
	}
	b, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		t.Fatalf("marshal meta: %v", err)
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("write meta.json: %v", err)
	}
	req.SeedMeta = true
}

func readEnvProbe(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read env probe %s: %v", path, err)
	}
	return string(data)
}

func probePATH(probe string) string {
	for _, line := range strings.Split(probe, "\n") {
		if strings.HasPrefix(line, "PATH=") {
			return strings.TrimPrefix(line, "PATH=")
		}
	}
	return ""
}

func pathHasPrefixDirs(pathValue string, dirs ...string) bool {
	if pathValue == "" || len(dirs) == 0 {
		return false
	}
	parts := strings.Split(pathValue, string(os.PathListSeparator))
	if len(parts) < len(dirs) {
		return false
	}
	for i, d := range dirs {
		if parts[i] != d {
			return false
		}
	}
	return true
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

func assertProbePATHPrefixed(t *testing.T, probe string, dirs ...string) {
	t.Helper()
	pathVal := probePATH(probe)
	if pathVal == "" {
		t.Fatalf("env probe missing PATH= line; probe:\n%s", probe)
	}
	if !pathHasPrefixDirs(pathVal, dirs...) {
		t.Fatalf("PATH does not start with %v;\nPATH=%s", dirs, pathVal)
	}
}

func assertMetaStringSliceEquals(t *testing.T, meta map[string]any, key string, want []string) {
	t.Helper()
	got := stringSliceField(meta, key)
	if len(got) != len(want) {
		t.Fatalf("meta[%q] len=%d want %d; got=%v want=%v\nmeta=%v", key, len(got), len(want), got, want, meta)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("meta[%q][%d]=%q want %q; full=%v", key, i, got[i], want[i], got)
		}
	}
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	repoRoot, err := findAgentProRoot(d.DOCTEST_ROOT)
	if err != nil {
		return err
	}
	req.RepoRoot = repoRoot
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
	req.Runner = defaultRunner
	req.Workspace = req.TempDir
	req.ExecTimeout = 60 * time.Second
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
