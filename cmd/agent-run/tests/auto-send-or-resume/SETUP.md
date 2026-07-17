# Scenario

**Feature**: `run --auto-send-or-resume` three-way branch (run / send / resume),
resume workspace resolution (`meta.workspace` / `--dir` / cwd), and Grok
resume `--dir` vs `info.cwd` relocate gate

```
# validation
agent-run run --auto-send-or-resume …  (gates)
  -> requires --session-id; mutex with --session-id-from-prompt
  -> run -h lists --auto-send-or-resume
  -> --new-terminal requires --auto-send-or-resume; run -h lists --new-terminal

# classify session
seed or missing meta + optional registry/ptywrap
  -> probeSessionStatus
  -> MODE = run | send | resume

# dispatch
MODE=run   -> agentui.Run create/re-run; no --resume
MODE=send  -> enqueue/deliver to terminal_session_id; live open denied
MODE=resume -> reclaim zombie if needed; provider --resume <id>; workspace rule

# --new-terminal
MODE=run|resume + --new-terminal -> iTerm2 ModeForceNew; strip flag; launcher exits
MODE=send + --new-terminal -> ignore; still send

# workspace
resume / auto→resume child cwd:
  --dir > meta.workspace > process cwd (+ warn)
  missing meta.workspace (no --dir) → exit 1 + path + --dir hint

# resume --dir vs Grok info.cwd (grok-tty; classic RED until implementer)
  --dir set + grok-tty → compare to summary info.cwd under config-home
  mismatch without --allow-relocate-resume-session-dir → exit 1
  mismatch with allow → RelocateCWD + meta.workspace update + continue
```

## Preconditions

- Repository contains `cmd/agent-run`.
- Nested DOCTEST root: does not inherit parent `cmd/agent-run/tests` Setup/Run.
- Repo root from this tree: `DOCTEST_ROOT/../../../..`
  (`auto-send-or-resume` → `tests` → `agent-run` → `cmd` → module root).
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Session-scoped build cache:
  `$TMPDIR/agent-run-auto-send-or-resume-doctest-<DOCTEST_SESSION_ID>/`
  shares the compiled `agent-run` binary across parallel leaves.
- `frontend-agent-run/dist` (and `frontend/dist` if present) may be absent
  (gitignored). Build Setup stubs a minimal `index.html` so `//go:embed dist`
  compiles; these leaves do not serve UI assets.
- No real Grok CLI required.
- Live send leaves use fake in-process ptywrap with WS scrollback **and** HTTP
  inject endpoints (`prepare-inject`, `input`).
- Argv/cwd probe leaves must **not** set `AGENT_RUN_GROK_TTY_COMMAND`.
- CLI stdout trailing newline contracts use assert templates with `\n` before
  the closing backtick where full stdout is matched.

## Steps

1. Root `Setup` resolves repo root, builds `agent-run` once per session, sets
   `AGENT_RUN_HOME`, default runner `grok-tty`, default `WorkDir` = TempDir.
2. Grouping / leaf `Setup` seeds meta/registry/fake-pty, writes fake runners,
   finalizes `req.Args`.
3. `Run` executes `agent-run` from `req.WorkDir` (CLI cwd) with optional meta
   and probe enrichment.
4. Leaf `Assert` checks exit code, stderr/stdout shape, argv/cwd probes, meta.

## Context

- Default runner: `grok-tty`.
- Resume ready ⇔ `runner_session_id ≠ ""` ∧ `runner.exited == true`.
- Auto flag long form only: `--auto-send-or-resume`.

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
	envOpenAttachInstant  = "AGENT_RUN_OPEN_ATTACH_INSTANT"
	envGrokTTYCommand     = "AGENT_RUN_GROK_TTY_COMMAND"
	defaultRunner         = "grok-tty"
	deadPIDSentinel       = 999999991
	defaultRegistryCreated = "2026-07-03T12:00:00Z"
)

var fakePTYWrapUpgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func sessionCacheDir() string {
	return filepath.Join(os.TempDir(), "agent-run-auto-send-or-resume-doctest-"+DOCTEST_SESSION_ID)
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
// //go:embed dist compiles.
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
		for _, rel := range []string{"frontend-agent-run/dist", "frontend/dist"} {
			if err := ensureStubDist(filepath.Join(repoRoot, rel)); err != nil {
				return fmt.Errorf("ensure %s stub: %w", rel, err)
			}
		}
		build := exec.Command("go", "build", "-buildvcs=false", "-o", agentRun, "./cmd/agent-run")
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

const (
	envIterm2Installed = "KOOL_ITERM2_INSTALLED"
	envIterm2ScriptOut = "KOOL_ITERM2_SCRIPT_OUT"
	envIterm2GOOS      = "KOOL_ITERM2_GOOS"
)

// applyIterm2TestHooks enables the iterm2 package test env: treat iTerm as
// installed and write AppleScript to ItermScriptOut instead of running osascript.
func applyIterm2TestHooks(req *Request) {
	if strings.TrimSpace(req.ItermScriptOut) == "" {
		return
	}
	setEnvKV(req, envIterm2Installed, "1")
	setEnvKV(req, envIterm2ScriptOut, req.ItermScriptOut)
	// Force darwin platform check inside iterm2 even if host GOOS differs.
	setEnvKV(req, envIterm2GOOS, "darwin")
}

// ensureItermScriptOutPath sets a default script-out path under TempDir when
// NewTerminal leaves enable capture without an explicit path.
func ensureItermScriptOutPath(req *Request) {
	if strings.TrimSpace(req.ItermScriptOut) != "" {
		return
	}
	if req.TempDir == "" {
		return
	}
	req.ItermScriptOut = filepath.Join(req.TempDir, "iterm-script.applescript")
}

func buildExecEnv(req *Request) []string {
	return append(os.Environ(), req.Env...)
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
	if req.WorkDir == "" && req.TempDir != "" {
		req.WorkDir = req.TempDir
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
	// Flat layout: sessions/<session_id>/meta.json (runner is meta field only).
	if err := store.CreateSession(req.SessionID, meta); err != nil {
		path := metaJSONPath(req.Home, req.SessionID)
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

// startFakePTYWrapServer starts an in-process fake ptywrap with:
// - WS /api/terminal scrollback (status / writable probes)
// - POST prepare-inject + input (send-queue drain)
func startFakePTYWrapServer(t *testing.T, req *Request) {
	t.Helper()
	if !req.StartFakePTYWrap {
		return
	}
	mux := http.NewServeMux()
	var inputMu sync.Mutex
	var inputs []string
	if req.FakePTYInjectLog == nil {
		log := make([]string, 0, 8)
		req.FakePTYInjectLog = &log
	}

	termID := req.TerminalSessionID
	if termID == "" {
		termID = "session-1"
	}

	mux.HandleFunc("/api/terminal/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, fmt.Sprintf(`[{"id":%q}]`, termID))
	})

	// Go 1.22+ method patterns used by production inject API.
	mux.HandleFunc("POST /api/terminal/sessions/{sessionID}/prepare-inject", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /api/terminal/sessions/{sessionID}/input", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		inputMu.Lock()
		inputs = append(inputs, string(body))
		*req.FakePTYInjectLog = append(*req.FakePTYInjectLog, string(body))
		inputMu.Unlock()
		w.WriteHeader(http.StatusOK)
	})
	// Fallback path-style handlers for older ServeMux matching.
	mux.HandleFunc("/api/terminal/sessions/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/prepare-inject") {
			w.WriteHeader(http.StatusOK)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/input") {
			body, _ := io.ReadAll(r.Body)
			inputMu.Lock()
			inputs = append(inputs, string(body))
			*req.FakePTYInjectLog = append(*req.FakePTYInjectLog, string(body))
			inputMu.Unlock()
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	})

	mux.HandleFunc("/api/terminal", func(w http.ResponseWriter, r *http.Request) {
		conn, err := fakePTYWrapUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		scrollback := req.FakePTYWrapScrollback
		if scrollback == "" {
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
			*req.FakePTYInjectLog = append(*req.FakePTYInjectLog, string(msg))
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

// seedBoundExitedDeadTerminal seeds meta with runner_session_id and no live registry.
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
	req.RegistryPID = 0
	if req.FakePTYWrapScrollback == "" {
		req.FakePTYWrapScrollback = "GROK_TTY_BANNER\nGrok › \n"
	}
	startFakePTYWrapServer(t, req)
	seedSessionMeta(t, req)
	writeRegistryEntry(t, req)
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

// writeArgvAndCwdRecordingRunner records argv and process cwd for workspace leaves.
func writeArgvAndCwdRecordingRunner(t *testing.T, dir, name, argvProbe, cwdProbe string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	script := fmt.Sprintf(`#!/bin/sh
pwd > %q
echo "ARGV_RECORD=$*" > %q
printf "GROK_TTY_BANNER\nGrok › "
read -r line || true
echo "Response: ${line:-done}"
exit 0
`, cwdProbe, argvProbe)
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write argv+cwd runner: %v", err)
	}
	return path
}

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
	req.GrokTTYCommand = ""
	req.Env = withoutEnvKey(req.Env, envGrokTTYCommand)
}

func installArgvCwdRunner(t *testing.T, req *Request) {
	t.Helper()
	binDir := filepath.Join(req.TempDir, "fake-bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("mkdir fake-bin: %v", err)
	}
	if req.ArgvProbePath == "" {
		req.ArgvProbePath = filepath.Join(req.TempDir, "argv-probe.log")
	}
	if req.CwdProbePath == "" {
		req.CwdProbePath = filepath.Join(req.TempDir, "cwd-probe.log")
	}
	req.RunnerScriptPath = writeArgvAndCwdRecordingRunner(t, binDir, "record-argv-cwd.sh", req.ArgvProbePath, req.CwdProbePath)
	req.AgentRunnerBinary = req.RunnerScriptPath
	req.GrokTTYCommand = ""
	req.Env = withoutEnvKey(req.Env, envGrokTTYCommand)
}

func fakeTUIRespondHi() string {
	return `sh -c 'printf "GROK_TTY_BANNER\nGrok › "; read line; echo "Response: $line"'`
}

func execCmdWithBase(t *testing.T, command string, args []string, dir string, fullEnv []string, timeout time.Duration) (*Response, error) {
	t.Helper()
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	if dir == "" {
		dir = ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	cmd.Env = fullEnv
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

func enrichMetaAfter(req *Request, resp *Response) {
	if req.SessionID == "" {
		return
	}
	path := metaJSONPath(req.Home, req.SessionID)
	if data, rErr := os.ReadFile(path); rErr == nil {
		var obj map[string]any
		if json.Unmarshal(data, &obj) == nil {
			resp.MetaAfter = obj
		}
	}
}

func enrichProbes(req *Request, resp *Response) {
	if p := strings.TrimSpace(req.ArgvProbePath); p != "" {
		if data, err := os.ReadFile(p); err == nil {
			resp.ArgvProbe = string(data)
		}
	}
	if p := strings.TrimSpace(req.CwdProbePath); p != "" {
		if data, err := os.ReadFile(p); err == nil {
			resp.CwdProbe = strings.TrimSpace(string(data))
		}
	}
}

func runAgentRun(t *testing.T, req *Request, args ...string) (*Response, error) {
	t.Helper()
	if len(args) == 0 {
		args = req.Args
	}
	applyOpenInstantAttach(req)
	applyGrokTTYCommand(req)
	applyIterm2TestHooks(req)
	workDir := req.WorkDir
	if workDir == "" {
		workDir = req.TempDir
	}
	resp, err := execCmdWithBase(t, req.AgentRun, args, workDir, buildExecEnv(req), req.ExecTimeout)
	if err != nil {
		return resp, err
	}
	if resp == nil {
		return resp, nil
	}
	switch req.Mode {
	case "read-meta":
		enrichMetaAfter(req, resp)
	case "read-probes":
		enrichProbes(req, resp)
		enrichMetaAfter(req, resp)
	}
	// Always attach probes when paths are set (leaves may assert without Mode).
	if req.Mode == "" {
		enrichProbes(req, resp)
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

func assertNotContains(t *testing.T, got, ban string) {
	t.Helper()
	if strings.Contains(got, ban) {
		t.Fatalf("must not contain %q in:\n%s", ban, got)
	}
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

func readItermScript(t *testing.T, req *Request) string {
	t.Helper()
	path := strings.TrimSpace(req.ItermScriptOut)
	if path == "" {
		path = filepath.Join(req.TempDir, "iterm-script.applescript")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read iTerm script %s: %v", path, err)
	}
	return string(data)
}

func assertItermForceNewScript(t *testing.T, script string) {
	t.Helper()
	if !strings.Contains(script, "create window") {
		t.Fatalf("ModeForceNew script must create a window; script:\n%s", script)
	}
	// ForceNew skips session scan / tab reuse branch.
	if strings.Contains(script, "create tab") {
		t.Fatalf("ModeForceNew script must not create tabs (session scan path); script:\n%s", script)
	}
}

func assertNoItermScript(t *testing.T, req *Request) {
	t.Helper()
	path := strings.TrimSpace(req.ItermScriptOut)
	if path == "" {
		path = filepath.Join(req.TempDir, "iterm-script.applescript")
	}
	if !fileExists(path) {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	if strings.TrimSpace(string(data)) != "" {
		t.Fatalf("expected no iTerm script (MODE=send ignores --new-terminal); path=%s content:\n%s", path, data)
	}
}

func assertNoInProcessProviderSpawn(t *testing.T, req *Request, resp *Response) {
	t.Helper()
	probePath := strings.TrimSpace(req.ArgvProbePath)
	if probePath == "" {
		return
	}
	if !fileExists(probePath) {
		return
	}
	data, err := os.ReadFile(probePath)
	if err != nil {
		return
	}
	if strings.TrimSpace(string(data)) != "" {
		t.Fatalf("launcher with --new-terminal must not spawn provider in-process; argv probe:\n%s\nstderr:\n%s\nstdout:\n%s", data, resp.Stderr, resp.Stdout)
	}
}

func canonicalPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	if eval, err := filepath.EvalSymlinks(abs); err == nil {
		return eval
	}
	return abs
}

func sendQueuePath(home, runner, terminalID string) string {
	return filepath.Join(home, "send-queue", runner, terminalID+".jsonl")
}

func Setup(t *testing.T, req *Request) error {
	req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.WorkDir = req.TempDir
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
	req.Env = withoutEnvKey(req.Env, envGrokTTYCommand)
	req.Runner = defaultRunner
	req.Workspace = req.TempDir
	return nil
}
```
