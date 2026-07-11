# Scenario

**Feature**: session-level `status` multi-layer probe, `resume` as run shortcut,
and `run --open` post-attach grok session bind

```
# status probe
seed meta + optional registry/ptywrap
  -> agent-run status <ref> [--json]
  -> session / process / terminal / runner.exited / resume.ready

# resume gate + run shortcut
seed meta (bound+exited | live | unbound | missing)
  -> agent-run resume [flags] <id> ["followup"]
  -> deny (exit 1) or run path with --resume <runner_session_id>

# open post-exit
agent-run run --open (+ instant attach hook + optional GROK_HOME seed)
  -> after attach: print/persist grok session or error not resolved
```

## Preconditions

- Repository contains `cmd/agent-run`.
- Nested DOCTEST root: does not inherit parent `cmd/agent-run/tests` Setup/Run.
- Repo root from this tree: `DOCTEST_ROOT/../../../..`
  (`status-resume` → `tests` → `agent-run` → `cmd` → module root).
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Session-scoped build cache: `$TMPDIR/agent-run-status-resume-doctest-<DOCTEST_SESSION_ID>/`
  shares the compiled `agent-run` binary across parallel leaves.
- `frontend-agent-run/dist` (and `frontend/dist` if present) may be absent
  (gitignored). Build Setup stubs a minimal `index.html` so `//go:embed dist`
  compiles; status/resume leaves do not serve UI assets.
- No real Grok CLI required for core contract leaves.
- Live terminal leaves use fake in-process ptywrap HTTP+WebSocket.
- Open lifecycle leaves set `AGENT_RUN_OPEN_ATTACH_INSTANT=1` so auto-attach
  returns without a controlling TTY.
- Argv-probe resume leaf must **not** set `AGENT_RUN_GROK_TTY_COMMAND` (hook
  replaces argv and would hide `--resume`); use `--agent-runner-binary` instead.

## Steps

1. Root `Setup` resolves repo root, builds `agent-run` once per session, sets
   `AGENT_RUN_HOME`, default runner `grok-tty`.
2. Grouping / leaf `Setup` seeds meta/registry/fake-pty, finalizes `req.Args`.
3. `Run` executes `agent-run` (optional JSON parse / meta re-read).
4. Leaf `Assert` checks exit code, stdout/stderr shape, JSON keys, argv probe,
   or persisted `runner_session_id`.

## Context

- Default runner: `grok-tty`.
- Default session ids are leaf-specific constants (e.g. `test-exited-s1`).
- Resume gate: `ready ⇔ runner_session_id ≠ "" ∧ runner.exited == true`.
- CLI status stdout (human and JSON) must end with trailing `\n`.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

const (
	envOpenAttachInstant   = "AGENT_RUN_OPEN_ATTACH_INSTANT"
	envGrokTTYCommand      = "AGENT_RUN_GROK_TTY_COMMAND"
	envGrokTTYGrokSession  = "AGENT_RUN_GROK_TTY_GROK_SESSION_ID"
	defaultRunner          = "grok-tty"
	deadPIDSentinel        = 999999991
	defaultRegistryCreated = "2026-07-03T12:00:00Z"
)

var fakePTYWrapUpgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func sessionCacheDir() string {
	return filepath.Join(os.TempDir(), "agent-run-status-resume-doctest-"+DOCTEST_SESSION_ID)
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

func buildOnce(t *testing.T) (agentRun string, err error) {
	t.Helper()
	cache := sessionCacheDir()
	agentRun = filepath.Join(cache, "agent-run")
	lock := filepath.Join(cache, "build.lock")
	ready := filepath.Join(cache, "binaries.ready")
	repoRoot := filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../.."))
	err = withFileLock(t, lock, func() error {
		if fileExists(ready) && fileExists(agentRun) {
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

func setEnvKV(req *Request, key, value string) {
	req.Env = withoutEnvKey(req.Env, key)
	req.Env = append(req.Env, key+"="+value)
}

func applyOpenInstantAttach(req *Request) {
	if !req.OpenInstantAttach {
		return
	}
	setEnvKV(req, envOpenAttachInstant, "1")
}

func applyGrokTTYCommand(req *Request) {
	if strings.TrimSpace(req.GrokTTYCommand) == "" {
		return
	}
	setEnvKV(req, envGrokTTYCommand, req.GrokTTYCommand)
}

func applyGrokHomeEnv(req *Request) {
	if strings.TrimSpace(req.GrokHome) == "" {
		return
	}
	setEnvKV(req, "GROK_HOME", req.GrokHome)
	if id := strings.TrimSpace(req.GrokSessionUUID); id != "" {
		setEnvKV(req, envGrokTTYGrokSession, id)
	}
}

func ensureDefaults(req *Request) {
	if req.Runner == "" {
		req.Runner = defaultRunner
	}
	if req.Workspace == "" && req.TempDir != "" {
		req.Workspace = req.TempDir
	}
	if req.Model == "" {
		req.Model = "test-model"
	}
	if req.MetaStatus == "" {
		req.MetaStatus = "finished"
	}
}

func openAgentStore(t *testing.T, req *Request) agentstorage.Store {
	t.Helper()
	store, err := agentstorage.NewFileStore(req.Home)
	if err != nil {
		t.Fatalf("NewFileStore(%q): %v", req.Home, err)
	}
	return store
}

func seedSessionMeta(t *testing.T, req *Request) {
	t.Helper()
	if !req.SeedMeta {
		return
	}
	ensureDefaults(req)
	if req.SessionID == "" {
		t.Fatal("SeedMeta requires SessionID")
	}
	store := openAgentStore(t, req)
	meta := agentstorage.SessionMeta{
		Runner:            req.Runner,
		SessionID:         req.SessionID,
		Status:            req.MetaStatus,
		RunnerSessionID:   req.RunnerSessionID,
		TerminalSessionID: req.TerminalSessionID,
		Workspace:         req.Workspace,
		Model:             req.Model,
		InitialPrompt:     req.InitialPrompt,
		CreatedAt:         "2026-07-03T12:00:00Z",
		UpdatedAt:         "2026-07-03T12:00:00Z",
	}
	if err := store.CreateSession(req.Runner, req.SessionID, meta); err != nil {
		// Allow overwrite for re-seed patterns
		path := metaJSONPath(req.Home, req.Runner, req.SessionID)
		if err2 := os.MkdirAll(filepath.Dir(path), 0755); err2 != nil {
			t.Fatalf("mkdir session dir: %v", err2)
		}
		b, mErr := json.MarshalIndent(meta, "", "  ")
		if mErr != nil {
			t.Fatalf("marshal meta: %v", mErr)
		}
		if wErr := os.WriteFile(path, b, 0644); wErr != nil {
			t.Fatalf("write meta.json: %v (create: %v)", wErr, err)
		}
		return
	}
}

func metaJSONPath(home, runner, sessionID string) string {
	return filepath.Join(home, "sessions", runner, sessionID, "meta.json")
}

func readMetaJSON(t *testing.T, home, runner, sessionID string) map[string]any {
	t.Helper()
	path := metaJSONPath(home, runner, sessionID)
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

func registryDir(home, runner string) string {
	return filepath.Join(home, runner+"-registry")
}

func registryPath(home, runner, terminalID string) string {
	return filepath.Join(registryDir(home, runner), terminalID+".json")
}

func resolveRegistryPID(req *Request) int {
	if req.RegistryPID < 0 {
		return deadPIDSentinel
	}
	if req.RegistryPID == 0 {
		return os.Getpid()
	}
	return req.RegistryPID
}

func writeRegistryEntry(t *testing.T, req *Request) {
	t.Helper()
	if !req.WriteRegistry {
		return
	}
	ensureDefaults(req)
	termID := req.TerminalSessionID
	if termID == "" {
		termID = "session-1"
		req.TerminalSessionID = termID
	}
	listenAddr := fmt.Sprintf("127.0.0.1:%d", req.FakePTYWrapPort)
	if req.FakePTYWrapPort <= 0 {
		if req.RegistryClosedPort {
			// Ephemeral closed port: bind then close so nothing listens.
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatalf("listen closed-port probe: %v", err)
			}
			port := ln.Addr().(*net.TCPAddr).Port
			_ = ln.Close()
			listenAddr = fmt.Sprintf("127.0.0.1:%d", port)
		} else {
			listenAddr = "127.0.0.1:59999"
		}
	}
	entry := RegistryEntryData{
		SessionID:  termID,
		ListenAddr: listenAddr,
		PID:        resolveRegistryPID(req),
		CreatedAt:  defaultRegistryCreated,
	}
	dir := registryDir(req.Home, req.Runner)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir registry: %v", err)
	}
	b, _ := json.Marshal(entry)
	path := registryPath(req.Home, req.Runner, termID)
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("write registry: %v", err)
	}
}

func startFakePTYWrapServer(t *testing.T, req *Request) {
	t.Helper()
	if !req.StartFakePTYWrap {
		return
	}
	mux := http.NewServeMux()
	var inputMu sync.Mutex
	var inputs []string

	mux.HandleFunc("/api/terminal/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `[{"id":"session-1"}]`)
	})

	mux.HandleFunc("/api/terminal", func(w http.ResponseWriter, r *http.Request) {
		conn, err := fakePTYWrapUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		scrollback := req.FakePTYWrapScrollback
		if scrollback == "" {
			// Idle/sendable-ish scrollback for "live, not exited" scenarios.
			scrollback = "GROK_TTY_BANNER\nGrok › \n"
		}
		if err := conn.WriteMessage(websocket.TextMessage, []byte(scrollback)); err != nil {
			return
		}
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			inputMu.Lock()
			inputs = append(inputs, string(msg))
			inputMu.Unlock()
		}
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("fake ptywrap listen: %v", err)
	}
	req.FakePTYWrapPort = ln.Addr().(*net.TCPAddr).Port
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	})
}

// seedBoundExitedDeadTerminal seeds meta with runner_session_id and no live registry
// (or closed-port registry). Models "finished + dead terminal + bound + exited".
func seedBoundExitedDeadTerminal(t *testing.T, req *Request) {
	t.Helper()
	req.SeedMeta = true
	ensureDefaults(req)
	if req.SessionID == "" {
		req.SessionID = "test-exited-s1"
	}
	if req.RunnerSessionID == "" {
		req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440000"
	}
	if req.MetaStatus == "" {
		req.MetaStatus = "finished"
	}
	if req.TerminalSessionID == "" {
		req.TerminalSessionID = "term-exited-1"
	}
	if req.InitialPrompt == "" {
		req.InitialPrompt = "prior turn"
	}
	seedSessionMeta(t, req)
	// Optional dead registry: closed port + dead PID.
	if req.WriteRegistry {
		req.RegistryPID = -1
		req.RegistryClosedPort = true
		writeRegistryEntry(t, req)
	}
}

// seedLiveBoundNotExited seeds meta + live registry + idle fake ptywrap.
func seedLiveBoundNotExited(t *testing.T, req *Request) {
	t.Helper()
	req.SeedMeta = true
	req.MetaStatus = "running"
	ensureDefaults(req)
	if req.SessionID == "" {
		req.SessionID = "test-live-s1"
	}
	if req.RunnerSessionID == "" {
		req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440111"
	}
	if req.TerminalSessionID == "" {
		req.TerminalSessionID = "term-live-1"
	}
	if req.InitialPrompt == "" {
		req.InitialPrompt = "live turn"
	}
	req.StartFakePTYWrap = true
	req.WriteRegistry = true
	req.RegistryPID = 0 // alive = current test process
	if req.FakePTYWrapScrollback == "" {
		req.FakePTYWrapScrollback = "GROK_TTY_BANNER\nGrok › \n"
	}
	startFakePTYWrapServer(t, req)
	seedSessionMeta(t, req)
	writeRegistryEntry(t, req)
}

// seedUnbound seeds meta without runner_session_id.
func seedUnbound(t *testing.T, req *Request) {
	t.Helper()
	req.SeedMeta = true
	req.RunnerSessionID = ""
	ensureDefaults(req)
	if req.SessionID == "" {
		req.SessionID = "test-unbound-s1"
	}
	if req.MetaStatus == "" {
		req.MetaStatus = "finished"
	}
	if req.InitialPrompt == "" {
		req.InitialPrompt = "never bound"
	}
	seedSessionMeta(t, req)
}

func fakeTUIRespondHi() string {
	return `sh -c 'printf "GROK_TTY_BANNER\nGrok › "; read line; echo "Response: $line"'`
}

func fakeTUIHoldSeconds(sec int) string {
	return fmt.Sprintf(`sh -c 'printf "GROK_TTY_BANNER\nGrok › "; sleep %d'`, sec)
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

func writeFakeGrokSessionDir(t *testing.T, grokHome, workspace, sessionUUID, prompt string) string {
	t.Helper()
	dir := grokSessionDir(grokHome, workspace, sessionUUID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir grok session: %v", err)
	}
	abs, _ := filepath.Abs(workspace)
	summary := map[string]any{
		"info": map[string]any{
			"cwd":       abs,
			"sessionId": sessionUUID,
			"openedAt":  time.Now().UTC().Format(time.RFC3339Nano),
		},
		"created_at": time.Now().UTC().Format(time.RFC3339Nano),
	}
	sb, _ := json.Marshal(summary)
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), sb, 0644); err != nil {
		t.Fatalf("write summary.json: %v", err)
	}
	userLine, _ := json.Marshal(map[string]any{
		"sessionUpdate": "user_message_chunk",
		"content":       map[string]any{"type": "text", "text": prompt},
	})
	agentLine, _ := json.Marshal(map[string]any{
		"sessionUpdate": "agent_message_chunk",
		"content":       map[string]any{"type": "text", "text": "OPEN_POST_EXIT_ASSISTANT"},
	})
	updatesPath := filepath.Join(dir, "updates.jsonl")
	body := string(userLine) + "\n" + string(agentLine) + "\n"
	if err := os.WriteFile(updatesPath, []byte(body), 0644); err != nil {
		t.Fatalf("write updates.jsonl: %v", err)
	}
	return updatesPath
}

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
		t.Fatalf("write argv runner: %v", err)
	}
	return path
}

func execCmd(t *testing.T, command string, args []string, dir string, env []string, timeout time.Duration) (*Response, error) {
	t.Helper()
	if timeout <= 0 {
		timeout = 45 * time.Second
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
	applyGrokTTYCommand(req)
	applyGrokHomeEnv(req)
	resp, err := execCmd(t, req.AgentRun, args, req.TempDir, req.Env, req.ExecTimeout)
	if err != nil {
		return resp, err
	}
	if resp == nil {
		return resp, nil
	}
	switch req.Mode {
	case "status-json":
		if resp.ExitCode == 0 && strings.TrimSpace(resp.Stdout) != "" {
			var obj map[string]any
			if jErr := json.Unmarshal([]byte(resp.Stdout), &obj); jErr != nil {
				return resp, fmt.Errorf("parse JSON stdout: %w\nstdout:\n%s", jErr, resp.Stdout)
			}
			resp.JSONBody = obj
		}
	case "read-meta":
		if req.SessionID != "" {
			runner := req.Runner
			if runner == "" {
				runner = defaultRunner
			}
			path := metaJSONPath(req.Home, runner, req.SessionID)
			if data, rErr := os.ReadFile(path); rErr == nil {
				var obj map[string]any
				if json.Unmarshal(data, &obj) == nil {
					resp.MetaAfter = obj
				}
			}
		}
	}
	return resp, nil
}

func assertExitCode(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp.ExitCode != want {
		t.Fatalf("expected exit code %d, got %d\nstderr:\n%s\nstdout:\n%s", want, resp.ExitCode, resp.Stderr, resp.Stdout)
	}
}

func assertSuccess(t *testing.T, resp *Response) {
	t.Helper()
	assertExitCode(t, resp, 0)
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

func jsonPathString(obj map[string]any, keys ...string) (string, bool) {
	var cur any = obj
	for _, k := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			return "", false
		}
		cur, ok = m[k]
		if !ok {
			return "", false
		}
	}
	switch v := cur.(type) {
	case string:
		return v, true
	case bool:
		if v {
			return "true", true
		}
		return "false", true
	case float64:
		return fmt.Sprintf("%v", v), true
	case nil:
		return "null", true
	default:
		return fmt.Sprint(v), true
	}
}

func jsonPathBool(obj map[string]any, keys ...string) (bool, bool) {
	var cur any = obj
	for _, k := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			return false, false
		}
		cur, ok = m[k]
		if !ok {
			return false, false
		}
	}
	b, ok := cur.(bool)
	return b, ok
}

func Setup(t *testing.T, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	if err := os.MkdirAll(req.Home, 0755); err != nil {
		return fmt.Errorf("mkdir home: %w", err)
	}
	cached, err := buildOnce(t)
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
	// Ensure argv-sensitive leaves are not polluted by ambient hooks.
	req.Env = withoutEnvKey(req.Env, envGrokTTYCommand)
	req.Runner = defaultRunner
	req.Workspace = req.TempDir
	return nil
}
```
