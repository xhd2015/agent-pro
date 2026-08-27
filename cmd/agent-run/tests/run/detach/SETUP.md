# Scenario

**Feature**: `agent-run run|resume --detach` starts a keep-alive TTY daemon and
exits after printing `session-id` + `terminal-id` (no attach/stream)

```
# validation
agent-run run --detach --open … -> error (mutually exclusive)
agent-run run --detach --json … -> error
agent-run run --detach --agent-runner fake-codex … -> TTY required
agent-run run --detach --no-submit … -> --no-submit requires --open

# run --detach (TTY + fake TUI)
agent-run run --agent-runner grok-tty --detach ["prompt"]
  -> keep-alive daemon registered
  -> soft grok bind (miss still exit 0)
  -> stdout:
       session-id: <agent storage id>
       terminal-id: <tty registry id>
  -> no attach; no event stream
  -> meta.status=running; registry TCP alive

# resume --detach
seed bound+exited -> agent-run resume --detach <id>
  -> reopen daemon; both ids; no attach

# auto
MODE=run|resume + --detach -> detach semantics
MODE=send + --detach -> note: ignored; send proceeds
```

## Preconditions

- Repository contains `cmd/agent-run` and `cmd/fake-codex`.
- Nested DOCTEST root: does not inherit parent `cmd/agent-run/tests` Setup/Run.
- Repo root from this tree: `d.DOCTEST_ROOT/../../../../..`
  (`detach` → `run` → `tests` → `agent-run` → `cmd` → module root).
- Each leaf uses isolated `AGENT_RUN_HOME` under `t.TempDir()`.
- Session-scoped build cache:
  `$TMPDIR/agent-run-run-detach-doctest-<d.DOCTEST_SESSION_ID>/`
  shares compiled `agent-run` + `fake-codex` across parallel leaves.
- `frontend-agent-run/dist` (and `frontend/dist` if present) may be absent
  (gitignored). Build Setup stubs a minimal `index.html` so `//go:embed dist`
  compiles.
- TTY lifecycle / resume / auto-run leaves set `AGENT_RUN_GROK_TTY_COMMAND` to a
  fake TUI that paints a banner and holds (keep-alive outlives parent).
- No real Grok CLI required. Soft-bind **miss** uses isolated `HOME` without
  grok sessions. Soft-bind **hit** is out of scope for this tree (optional/slow).
- Live MODE=send leaf uses in-process fake ptywrap (WS + inject HTTP).

## Steps

1. Root `Setup` resolves repo root, builds binaries once per session, sets
   `AGENT_RUN_HOME` + `PATH`, default runner `grok-tty`.
2. Grouping `Setup` narrows outcome class (help / reject / prompt-policy /
   tty-lifecycle / resume / auto).
3. Leaf `Setup` finalizes flags, prompt, TTY hooks, and optional meta/registry
   seeds.
4. `Run` executes `agent-run`; optional post-read of registry / meta from printed
   ids.
5. Leaf `Assert` checks exit code, id lines, silence, registry, meta.status,
   exclusivity errors, or live-send ignore note.

## Context

- Detach success stdout (product contract C):
  ```
  session-id: <id>
  terminal-id: <id>
  ```
  Both lines always; may be equal when `--session-id` is used.
- Soft bind budget is a product concern (1 minute); tests assert soft miss still
  exits 0 without requiring a full-minute wall clock.
- `--detach` does not use `AGENT_RUN_OPEN_ATTACH_INSTANT` (no attach).

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
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	envGrokTTYCommand  = "AGENT_RUN_GROK_TTY_COMMAND"
	envCodexTTYCommand = "AGENT_RUN_CODEX_TTY_COMMAND"
	defaultRunner      = "grok-tty"
	deadPIDSentinel    = 999999991
	defaultRegistryCreated = "2026-07-03T12:00:00Z"
)

var fakePTYWrapUpgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// modernGrokIdleScrollback is section-judge idle chrome (Worked for + boxed ❯).
// Legacy "Grok ›" is no longer CheckWritable-ready after status_above_composer.
func modernGrokIdleScrollback() string {
	return "" +
		"GROK_TTY_BANNER\n" +
		" ⎇ master worktree ~/.wrk/… 1K / 10K\n" +
		"    Worked for 1.0s                                        stop  [hooks: 1]\n" +
		" ╭--------------------------------------------------------------------------╮\n" +
		" │ ❯                                                                        │\n" +
		" ╰----------------------------------------- Grok 4.5 (high) · always-approve -╯\n" +
		" Shift+Tab:mode  │  Ctrl+.:shortcuts\n"
}

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "agent-run-run-detach-doctest-"+d.DOCTEST_SESSION_ID)
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

func buildOnce(t *testing.T, d *session.Doctest) (agentRun, fakeCodex string, err error) {
	t.Helper()
	cache := sessionCacheDir(d)
	agentRun = filepath.Join(cache, "agent-run")
	fakeCodex = filepath.Join(cache, "fake-codex")
	lock := filepath.Join(cache, "build.lock")
	ready := filepath.Join(cache, "binaries.ready")
	repoRoot, err := findAgentProRoot(d.DOCTEST_ROOT)
	if err != nil {
		return "", "", err
	}
	err = withFileLock(t, lock, func() error {
		if fileExists(ready) && fileExists(agentRun) && fileExists(fakeCodex) {
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
		build2 := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", fakeCodex, "./fake-codex")
		build2.Dir = repoRoot
		if out, err := build2.CombinedOutput(); err != nil {
			return fmt.Errorf("build fake-codex: %w\n%s", err, string(out))
		}
		return os.WriteFile(ready, []byte("ok"), 0644)
	})
	return agentRun, fakeCodex, err
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

func fakeTUIRespondHi() string {
	return `sh -c 'printf "GROK_TTY_BANNER\nGrok › "; read line; echo "Response: $line"'`
}

func fakeTUIHoldSeconds(sec int) string {
	return fmt.Sprintf(`sh -c 'printf "GROK_TTY_BANNER\nGrok › "; sleep %d'`, sec)
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
	}
}

func registryDir(home, runner string) string {
	return filepath.Join(home, runner+"-registry")
}

func registryPath(home, runner, sessionID string) string {
	return filepath.Join(registryDir(home, runner), sessionID+".json")
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
	entry := RegistryEntry{
		SessionID:  termID,
		ListenAddr: listenAddr,
		PID:        resolveRegistryPID(req),
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
	mux.HandleFunc("POST /api/terminal/sessions/{sessionID}/prepare-inject", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /api/terminal/sessions/{sessionID}/input", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		inputMu.Lock()
		*req.FakePTYInjectLog = append(*req.FakePTYInjectLog, string(body))
		inputMu.Unlock()
		w.WriteHeader(http.StatusOK)
	})
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
			scrollback = modernGrokIdleScrollback()
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
		req.FakePTYWrapScrollback = modernGrokIdleScrollback()
	}
	startFakePTYWrapServer(t, req)
	seedSessionMeta(t, req)
	writeRegistryEntry(t, req)
}

func applyGrokTTYCommand(req *Request) {
	if strings.TrimSpace(req.GrokTTYCommand) != "" {
		setEnvKV(req, envGrokTTYCommand, req.GrokTTYCommand)
	}
	if strings.TrimSpace(req.CodexTTYCommand) != "" {
		setEnvKV(req, envCodexTTYCommand, req.CodexTTYCommand)
	}
}

func applyNoGrokHomeEnv(req *Request) {
	if !req.NoGrokHomeEnv {
		return
	}
	req.Env = withoutEnvKey(req.Env, "GROK_HOME")
	req.Env = withoutEnvKey(req.Env, "AGENT_RUNNER_CONFIG_HOME")
	req.Env = withoutEnvKey(req.Env, "AGENT_RUN_GROK_TTY_GROK_SESSION_ID")
}

func buildExecEnv(req *Request) []string {
	return append(os.Environ(), req.Env...)
}

func execCmdWithBase(t *testing.T, command string, args []string, dir string, fullEnv []string, timeout time.Duration) (*Response, error) {
	t.Helper()
	if timeout <= 0 {
		timeout = 60 * time.Second
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

// parseLabeledID extracts first non-empty value for a "label: value" line.
func parseLabeledID(text, label string) (string, bool) {
	prefix := label + ":"
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(strings.ToLower(line), strings.ToLower(prefix)) {
			// also allow exact prefix match without lowercasing the value side
			if !strings.HasPrefix(line, prefix) {
				continue
			}
		}
		// Prefer exact prefix (product uses lowercase labels).
		if !strings.HasPrefix(line, prefix) {
			// case-insensitive fallback
			if !strings.HasPrefix(strings.ToLower(line), strings.ToLower(prefix)) {
				continue
			}
			// re-slice using original colon position
			idx := strings.Index(line, ":")
			if idx < 0 {
				continue
			}
			rest := strings.TrimSpace(line[idx+1:])
			if rest == "" || strings.Contains(rest, " ") {
				continue
			}
			if matched, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`, rest); matched {
				return rest, true
			}
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

func parseDetachIDs(stdout string) (sessionID, terminalID string, ok bool) {
	sessionID, okS := parseLabeledID(stdout, "session-id")
	terminalID, okT := parseLabeledID(stdout, "terminal-id")
	return sessionID, terminalID, okS && okT
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

// forbiddenDetachNoise is discovery/event stream noise that must not appear
// during --detach (parent exits without streaming a turn).
func forbiddenDetachNoise(combined string) []string {
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

func assertDetachIDsOnStdout(t *testing.T, resp *Response) (sessionID, terminalID string) {
	t.Helper()
	sessionID, terminalID, ok := parseDetachIDs(resp.Stdout)
	if !ok {
		// fall back to response fields filled by Run
		sessionID = resp.SessionID
		terminalID = resp.TerminalID
	}
	if sessionID == "" || terminalID == "" {
		t.Fatalf("stdout must print both session-id and terminal-id; got session=%q terminal=%q\nstdout:\n%s\nstderr:\n%s",
			sessionID, terminalID, resp.Stdout, resp.Stderr)
	}
	// Prefer labeled lines present even if values were also parsed elsewhere.
	if !strings.Contains(resp.Stdout, "session-id:") {
		t.Fatalf("stdout missing session-id: label\nstdout:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "terminal-id:") {
		t.Fatalf("stdout missing terminal-id: label\nstdout:\n%s", resp.Stdout)
	}
	return sessionID, terminalID
}

func assertNotUnknownFlag(t *testing.T, combined, flag string) {
	t.Helper()
	lower := strings.ToLower(combined)
	if (strings.Contains(lower, "unrecognized flag") ||
		strings.Contains(lower, "unknown flag") ||
		strings.Contains(lower, "flag provided but not defined") ||
		strings.Contains(lower, "unknown option")) &&
		strings.Contains(lower, strings.TrimPrefix(flag, "--")) {
		t.Fatalf("want product error for %s, got unknown/unrecognized flag style:\n%s", flag, combined)
	}
}

func sendQueuePath(home, runner, terminalID string) string {
	return filepath.Join(home, "send-queue", runner, terminalID+".jsonl")
}

func runAgentRun(t *testing.T, req *Request, args ...string) (*Response, error) {
	t.Helper()
	if len(args) == 0 {
		args = req.Args
	}
	applyGrokTTYCommand(req)
	applyNoGrokHomeEnv(req)
	workDir := req.WorkDir
	if workDir == "" {
		workDir = req.TempDir
	}
	resp, err := execCmdWithBase(t, req.AgentRun, args, workDir, buildExecEnv(req), req.ExecTimeout)
	if err != nil {
		return resp, err
	}
	if resp != nil {
		// Always parse detach ids when present.
		if sid, ok := parseLabeledID(resp.Stdout, "session-id"); ok {
			resp.SessionID = sid
		}
		if tid, ok := parseLabeledID(resp.Stdout, "terminal-id"); ok {
			resp.TerminalID = tid
		}
		runner := req.Runner
		if runner == "" {
			runner = defaultRunner
		}
		switch req.Mode {
		case "detach-registry-after", "read-meta+registry":
			id := resp.TerminalID
			if id == "" {
				id = resp.SessionID
			}
			if id != "" {
				if entry, rerr := readRegistryEntryOptional(req.Home, runner, id); rerr == nil {
					resp.RegistryEntry = entry
				} else if resp.SessionID != "" && resp.SessionID != id {
					if entry, rerr := readRegistryEntryOptional(req.Home, runner, resp.SessionID); rerr == nil {
						resp.RegistryEntry = entry
					}
				}
			}
			fallthrough
		case "read-meta":
			sid := req.SessionID
			if sid == "" {
				sid = resp.SessionID
			}
			if sid != "" {
				path := metaJSONPath(req.Home, sid)
				if data, rErr := os.ReadFile(path); rErr == nil {
					var obj map[string]any
					if json.Unmarshal(data, &obj) == nil {
						resp.MetaAfter = obj
					}
				}
			}
		}
	}
	return resp, nil
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	repoRoot, err := findAgentProRoot(d.DOCTEST_ROOT)
	if err != nil {
		return err
	}
	req.RepoRoot = repoRoot
	req.TempDir = t.TempDir()
	req.WorkDir = req.TempDir
	req.Home = filepath.Join(req.TempDir, ".agent-run")
	if err := os.MkdirAll(req.Home, 0755); err != nil {
		return fmt.Errorf("mkdir home: %w", err)
	}
	agentRun, fakeCodex, err := buildOnce(t, d)
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
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	req.Runner = defaultRunner
	req.Workspace = req.TempDir
	req.Args = []string{"run"}
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
