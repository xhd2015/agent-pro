# Scenario

**Feature**: resume `--dir` ≠ Grok cwd with `--allow-relocate-resume-session-dir`
→ warn, RelocateCWD, update meta.workspace, continue resume

```
Grok session under encode(ws-old); info.cwd = abs(ws-old)
--dir = ws-new; --allow-relocate-resume-session-dir
  -> warning on stderr
  -> sessions.RelocateCWD(runnerSessionID, dir, {GrokHome})
  -> session under encode(ws-new); info.cwd = abs(ws-new)
  -> meta.workspace = --dir
  -> exit 0; argv has --resume <runner_session_id>
```

## Preconditions

- Both workspaces exist; session inactive.
- Effective Grok home via `--agent-runner-config-home`.
- Fake argv/cwd runner for successful resume continuation.

## Steps

1. Create ws-old + ws-new.
2. Seed inactive Grok session under encode(ws-old).
3. Seed bound+exited meta (workspace=ws-old).
4. Resume with `--dir` ws-new + `--allow-relocate-resume-session-dir`.

```go
import (
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	oldWS := filepath.Join(req.TempDir, "ws-old")
	newWS := filepath.Join(req.TempDir, "ws-new")
	mustMkdirWS(t, oldWS, "old")
	mustMkdirWS(t, newWS, "new")

	ensureGrokHome(t, req)
	req.GrokSessionCwd = absPathNoEval(t, oldWS)
	req.DirOverride = newWS
	req.Workspace = oldWS

	req.SessionID = "relocate-allow-r3"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440r03"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-relocate-r3"
	req.InitialPrompt = "created in ws-old for allow relocate"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)

	req.GrokSessionDir = seedInactiveGrokSession(t, req.GrokHome, req.GrokSessionCwd, req.RunnerSessionID)

	installArgvCwdRunner(t, req)
	req.FollowupPrompt = "hi after allow relocate"
	req.Args = []string{
		"resume",
		"--dir", req.DirOverride,
		"--allow-relocate-resume-session-dir",
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
