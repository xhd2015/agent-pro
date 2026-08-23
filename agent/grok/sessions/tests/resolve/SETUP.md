# Scenario

**Feature**: resolve contextual grok session id via ancestor walk

```
# harness fills Request (args, pid, procs, open files)
sessions.RunResolve(args, ResolveOpts{…inject…})
  -> FindAncestorGrok + ResolveFromAncestors
doctest <- stdout/stderr + returned error
```

## Preconditions

- Package exports `RunResolve`, `ResolveOpts`, `ResolveDetails`,
  `ResolveHelp`, and `ResolveCommandHelpLine`.
- Tests never list live processes or call live lsof.
- Session id comes only from open-file paths under `/.grok/sessions/…/<uuid>/…`.

## Steps

1. Root `Setup` allocates temp dir / grok home and default start pid.
2. Leaf `Setup` seeds procs, open files, and args.
3. Root `Run` (in DOCTEST.md) calls `RunResolve` with injectable ListProcs/Lsof.
4. Leaf `Assert` checks stdout/stderr/error.

## Context

- Default fixture session: `019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa`
- Nested nearer session: `019f283b-bbbb-7bbb-bbbb-bbbbbbbbbbbb`
- Alternate `--pid` session: `019f283b-cccc-7ccc-cccc-cccccccccccc`
- Default chain: grok `4242` → bash `5000` → start `6000`
- Grok cmdline includes ignored `--resume` / `--session-id` decoys.

```go
import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

const (
	fixtureSessionID     = "019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa"
	fixtureNearSessionID = "019f283b-bbbb-7bbb-bbbb-bbbbbbbbbbbb"
	fixtureAltSessionID  = "019f283b-cccc-7ccc-cccc-cccccccccccc"
	fixtureTabSessionID  = "019f283b-dddd-7ddd-dddd-dddddddddddd"
	wrongResumeSessionID = "00000000-0000-0000-0000-000000000000"
	wrongFlagSessionID   = "11111111-1111-1111-1111-111111111111"

	pidGrok     = 4242
	pidBash     = 5000
	pidStart    = 6000
	pidMainGrok = 3000
	pidAltGrok  = 7000
	pidAltStart = 7100

	pidTabGrok1 = 8100
	pidTabGrok2 = 8200
	pidTabGrok3 = 8300
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.TempDir == "" {
		req.TempDir = t.TempDir()
	}
	if req.GrokHome == "" {
		req.GrokHome = filepath.Join(req.TempDir, ".grok")
	}
	if req.PID == 0 {
		req.PID = pidStart
	}
	if req.OpenFiles == nil {
		req.OpenFiles = map[int][]string{}
	}
	if req.Args == nil {
		req.Args = []string{}
	}
	return nil
}

func grokCmdWithIgnoredFlags() string {
	return "/usr/local/bin/grok --resume " + wrongResumeSessionID + " --session-id " + wrongFlagSessionID
}

func grokSessionPath(sessionID string) string {
	return "/Users/fixture/.grok/sessions/%2Ftmp%2Fproj/" + sessionID + "/events.jsonl"
}

func defaultAncestorChain() []FixtureProc {
	return []FixtureProc{
		{PID: pidGrok, PPID: 1, Cmd: grokCmdWithIgnoredFlags()},
		{PID: pidBash, PPID: pidGrok, Cmd: "/bin/bash"},
		{PID: pidStart, PPID: pidBash, Cmd: "/usr/local/bin/agent-pro"},
	}
}

func seedHit(req *Request, sessionID string, grokPID int) {
	if len(req.Procs) == 0 {
		req.Procs = defaultAncestorChain()
	}
	req.OpenFiles[grokPID] = []string{grokSessionPath(sessionID)}
}

// seedTabWindow installs a 3-tab window; current is tab 1 (/dev/ttys101).
// Tab 2 hosts fixtureTabSessionID on /dev/ttys102; tab 3 is bash-only.
func seedTabWindow(req *Request) {
	req.CurrentSessionID = "w0t1p0:CURRENT-TAB-UUID"
	req.ControllingTTY = "/dev/ttys101"
	req.ITerm = []iterm2.SessionRef{
		{WindowID: "100", WindowName: "work", TabIndex: 1, SessionID: "w0t1p0:CURRENT-TAB-UUID", TTY: "/dev/ttys101", Name: "current"},
		{WindowID: "100", WindowName: "work", TabIndex: 2, SessionID: "w0t2p0:TAB2-UUID", TTY: "/dev/ttys102", Name: "grok-tab"},
		{WindowID: "100", WindowName: "work", TabIndex: 3, SessionID: "w0t3p0:TAB3-UUID", TTY: "/dev/ttys103", Name: "bash-only"},
	}
	req.FocusProcs = []sessions.FocusProc{
		{PID: pidTabGrok1, PPID: 1, TTY: "/dev/ttys101", Cmd: "/usr/local/bin/bash"},
		{PID: pidTabGrok2, PPID: 1, TTY: "/dev/ttys102", Cmd: "/usr/local/bin/grok"},
		{PID: 9100, PPID: 1, TTY: "/dev/ttys103", Cmd: "/bin/bash"},
	}
	req.OpenFiles[pidTabGrok2] = []string{grokSessionPath(fixtureTabSessionID)}
}

func seedTabCurrentIsLast(req *Request) {
	seedTabWindow(req)
	req.CurrentSessionID = "w0t3p0:TAB3-UUID"
	req.ControllingTTY = "/dev/ttys103"
}

func seedTabCurrentIsSecond(req *Request) {
	seedTabWindow(req)
	req.CurrentSessionID = "w0t2p0:TAB2-UUID"
	req.ControllingTTY = "/dev/ttys102"
}

func seedTabLeftHit(req *Request) {
	seedTabCurrentIsSecond(req)
	req.FocusProcs = []sessions.FocusProc{
		{PID: pidTabGrok2, PPID: 1, TTY: "/dev/ttys101", Cmd: "/usr/local/bin/grok"},
		{PID: 9100, PPID: 1, TTY: "/dev/ttys102", Cmd: "/bin/bash"},
		{PID: 9101, PPID: 1, TTY: "/dev/ttys103", Cmd: "/bin/bash"},
	}
	req.OpenFiles = map[int][]string{
		pidTabGrok2: {grokSessionPath(fixtureTabSessionID)},
	}
}

func seedTabMultiGrok(req *Request) {
	seedTabWindow(req)
	req.FocusProcs = append(req.FocusProcs, sessions.FocusProc{
		PID: pidTabGrok3, PPID: 1, TTY: "/dev/ttys102", Cmd: "/usr/local/bin/grok",
	})
	req.OpenFiles[pidTabGrok3] = []string{grokSessionPath(fixtureNearSessionID)}
}

// seedTabWrappedGrok models agent-run: shell on tab TTY, grok child on a
// different PTY (or ??). Tab match uses the inverse of focus's TTY tree walk.
func seedTabWrappedGrok(req *Request) {
	req.CurrentSessionID = "w0t1p0:CURRENT-TAB-UUID"
	req.ControllingTTY = "/dev/ttys101"
	req.ITerm = []iterm2.SessionRef{
		{WindowID: "100", WindowName: "work", TabIndex: 1, SessionID: "w0t1p0:CURRENT-TAB-UUID", TTY: "/dev/ttys101", Name: "current"},
		{WindowID: "100", WindowName: "work", TabIndex: 2, SessionID: "w0t2p0:TAB2-UUID", TTY: "/dev/ttys102", Name: "wrapped-grok"},
	}
	const (
		pidShell   = 9001
		pidServe   = 9002
		pidWrapped = 9003
	)
	req.FocusProcs = []sessions.FocusProc{
		{PID: pidShell, PPID: 1, TTY: "/dev/ttys102", Cmd: "/bin/bash"},
		{PID: pidServe, PPID: pidShell, TTY: "??", Cmd: "/usr/local/bin/agent-run __serve -- /usr/local/bin/grok"},
		{PID: pidWrapped, PPID: pidServe, TTY: "/dev/ttys199", Cmd: "/usr/local/bin/grok --always-approve"},
	}
	req.OpenFiles = map[int][]string{
		pidWrapped: {grokSessionPath(fixtureTabSessionID)},
	}
}

func assertNoHarnessErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run harness error: %v", err)
	}
}

func assertOK(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err != nil {
		t.Fatalf("unexpected error: %v", resp.Err)
	}
}

func assertErrContains(t *testing.T, resp *Response, sub string) {
	t.Helper()
	if resp.Err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(resp.Err.Error(), sub) {
		t.Fatalf("error %q does not contain %q", resp.Err.Error(), sub)
	}
}

func assertStdoutExact(t *testing.T, got string, lines ...string) {
	t.Helper()
	want := strings.Join(lines, "\n") + "\n"
	if got != want {
		t.Fatalf("stdout mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func assertJSONDetails(t *testing.T, stdout string, want ResolveDetailsExpect) {
	t.Helper()
	var got struct {
		SessionID   string `json:"session_id"`
		StartPID    int    `json:"start_pid"`
		AncestorPID int    `json:"ancestor_pid"`
		RunnerPID   int    `json:"runner_pid"`
		Source      string `json:"source"`
		Confidence  string `json:"confidence"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("json decode: %v\n%s", err, stdout)
	}
	if got.SessionID != want.SessionID {
		t.Fatalf("session_id = %q, want %q", got.SessionID, want.SessionID)
	}
	if got.StartPID != want.StartPID {
		t.Fatalf("start_pid = %d, want %d", got.StartPID, want.StartPID)
	}
	if got.AncestorPID != want.AncestorPID {
		t.Fatalf("ancestor_pid = %d, want %d", got.AncestorPID, want.AncestorPID)
	}
	if got.RunnerPID != want.RunnerPID {
		t.Fatalf("runner_pid = %d, want %d", got.RunnerPID, want.RunnerPID)
	}
	if got.Source == "" {
		t.Fatal("source is empty")
	}
	if got.Confidence != "hard" {
		t.Fatalf("confidence = %q, want hard", got.Confidence)
	}
}

type ResolveDetailsExpect struct {
	SessionID   string
	StartPID    int
	AncestorPID int
	RunnerPID   int
}

var _ = assert.Output
```
