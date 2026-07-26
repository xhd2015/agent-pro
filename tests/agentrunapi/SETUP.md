# Scenario

**Feature**: in-process auto-send-or-resume library (`Classify` + `AutoSendOrResume`)

```
# classify
Store + sessionID + Probe?
  -> Classify -> Mode(run|send|resume) + meta + found

# auto
Opts(SessionID, Open/Detach, Store, Probe?, hooks?)
  -> validate
  -> Classify
  -> RunSession | SendLive | ResumeSession  (hooks or production)

# source wire
cmd/agent-run/*.go import pkgs/agentrunapi
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agentrunapi` (P1 surface).
- Exports: `Mode` (`ModeRun`/`ModeSend`/`ModeResume`), `ProbeReport`, `ProbeFunc`,
  `LifecycleProbe`, `EmptyProbe`, `Classify`, `Opts`, `AutoSendOrResume` —
  see root DOCTEST.md planned API.
- Unit leaves use temp `agentstorage.NewFileStore` + injectable `Probe` /
  dispatch hooks. **No real agent-run binary, PATH LookPath, iTerm, or grok.**
- Nil probe → production `LifecycleProbe` (not EmptyProbe). Empty store + seeded
  meta without TTY registry still classifies ModeRun in unit leaves.
- `NewTerminal=false` on all auto unit leaves (ForceNew / FollowUp are P2 nested).
- P2 WaitReady/FollowUp: nested root `wait-driver/` (own DOCTEST.md; not this Run).
- Harness stores API errors on `Response.ErrString` (harness `error` is nil for
  expected API failures).

## Steps

1. Root `Setup` creates isolated store home under `t.TempDir()`.
2. Grouping `Setup` sets `req.Mode`.
3. Leaf `Setup` fills session/probe/validation fields.
4. `Run` calls package APIs or scans CLI sources; leaf `Assert` checks outcomes.

## Context

- Default session id for seeded leaves: `sess-api-1` (leaves may override).
- Default prompt for auto dispatch: `hello agentrunapi`.
- `d.DOCTEST_ROOT` is this tree (`tests/agentrunapi`); module root is `../..`.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Home == "" {
		req.Home = filepath.Join(t.TempDir(), ".agent-run")
	}
	if req.SessionID == "" {
		req.SessionID = "sess-api-1"
	}
	if req.Prompt == "" {
		req.Prompt = "hello agentrunapi"
	}
	return nil
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected harness error: %v", err)
	}
}

func assertAPIError(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil || resp.ErrString == "" {
		t.Fatal("expected API error, got nil/empty")
	}
}

func assertNoAPIError(t *testing.T, resp *Response) {
	t.Helper()
	if resp != nil && resp.ErrString != "" {
		t.Fatalf("unexpected API error: %s", resp.ErrString)
	}
}

func assertEqual(t *testing.T, field string, got, want any) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %#v, want %#v", field, got, want)
	}
}

func assertContainsFold(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(strings.ToLower(got), strings.ToLower(want)) {
		t.Fatalf("missing %q in %q", want, got)
	}
}

func assertZeroHooks(t *testing.T, resp *Response) {
	t.Helper()
	if resp.RunCalls != 0 || resp.SendCalls != 0 || resp.ResumeCalls != 0 {
		t.Fatalf("expected zero dispatch hooks, got run=%d send=%d resume=%d",
			resp.RunCalls, resp.SendCalls, resp.ResumeCalls)
	}
}
```
