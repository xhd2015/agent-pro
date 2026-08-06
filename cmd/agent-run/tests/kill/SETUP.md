# Scenario

**Feature**: `agent-run kill` / `agent-run tty kill` — stop a live TTY session by registry id

```
# help / args validation (L2 Mode handle)
agent-run kill --help -> usage, session-id, --dry-run
agent-run kill -> missing session-id error
agent-run kill no-such-id -> not found (exit 1)

# live registry fixture under AGENT_RUN_HOME/<runner>-registry/
seed sleep PID + registry JSON
  -> agent-run kill [--dry-run] <session-id>
  -> stopped | dry-run: would stop | warning: not running
  -> ReclaimSessionID-style TERM→KILL + registry remove (product)
```

## Preconditions

- Inherits root harness: `Request` / `Response` / `Run` (Mode `"handle"` preferred).
- Each leaf gets isolated `AGENT_RUN_HOME` under `t.TempDir()` from root Setup.
- Live fixtures spawn a long-lived `sleep` process and write a
  `grok-tty-registry/<session-id>.json` entry (pid + listen_addr). Product kill
  should reclaim via registry read (not LookupSession TCP pruning).
- Optional session meta (`sessions/<id>/meta.json`) is seeded for known live
  sessions so implementers can distinguish unknown vs already-stopped.
- No production code under this tree; Classic TDD — leaves stay **RED** until
  `kill` / `tty kill` land in `agentruncli.Handle`.

## Steps

1. Root Setup prepares home / env (inherited).
2. Grouping / leaf Setup sets Mode handle, Args, and optional live fixture.
3. `Run` calls `agentruncli.Handle` in-process (Mode handle).
4. Assert checks exit code, stdout/stderr contract, process liveness, registry.

## Context

- Success stdout: `stopped <session-id>\n`
- Dry-run stdout: `dry-run: would stop <session-id>\n`
- Idempotent not-running: exit 0, stderr `warning: session <session-id> not running`
- Unknown session: exit non-zero, error text (not found / expired)

```go
import (
	"encoding/json"
	"fmt"
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

const (
	killRegistryDir     = "grok-tty-registry"
	killFixturePIDFile  = "kill-fixture.pid"
	killDefaultListen   = "127.0.0.1:1" // closed; reclaim reads registry by id, not TCP
	killDefaultCreated  = "2026-07-03T12:00:00Z"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Default all kill leaves to L2 in-process Handle.
	req.Mode = "handle"
	return nil
}

func killRegistryPath(home, sessionID string) string {
	return filepath.Join(home, killRegistryDir, sessionID+".json")
}

func writeKillRegistryEntry(t *testing.T, home, sessionID string, pid int, listenAddr string) {
	t.Helper()
	if listenAddr == "" {
		listenAddr = killDefaultListen
	}
	dir := filepath.Join(home, killRegistryDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir registry: %v", err)
	}
	entry := map[string]any{
		"session_id":  sessionID,
		"listen_addr": listenAddr,
		"pid":         pid,
		"created_at":  killDefaultCreated,
	}
	b, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal registry: %v", err)
	}
	path := killRegistryPath(home, sessionID)
	if err := os.WriteFile(path, append(b, '\n'), 0644); err != nil {
		t.Fatalf("write registry %s: %v", path, err)
	}
}

func seedKillSessionMeta(t *testing.T, req *Request, sessionID string) {
	t.Helper()
	store, err := agentstorage.NewFileStore(req.Home)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	meta := agentstorage.SessionMeta{
		Runner:            "grok-tty",
		SessionID:         sessionID,
		Status:            "running",
		TerminalSessionID: sessionID,
	}
	if err := store.CreateSession(sessionID, meta); err != nil {
		t.Fatalf("CreateSession(%q): %v", sessionID, err)
	}
}

// startLiveKillFixture spawns sleep, writes grok-tty registry + optional meta,
// records PID under req.TempDir for asserts. Sets req.SessionID.
func startLiveKillFixture(t *testing.T, req *Request, sessionID string, withMeta bool) {
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
	// Brief settle so process is visible to kill(2).
	time.Sleep(20 * time.Millisecond)
	if !processAlive(pid) {
		t.Fatalf("fixture pid %d not alive after start", pid)
	}
	writeKillRegistryEntry(t, req.Home, sessionID, pid, killDefaultListen)
	if withMeta {
		seedKillSessionMeta(t, req, sessionID)
	}
	req.SessionID = sessionID
	pidPath := filepath.Join(req.TempDir, killFixturePIDFile)
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(pid)), 0644); err != nil {
		t.Fatalf("write fixture pid: %v", err)
	}
}

func fixturePID(t *testing.T, req *Request) int {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(req.TempDir, killFixturePIDFile))
	if err != nil {
		t.Fatalf("read fixture pid: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		t.Fatalf("parse fixture pid %q: %v", b, err)
	}
	return pid
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, syscall.Signal(0))
	return err == nil
}

func registryExists(home, sessionID string) bool {
	_, err := os.Stat(killRegistryPath(home, sessionID))
	return err == nil
}

func assertTrailingNewline(t *testing.T, s, label string) {
	t.Helper()
	if s == "" || !strings.HasSuffix(s, "\n") {
		t.Fatalf("%s: expected trailing newline, got %q", label, s)
	}
}

func assertStdoutLine(t *testing.T, stdout, wantLine string) {
	t.Helper()
	// Exact single-line CLI message + trailing newline.
	want := wantLine
	if !strings.HasSuffix(want, "\n") {
		want += "\n"
	}
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

// firstKillBestEffort runs kill once (for double-kill Setup). Does not fail
// Setup when kill is still unimplemented (Classic TDD RED).
func firstKillBestEffort(t *testing.T, req *Request, sessionID string) {
	t.Helper()
	r := *req
	r.Mode = "handle"
	r.Args = []string{"kill", sessionID}
	// Copy env slice so Handle isolation stays per-call.
	r.Env = append([]string(nil), req.Env...)
	resp, err := runHandleInProcess(t, &r)
	if err != nil {
		t.Logf("first kill setup err (ok while RED): %v", err)
		return
	}
	if resp != nil && resp.ExitCode != 0 {
		t.Logf("first kill setup exit=%d stderr=%q (ok while RED)", resp.ExitCode, resp.Stderr)
	}
}
```
