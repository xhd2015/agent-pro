# Scenario

**Feature**: OnTTYStarted fires once when a session first gets a live TTY

```
# newly established
AutoSendOrResume(ModeRun, OnTTYStarted?)
  -> after successful first-open dispatch
  -> OnTTYStarted once | nil no-op

# already live
AutoSendOrResume(ModeRun) then AutoSendOrResume(ModeSend)
  -> total OnTTYStarted calls stay 1
```

## Preconditions

- Nested DOCTEST root under `tests/agentrunapi/on-tty-started/` (does not
  inherit parent `tests/agentrunapi` Setup/Run).
- Product API under design (RED until implementer):
  - `agentrunapi.TTYStartedInfo` with `SessionID`, `Runner`, `Workspace`
  - `agentrunapi.Opts.OnTTYStarted func(TTYStartedInfo)`
  - `AutoSendOrResume` invokes the hook once on newly established TTY
    (ModeRun success path with dispatch hooks in these leaves); never on
    ModeSend live follow-up.
- L2 only: temp `agentstorage.NewFileStore`, injectable `Probe` / `RunSession` /
  `SendLive`. No real agent-run binary, iTerm, or network.
- Parallel-safe: no `t.Setenv` / `os.Chdir`; each leaf owns temp Home.

## Steps

1. Root `Setup` seeds default session identity and workspace/runner.
2. Branch `Setup` sets `Op` (newly-established | follow-up-live).
3. Leaf `Setup` sets `InstallHook` and any session seed fields.
4. Root `Run` exercises AutoSendOrResume; leaf `Assert` checks hook counts.

## Context

- Default session id: `sess-tty-started-1` (leaves may override).
- Default runner: `grok-tty`.
- Default workspace: `/tmp/ws-on-tty-started`.
- `d.DOCTEST_ROOT` is this tree; module root is `../../..`.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	if req.Home == "" {
		req.Home = filepath.Join(t.TempDir(), ".agent-run")
	}
	if req.SessionID == "" {
		req.SessionID = "sess-tty-started-1"
	}
	if req.Prompt == "" {
		req.Prompt = "hello on-tty-started"
	}
	if req.Runner == "" {
		req.Runner = "grok-tty"
	}
	if req.Workspace == "" {
		req.Workspace = "/tmp/ws-on-tty-started"
	}
	return nil
}
```
