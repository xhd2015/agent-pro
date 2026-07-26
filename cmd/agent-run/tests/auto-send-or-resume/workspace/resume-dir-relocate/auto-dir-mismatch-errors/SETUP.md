# Scenario

**Feature**: auto→resume path enforces the same `--dir` vs Grok cwd mismatch error

```
bound+exited meta; Grok session under encode(ws-old)
run --auto-send-or-resume --session-id ID --dir ws-new … (no allow flag)
  -> MODE=resume; exit 1; stderr mismatch + --allow-relocate-resume-session-dir
  -> Grok session not moved
```

## Preconditions

- Same fixture shape as `dir-mismatch-errors`, but invocation is
  `run --auto-send-or-resume` (not `resume` subcommand).

## Steps

1. Create ws-old + ws-new; seed inactive Grok session at encode(ws-old).
2. Seed bound+exited meta so auto classifies MODE=resume.
3. Run auto with `--dir` ws-new (no allow flag).

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

	req.SessionID = "relocate-auto-mismatch-r5"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440r05"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-relocate-r5"
	req.InitialPrompt = "created in ws-old auto"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)

	req.GrokSessionDir = seedInactiveGrokSession(t, req.GrokHome, req.GrokSessionCwd, req.RunnerSessionID)

	installArgvCwdRunner(t, req)
	req.FollowupPrompt = "hi auto mismatch"
	req.Args = []string{
		"run",
		"--auto-send-or-resume",
		"--session-id", req.SessionID,
		"--dir", req.DirOverride,
		"--agent-runner-config-home", req.GrokHome,
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.FollowupPrompt,
	}
	req.ExecTimeout = 60 * time.Second
	req.Mode = "read-probes"
	return nil
}
```
