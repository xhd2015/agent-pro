# Scenario

**Feature**: `resume --no-submit` without `--open` is rejected (unchanged gate)

```
agent-run resume --no-submit <session-id> "x"
  -> exit ≠ 0
  -> error: --no-submit requires --open
```

## Steps

1. Seed bound+exited meta (so failure is flag pairing, not resume gate).
2. Run resume with `--no-submit` and followup, without `--open`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "test-resume-nosubmit-s1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440711"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-resume-nosubmit-1"
	req.InitialPrompt = "prior"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)

	req.FollowupPrompt = "x"
	req.Args = []string{
		"resume",
		"--no-submit",
		req.SessionID,
		req.FollowupPrompt,
	}
	return nil
}
```
