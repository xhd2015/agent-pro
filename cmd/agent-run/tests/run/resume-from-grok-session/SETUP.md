# Scenario

**Feature**: `run --resume-from-grok-session` — shared harness for P1 validation,
P2 headless create/pre-bind/resume-argv, and P3 open/detach/help/mutex: isolated
`AGENT_RUN_HOME` + `GROK_HOME`, binary build, Grok `summary.json` seeding,
agent-run meta seed, argv-recorder + long-hold fake runners

```
# build + isolate homes
build agent-run
  -> AGENT_RUN_HOME = temp/.agent-run
  -> GROK_HOME = temp/grok-home

# optional Grok fixture (sessions.Find / Info)
GROK_HOME/sessions/<url.PathEscape(abs cwd)>/<uuid>/summary.json
  info.id = uuid
  info.cwd = workspace

# optional agent-run mapping / collision seed
AGENT_RUN_HOME/sessions/<id>/meta.json
  runner = grok|grok-tty|…
  runner_session_id = uuid | empty

# P2 fake runner (no AGENT_RUN_GROK_TTY_COMMAND)
--agent-runner-binary record-argv.sh -> ARGV_RECORD=$*  (+ GROK_TTY_BANNER)

# P3 hold runner (proves detach/open return without waiting full sleep)
--agent-runner-binary hold.sh -> banner + sleep N

# CLI under test
agent-run run [--session-id ID] [--agent-runner R] [--agent-runner-binary BIN]
              [--dir DIR] [--open|--detach] [--auto-send-or-resume]
              --resume-from-grok-session <id> ["followup"]
  -> P1 error | P2 headless | P3 open/detach/mutex
```

## Preconditions

- Repository contains `cmd/agent-run`.
- Each leaf uses isolated `AGENT_RUN_HOME` and `GROK_HOME` under `t.TempDir()`.
- Grok summary layout matches `agent/grok/sessions` (`info.id`, `info.cwd`,
  timestamps so Find/List accept the file).
- Nested root: no inheritance from parent `cmd/agent-run/tests` (own `DOCTEST.md`).
- P1 leaves do not need a real Grok binary; P2 uses argv-recorder; P3 detach/open
  use a long-hold binary so missing mode wiring times out (classic RED).

## Steps

1. Root `Setup` builds `agent-run`, sets isolated homes + env, default workdir.
2. Leaf `Setup` seeds fixtures / fake runner (when needed) and finalizes `req.Args`.
3. `Run` executes `agent-run` with `req.Args` at `req.WorkDir`.
4. Leaf `Assert` checks exit code, error substrings, meta, argv, or detach ids.

## Context

- Shared fixture UUID default: `550e8400-e29b-41d4-a716-446655440a01`.
- Process `cmd.Dir` is `req.WorkDir` (defaults to `TempDir`).
- `GROK_HOME` is always set on `req.Env` so lookups never touch the developer home.
- `AGENT_RUN_GROK_TTY_COMMAND` is stripped so binary/argv paths are exercised.
- Open leaves set `AGENT_RUN_OPEN_ATTACH_INSTANT=1` (same hook as `run/open/`).

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
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/doctest/session"
	"path"
)

// Default provider UUID for leaves that need a known Grok session id.
const defaultGrokSessionID = "550e8400-e29b-41d4-a716-446655440a01"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.WorkDir = req.TempDir
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	if err := os.MkdirAll(req.Home, 0755); err != nil {
		return fmt.Errorf("mkdir AGENT_RUN_HOME: %w", err)
	}
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		return fmt.Errorf("mkdir GROK_HOME: %w", err)
	}
	req.AgentRun = filepath.Join(req.TempDir, "bin", "agent-run")
	if err := os.MkdirAll(filepath.Dir(req.AgentRun), 0755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}
	build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", req.AgentRun, "./agent-run")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build agent-run: %w\n%s", err, string(out))
	}
	req.Env = append(req.Env,
		"AGENT_RUN_HOME="+req.Home,
		"GROK_HOME="+req.GrokHome,
		"PATH="+filepath.Dir(req.AgentRun)+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	// Avoid accidental hook substitution if present in the ambient env.
	req.Env = withoutEnvKey(req.Env, "AGENT_RUN_GROK_TTY_COMMAND")
	if req.GrokSessionID == "" {
		req.GrokSessionID = defaultGrokSessionID
	}
	if req.ExecTimeout <= 0 {
		req.ExecTimeout = 30 * time.Second
	}
	return nil
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

func encodedGrokCwd(workspace string) string {
	abs, err := filepath.Abs(workspace)
	if err != nil {
		abs = workspace
	}
	return url.PathEscape(abs)
}

func absPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatalf("abs %q: %v", p, err)
	}
	return abs
}

func grokSessionDir(grokHome, workspace, sessionID string) string {
	return filepath.Join(grokHome, "sessions", encodedGrokCwd(workspace), sessionID)
}

// seedGrokSession writes a Find-compatible summary.json under GROK_HOME.
// Uses info.id (not sessionId) so agent/grok/sessions.Find resolves the id.
// Returns the absolute session directory path.
func seedGrokSession(t *testing.T, grokHome, workspace, sessionID string) string {
	t.Helper()
	absCwd := absPath(t, workspace)
	dir := grokSessionDir(grokHome, absCwd, sessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir grok session dir: %v", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	summary := map[string]any{
		"info": map[string]any{
			"id":  sessionID,
			"cwd": absCwd,
		},
		"generated_title":   "resume-from-grok-session P1 fixture",
		"created_at":        now,
		"updated_at":        now,
		"last_active_at":    now,
		"num_messages":      1,
		"num_chat_messages": 1,
	}
	b, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), b, 0644); err != nil {
		t.Fatalf("write summary.json: %v", err)
	}
	return dir
}

func openAgentStore(t *testing.T, req *Request) agentstorage.Store {
	t.Helper()
	store, err := agentstorage.NewFileStore(req.Home)
	if err != nil {
		t.Fatalf("NewFileStore(%q): %v", req.Home, err)
	}
	return store
}

// seedMappedMeta creates flat agent-run meta with runner + runner_session_id
// so the "already mapped" gate can fire for the given Grok id.
func seedMappedMeta(t *testing.T, req *Request, runner, sessionID, runnerSessionID string) {
	t.Helper()
	store := openAgentStore(t, req)
	meta := agentstorage.SessionMeta{
		Runner:          runner,
		SessionID:       sessionID,
		Status:          "finished",
		RunnerSessionID: runnerSessionID,
		Workspace:       req.GrokCWD,
	}
	if meta.Workspace == "" {
		meta.Workspace = req.WorkDir
	}
	if err := store.CreateSession(sessionID, meta); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
}

func mustMkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	// Project-like marker (avoid bare-tmp heuristics if any).
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("resume-from-grok-session fixture\n"), 0644); err != nil {
		t.Fatalf("write README in %s: %v", dir, err)
	}
}

// writeArgvRecordingRunner writes a fake grok-tty binary that records argv then
// prints the banner and exits (same pattern as agent-runner-binary / status-resume).
func writeArgvRecordingRunner(t *testing.T, dir, name, probePath string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	script := fmt.Sprintf(`#!/bin/sh
echo "ARGV_RECORD=$*" > %q
printf "GROK_TTY_BANNER\nGrok › "
read -r line || true
echo "Response: ${line:-done}"
exit 0
`, probePath)
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write argv runner %s: %v", path, err)
	}
	return path
}

// installArgvRunner installs record-argv.sh under TempDir/fake-bin and sets
// AgentRunnerBinary + ArgvProbePath. Clears AGENT_RUN_GROK_TTY_COMMAND.
func installArgvRunner(t *testing.T, req *Request) {
	t.Helper()
	binDir := filepath.Join(req.TempDir, "fake-bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("mkdir fake-bin: %v", err)
	}
	if req.ArgvProbePath == "" {
		req.ArgvProbePath = filepath.Join(req.TempDir, "argv-probe.log")
	}
	req.RunnerScriptPath = writeArgvRecordingRunner(t, binDir, "record-argv.sh", req.ArgvProbePath)
	req.AgentRunnerBinary = req.RunnerScriptPath
	req.Env = withoutEnvKey(req.Env, "AGENT_RUN_GROK_TTY_COMMAND")
}

const envOpenAttachInstant = "AGENT_RUN_OPEN_ATTACH_INSTANT"

func setEnvKV(req *Request, key, value string) {
	req.Env = withoutEnvKey(req.Env, key)
	req.Env = append(req.Env, key+"="+value)
}

// writeHoldRunner paints the TTY banner then sleeps (keep-alive / detach/open probes).
// holdSec should exceed ExecTimeout when proving that --detach/--open return early.
func writeHoldRunner(t *testing.T, dir, name string, holdSec int) string {
	t.Helper()
	if holdSec <= 0 {
		holdSec = 120
	}
	path := filepath.Join(dir, name)
	script := fmt.Sprintf(`#!/bin/sh
printf "GROK_TTY_BANNER\nGrok › "
sleep %d
exit 0
`, holdSec)
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write hold runner %s: %v", path, err)
	}
	return path
}

// installHoldRunner installs a long-sleep fake binary and clears hook env.
func installHoldRunner(t *testing.T, req *Request, holdSec int) {
	t.Helper()
	binDir := filepath.Join(req.TempDir, "fake-bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("mkdir fake-bin: %v", err)
	}
	req.RunnerScriptPath = writeHoldRunner(t, binDir, "hold.sh", holdSec)
	req.AgentRunnerBinary = req.RunnerScriptPath
	req.Env = withoutEnvKey(req.Env, "AGENT_RUN_GROK_TTY_COMMAND")
}

// applyOpenInstantEnv sets AGENT_RUN_OPEN_ATTACH_INSTANT=1 when requested.
func applyOpenInstantEnv(req *Request) {
	if req.OpenInstantAttach {
		setEnvKV(req, envOpenAttachInstant, "1")
	} else {
		req.Env = withoutEnvKey(req.Env, envOpenAttachInstant)
	}
}

// parseDetachIDs extracts labeled session-id / terminal-id lines from detach stdout.
func parseDetachIDs(stdout string) (sessionID, terminalID string, ok bool) {
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "session-id:"):
			sessionID = strings.TrimSpace(strings.TrimPrefix(line, "session-id:"))
		case strings.HasPrefix(line, "terminal-id:"):
			terminalID = strings.TrimSpace(strings.TrimPrefix(line, "terminal-id:"))
		}
	}
	return sessionID, terminalID, sessionID != "" && terminalID != ""
}

func assertDetachIDsOnStdout(t *testing.T, resp *Response) (sessionID, terminalID string) {
	t.Helper()
	sessionID, terminalID, ok := parseDetachIDs(resp.Stdout)
	if !ok {
		t.Fatalf("stdout must print both session-id and terminal-id\nstdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "session-id:") {
		t.Fatalf("stdout missing session-id: label\nstdout:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "terminal-id:") {
		t.Fatalf("stdout missing terminal-id: label\nstdout:\n%s", resp.Stdout)
	}
	return sessionID, terminalID
}

func registryPath(home, runner, terminalID string) string {
	return filepath.Join(home, runner+"-registry", terminalID+".json")
}

// seedExistingAgentSession creates a flat meta without mapping the import UUID
// (used by session-id-already-exists — collision on agent-run id only).
func seedExistingAgentSession(t *testing.T, req *Request, sessionID, runner string) {
	t.Helper()
	store := openAgentStore(t, req)
	meta := agentstorage.SessionMeta{
		Runner:    runner,
		SessionID: sessionID,
		Status:    "finished",
		Workspace: req.WorkDir,
	}
	if err := store.CreateSession(sessionID, meta); err != nil {
		t.Fatalf("CreateSession existing: %v", err)
	}
}

func sessionMetaPath(home, sessionID string) string {
	return filepath.Join(home, "sessions", sessionID, "meta.json")
}

func readSessionMetaFile(t *testing.T, home, sessionID string) map[string]any {
	t.Helper()
	path := sessionMetaPath(home, sessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read meta %s: %v", path, err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse meta %s: %v\n%s", path, err, string(data))
	}
	return m
}

// runArgs builds:
//
//	run [--agent-runner R] [--session-id ID] [--agent-runner-binary BIN]
//	    [--dir D] [--open] [--detach] [--auto-send-or-resume]
//	    --resume-from-grok-session ID [extra...]
func runArgs(req *Request, resumeID string, extra ...string) []string {
	args := []string{"run"}
	if strings.TrimSpace(req.AgentRunner) != "" {
		args = append(args, "--agent-runner", req.AgentRunner)
	}
	if strings.TrimSpace(req.SessionID) != "" {
		args = append(args, "--session-id", req.SessionID)
	}
	if strings.TrimSpace(req.AgentRunnerBinary) != "" {
		args = append(args, "--agent-runner-binary", req.AgentRunnerBinary)
	}
	if strings.TrimSpace(req.DirFlag) != "" {
		args = append(args, "--dir", req.DirFlag)
	}
	if req.AutoSendOrResume {
		args = append(args, "--auto-send-or-resume")
	}
	if req.OpenFlag {
		args = append(args, "--open")
	}
	if req.DetachFlag {
		args = append(args, "--detach")
	}
	if req.ForkFlag {
		args = append(args, "--fork")
	}
	args = append(args, "--resume-from-grok-session", resumeID)
	args = append(args, extra...)
	return args
}

// setupValidImport seeds a Grok session at WorkDir and optional argv-recorder
// for P2 success leaves. Does not create agent-run mapping.
func setupValidImport(t *testing.T, req *Request, withArgvRunner bool) {
	t.Helper()
	req.GrokCWD = absPath(t, req.WorkDir)
	seedGrokSession(t, req.GrokHome, req.GrokCWD, req.GrokSessionID)
	if withArgvRunner {
		installArgvRunner(t, req)
	}
	if req.ExecTimeout < 60*time.Second {
		req.ExecTimeout = 60 * time.Second
	}
}

// setupDetachOrOpenImport seeds Grok + long-hold binary for P3 mode leaves.
// holdSec should exceed timeout so headless-without-mode fails by timeout.
func setupDetachOrOpenImport(t *testing.T, req *Request, holdSec int, timeout time.Duration) {
	t.Helper()
	req.GrokCWD = absPath(t, req.WorkDir)
	seedGrokSession(t, req.GrokHome, req.GrokCWD, req.GrokSessionID)
	installHoldRunner(t, req, holdSec)
	applyOpenInstantEnv(req)
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	req.ExecTimeout = timeout
}

func execCmd(t *testing.T, command string, args []string, dir string, env []string, timeout time.Duration) (*Response, error) {
	t.Helper()
	if timeout <= 0 {
		timeout = 30 * time.Second
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
	dir := req.WorkDir
	if dir == "" {
		dir = req.TempDir
	}
	return execCmd(t, req.AgentRun, args, dir, req.Env, req.ExecTimeout)
}

func assertSuccess(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d\nstderr:\n%s\nstdout:\n%s",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
}

func assertExitCode(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp.ExitCode != want {
		t.Fatalf("expected exit code %d, got %d\nstderr:\n%s\nstdout:\n%s",
			want, resp.ExitCode, resp.Stderr, resp.Stdout)
	}
}

func assertNonZeroExit(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0\nstderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
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
	lower := strings.ToLower(got)
	for _, w := range wants {
		if strings.Contains(lower, strings.ToLower(w)) {
			return
		}
	}
	t.Fatalf("none of %v found in:\n%s", wants, got)
}

func combinedOut(resp *Response) string {
	return resp.Stderr + "\n" + resp.Stdout
}
```
