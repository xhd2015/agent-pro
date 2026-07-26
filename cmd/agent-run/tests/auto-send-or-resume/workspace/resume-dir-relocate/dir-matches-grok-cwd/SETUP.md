# Scenario

**Feature**: resume `--dir` equal to Grok `info.cwd` → success; no relocate

```
Grok session under encode(ws-match); info.cwd = abs(ws-match)
meta.workspace = ws-match; runner bound+exited (grok-tty)
  -> agent-run resume --dir ws-match --agent-runner-config-home GROK_HOME …
  -> exit 0; session still under encode(ws-match); no relocate warning required
  -> argv has --resume <runner_session_id>
```

## Preconditions

- Inactive Grok session (empty `active_sessions.json`).
- `--dir` path exists and is a directory.
- Canonical equality: same Abs path for `--dir` and seeded `info.cwd`.

## Steps

1. Create `ws-match` under TempDir.
2. Seed Grok home with session at encode(ws-match) / runner_session_id.
3. Seed agent-run meta (bound+exited) with workspace=ws-match.
4. Resume with `--dir` = ws-match + `--agent-runner-config-home` + argv runner.

```go
import (
	"path/filepath"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	ws := filepath.Join(req.TempDir, "ws-match")
	mustMkdirWS(t, ws, "match")

	ensureGrokHome(t, req)
	req.GrokSessionCwd = absPathNoEval(t, ws)
	req.DirOverride = ws
	req.Workspace = ws

	req.SessionID = "relocate-match-r1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440r01"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-relocate-r1"
	req.InitialPrompt = "created in ws-match"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)

	req.GrokSessionDir = seedInactiveGrokSession(t, req.GrokHome, req.GrokSessionCwd, req.RunnerSessionID)

	installArgvCwdRunner(t, req)
	req.FollowupPrompt = "hi match cwd"
	req.Args = []string{
		"resume",
		"--dir", req.DirOverride,
		"--agent-runner-config-home", req.GrokHome,
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.SessionID,
		req.FollowupPrompt,
	}
	req.ExecTimeout = 60 * time.Second
	req.Mode = "read-probes"
	return nil
}
```
