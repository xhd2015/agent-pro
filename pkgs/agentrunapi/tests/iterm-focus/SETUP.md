# Scenario

**Feature**: resolve agent-run session_id to iTerm and focus (P2)

```
# pure TTYs
[]ProcRow + rootPID -> CollectTTYsFromTree -> real TTYs (ancestors+descendants)

# library
session_id + Store + ListProcs + ListITerm
  -> FindITermForSession / FocusSession -> candidates | focus (unless DryRun)

# CLI
agent-run focus [flags] -> agentruncli.RunFocus(writers) -> library path
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agentrunapi` gains P2 APIs locked in
  root `DOCTEST.md` (`ProcRow`, `FocusCandidate`, `FocusOpts`,
  `CollectTTYsFromTree`, `FindITermForSession`, `FocusSession`).
- Package `github.com/xhd2015/agent-pro/pkgs/agentruncli` gains `RunFocus` and
  `Handle` switch case `"focus"`.
- Depends on P1 `iterm2.SessionRef` / `FindByTTY` / `Focus` (implementer go.mod
  replace; designer does not add replace).
- Classic TDD: leaves RED until those symbols exist and behave as specified.
- Parallel-safe: no `t.Setenv` / `os.Chdir`; inject Store via `NewFileStore(home)`
  under `t.TempDir()`; inject ListProcs / ListITerm / FocusITerm.

## Steps

1. Root `Setup` creates isolated `req.Home` under `t.TempDir()` and defaults
   runner/term when empty.
2. Leaf `Setup` sets `Phase`, session seed flags, procs, iTerm refs, dry-run,
   index, or CLI args.
3. `Run` calls product APIs; leaf `Assert` checks outcomes / focus call log.

## Context

- Candidate `Index` is **0-based**; CLI `--index N` selects the same.
- Registry path: `{home}/{runner}-registry/{terminal_session_id}.json` with at
  least `session_id`, `pid` fields (TTY may be absent/`??`).
- Serve PID often has TTY `??`; real TTYs come from tree walk.

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.Home == "" {
		req.Home = filepath.Join(t.TempDir(), ".agent-run")
	}
	if req.Runner == "" {
		req.Runner = "grok-tty"
	}
	return nil
}

// writeRegistryJSON writes a minimal registry entry for terminalSessionID.
func writeRegistryJSON(dir, terminalSessionID string, pid int) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	entry := map[string]any{
		"session_id":  terminalSessionID,
		"listen_addr": "127.0.0.1:9",
		"pid":         pid,
		"created_at":  "2026-01-01T00:00:00Z",
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, terminalSessionID+".json"), data, 0o644)
}

// singleMatchFixtures: serve PID with ?? TTY; ancestor holds /dev/ttys148;
// one matching iTerm ref. Used by single-match leaves.
func singleMatchFixtures(req *Request) {
	req.SessionID = "sess-single"
	req.TermID = "term-single"
	req.RootPID = 200
	req.RegistryPID = 200
	req.SeedSession = true
	req.SeedRegistry = true
	req.Procs = []agentrunapi.ProcRow{
		{PID: 1, PPID: 0, TTY: "??", Cmd: "kernel_task"},
		{PID: 100, PPID: 1, TTY: "/dev/ttys148", Cmd: "/bin/zsh"},
		{PID: 200, PPID: 100, TTY: "??", Cmd: "agent-run serve"},
		{PID: 201, PPID: 200, TTY: "", Cmd: "grok"},
	}
	req.ITermRefs = []iterm2.SessionRef{
		{WindowID: "win-1", TabIndex: 2, SessionID: "iterm-s1", TTY: "/dev/ttys148", Name: "agent"},
		{WindowID: "win-2", TabIndex: 1, SessionID: "iterm-other", TTY: "/dev/ttys999", Name: "other"},
	}
}

// multiMatchFixtures: two real TTYs in tree map to two iTerm sessions.
func multiMatchFixtures(req *Request) {
	req.SessionID = "sess-multi"
	req.TermID = "term-multi"
	req.RootPID = 300
	req.RegistryPID = 300
	req.SeedSession = true
	req.SeedRegistry = true
	req.Procs = []agentrunapi.ProcRow{
		{PID: 50, PPID: 1, TTY: "/dev/ttys010", Cmd: "zsh"},
		{PID: 300, PPID: 50, TTY: "??", Cmd: "agent-run serve"},
		{PID: 301, PPID: 300, TTY: "/dev/ttys011", Cmd: "grok"},
	}
	req.ITermRefs = []iterm2.SessionRef{
		{WindowID: "w-a", TabIndex: 1, SessionID: "ia", TTY: "/dev/ttys010", Name: "A"},
		{WindowID: "w-b", TabIndex: 3, SessionID: "ib", TTY: "/dev/ttys011", Name: "B"},
	}
}

func assertNoFocus(t *testing.T, resp *Response) {
	t.Helper()
	if len(resp.FocusCalls) != 0 {
		t.Fatalf("expected FocusITerm never called, got %d calls: %+v", len(resp.FocusCalls), resp.FocusCalls)
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected non-nil error")
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
```
