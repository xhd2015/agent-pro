# Scenario

**Feature**: resume `--dir` ≠ Grok `info.cwd` without allow flag → error; no move

```
Grok session under encode(ws-old); info.cwd = abs(ws-old)
--dir = ws-new (exists, ≠ grok cwd); NO --allow-relocate-resume-session-dir
  -> agent-run resume --dir ws-new --agent-runner-config-home GROK_HOME …
  -> exit 1; stderr explains cwd mismatch + hints --allow-relocate-resume-session-dir
  -> session still under encode(ws-old); info.cwd unchanged
```

## Preconditions

- Both ws-old and ws-new exist as directories.
- Grok session inactive.
- meta.workspace seeded as ws-old (historical create cwd).
- Must not relocate on error path.

## Steps

1. Create ws-old + ws-new.
2. Seed Grok session at encode(ws-old).
3. Seed bound+exited meta (workspace=ws-old).
4. Resume with `--dir` ws-new only (no allow flag).

```go
import (
	"path/filepath"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	oldWS := filepath.Join(req.TempDir, "ws-old")
	newWS := filepath.Join(req.TempDir, "ws-new")
	mustMkdirWS(t, oldWS, "old")
	mustMkdirWS(t, newWS, "new")

	ensureGrokHome(t, req)
	req.GrokSessionCwd = absPathNoEval(t, oldWS)
	req.DirOverride = newWS
	req.Workspace = oldWS

	req.SessionID = "relocate-mismatch-r2"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440r02"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-relocate-r2"
	req.InitialPrompt = "created in ws-old"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)

	req.GrokSessionDir = seedInactiveGrokSession(t, req.GrokHome, req.GrokSessionCwd, req.RunnerSessionID)

	installArgvCwdRunner(t, req)
	req.FollowupPrompt = "hi mismatch no allow"
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
