# Scenario

**Feature**: resolve nearest ancestor grok and its session from injectable snapshot + Lsof

```
# caller supplies start pid + fixture procs / open files
caller: FindAncestorGrok(pid) + ResolveFromAncestors(pid)
  -> walk start then PPID; first IsGrokRunner wins
  -> ResolveFromPID(grokPID) for session from open files
doctest <- Ancestor+ok + Result | error (pid not found)
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/procresolve` will export
  `FindAncestorGrok` and `ResolveFromAncestors` (RED until then).
- **No live processes, no real `lsof`** — every leaf injects `ListProcs` /
  `Lsof` via `Request.Procs` and `Request.OpenFiles`.
- Session ids are parsed from open-file paths only (not cmdline flags).

## Steps

1. Root `Setup` seeds default `MaxDepth` and empty open-files map.
2. Leaf `Setup` installs fixture procs, open paths, and start `PID`.
3. Root `Run` calls both new APIs with the same injectors.
4. Leaf `Assert` checks ancestor ok/pid plus Kind / SessionID / error.

## Context

- Default `MaxDepth` = 16 (used only by the delegated `ResolveFromPID`).
- Fixture homes (paths in Lsof are absolute):
  - GrokHome: `/tmp/fake-grok-home`
  - CodexHome: `/tmp/fake-codex-home`
- Canonical fixture session ids:
  - Grok (nearest / default): `019fabcdef-1234-5678-9abc-def012345678`
  - Grok (outer / main): `019fabcdef-0000-0000-0000-000000000001`
  - Grok (descendant decoy): `019fabcdef-9999-9999-9999-999999999999`
- Default chain pids: grok `4242` → bash `5000` → start `6000`.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
)

const (
	fixtureGrokSessionID      = "019fabcdef-1234-5678-9abc-def012345678"
	fixtureMainGrokSessionID  = "019fabcdef-0000-0000-0000-000000000001"
	fixtureDecoyGrokSessionID = "019fabcdef-9999-9999-9999-999999999999"
	wrongResumeSessionID      = "00000000-0000-0000-0000-000000000000"
	wrongFlagSessionID        = "11111111-1111-1111-1111-111111111111"

	pidGrok     = 4242
	pidBash     = 5000
	pidStart    = 6000
	pidMainGrok = 3000
	pidUpdate   = 4300
	pidDecoy    = 7000
	pidCodex    = 2000
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.MaxDepth == 0 {
		req.MaxDepth = 16
	}
	if req.OpenFiles == nil {
		req.OpenFiles = map[int][]string{}
	}
	if req.GrokHome == "" {
		req.GrokHome = "/tmp/fake-grok-home"
	}
	if req.CodexHome == "" {
		req.CodexHome = "/tmp/fake-codex-home"
	}
	return nil
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertResult(t *testing.T, resp *Response) *procresolve.Result {
	t.Helper()
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.Result == nil {
		t.Fatal("Result is nil")
	}
	return resp.Result
}

func assertEqualString(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %q, want %q", field, got, want)
	}
}

func assertEqualInt(t *testing.T, field string, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %d, want %d", field, got, want)
	}
}

func assertAncestor(t *testing.T, resp *Response, wantPID int) {
	t.Helper()
	if resp == nil {
		t.Fatal("response is nil")
	}
	if !resp.AncestorOK {
		t.Fatalf("FindAncestorGrok ok=false, want pid %d", wantPID)
	}
	assertEqualInt(t, "Ancestor.PID", resp.Ancestor.PID, wantPID)
	if !strings.Contains(resp.Ancestor.Cmd, "grok") {
		t.Fatalf("Ancestor.Cmd: got %q, want substring %q", resp.Ancestor.Cmd, "grok")
	}
}

func assertNoAncestor(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.AncestorOK {
		t.Fatalf("FindAncestorGrok ok=true, want false; ancestor=%+v", resp.Ancestor)
	}
	if resp.Ancestor.PID != 0 {
		t.Fatalf("Ancestor.PID: got %d, want 0 on miss", resp.Ancestor.PID)
	}
}

func assertHardGrokHit(t *testing.T, r *procresolve.Result, grokPID int, sessionID string) {
	t.Helper()
	assertEqualInt(t, "InputPID", r.InputPID, grokPID)
	assertEqualString(t, "Kind", r.Kind, "grok")
	assertEqualString(t, "SessionID", r.SessionID, sessionID)
	assertEqualString(t, "Confidence", r.Confidence, "hard")
	assertEqualInt(t, "RunnerPID", r.RunnerPID, grokPID)
	if !strings.Contains(r.RunnerCmd, "grok") {
		t.Fatalf("RunnerCmd: got %q, want substring %q", r.RunnerCmd, "grok")
	}
	if r.SessionID == wrongResumeSessionID || r.SessionID == wrongFlagSessionID {
		t.Fatalf("SessionID must not come from cmdline flags: %q", r.SessionID)
	}
	if r.SessionID == fixtureDecoyGrokSessionID {
		t.Fatalf("SessionID must not come from a descendant decoy grok: %q", r.SessionID)
	}
}

func assertNone(t *testing.T, r *procresolve.Result, startPID int) {
	t.Helper()
	assertEqualInt(t, "InputPID", r.InputPID, startPID)
	assertEqualString(t, "Kind", r.Kind, "none")
	assertEqualString(t, "SessionID", r.SessionID, "")
	assertEqualString(t, "Confidence", r.Confidence, "")
	assertEqualInt(t, "RunnerPID", r.RunnerPID, 0)
}

func grokSessionPath(uuid string) string {
	return "/tmp/fake-grok-home/.grok/sessions/2026-07/" + uuid + "/events.jsonl"
}

func grokCmdWithIgnoredFlags() string {
	return "/usr/local/bin/grok --resume " + wrongResumeSessionID + " --session-id " + wrongFlagSessionID
}
```
