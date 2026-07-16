# Scenario

**Feature**: live session + auto + `--detach` ignores detach and still sends

```
seed live sendable terminal
  -> run --auto-send-or-resume --detach --session-id ID "followup"
  -> exit 0; stdout msg_N; optional stderr note: --detach ignored…
  -> no new detach daemon create path
```

## Steps

1. Seed live bound not-exited session (fake ptywrap inject).
2. Run auto + detach + prompt.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "auto-live-detach-s1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440c11"
	req.TerminalSessionID = "term-auto-live-detach-s1"
	req.InitialPrompt = "prior live detach"
	req.FollowupPrompt = "live detach send followup"
	seedLiveBoundNotExited(t, req)

	req.Args = []string{
		"run",
		"--auto-send-or-resume",
		"--detach",
		"--session-id", req.SessionID,
		req.FollowupPrompt,
	}
	req.ExecTimeout = 45 * time.Second
	return nil
}
```
