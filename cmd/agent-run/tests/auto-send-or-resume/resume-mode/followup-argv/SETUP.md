# Scenario

**Feature**: exited session + auto + followup → resume with `--resume <runner_session_id>` in argv

```
seed finished bound+exited meta
  -> run --auto-send-or-resume --session-id ID --agent-runner-binary REC "followup"
  -> exit 0; argv probe has --resume <runner_session_id>
```

## Steps

1. Seed bound exited meta (no live terminal).
2. Write argv-recording fake runner; do **not** set `AGENT_RUN_GROK_TTY_COMMAND`.
3. Run auto with followup prompt.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "auto-resume-d1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440d11"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-auto-resume-d1"
	req.InitialPrompt = "prior turn d1"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)

	installArgvRunner(t, req)
	req.FollowupPrompt = "resume followup d1"
	req.Args = []string{
		"run",
		"--auto-send-or-resume",
		"--session-id", req.SessionID,
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.FollowupPrompt,
	}
	req.ExecTimeout = 60 * time.Second
	req.Mode = "read-probes"
	return nil
}
```
