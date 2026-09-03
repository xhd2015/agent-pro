# Scenario

**Feature**: resolve session id from pid using injectable process snapshot and Lsof

```
# caller supplies pid + fixture procs / open files
caller: ResolveFromPID(pid, Options{ListProcs, Lsof, MaxDepth, homes})
  -> build tree (input + descendants)
  -> classify roles; pick grok/codex candidates (deeper first)
  -> Lsof each candidate; parse session uuid from open path
doctest <- Result | error (pid not found)
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/procresolve` exports:
  - `ResolveFromPID(pid int, opts Options) (*Result, error)`
  - types `Options`, `Proc`, `ProcNode`, `Result` per root DOCTEST.md DSN
- **No live processes, no real `lsof`, no real agent-run** — every leaf injects
  `ListProcs` / `Lsof` via harness `Request.Procs` and `Request.OpenFiles`.
- Session ids are parsed from open-file paths only (not cmdline flags).
- Unknown pid (not present in snapshot) → **error** containing `pid not found`.

## Steps

1. Root `Setup` seeds default `MaxDepth` and empty open-files map.
2. Grouping / leaf `Setup` installs fixture procs, open paths, and input `PID`.
3. Root `Run` maps fixtures into `Options` injectors and calls `ResolveFromPID`.
4. Leaf `Assert` checks Kind / SessionID / Source / Confidence / Tree / error.

## Context

- Default `MaxDepth` = 16 (large enough for all fixture trees).
- Fixture homes (optional for path parse when paths are absolute):
  - GrokHome: `/tmp/fake-grok-home`
  - CodexHome: `/tmp/fake-codex-home`
- Canonical fixture session ids (locked for asserts):
  - Grok UUID: `019fabcdef-1234-5678-9abc-def012345678`
  - Codex UUID: `a1b2c3d4-e5f6-7890-abcd-ef1234567890`

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
)

// Fixture session ids shared by hit leaves.
const (
	fixtureGrokSessionID  = "019fabcdef-1234-5678-9abc-def012345678"
	fixtureCodexSessionID = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
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

func assertContainsString(t *testing.T, field, got, substr string) {
	t.Helper()
	if substr == "" {
		t.Fatalf("%s: empty substr", field)
	}
	if !strings.Contains(got, substr) {
		t.Fatalf("%s: got %q, want substring %q", field, got, substr)
	}
}

func findNodeByPID(tree []procresolve.ProcNode, pid int) (procresolve.ProcNode, bool) {
	for _, n := range tree {
		if n.PID == pid {
			return n, true
		}
	}
	return procresolve.ProcNode{}, false
}

func assertRole(t *testing.T, tree []procresolve.ProcNode, pid int, wantRole string) {
	t.Helper()
	n, ok := findNodeByPID(tree, pid)
	if !ok {
		t.Fatalf("tree missing pid %d (want role %q); tree=%+v", pid, wantRole, tree)
	}
	if n.Role != wantRole {
		t.Fatalf("pid %d Role: got %q, want %q", pid, n.Role, wantRole)
	}
}

// grokSessionPath returns a primary Grok session open path (events.jsonl).
// Directory-only session paths are not hard hits for ResolveFromPID.
func grokSessionPath(uuid string) string {
	return "/tmp/fake-grok-home/.grok/sessions/2026-07/" + uuid + "/events.jsonl"
}

// codexRolloutPath returns a realistic open path under .codex/sessions with uuid.
func codexRolloutPath(uuid string) string {
	return "/tmp/fake-codex-home/.codex/sessions/2026/07/25/rollout-2026-07-25T12-00-00-" + uuid + ".jsonl"
}
```
