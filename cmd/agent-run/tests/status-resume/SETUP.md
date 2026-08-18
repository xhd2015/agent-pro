# Scenario

**Feature**: session-level `status` multi-layer probe, `resume` as run shortcut,
and `run --open` grok session bind (post-exit finalize + background bind wait)

```
# status probe
seed meta + optional registry/ptywrap
  -> agent-run status <ref> [--json]
  -> session / process / terminal / runner.exited / resume.ready
     (runner.status may be binding|bound|unbound)
  # zombie keep-alive: process alive + terminal reachable + exit scrollback
  #   => exited true / resume ready (not reachable-alone = still active)

# resume gate + run shortcut (+ zombie terminal reclaim)
seed meta (bound+exited | live | unbound | missing)
  [+ optional zombie registry: alive PID + reachable + exit scrollback]
  -> agent-run resume [flags] <id> ["followup"]
  -> deny (exit 1; live → send, not already-in-use)
  -> or reclaim zombie terminal id then run path with --resume <runner_session_id>
  -> --no-submit without --open → error

# open post-exit finalize
agent-run run --open (+ instant attach hook + optional GROK_HOME seed)
  -> after attach: print/persist grok session or error not resolved

# open background bind
agent-run run --open (+ instant attach + delayed GROK_HOME materialization)
  -> bind worker runs without blocking attach
  -> detach then ALWAYS wait for bind
  -> stderr grok session + meta.runner_session_id (or hard-fail error)
```

## Preconditions

- Repository contains `cmd/agent-run`.
- Nested DOCTEST root: does not inherit parent `cmd/agent-run/tests` Setup/Run.
- Repo root from this tree: `d.DOCTEST_ROOT/../../../..`
  (`status-resume` → `tests` → `agent-run` → `cmd` → module root).
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Session-scoped build cache: `$TMPDIR/agent-run-status-resume-doctest-<d.DOCTEST_SESSION_ID>/`
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
	"runtime"
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
	"github.com/xhd2015/doctest/session"
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

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "agent-run-status-resume-doctest-"+d.DOCTEST_SESSION_ID)
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
		// Satisfy //go:embed dist in frontend-agent-run (and frontend if linked).
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
	if req.NoGrokHomeEnv {
		// O1: strip ambient hard-require env so only non-empty prompt forces hard wait.
		// GrokHome is still used for materialization under isolated HOME/.grok.
		req.Env = withoutEnvKey(req.Env, "GROK_HOME")
		req.Env = withoutEnvKey(req.Env, "AGENT_RUNNER_CONFIG_HOME")
		req.Env = withoutEnvKey(req.Env, envGrokTTYGrokSession)
		return
	}
	if strings.TrimSpace(req.GrokHome) == "" {
		return
	}
	setEnvKV(req, "GROK_HOME", req.GrokHome)
	if id := strings.TrimSpace(req.GrokSessionUUID); id != "" {
		setEnvKV(req, envGrokTTYGrokSession, id)
	}
}

// buildExecEnv builds the child process environment. When NoGrokHomeEnv is set,
// strip ambient GROK_HOME / config-home / session-id hook from the parent environ
// so they cannot force hard-require or pin discovery.
func buildExecEnv(req *Request) []string {
	base := os.Environ()
	if req.NoGrokHomeEnv {
		base = withoutEnvKey(base, "GROK_HOME")
		base = withoutEnvKey(base, "AGENT_RUNNER_CONFIG_HOME")
		base = withoutEnvKey(base, envGrokTTYGrokSession)
	}
	return append(base, req.Env...)
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
	// Flat layout: sessions/<session_id>/meta.json (runner is meta field only).
	if err := store.CreateSession(req.SessionID, meta); err != nil {
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

// seedExtraSessionMeta writes an additional flat session without mutating req identity fields.
// Used for ambiguous / multi-session lookup fixtures (e.g. two metas share runner_session_id).
func seedExtraSessionMeta(t *testing.T, req *Request, sessionID, runner, runnerSessionID, metaStatus, terminalID, prompt string) {
	t.Helper()
	ensureDefaults(req)
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		t.Fatal("seedExtraSessionMeta requires sessionID")
	}
	if runner == "" {
		runner = defaultRunner
	}
	if metaStatus == "" {
		metaStatus = "finished"
	}
	store := openAgentStore(t, req)
	meta := agentstorage.SessionMeta{
		Runner:            runner,
		SessionID:         sessionID,
		Status:            metaStatus,
		RunnerSessionID:   runnerSessionID,
		TerminalSessionID: terminalID,
		Workspace:         req.Workspace,
		Model:             req.Model,
		InitialPrompt:     prompt,
		CreatedAt:         "2026-07-03T12:00:00Z",
		UpdatedAt:         "2026-07-03T12:00:00Z",
	}
	if err := store.CreateSession(sessionID, meta); err != nil {
		path := metaJSONPath(req.Home, runner, sessionID)
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
	}
}

func metaJSONPath(home, runner, sessionID string) string {
	// Flat layout: sessions/<session_id>/meta.json. runner kept for call-site compat.
	_ = runner
	return filepath.Join(home, "sessions", sessionID, "meta.json")
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

// zombieServeExitScrollback is the post-/exit keep-alive serve scrollback:
// resume footer + [Terminal exited], no idle Grok prompt (sendable: no).
func zombieServeExitScrollback(runnerSessionID string) string {
	id := strings.TrimSpace(runnerSessionID)
	if id == "" {
		id = "550e8400-e29b-41d4-a716-446655440222"
	}
	return "Resume this session with:\n  grok --resume " + id + "\n[Terminal exited]\n"
}

// startDetachedSleepPID starts a long-lived child process used as a keep-alive
// serve PID in reclaim fixtures. Prefer this over RegistryPID=0 (test process)
// when production code may tear down the registry PID on zombie reclaim.
func startDetachedSleepPID(t *testing.T) int {
	t.Helper()
	cmd := exec.Command("sleep", "3600")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start detached sleep (zombie serve PID): %v", err)
	}
	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_, _ = cmd.Process.Wait()
		}
	})
	return cmd.Process.Pid
}

// seedZombieServeAfterExit models keep-alive __serve__ still TCP-reachable and
// process alive after the agent /exit (child gone), with exit markers in
// scrollback. Status must report runner.exited true and resume.ready yes when
// bound — not treat reachable alone as still active.
//
// RegistryPID: leave 0 for os.Getpid() (status leaves); set >0 before call for a
// detached serve PID (resume reclaim leaves — safe if reclaim kills the PID).
func seedZombieServeAfterExit(t *testing.T, req *Request) {
	t.Helper()
	req.SeedMeta = true
	req.MetaStatus = "running"
	ensureDefaults(req)
	if req.SessionID == "" {
		req.SessionID = "test-zombie-s1"
	}
	if req.RunnerSessionID == "" {
		req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440222"
	}
	if req.TerminalSessionID == "" {
		req.TerminalSessionID = "term-zombie-1"
	}
	if req.InitialPrompt == "" {
		req.InitialPrompt = "zombie after exit"
	}
	req.StartFakePTYWrap = true
	req.WriteRegistry = true
	// Preserve caller-set positive PID (detached zombie serve). Negative → force
	// alive default (0 = test process via resolveRegistryPID).
	if req.RegistryPID < 0 {
		req.RegistryPID = 0
	}
	if req.FakePTYWrapScrollback == "" {
		req.FakePTYWrapScrollback = zombieServeExitScrollback(req.RunnerSessionID)
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
	path, err := writeFakeGrokSessionDirErr(grokHome, workspace, sessionUUID, prompt)
	if err != nil {
		t.Fatalf("write fake grok session: %v", err)
	}
	return path
}

// writeFakeGrokSessionDirAtCwd seeds a session under sessionCwd's encoded path
// (and summary.info.cwd), which may differ from the agent-run workspace.
func writeFakeGrokSessionDirAtCwd(t *testing.T, grokHome, sessionCwd, sessionUUID, prompt string) string {
	t.Helper()
	path, err := writeFakeGrokSessionDirErr(grokHome, sessionCwd, sessionUUID, prompt)
	if err != nil {
		t.Fatalf("write fake grok session at cwd: %v", err)
	}
	return path
}

func writeFakeGrokSessionDirErr(grokHome, sessionCwd, sessionUUID, prompt string) (string, error) {
	dir := grokSessionDir(grokHome, sessionCwd, sessionUUID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("mkdir grok session: %w", err)
	}
	abs, _ := filepath.Abs(sessionCwd)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	summary := map[string]any{
		"info": map[string]any{
			"cwd":       abs,
			"sessionId": sessionUUID,
			"openedAt":  now,
		},
		"created_at": now,
	}
	sb, err := json.Marshal(summary)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), sb, 0644); err != nil {
		return "", fmt.Errorf("write summary.json: %w", err)
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
		return "", fmt.Errorf("write updates.jsonl: %w", err)
	}
	return updatesPath, nil
}

// openPromptForDiscover picks the prompt text used when materializing delayed
// GROK_HOME session files (must match the open inject / DiscoverSession match).
func openPromptForDiscover(req *Request) string {
	if p := strings.TrimSpace(req.OpenPrompt); p != "" {
		return p
	}
	if p := strings.TrimSpace(req.InitialPrompt); p != "" {
		return p
	}
	if n := len(req.Args); n > 0 {
		last := strings.TrimSpace(req.Args[n-1])
		if last != "" && !strings.HasPrefix(last, "-") {
			return last
		}
	}
	return "open bind"
}

// scheduleDelayedGrokMaterialize starts a background write of summary.json +
// updates.jsonl after req.GrokMaterializeDelay. No-op when delay is zero.
// Session path uses GrokSessionCwd when set, else Workspace / TempDir.
func scheduleDelayedGrokMaterialize(t *testing.T, req *Request) {
	t.Helper()
	if req.GrokMaterializeDelay <= 0 {
		return
	}
	if strings.TrimSpace(req.GrokHome) == "" || strings.TrimSpace(req.GrokSessionUUID) == "" {
		t.Fatal("GrokMaterializeDelay requires GrokHome and GrokSessionUUID")
	}
	delay := req.GrokMaterializeDelay
	grokHome := req.GrokHome
	sessionCwd := strings.TrimSpace(req.GrokSessionCwd)
	if sessionCwd == "" {
		sessionCwd = req.Workspace
	}
	if sessionCwd == "" {
		sessionCwd = req.TempDir
	}
	uuid := req.GrokSessionUUID
	prompt := openPromptForDiscover(req)
	go func() {
		time.Sleep(delay)
		path, err := writeFakeGrokSessionDirErr(grokHome, sessionCwd, uuid, prompt)
		if err != nil {
			// Best-effort log; Assert will fail if discovery never succeeds.
			fmt.Fprintf(os.Stderr, "doctest delayed grok materialize: %v\n", err)
			return
		}
		req.GrokUpdatesPath = path
	}()
}

// findMetaRunnerSessionID walks flat sessions/<id>/ for meta.json whose
// runner_session_id equals want (or any non-empty id when want == "").
// When runner is non-empty, only metas with meta.runner == runner match.
func findMetaRunnerSessionID(t *testing.T, home, runner, want string) (metaPath, gotID string, ok bool) {
	t.Helper()
	root := filepath.Join(home, "sessions")
	_ = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info == nil || info.IsDir() {
			return nil
		}
		if info.Name() != "meta.json" {
			return nil
		}
		data, rErr := os.ReadFile(path)
		if rErr != nil {
			return nil
		}
		var meta map[string]any
		if json.Unmarshal(data, &meta) != nil {
			return nil
		}
		if runner != "" {
			if r, _ := meta["runner"].(string); strings.TrimSpace(r) != runner {
				return nil
			}
		}
		id, _ := meta["runner_session_id"].(string)
		id = strings.TrimSpace(id)
		if id == "" {
			return nil
		}
		if want == "" || id == want {
			metaPath = path
			gotID = id
			ok = true
			return io.EOF // stop walk
		}
		return nil
	})
	return metaPath, gotID, ok
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
	return execCmdWithBase(t, command, args, dir, append(os.Environ(), env...), timeout)
}

// execCmdWithBase is like execCmd but uses a pre-built base environment slice
// (used so NoGrokHomeEnv can strip ambient keys before appending req.Env).
func execCmdWithBase(t *testing.T, command string, args []string, dir string, fullEnv []string, timeout time.Duration) (*Response, error) {
	t.Helper()
	if timeout <= 0 {
		timeout = 45 * time.Second
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

func runOpenWithMidStatusProbe(t *testing.T, req *Request, args []string) (*Response, error) {
	t.Helper()
	timeout := req.ExecTimeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	scheduleDelayedGrokMaterialize(t, req)
	applyOpenInstantAttach(req)
	applyGrokTTYCommand(req)
	applyGrokHomeEnv(req)

	cmd := exec.CommandContext(ctx, req.AgentRun, args...)
	cmd.Dir = req.TempDir
	cmd.Env = buildExecEnv(req)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	if err := cmd.Start(); err != nil {
		return &Response{Stdout: stdout.String(), Stderr: stderr.String(), Err: err, Elapsed: time.Since(start)}, err
	}

	runner := req.Runner
	if runner == "" {
		runner = defaultRunner
	}
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, fmt.Errorf("open-status-mid requires Request.SessionID (use --session)")
	}
	metaPath := metaJSONPath(req.Home, runner, sessionID)

	// Wait until open creates session meta (bind may still be in progress).
	deadline := time.Now().Add(15 * time.Second)
	for {
		if _, err := os.Stat(metaPath); err == nil {
			break
		}
		if time.Now().After(deadline) {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			return &Response{
				Stdout:  stdout.String(),
				Stderr:  stderr.String(),
				Elapsed: time.Since(start),
			}, fmt.Errorf("timeout waiting for session meta %s\nstderr:\n%s", metaPath, stderr.String())
		}
		select {
		case <-ctx.Done():
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			return &Response{Stdout: stdout.String(), Stderr: stderr.String(), Elapsed: time.Since(start)}, ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}

	probeAfter := req.StatusProbeAfter
	if probeAfter <= 0 {
		probeAfter = 400 * time.Millisecond
	}
	select {
	case <-ctx.Done():
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return &Response{Stdout: stdout.String(), Stderr: stderr.String(), Elapsed: time.Since(start)}, ctx.Err()
	case <-time.After(probeAfter):
	}

	probeArgs := req.StatusProbeArgs
	if len(probeArgs) == 0 {
		probeArgs = []string{"status", "--json", sessionID}
	}
	probeResp, probeErr := execCmdWithBase(t, req.AgentRun, probeArgs, req.TempDir, buildExecEnv(req), 15*time.Second)
	if probeErr != nil && probeResp == nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, fmt.Errorf("status probe: %w", probeErr)
	}

	waitErr := cmd.Wait()
	elapsed := time.Since(start)
	resp := &Response{
		Stdout:  stdout.String(),
		Stderr:  stderr.String(),
		Elapsed: elapsed,
	}
	if waitErr == nil {
		resp.ExitCode = 0
	} else if ctx.Err() != nil {
		resp.Err = ctx.Err()
		return resp, ctx.Err()
	} else {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			resp.ExitCode = exitErr.ExitCode()
		} else {
			resp.Err = waitErr
			return resp, waitErr
		}
	}
	if probeResp != nil {
		resp.StatusProbeStdout = probeResp.Stdout
		resp.StatusProbeStderr = probeResp.Stderr
		resp.StatusProbeExit = probeResp.ExitCode
		if strings.TrimSpace(probeResp.Stdout) != "" {
			var obj map[string]any
			if jErr := json.Unmarshal([]byte(probeResp.Stdout), &obj); jErr == nil {
				resp.StatusProbeJSON = obj
			}
		}
	}
	enrichMetaAfter(req, resp)
	return resp, nil
}

func runAgentRun(t *testing.T, req *Request, args ...string) (*Response, error) {
	t.Helper()
	if len(args) == 0 {
		args = req.Args
	}
	if req.Mode == "open-status-mid" {
		return runOpenWithMidStatusProbe(t, req, args)
	}
	scheduleDelayedGrokMaterialize(t, req)
	applyOpenInstantAttach(req)
	applyGrokTTYCommand(req)
	applyGrokHomeEnv(req)
	start := time.Now()
	resp, err := execCmdWithBase(t, req.AgentRun, args, req.TempDir, buildExecEnv(req), req.ExecTimeout)
	if resp != nil {
		resp.Elapsed = time.Since(start)
	}
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
		enrichMetaAfter(req, resp)
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
	// Ensure argv-sensitive leaves are not polluted by ambient hooks.
	req.Env = withoutEnvKey(req.Env, envGrokTTYCommand)
	req.Runner = defaultRunner
	req.Workspace = req.TempDir
	return nil
}
```
