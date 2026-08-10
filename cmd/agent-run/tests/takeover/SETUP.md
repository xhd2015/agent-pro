# Scenario

**Feature**: `agent-run takeover` — adopt a provider session into agent-run

```
# P1 — help + flag / arg validation (L2 Mode handle)
agent-run --help -> lists takeover
agent-run takeover --help -> usage, <session-id>, --grok, --codex, --agent-runner, --dry-run
agent-run takeover -> missing session-id (non-zero)
agent-run takeover --grok --codex <id> -> mutually exclusive
…

# P2 — Grok lifecycle (L2 Mode handle + injectable hooks)
agent-run takeover --grok <provider-uuid>
  -> resolve provider session under GROK_HOME
  -> probe live PIDs (ListProcs/Lsof hooks)
  -> already managed? warning + exit 0 (no kill, no iTerm)
  -> live native? kill + iTerm ForceNew resume/import
  -> not running? import or mapped-resume + iTerm ForceNew
  -> --dry-run? plan only (no kill, no meta, no iTerm)

# P3 — Codex lifecycle (parity with P2; --codex / codex-tty)
agent-run takeover --codex <provider-uuid>
  -> resolve under CODEX_HOME (sessions/YYYY/MM/DD/rollout-*-<uuid>.jsonl)
  -> same managed / kill / import / resume / dry-run matrix

# P4 — Auto-detect when --grok/--codex/--agent-runner omitted
agent-run takeover <provider-uuid>
  -> Find under GROK_HOME then CODEX_HOME
  -> exactly one hit -> run that provider lifecycle
  -> both hit -> ambiguous error
  -> neither -> not found / cannot resolve
```

## Preconditions

- Inherits root harness: `Request` / `Response` / `Run` (Mode `"handle"` preferred).
- Each leaf gets isolated `AGENT_RUN_HOME` under `t.TempDir()` from root Setup.
- Takeover root Setup isolates `GROK_HOME` (`TempDir/grok-home`) and `CODEX_HOME`
  (`TempDir/.codex` — path contains `/.codex/sessions/` for procresolve hard-hit)
  via `req.Env`. P1 leaves do not require provider fixtures; empty homes are fine.
- **No product process kill of real host PIDs** for synthetic hook PIDs — Kill is
  recorded to a log file via injectable hooks (see Context).
- Parallel-safe: no `t.Setenv` / `os.Setenv` / `t.Chdir` in leaves; use `req.Env`
  (Handle applies env under `handleStdoutMu`) and temp dirs only.

## Steps

1. Root Setup prepares home / env (inherited) + GROK_HOME + CODEX_HOME + Mode handle.
2. Grouping / leaf Setup seeds fixtures (Grok/Codex, meta, registry, hooks JSON)
   and finalizes `req.Args`.
3. `Run` calls `agentruncli.Handle` in-process (Mode handle).
4. Assert checks exit code, stdout/stderr, meta, iTerm script, kill log.

## Context

### CLI

```text
agent-run takeover [OPTIONS] <session-id>
```

- `<session-id>` = **provider** session UUID (Grok/Codex), not agent-run bare id.
- `--grok` / `--codex` alias `--agent-runner=grok-tty` / `codex-tty`.
- P2 leaves are **Grok** (`--grok`); P3 leaves are **Codex** (`--codex`).
- P4 leaves omit runner flags and rely on provider home lookup (auto-detect).

### Injectable deps (implementer contract — L2)

Product should accept injectable takeover deps (package hooks or options):

| Hook | Role |
|------|------|
| `ListProcs` | `[]procresolve.Proc` snapshot |
| `Lsof` | open files per pid |
| `Kill` | `kill(pid, sig)` — tests record to kill-log; no host harm |
| `WaitDead` | wait until pid gone (or timeout) |
| `OpenIterm` | ForceNew window; production uses `iterm2` + `KOOL_ITERM2_*` |

**Test wiring (env, parallel-safe via `req.Env`):**

| Env | Purpose |
|-----|---------|
| `GROK_HOME` | isolated Grok provider home (set by root Setup) |
| `CODEX_HOME` | isolated Codex provider home (`TempDir/.codex`; product: `CodexHome()` / `CodexHomeForRunner`) |
| `AGENT_RUN_HOME` | isolated agent-run home (parent Setup) |
| `AGENT_RUN_TAKEOVER_TEST_HOOKS` | path to JSON snapshot (procs + open_files + kill_log) |
| `KOOL_ITERM2_INSTALLED=1` | treat iTerm as installed |
| `KOOL_ITERM2_SCRIPT_OUT` | capture AppleScript instead of osascript |
| `KOOL_ITERM2_GOOS=darwin` | platform check inside iterm2 package |

**Hooks JSON shape** (written by `writeTakeoverHooks`):

```json
{
  "procs": [{"pid": 9001, "ppid": 1, "cmd": "/usr/local/bin/grok"}],
  "open_files": {"9001": ["/abs/.../sessions/<esc-cwd>/<uuid>/events.jsonl"]},
  "kill_log": "/abs/.../takeover-kill.log"
}
```

Codex open-file hard hits use rollout paths under `…/.codex/sessions/YYYY/MM/DD/rollout-…-<uuid>.jsonl`
(basename `codex` on the process cmd). Registry for codex-tty: `codex-tty-registry/`.

Implementer: when `AGENT_RUN_TAKEOVER_TEST_HOOKS` is set, wire `ListProcs`/`Lsof`
from the file and append `SIGTERM`/`SIGKILL` lines to `kill_log` instead of (or
before) real signals for synthetic PIDs. Prefer package-level `TakeoverDeps` for
unit injectability; env snapshot is the L2 Handle seam. Codex path must honor
`CODEX_HOME` (and `agent/codex/sessions.Find`).

### Expected CLI messages (P2/P3)

```text
# success kill+open
killed pid N (grok|codex)
session-id: …
provider: …
opened new iTerm2 window

# already managed
warning: session … is already managed by agent-run …; nothing to take over

# dry-run
dry-run: would kill pid …
dry-run: would open iTerm2 with: …
```

```go
import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/doctest/session"
)

// takeoverFixtureSessionID is a stable provider-shaped UUID for validation and
// Grok lifecycle leaves (provider session id, not agent-run bare id).
const takeoverFixtureSessionID = "550e8400-e29b-41d4-a716-446655440abc"

// takeoverCodexFixtureSessionID is the default Codex provider UUID for P3 leaves.
const takeoverCodexFixtureSessionID = "660e8400-e29b-41d4-a716-446655440c0d"

// takeoverAutoDetectSharedID is used when both GROK_HOME and CODEX_HOME must
// contain the same provider UUID (ambiguous auto-detect).
const takeoverAutoDetectSharedID = "770e8400-e29b-41d4-a716-446655440ad0"

const (
	envTakeoverTestHooks = "AGENT_RUN_TAKEOVER_TEST_HOOKS"
	envIterm2Installed   = "KOOL_ITERM2_INSTALLED"
	envIterm2ScriptOut   = "KOOL_ITERM2_SCRIPT_OUT"
	envIterm2GOOS        = "KOOL_ITERM2_GOOS"

	takeoverGrokHomeDir    = "grok-home"
	// .codex so open-file paths contain "/.codex/sessions/" (procresolve hard-hit).
	takeoverCodexHomeDir   = ".codex"
	takeoverKillLogFile    = "takeover-kill.log"
	takeoverHooksFile      = "takeover-hooks.json"
	takeoverItermScript    = "iterm-script.applescript"
	takeoverRegistryDir    = "grok-tty-registry"
	takeoverCodexRegistryDir = "codex-tty-registry"
	takeoverFixturePIDFile = "takeover-fixture.pid"
	takeoverDefaultListen  = "127.0.0.1:1"
	takeoverDefaultCreated = "2026-07-03T12:00:00Z"
)

// takeoverHooksDoc is the JSON snapshot for injectable ListProcs/Lsof/Kill.
type takeoverHooksDoc struct {
	Procs     []takeoverHookProc  `json:"procs"`
	OpenFiles map[string][]string `json:"open_files"` // pid string -> paths
	KillLog   string              `json:"kill_log"`
}

type takeoverHookProc struct {
	PID  int    `json:"pid"`
	PPID int    `json:"ppid"`
	Cmd  string `json:"cmd"`
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Default all takeover leaves to L2 in-process Handle.
	req.Mode = "handle"
	if req.TempDir == "" {
		// Parent root Setup should have filled TempDir; keep P1-safe no-op.
		return nil
	}
	grokHome := takeoverGrokHome(req)
	if err := os.MkdirAll(grokHome, 0755); err != nil {
		return fmt.Errorf("mkdir GROK_HOME: %w", err)
	}
	setTakeoverEnv(req, "GROK_HOME", grokHome)
	codexHome := takeoverCodexHome(req)
	if err := os.MkdirAll(codexHome, 0755); err != nil {
		return fmt.Errorf("mkdir CODEX_HOME: %w", err)
	}
	// Product should resolve via agenttty.CodexHome() / CodexHomeForRunner (CODEX_HOME).
	setTakeoverEnv(req, "CODEX_HOME", codexHome)
	return nil
}

func takeoverGrokHome(req *Request) string {
	return filepath.Join(req.TempDir, takeoverGrokHomeDir)
}

func takeoverCodexHome(req *Request) string {
	return filepath.Join(req.TempDir, takeoverCodexHomeDir)
}

func setTakeoverEnv(req *Request, key, value string) {
	prefix := key + "="
	out := make([]string, 0, len(req.Env)+1)
	for _, e := range req.Env {
		if strings.HasPrefix(e, prefix) {
			continue
		}
		out = append(out, e)
	}
	req.Env = append(out, key+"="+value)
}

func withoutTakeoverEnvKey(env []string, key string) []string {
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

func assertTrailingNewline(t *testing.T, s, label string) {
	t.Helper()
	if s == "" || !strings.HasSuffix(s, "\n") {
		t.Fatalf("%s: expected trailing newline, got %q", label, s)
	}
}

func assertContainsAny(t *testing.T, got string, options ...string) {
	t.Helper()
	lower := strings.ToLower(got)
	for _, opt := range options {
		if strings.Contains(lower, strings.ToLower(opt)) {
			return
		}
	}
	t.Fatalf("expected one of %v in:\n%s", options, got)
}

// assertTakeoverRecognized fails while Handle still returns unknown command: takeover.
// Classic TDD: validation leaves stay RED until the subcommand is registered.
func assertTakeoverRecognized(t *testing.T, stderr string) {
	t.Helper()
	lower := strings.ToLower(stderr)
	if strings.Contains(lower, "unknown command") {
		t.Fatalf("takeover not implemented yet (unknown command); want command-specific validation, got:\n%s", stderr)
	}
}

// assertTakeoverActionImplemented fails while the action body is still stubbed
// ("takeover: not implemented" / "codex support is not implemented yet").
// Lifecycle leaves use this so P2/P3 stay RED until the implementer lands
// behavior (without weakening P1 validation).
func assertTakeoverActionImplemented(t *testing.T, combined string) {
	t.Helper()
	if strings.Contains(strings.ToLower(combined), "not implemented") {
		t.Fatalf("takeover action still stubbed (not implemented); lifecycle contract not met:\n%s", combined)
	}
}

func absPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatalf("abs %q: %v", p, err)
	}
	return abs
}

func encodedGrokCwd(workspace string) string {
	abs, err := filepath.Abs(workspace)
	if err != nil {
		abs = workspace
	}
	return url.PathEscape(abs)
}

func grokSessionDir(grokHome, workspace, sessionID string) string {
	return filepath.Join(grokHome, "sessions", encodedGrokCwd(workspace), sessionID)
}

// seedGrokSession writes a Find-compatible summary.json under GROK_HOME.
// Returns absolute session directory path.
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
		"generated_title":   "takeover P2 fixture",
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
	// Touch an events path so open-file hard-hit fixtures can reference a real path.
	ev := filepath.Join(dir, "events.jsonl")
	if err := os.WriteFile(ev, []byte("{}\n"), 0644); err != nil {
		t.Fatalf("write events.jsonl: %v", err)
	}
	return dir
}

func grokOpenPath(grokHome, workspace, sessionID string) string {
	return filepath.Join(grokSessionDir(grokHome, workspace, sessionID), "events.jsonl")
}

// codexRolloutPath returns the conventional rollout path under CODEX_HOME.
// Layout: sessions/YYYY/MM/DD/rollout-<ts>-<uuid>.jsonl
func codexRolloutPath(codexHome, sessionID string) string {
	return filepath.Join(
		codexHome, "sessions", "2026", "07", "02",
		"rollout-2026-07-02T12-01-53-"+sessionID+".jsonl",
	)
}

// seedCodexSession writes a Find-compatible rollout jsonl under CODEX_HOME.
// First line is session_meta with payload.id + cwd (agent/codex/sessions.Find).
// Returns absolute rollout path (also used for open-file hard-hit fixtures).
func seedCodexSession(t *testing.T, codexHome, workspace, sessionID string) string {
	t.Helper()
	absCwd := absPath(t, workspace)
	path := codexRolloutPath(codexHome, sessionID)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir codex rollout dir: %v", err)
	}
	// session_meta uses payload.id (not session_id) for parseSessionMeta.
	metaLine, err := json.Marshal(map[string]any{
		"type": "session_meta",
		"payload": map[string]any{
			"id":        sessionID,
			"timestamp": "2026-07-02T12:01:53.000Z",
			"cwd":       absCwd,
		},
	})
	if err != nil {
		t.Fatalf("marshal session_meta: %v", err)
	}
	content := string(metaLine) + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write rollout %s: %v", path, err)
	}
	return path
}

// codexOpenPath is the open-file path for LivePIDs hard-hit (rollout file).
// When CODEX_HOME is TempDir/.codex the path contains "/.codex/sessions/".
func codexOpenPath(codexHome, sessionID string) string {
	return codexRolloutPath(codexHome, sessionID)
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
// binding the provider UUID.
func seedMappedMeta(t *testing.T, req *Request, runner, agentSessionID, providerSessionID, status string) {
	t.Helper()
	if status == "" {
		status = "finished"
	}
	store := openAgentStore(t, req)
	ws := absPath(t, req.TempDir)
	meta := agentstorage.SessionMeta{
		Runner:            runner,
		SessionID:         agentSessionID,
		Status:            status,
		RunnerSessionID:   providerSessionID,
		TerminalSessionID: agentSessionID,
		Workspace:         ws,
	}
	if err := store.CreateSession(agentSessionID, meta); err != nil {
		t.Fatalf("CreateSession(%q): %v", agentSessionID, err)
	}
	req.SessionID = agentSessionID
}

func metaJSONPath(home, sessionID string) string {
	return filepath.Join(home, "sessions", sessionID, "meta.json")
}

func metaExists(home, sessionID string) bool {
	_, err := os.Stat(metaJSONPath(home, sessionID))
	return err == nil
}

func killLogPath(req *Request) string {
	return filepath.Join(req.TempDir, takeoverKillLogFile)
}

func hooksFilePath(req *Request) string {
	return filepath.Join(req.TempDir, takeoverHooksFile)
}

func itermScriptPath(req *Request) string {
	return filepath.Join(req.TempDir, takeoverItermScript)
}

// applyIterm2TestHooks enables iterm2 package test env via req.Env (no t.Setenv).
func applyIterm2TestHooks(req *Request) {
	path := itermScriptPath(req)
	setTakeoverEnv(req, envIterm2Installed, "1")
	setTakeoverEnv(req, envIterm2ScriptOut, path)
	setTakeoverEnv(req, envIterm2GOOS, "darwin")
	_ = os.Remove(path)
}

// writeTakeoverHooks writes the injectable proc/lsof/kill snapshot and sets
// AGENT_RUN_TAKEOVER_TEST_HOOKS on req.Env.
func writeTakeoverHooks(t *testing.T, req *Request, procs []takeoverHookProc, openFiles map[int][]string) {
	t.Helper()
	of := make(map[string][]string, len(openFiles))
	for pid, paths := range openFiles {
		of[strconv.Itoa(pid)] = paths
	}
	klog := killLogPath(req)
	_ = os.Remove(klog)
	doc := takeoverHooksDoc{
		Procs:     procs,
		OpenFiles: of,
		KillLog:   klog,
	}
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("marshal hooks: %v", err)
	}
	path := hooksFilePath(req)
	if err := os.WriteFile(path, append(b, '\n'), 0644); err != nil {
		t.Fatalf("write hooks %s: %v", path, err)
	}
	setTakeoverEnv(req, envTakeoverTestHooks, path)
}

// writeEmptyTakeoverHooks installs an empty proc snapshot (not running) + kill log path.
func writeEmptyTakeoverHooks(t *testing.T, req *Request) {
	t.Helper()
	writeTakeoverHooks(t, req, nil, nil)
}

func readKillLog(t *testing.T, req *Request) string {
	t.Helper()
	b, err := os.ReadFile(killLogPath(req))
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatalf("read kill log: %v", err)
	}
	return string(b)
}

func assertNoKillLog(t *testing.T, req *Request) {
	t.Helper()
	got := strings.TrimSpace(readKillLog(t, req))
	if got != "" {
		t.Fatalf("expected empty kill log (no kill), got:\n%s", got)
	}
}

func assertKillLogMentionsPID(t *testing.T, req *Request, pid int) {
	t.Helper()
	got := readKillLog(t, req)
	if got == "" {
		t.Fatalf("expected kill log to mention pid %d; log missing/empty", pid)
	}
	if !strings.Contains(got, strconv.Itoa(pid)) {
		t.Fatalf("kill log missing pid %d:\n%s", pid, got)
	}
}

func readItermScript(t *testing.T, req *Request) string {
	t.Helper()
	path := itermScriptPath(req)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read iTerm script %s: %v", path, err)
	}
	return string(b)
}

func assertItermForceNewScript(t *testing.T, script string) {
	t.Helper()
	if !strings.Contains(script, "create window") {
		t.Fatalf("ModeForceNew script must create a window; script:\n%s", script)
	}
	if strings.Contains(script, "create tab") {
		t.Fatalf("ModeForceNew script must not create tabs; script:\n%s", script)
	}
}

func assertNoItermScript(t *testing.T, req *Request) {
	t.Helper()
	path := itermScriptPath(req)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Fatalf("stat iTerm script: %v", err)
	}
	if strings.TrimSpace(string(b)) != "" {
		t.Fatalf("expected no iTerm script content; path=%s content:\n%s", path, b)
	}
}

func takeoverRegistryPath(home, sessionID string) string {
	return filepath.Join(home, takeoverRegistryDir, sessionID+".json")
}

func writeTakeoverRegistryEntry(t *testing.T, home, sessionID string, pid int) {
	t.Helper()
	writeTakeoverRegistryEntryIn(t, home, takeoverRegistryDir, sessionID, pid)
}

// writeTakeoverCodexRegistryEntry writes under codex-tty-registry/ (codex-tty provider).
func writeTakeoverCodexRegistryEntry(t *testing.T, home, sessionID string, pid int) {
	t.Helper()
	writeTakeoverRegistryEntryIn(t, home, takeoverCodexRegistryDir, sessionID, pid)
}

func writeTakeoverRegistryEntryIn(t *testing.T, home, registryDir, sessionID string, pid int) {
	t.Helper()
	dir := filepath.Join(home, registryDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir registry %s: %v", registryDir, err)
	}
	entry := map[string]any{
		"session_id":  sessionID,
		"listen_addr": takeoverDefaultListen,
		"pid":         pid,
		"created_at":  takeoverDefaultCreated,
	}
	b, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal registry: %v", err)
	}
	path := filepath.Join(dir, sessionID+".json")
	if err := os.WriteFile(path, append(b, '\n'), 0644); err != nil {
		t.Fatalf("write registry %s: %v", path, err)
	}
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, syscall.Signal(0))
	return err == nil
}

// startLiveSleepFixture starts sleep, records pid under TempDir for asserts.
func startLiveSleepFixture(t *testing.T, req *Request) int {
	t.Helper()
	cmd := exec.Command("sleep", "600")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start sleep fixture: %v", err)
	}
	pid := cmd.Process.Pid
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})
	time.Sleep(20 * time.Millisecond)
	if !processAlive(pid) {
		t.Fatalf("fixture pid %d not alive after start", pid)
	}
	pidPath := filepath.Join(req.TempDir, takeoverFixturePIDFile)
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(pid)), 0644); err != nil {
		t.Fatalf("write fixture pid: %v", err)
	}
	return pid
}

func fixturePID(t *testing.T, req *Request) int {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(req.TempDir, takeoverFixturePIDFile))
	if err != nil {
		t.Fatalf("read fixture pid: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		t.Fatalf("parse fixture pid %q: %v", b, err)
	}
	return pid
}

// combinedOut lowercases stdout+stderr for flexible matching.
func combinedOut(resp *Response) string {
	if resp == nil {
		return ""
	}
	return resp.Stderr + "\n" + resp.Stdout
}

// listAgentSessionIDs returns bare session ids under AGENT_RUN_HOME/sessions.
func listAgentSessionIDs(t *testing.T, home string) []string {
	t.Helper()
	dir := filepath.Join(home, "sessions")
	ents, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("readdir sessions: %v", err)
	}
	var ids []string
	for _, e := range ents {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			ids = append(ids, e.Name())
		}
	}
	return ids
}
```
