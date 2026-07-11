# Scenario

**Feature**: resume denied when runner still live (`exited != true`)

```
live bound session -> agent-run resume <id> "followup"
  -> exit 1; cannot resume / exited false / hint send
```

## Steps

1. Seed live bound not-exited session (fake ptywrap + registry).
2. Run `resume <id> "followup"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "test-resume-live-s1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440411"
	req.TerminalSessionID = "term-resume-live-1"
	req.InitialPrompt = "still live"
	seedLiveBoundNotExited(t, req)
	req.FollowupPrompt = "should not resume"
	req.Args = []string{"resume", req.SessionID, req.FollowupPrompt}
	return nil
}
```
